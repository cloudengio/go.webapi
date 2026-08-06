// Copyright 2026 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package githubcmd

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strings"

	"cloudeng.io/webapi/clients/github"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/operations/apicrawlcmd"
	gogithub "github.com/google/go-github/v89/github"
)

// Command implements the GitHub Actions API command line operations.
type Command struct {
	config apicrawlcmd.Crawl[Service]
}

// NewCommand returns a new Command for GitHub Actions API commands.
func NewCommand(ctx context.Context, config apicrawlcmd.Crawl[Service]) (*Command, error) {
	return &Command{config: config}, nil
}

// ListRunsFlags are the flags for the ListRuns command.
type ListRunsFlags struct {
	Branch   string `subcmd:"branch,,'filter runs by branch name'"`
	Status   string `subcmd:"status,,'filter by status: completed, in_progress, queued, etc.'"`
	Event    string `subcmd:"event,,'filter by triggering event: push, pull_request, schedule, etc.'"`
	Actor    string `subcmd:"actor,,'filter by the login of the user who triggered the run'"`
	PageSize int    `subcmd:"size,30,'number of items per page (max 100)'"`
}

// ListJobsFlags are the flags for the ListJobs command.
type ListJobsFlags struct {
	Filter   string `subcmd:"filter,latest,'jobs to include: latest (default) or all attempts'"`
	PageSize int    `subcmd:"size,30,'number of items per page (max 100)'"`
}

// ListRunnersFlags are the flags for the ListRunners command.
type ListRunnersFlags struct {
	PageSize int `subcmd:"size,30,'number of items per page (max 100)'"`
}

func (c *Command) cfg() apicrawlcmd.Crawl[Service] {
	return c.config
}

func (c *Command) perPage(flagSize int) int {
	if flagSize > 0 {
		return flagSize
	}
	if c.cfg().Service.PerPage > 0 {
		return c.cfg().Service.PerPage
	}
	return DefaultPageSize
}

// GetRuns returns an iterator over the workflow runs for each run ID supplied as
// an argument. Each run is yielded with any error encountered while fetching it;
// the caller decides whether to stop on error.
func (c *Command) GetRuns(ctx context.Context, args []string) iter.Seq2[gogithub.WorkflowRun, error] {
	return func(yield func(gogithub.WorkflowRun, error) bool) {
		opts, err := OptionsForEndpoint(c.cfg())
		if err != nil {
			yield(gogithub.WorkflowRun{}, err)
			return
		}
		ep := operations.NewEndpoint[gogithub.WorkflowRun](opts...)
		svc := c.cfg().Service
		for _, id := range args {
			u := fmt.Sprintf("%s/repos/%s/%s/actions/runs/%s",
				github.APIHost,
				url.PathEscape(svc.Owner), url.PathEscape(svc.Repo), url.PathEscape(id))
			run, _, _, err := ep.Get(ctx, u)
			if err != nil {
				err = fmt.Errorf("get run %s: %w", id, err)
			}
			if !yield(run, err) {
				return
			}
		}
	}
}

func getPageSize(ps int) int {
	if ps < 1 || ps > 100 {
		return DefaultPageSize
	}
	return ps
}

// ListRuns returns an iterator over all workflow runs for the configured repo
// and a function that, once iteration has completed, reports the detail of the
// first error encountered (see operations.Scanner.ErrDetail): the response body
// and request that caused it along with the error itself. Optional filters can
// be applied via ListRunsFlags and pagination is handled transparently.
func (c *Command) ListRuns(ctx context.Context, fv ListRunsFlags) (iter.Seq[gogithub.WorkflowRun], func() ([]byte, *http.Request, error)) {
	pageSize := getPageSize(fv.PageSize)
	var (
		body []byte
		req  *http.Request
		err  error
	)
	seq := func(yield func(gogithub.WorkflowRun) bool) {
		opts, oerr := OptionsForEndpoint(c.cfg())
		if oerr != nil {
			err = oerr
			return
		}
		svc := c.cfg().Service
		filter := github.RunsFilter{
			Actor:  fv.Actor,
			Branch: fv.Branch,
			Event:  fv.Event,
			Status: fv.Status,
		}
		scanner := github.NewRunsScanner(svc.Owner, svc.Repo, c.perPage(pageSize), filter, opts...)
		for scanner.Scan(ctx) {
			page := scanner.Response()
			for _, run := range page.WorkflowRuns {
				if !yield(*run) {
					return
				}
			}
		}
		body, req, err = scanner.ErrDetail()
	}
	return seq, func() ([]byte, *http.Request, error) { return body, req, err }
}

// GetJobs returns an iterator over the jobs for each job ID supplied as an
// argument. Each job is yielded with any error encountered while fetching it;
// the caller decides whether to stop on error.
func (c *Command) GetJobs(ctx context.Context, args []string) iter.Seq2[gogithub.WorkflowJob, error] {
	return func(yield func(gogithub.WorkflowJob, error) bool) {
		opts, err := OptionsForEndpoint(c.cfg())
		if err != nil {
			yield(gogithub.WorkflowJob{}, err)
			return
		}
		ep := operations.NewEndpoint[gogithub.WorkflowJob](opts...)
		svc := c.cfg().Service
		for _, id := range args {
			u := fmt.Sprintf("%s/repos/%s/%s/actions/jobs/%s",
				github.APIHost,
				url.PathEscape(svc.Owner), url.PathEscape(svc.Repo), url.PathEscape(id))
			job, _, _, err := ep.Get(ctx, u)
			if err != nil {
				err = fmt.Errorf("get job %s: %w", id, err)
			}
			if !yield(job, err) {
				return
			}
		}
	}
}

// ListJobs returns an iterator over all jobs for the specified workflow run ID
// and a function that, once iteration has completed, reports the detail of the
// first error encountered (see operations.Scanner.ErrDetail): the response body
// and request that caused it along with the error itself. Pagination is handled
// transparently.
func (c *Command) ListJobs(ctx context.Context, fv ListJobsFlags, runID int64) (iter.Seq[gogithub.WorkflowJob], func() ([]byte, *http.Request, error)) {
	pageSize := getPageSize(fv.PageSize)
	var (
		body []byte
		req  *http.Request
		err  error
	)
	seq := func(yield func(gogithub.WorkflowJob) bool) {
		opts, oerr := OptionsForEndpoint(c.cfg())
		if oerr != nil {
			err = oerr
			return
		}
		svc := c.cfg().Service
		scanner := github.NewJobsScanner(svc.Owner, svc.Repo, runID, fv.Filter, c.perPage(pageSize), opts...)
		for scanner.Scan(ctx) {
			page := scanner.Response()
			for _, job := range page.Jobs {
				if !yield(*job) {
					return
				}
			}
		}
		body, req, err = scanner.ErrDetail()
	}
	return seq, func() ([]byte, *http.Request, error) { return body, req, err }
}

// CreateWebhookFlags are the flags for the CreateWebhook command.
type CreateWebhookFlags struct {
	URL         string `subcmd:"url,,'webhook delivery URL (required)'"`
	ContentType string `subcmd:"content-type,json,'payload content type: json or form'"`
	Events      string `subcmd:"events,push,'comma-separated list of events to trigger on'"`
	Inactive    bool   `subcmd:"inactive,,'create the webhook in an inactive state'"`
}

// CreateWebhook creates a new HTTP webhook for the configured owner/repo and
// prints the resulting webhook ID, URL, active state, and events to stdout.
func (c *Command) CreateWebhook(ctx context.Context, secret string, fv CreateWebhookFlags) (gogithub.Hook, error) {
	if fv.URL == "" {
		return gogithub.Hook{}, fmt.Errorf("--url is required")
	}
	opts, err := OptionsForEndpoint(c.cfg())
	if err != nil {
		return gogithub.Hook{}, err
	}
	svc := c.cfg().Service
	events := strings.Split(fv.Events, ",")
	request := &gogithub.Hook{
		Name:   gogithub.Ptr("web"),
		Active: gogithub.Ptr(!fv.Inactive),
		Events: events,
		Config: &gogithub.HookConfig{
			URL:         gogithub.Ptr(fv.URL),
			ContentType: gogithub.Ptr(fv.ContentType),
			Secret:      gogithub.Ptr(secret),
		},
	}
	hook, err := github.CreateWebhook(ctx, svc.Owner, svc.Repo, request, opts...)
	if err != nil {
		return gogithub.Hook{}, err
	}

	return hook, nil
}

// CreateRegistrationTokenFlags are the flags for the CreateRegistrationToken
// command.
type CreateRegistrationTokenFlags struct{}

// CreateRegistrationToken requests a new self-hosted runner registration token
// for the configured owner/repo and prints the token and its expiry to stdout.
func (c *Command) CreateRegistrationToken(ctx context.Context) (gogithub.RegistrationToken, error) {
	opts, err := OptionsForEndpoint(c.cfg())
	if err != nil {
		return gogithub.RegistrationToken{}, err
	}
	svc := c.cfg().Service
	tok, err := github.CreateRegistrationToken(ctx, svc.Owner, svc.Repo, opts...)
	if err != nil {
		return gogithub.RegistrationToken{}, err
	}
	return tok, nil
}

const DefaultPageSize = 30

// ListRunners returns an iterator over all self-hosted runners for the
// configured repo and a function that, once iteration has completed, reports the
// detail of the first error encountered (see operations.Scanner.ErrDetail): the
// response body and request that caused it along with the error itself.
// Pagination is handled transparently.
func (c *Command) ListRunners(ctx context.Context, fv ListRunnersFlags) (iter.Seq[gogithub.Runner], func() ([]byte, *http.Request, error)) {
	pageSize := getPageSize(fv.PageSize)
	var (
		body []byte
		req  *http.Request
		err  error
	)
	seq := func(yield func(gogithub.Runner) bool) {
		opts, oerr := OptionsForEndpoint(c.cfg())
		if oerr != nil {
			err = oerr
			return
		}
		svc := c.cfg().Service
		scanner := github.NewRunnersScanner(svc.Owner, svc.Repo, c.perPage(pageSize), opts...)
		for scanner.Scan(ctx) {
			page := scanner.Response()
			for _, runner := range page.Runners {
				if !yield(*runner) {
					return
				}
			}
		}
		body, req, err = scanner.ErrDetail()
	}
	return seq, func() ([]byte, *http.Request, error) { return body, req, err }
}
