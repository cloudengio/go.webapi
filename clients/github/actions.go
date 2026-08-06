// Copyright 2026 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"cloudeng.io/text/textutil"
	"cloudeng.io/webapi/operations"
	gogithub "github.com/google/go-github/v89/github"
)

const APIHost = "https://api.github.com"

// linkNextRE matches the URL with rel="next" in a GitHub Link response header.
var linkNextRE = regexp.MustCompile(`<([^>]*)>;\s*rel="next"`)

// parseLinkHeader extracts the URL with rel="next" from a GitHub Link header.
// Returns an empty string if no next page link is present.
func parseLinkHeader(header string) string {
	if m := linkNextRE.FindStringSubmatch(header); len(m) > 1 {
		return m[1]
	}
	return ""
}

// appendPerPage appends per_page=n to the URL, using ? or & as appropriate.
func appendPerPage(u string, perPage int) string {
	if perPage <= 0 {
		return u
	}
	sep := "?"
	if strings.Contains(u, "?") {
		sep = "&"
	}
	return u + sep + "per_page=" + strconv.Itoa(perPage)
}

// appendQueryParam appends key=value to u if value is non-empty.
func appendQueryParam(u, key, value string) string {
	if value == "" {
		return u
	}
	sep := "?"
	if strings.Contains(u, "?") {
		sep = "&"
	}
	return u + sep + key + "=" + url.QueryEscape(value)
}

// RunsFilter holds optional server-side filter parameters for listing workflow runs.
type RunsFilter struct {
	Actor  string
	Branch string
	Event  string
	Status string
}

// linkPaginator implements operations.Paginator[T] using GitHub's Link header.
type linkPaginator[T any] struct {
	initialURL string
	perPage    int
}

func (p *linkPaginator[T]) Next(ctx context.Context, _ T, r *http.Response) (*http.Request, bool, error) {
	if r == nil {
		req, err := http.NewRequestWithContext(ctx, "GET", appendPerPage(p.initialURL, p.perPage), nil)
		return req, false, err
	}
	next := parseLinkHeader(r.Header.Get("Link"))
	if next == "" {
		return nil, true, nil
	}
	req, err := http.NewRequestWithContext(ctx, "GET", next, nil)
	return req, false, err
}

// NewRunsScanner returns an operations.Scanner that iterates over workflow runs
// for the specified owner and repo, one page at a time. Non-empty fields in
// filter are sent as query parameters so the GitHub API performs server-side
// filtering before any results are returned.
func NewRunsScanner(owner, repo string, perPage int, filter RunsFilter, opts ...operations.Option) *operations.Scanner[gogithub.WorkflowRuns] {
	u := fmt.Sprintf("%s/repos/%s/%s/actions/runs",
		APIHost,
		url.PathEscape(owner), url.PathEscape(repo))
	u = appendQueryParam(u, "actor", filter.Actor)
	u = appendQueryParam(u, "branch", filter.Branch)
	u = appendQueryParam(u, "event", filter.Event)
	u = appendQueryParam(u, "status", filter.Status)
	return operations.NewScanner(
		&linkPaginator[gogithub.WorkflowRuns]{initialURL: u, perPage: perPage}, opts...)
}

// NewJobsScanner returns an operations.Scanner that iterates over jobs for the
// specified workflow run. filter may be "latest" (the default) or "all" to
// include jobs from all prior run attempts.
func NewJobsScanner(owner, repo string, runID int64, filter string, perPage int, opts ...operations.Option) *operations.Scanner[gogithub.Jobs] {
	u := fmt.Sprintf("%s/repos/%s/%s/actions/runs/%d/jobs",
		APIHost,
		url.PathEscape(owner), url.PathEscape(repo), runID)
	if filter != "" {
		u = u + "?filter=" + filter
	}
	return operations.NewScanner(
		&linkPaginator[gogithub.Jobs]{initialURL: u, perPage: perPage}, opts...)
}

// GetWorkflowJob returns the job with the specified ID via the
// /repos/{owner}/{repo}/actions/jobs/{job_id} endpoint. Unlike the jobs
// returned by NewJobsScanner, which are scoped to a single workflow run, a job
// is addressed here by its own ID, as reported by, for example, the workflow_job
// webhook event.
func GetWorkflowJob(ctx context.Context, owner, repo string, jobID int64, opts ...operations.Option) (gogithub.WorkflowJob, error) {
	u := fmt.Sprintf("%s/repos/%s/%s/actions/jobs/%d",
		APIHost, url.PathEscape(owner), url.PathEscape(repo), jobID)
	ep := operations.NewEndpoint[gogithub.WorkflowJob](opts...)
	job, body, _, err := ep.Get(ctx, u)
	if err != nil {
		truncatedBody := textutil.Head(body, '\n', 5)
		return gogithub.WorkflowJob{}, fmt.Errorf("failed to get job %v: %q: %s: %w", jobID, u, truncatedBody, err)
	}
	return job, nil
}

// RerunWorkflowJob requests that the job with the specified ID be rerun, via
// the /repos/{owner}/{repo}/actions/jobs/{job_id}/rerun endpoint. GitHub reruns
// the job along with any jobs in the same workflow run that depend on it, and
// responds 201 Created with an empty body, so there is nothing to return beyond
// any error. The rerun is queued asynchronously: a successful return means
// GitHub accepted the request, not that the job has started.
func RerunWorkflowJob(ctx context.Context, owner, repo string, jobID int64, opts ...operations.Option) error {
	u := fmt.Sprintf("%s/repos/%s/%s/actions/jobs/%d/rerun",
		APIHost, url.PathEscape(owner), url.PathEscape(repo), jobID)
	ep := operations.NewPutEndpoint[struct{}, struct{}](opts...)
	_, body, _, err := ep.Post(ctx, u, struct{}{})
	if err == nil {
		return nil
	}
	// GitHub returns 201 Created for this endpoint; Post treats any non-200
	// status as an error but still returns the pre-read body bytes.
	if opErr, ok := err.(*operations.Error); ok && opErr.StatusCode == http.StatusCreated {
		return nil
	}
	truncatedBody := textutil.Head(body, '\n', 5)
	return fmt.Errorf("failed to rerun job %v: %q: %s: %w", jobID, u, truncatedBody, err)
}

// NewRunnersScanner returns an operations.Scanner that iterates over self-hosted
// runners registered for the specified owner and repo.
func NewRunnersScanner(owner, repo string, perPage int, opts ...operations.Option) *operations.Scanner[gogithub.Runners] {
	u := fmt.Sprintf("%s/repos/%s/%s/actions/runners", APIHost,
		url.PathEscape(owner), url.PathEscape(repo))
	return operations.NewScanner(
		&linkPaginator[gogithub.Runners]{initialURL: u, perPage: perPage}, opts...)
}

// CreateRegistrationToken requests a new runner registration token for the
// given owner/repo. Options (including WithAuth) follow the same pattern as
// NewRunsScanner, NewRunnersScanner, and the other functions in this package.
func CreateRegistrationToken(ctx context.Context, owner, repo string, opts ...operations.Option) (gogithub.RegistrationToken, error) {
	u := fmt.Sprintf("%s/repos/%s/%s/actions/runners/registration-token",
		APIHost, url.PathEscape(owner), url.PathEscape(repo))
	ep := operations.NewPutEndpoint[struct{}, gogithub.RegistrationToken](opts...)
	tok, body, _, err := ep.Post(ctx, u, struct{}{})
	if err == nil {
		return tok, nil
	}
	// GitHub returns 201 Created for this endpoint; Post treats any non-200
	// status as an error but still returns the pre-read body bytes.
	if opErr, ok := err.(*operations.Error); ok && opErr.StatusCode == http.StatusCreated {
		if jsonErr := json.Unmarshal(body, &tok); jsonErr != nil {
			return gogithub.RegistrationToken{}, jsonErr
		}
		return tok, nil
	}
	truncatedBody := textutil.Head(body, '\n', 5)
	return gogithub.RegistrationToken{}, fmt.Errorf("failed to create registration token: %q: %s: %w", u, truncatedBody, err)
}

// CreateWebhook creates a new webhook for the given owner/repo. The Name field
// in the request must be "web" for HTTP webhooks. GitHub returns 201 Created
// on success.
func CreateWebhook(ctx context.Context, owner, repo string, request *gogithub.Hook, opts ...operations.Option) (gogithub.Hook, error) {
	u := fmt.Sprintf("%s/repos/%s/%s/hooks",
		APIHost, url.PathEscape(owner), url.PathEscape(repo))
	ep := operations.NewPutEndpoint[*gogithub.Hook, gogithub.Hook](opts...)
	hook, body, _, err := ep.Post(ctx, u, request)
	if err == nil {
		return hook, nil
	}
	// GitHub returns 201 Created on success; Post treats any non-200 status as
	// an error but still returns the pre-read body bytes.
	if opErr, ok := err.(*operations.Error); ok && opErr.StatusCode == http.StatusCreated {
		if jsonErr := json.Unmarshal(body, &hook); jsonErr != nil {
			return gogithub.Hook{}, jsonErr
		}
		return hook, nil
	}
	truncatedBody := textutil.Head(body, '\n', 5)
	return gogithub.Hook{}, fmt.Errorf("failed to create webhook: %q: %s: %w", u, truncatedBody, err)
}

// ErrRelayClosed is returned by ReadWebhookEvent when the relay closes the
// long-poll connection without delivering an event (for example when the relay
// is shutting down). Callers looping over ReadWebhookEvent can use it to
// distinguish a clean relay shutdown from a transport or decode error.
var ErrRelayClosed = errors.New("relay closed the connection without delivering an event")

// ErrUnexpectedEvent is returned by ReadWorkflowRunEvent and ReadWorkflowJobEvent
// when the relay delivers an event whose type does not match the one the caller
// asked for. It is wrapped with the delivered event type so callers can match it
// with errors.Is and skip the delivery.
var ErrUnexpectedEvent = errors.New("unexpected webhook event type")

// ReadWebhookEvent performs a single long-poll ("hanging read") GET against a
// webhooks.Relay polling endpoint at relayURL and returns the GitHub event type
// together with the raw, already-verified webhook payload. The relay validates
// the webhook signature before queuing deliveries, so the returned payload is
// already authenticated.
//
// eventType is taken from the X-GitHub-Event header the relay forwards; it is
// empty if the relay is not configured to forward that header. Callers can
// switch on eventType to demultiplex different event types (for example
// workflow_run vs workflow_job) from a single relay.
//
// The request is issued via an operations.Endpoint, so it shares the same
// rate-control, retry/backoff, auth, HTTP client, and logging behaviour as the
// rest of the package; configure them via opts (for example
// operations.WithHTTPClient or operations.WithRateController). Because a hanging
// read can block indefinitely, any configured HTTP client should not impose a
// request timeout — use ctx to bound or cancel the call.
//
// It blocks until the relay delivers a payload, ctx is cancelled, or the
// request fails, and is intended to be called in a loop to consume a stream of
// events. ErrRelayClosed is returned if the relay responds with an empty body,
// which indicates a clean shutdown.
func ReadWebhookEvent(ctx context.Context, relayURL string, opts ...operations.Option) (eventType string, payload json.RawMessage, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, relayURL, nil)
	if err != nil {
		return "", nil, err
	}
	ep := operations.NewEndpoint[json.RawMessage](opts...)
	payload, body, encoding, resp, err := ep.IssueRequest(ctx, req)
	if err != nil {
		return "", nil, err
	}
	if encoding != operations.JSONEncoding {
		return "", nil, fmt.Errorf("unexpected content encoding %v from relay %q", encoding, relayURL)
	}
	if len(body) == 0 {
		return "", nil, ErrRelayClosed
	}
	return resp.Header.Get("X-GitHub-Event"), payload, nil
}

// ReadWorkflowRunEvent performs a single long-poll ("hanging read") GET against
// a webhooks.Relay polling endpoint at relayURL and decodes the delivered
// payload as a workflow_run webhook event. It is a convenience wrapper around
// ReadWebhookEvent for callers that consume only workflow_run events; see that
// function for the blocking, options, and shutdown semantics.
//
// If the relay reports an event type (via the forwarded X-GitHub-Event header)
// that is not workflow_run, it returns ErrUnexpectedEvent wrapped with the
// delivered type so the caller can skip it. When the relay does not forward the
// event type the payload is decoded as workflow_run unconditionally.
func ReadWorkflowRunEvent(ctx context.Context, relayURL string, opts ...operations.Option) (gogithub.WorkflowRunEvent, error) {
	eventType, body, err := ReadWebhookEvent(ctx, relayURL, opts...)
	if err != nil {
		return gogithub.WorkflowRunEvent{}, err
	}
	if eventType != "" && eventType != "workflow_run" {
		return gogithub.WorkflowRunEvent{}, fmt.Errorf("%w: %q", ErrUnexpectedEvent, eventType)
	}
	var event gogithub.WorkflowRunEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return gogithub.WorkflowRunEvent{}, fmt.Errorf("failed to decode workflow_run event from relay %q: %w", relayURL, err)
	}
	return event, nil
}

// ReadWorkflowJobEvent performs a single long-poll ("hanging read") GET against
// a webhooks.Relay polling endpoint at relayURL and decodes the delivered
// payload as a workflow_job webhook event. It is a convenience wrapper around
// ReadWebhookEvent for callers that consume only workflow_job events; see that
// function for the blocking, options, and shutdown semantics.
//
// If the relay reports an event type (via the forwarded X-GitHub-Event header)
// that is not workflow_job, it returns ErrUnexpectedEvent wrapped with the
// delivered type so the caller can skip it. When the relay does not forward the
// event type the payload is decoded as workflow_job unconditionally.
func ReadWorkflowJobEvent(ctx context.Context, relayURL string, opts ...operations.Option) (gogithub.WorkflowJobEvent, error) {
	eventType, body, err := ReadWebhookEvent(ctx, relayURL, opts...)
	if err != nil {
		return gogithub.WorkflowJobEvent{}, err
	}
	if eventType != "" && eventType != "workflow_job" {
		return gogithub.WorkflowJobEvent{}, fmt.Errorf("%w: %q", ErrUnexpectedEvent, eventType)
	}
	var event gogithub.WorkflowJobEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return gogithub.WorkflowJobEvent{}, fmt.Errorf("failed to decode workflow_job event from relay %q: %w", relayURL, err)
	}
	return event, nil
}
