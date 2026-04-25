// Copyright 2026 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"cloudeng.io/webapi/operations"
)


const APIHost = "https://api.github.com"

// Actor represents a GitHub user or app that triggered a workflow run.
type Actor struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
	Type  string `json:"type"`
}

// HeadCommit contains information about the commit that triggered a run.
type HeadCommit struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Author    struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"author"`
}

// WorkflowRun represents a single GitHub Actions workflow run.
type WorkflowRun struct {
	ID           int64      `json:"id"`
	Name         string     `json:"name"`
	HeadBranch   string     `json:"head_branch"`
	HeadSHA      string     `json:"head_sha"`
	RunNumber    int        `json:"run_number"`
	RunAttempt   int        `json:"run_attempt"`
	Status       string     `json:"status"`
	Conclusion   string     `json:"conclusion"`
	WorkflowID   int64      `json:"workflow_id"`
	WorkflowName string     `json:"workflow_name"`
	URL          string     `json:"url"`
	HTMLURL      string     `json:"html_url"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
	RunStartedAt *time.Time `json:"run_started_at"`
	Event        string     `json:"event"`
	Actor        Actor      `json:"actor"`
	HeadCommit   HeadCommit `json:"head_commit"`
}

// WorkflowRunsResponse is the response from the list workflow runs endpoint.
type WorkflowRunsResponse struct {
	TotalCount   int           `json:"total_count"`
	WorkflowRuns []WorkflowRun `json:"workflow_runs"`
}

// Step represents a single step within a GitHub Actions job.
type Step struct {
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	Conclusion  string     `json:"conclusion"`
	Number      int        `json:"number"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

// Job represents a single GitHub Actions job within a workflow run.
type Job struct {
	ID              int64      `json:"id"`
	RunID           int64      `json:"run_id"`
	Name            string     `json:"name"`
	Status          string     `json:"status"`
	Conclusion      string     `json:"conclusion"`
	StartedAt       *time.Time `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at"`
	HTMLURL         string     `json:"html_url"`
	Steps           []Step     `json:"steps"`
	RunnerName      string     `json:"runner_name"`
	RunnerGroupName string     `json:"runner_group_name"`
	WorkflowName    string     `json:"workflow_name"`
	HeadBranch      string     `json:"head_branch"`
	HeadSHA         string     `json:"head_sha"`
}

// JobsResponse is the response from the list jobs for a workflow run endpoint.
type JobsResponse struct {
	TotalCount int   `json:"total_count"`
	Jobs       []Job `json:"jobs"`
}

// RunnerLabel represents a label assigned to a self-hosted runner.
type RunnerLabel struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// Runner represents a GitHub Actions self-hosted runner.
type Runner struct {
	ID     int64         `json:"id"`
	Name   string        `json:"name"`
	OS     string        `json:"os"`
	Status string        `json:"status"`
	Busy   bool          `json:"busy"`
	Labels []RunnerLabel `json:"labels"`
}

// RunnersResponse is the response from the list runners endpoint.
type RunnersResponse struct {
	TotalCount int      `json:"total_count"`
	Runners    []Runner `json:"runners"`
}

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
func NewRunsScanner(owner, repo string, perPage int, filter RunsFilter, opts ...operations.Option) *operations.Scanner[WorkflowRunsResponse] {
	u := fmt.Sprintf("%s/repos/%s/%s/actions/runs",
		APIHost,
		url.PathEscape(owner), url.PathEscape(repo))
	u = appendQueryParam(u, "actor", filter.Actor)
	u = appendQueryParam(u, "branch", filter.Branch)
	u = appendQueryParam(u, "event", filter.Event)
	u = appendQueryParam(u, "status", filter.Status)
	return operations.NewScanner(
		&linkPaginator[WorkflowRunsResponse]{initialURL: u, perPage: perPage}, opts...)
}

// NewJobsScanner returns an operations.Scanner that iterates over jobs for the
// specified workflow run. filter may be "latest" (the default) or "all" to
// include jobs from all prior run attempts.
func NewJobsScanner(owner, repo string, runID int64, filter string, perPage int, opts ...operations.Option) *operations.Scanner[JobsResponse] {
	u := fmt.Sprintf("%s/repos/%s/%s/actions/runs/%d/jobs",
		APIHost,
		url.PathEscape(owner), url.PathEscape(repo), runID)
	if filter != "" {
		u = u + "?filter=" + filter
	}
	return operations.NewScanner(
		&linkPaginator[JobsResponse]{initialURL: u, perPage: perPage}, opts...)
}

// NewRunnersScanner returns an operations.Scanner that iterates over self-hosted
// runners registered for the specified owner and repo.
func NewRunnersScanner(owner, repo string, perPage int, opts ...operations.Option) *operations.Scanner[RunnersResponse] {
	u := fmt.Sprintf("%s/repos/%s/%s/actions/runners", APIHost,
		url.PathEscape(owner), url.PathEscape(repo))
	return operations.NewScanner(
		&linkPaginator[RunnersResponse]{initialURL: u, perPage: perPage}, opts...)
}

// RegistrationToken is the response from the runner registration-token endpoint.
type RegistrationToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// CreateRegistrationToken requests a new runner registration token for the
// given owner/repo. Options (including WithAuth) follow the same pattern as
// NewRunsScanner, NewRunnersScanner, and the other functions in this package.
func CreateRegistrationToken(ctx context.Context, owner, repo string, opts ...operations.Option) (RegistrationToken, error) {
	u := fmt.Sprintf("%s/repos/%s/%s/actions/runners/registration-token",
		APIHost, url.PathEscape(owner), url.PathEscape(repo))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, nil)
	if err != nil {
		return RegistrationToken{}, err
	}
	ep := operations.NewEndpoint[RegistrationToken](opts...)
	tok, body, _, _, err := ep.IssueRequest(ctx, req)
	if err == nil {
		return tok, nil
	}
	// GitHub returns 201 Created for this endpoint; IssueRequest treats any
	// non-200 status as an error but still returns the pre-read body bytes.
	if opErr, ok := err.(*operations.Error); ok && opErr.StatusCode == http.StatusCreated {
		if jsonErr := json.Unmarshal(body, &tok); jsonErr != nil {
			return RegistrationToken{}, jsonErr
		}
		return tok, nil
	}
	return RegistrationToken{}, err
}
