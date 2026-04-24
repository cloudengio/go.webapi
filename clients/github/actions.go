// Copyright 2026 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"fmt"
	"net/http"
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
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	RunStartedAt time.Time  `json:"run_started_at"`
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
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	Conclusion  string    `json:"conclusion"`
	Number      int       `json:"number"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
}

// Job represents a single GitHub Actions job within a workflow run.
type Job struct {
	ID              int64     `json:"id"`
	RunID           int64     `json:"run_id"`
	Name            string    `json:"name"`
	Status          string    `json:"status"`
	Conclusion      string    `json:"conclusion"`
	StartedAt       time.Time `json:"started_at"`
	CompletedAt     time.Time `json:"completed_at"`
	HTMLURL         string    `json:"html_url"`
	Steps           []Step    `json:"steps"`
	RunnerName      string    `json:"runner_name"`
	RunnerGroupName string    `json:"runner_group_name"`
	WorkflowName    string    `json:"workflow_name"`
	HeadBranch      string    `json:"head_branch"`
	HeadSHA         string    `json:"head_sha"`
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

// runsLinkPaginator implements operations.Paginator[WorkflowRunsResponse] using
// GitHub's Link header pagination.
type runsLinkPaginator struct {
	initialURL string
	perPage    int
}

func (p *runsLinkPaginator) Next(_ context.Context, _ WorkflowRunsResponse, r *http.Response) (*http.Request, bool, error) {
	if r == nil {
		req, err := http.NewRequest("GET", appendPerPage(p.initialURL, p.perPage), nil)
		return req, false, err
	}
	next := parseLinkHeader(r.Header.Get("Link"))
	if next == "" {
		return nil, true, nil
	}
	req, err := http.NewRequest("GET", next, nil)
	return req, false, err
}

// jobsLinkPaginator implements operations.Paginator[JobsResponse].
type jobsLinkPaginator struct {
	initialURL string
	perPage    int
}

func (p *jobsLinkPaginator) Next(_ context.Context, _ JobsResponse, r *http.Response) (*http.Request, bool, error) {
	if r == nil {
		req, err := http.NewRequest("GET", appendPerPage(p.initialURL, p.perPage), nil)
		return req, false, err
	}
	next := parseLinkHeader(r.Header.Get("Link"))
	if next == "" {
		return nil, true, nil
	}
	req, err := http.NewRequest("GET", next, nil)
	return req, false, err
}

// runnersLinkPaginator implements operations.Paginator[RunnersResponse].
type runnersLinkPaginator struct {
	initialURL string
	perPage    int
}

func (p *runnersLinkPaginator) Next(_ context.Context, _ RunnersResponse, r *http.Response) (*http.Request, bool, error) {
	if r == nil {
		req, err := http.NewRequest("GET", appendPerPage(p.initialURL, p.perPage), nil)
		return req, false, err
	}
	next := parseLinkHeader(r.Header.Get("Link"))
	if next == "" {
		return nil, true, nil
	}
	req, err := http.NewRequest("GET", next, nil)
	return req, false, err
}

// NewRunsScanner returns an operations.Scanner that iterates over workflow runs
// for the specified owner and repo, one page at a time. Additional query
// parameters (branch, status, event, actor, etc.) can be appended to the URL
// by the caller before passing it to the scanner; alternatively pass them via
// the opts slice using a custom paginator.
func NewRunsScanner(owner, repo string, perPage int, opts ...operations.Option) *operations.Scanner[WorkflowRunsResponse] {
	u := fmt.Sprintf("%s/repos/%s/%s/actions/runs", APIHost, owner, repo)
	return operations.NewScanner[WorkflowRunsResponse](
		&runsLinkPaginator{initialURL: u, perPage: perPage}, opts...)
}

// NewJobsScanner returns an operations.Scanner that iterates over jobs for the
// specified workflow run. filter may be "latest" (the default) or "all" to
// include jobs from all prior run attempts.
func NewJobsScanner(owner, repo string, runID int64, filter string, perPage int, opts ...operations.Option) *operations.Scanner[JobsResponse] {
	u := fmt.Sprintf("%s/repos/%s/%s/actions/runs/%d/jobs", APIHost, owner, repo, runID)
	if filter != "" {
		u = u + "?filter=" + filter
	}
	return operations.NewScanner[JobsResponse](
		&jobsLinkPaginator{initialURL: u, perPage: perPage}, opts...)
}

// NewRunnersScanner returns an operations.Scanner that iterates over self-hosted
// runners registered for the specified owner and repo.
func NewRunnersScanner(owner, repo string, perPage int, opts ...operations.Option) *operations.Scanner[RunnersResponse] {
	u := fmt.Sprintf("%s/repos/%s/%s/actions/runners", APIHost, owner, repo)
	return operations.NewScanner[RunnersResponse](
		&runnersLinkPaginator{initialURL: u, perPage: perPage}, opts...)
}
