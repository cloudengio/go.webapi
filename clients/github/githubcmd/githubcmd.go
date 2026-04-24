// Copyright 2026 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package githubcmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloudeng.io/webapi/clients/github"
	"cloudeng.io/webapi/operations"
	"cloudeng.io/webapi/operations/apicrawlcmd"
	"gopkg.in/yaml.v3"
)

// Command implements the GitHub Actions API command line operations.
type Command struct {
	state apicrawlcmd.State[Service]
}

// NewCommand returns a new Command for GitHub Actions API commands.
func NewCommand(ctx context.Context, config apicrawlcmd.Crawl[yaml.Node], resources apicrawlcmd.Resources) (*Command, error) {
	state, err := apicrawlcmd.NewState[Service](ctx, config, resources)
	if err != nil {
		return nil, err
	}
	return &Command{state: state}, nil
}

// GetRunFlags are the flags for the GetRun command.
type GetRunFlags struct{}

// ListRunsFlags are the flags for the ListRuns command.
type ListRunsFlags struct {
	Branch   string `subcmd:"branch,,'filter runs by branch name'"`
	Status   string `subcmd:"status,,'filter by status: completed, in_progress, queued, etc.'"`
	Event    string `subcmd:"event,,'filter by triggering event: push, pull_request, schedule, etc.'"`
	Actor    string `subcmd:"actor,,'filter by the login of the user who triggered the run'"`
	PageSize int    `subcmd:"size,30,'number of items per page (max 100)'"`
}

// GetJobFlags are the flags for the GetJob command.
type GetJobFlags struct{}

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
	return c.state.Config
}

func (c *Command) perPage(flagSize int) int {
	if flagSize > 0 {
		return flagSize
	}
	if c.cfg().Service.PerPage > 0 {
		return c.cfg().Service.PerPage
	}
	return 30
}

// GetRun retrieves the workflow run for each run ID supplied as an argument
// and prints it to stdout.
func (c *Command) GetRun(ctx context.Context, _ *GetRunFlags, args []string) error {
	opts, err := OptionsForEndpoint(c.cfg())
	if err != nil {
		return err
	}
	ep := operations.NewEndpoint[github.WorkflowRun](opts...)
	svc := c.cfg().Service
	for _, id := range args {
		u := fmt.Sprintf("%s/repos/%s/%s/actions/runs/%s", github.APIHost, svc.Owner, svc.Repo, id)
		run, _, _, err := ep.Get(ctx, u)
		if err != nil {
			return fmt.Errorf("get run %s: %w", id, err)
		}
		fmt.Printf("%d\t%s\t%s\t%s\t%s\t%s\n",
			run.ID, run.Name, run.HeadBranch, run.Event, run.Status, run.Conclusion)
	}
	return nil
}

// ListRuns iterates over all workflow runs for the configured repo and prints
// each run to stdout. Runs are retrieved using the scanner/paginator pattern
// and optional filters can be applied via ListRunsFlags.
func (c *Command) ListRuns(ctx context.Context, fv *ListRunsFlags) error {
	opts, err := OptionsForEndpoint(c.cfg())
	if err != nil {
		return err
	}
	svc := c.cfg().Service
	filter := github.RunsFilter{
		Actor:  fv.Actor,
		Branch: fv.Branch,
		Event:  fv.Event,
		Status: fv.Status,
	}
	scanner := github.NewRunsScanner(svc.Owner, svc.Repo, c.perPage(fv.PageSize), filter, opts...)
	for scanner.Scan(ctx) {
		page := scanner.Response()
		for _, run := range page.WorkflowRuns {
			fmt.Printf("%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
				run.ID, run.Name, run.HeadBranch, run.Event,
				run.Status, run.Conclusion, ft(run.CreatedAt))
		}
	}
	return scanner.Err()
}

func ft(t *time.Time) string {
	if t != nil {
		return t.Format("2006-01-02T15:04:05Z")
	}
	return "N/A"
}

// GetJob retrieves the job for each job ID supplied as an argument and prints
// it to stdout.
func (c *Command) GetJob(ctx context.Context, _ *GetJobFlags, args []string) error {
	opts, err := OptionsForEndpoint(c.cfg())
	if err != nil {
		return err
	}
	ep := operations.NewEndpoint[github.Job](opts...)
	svc := c.cfg().Service
	for _, id := range args {
		u := fmt.Sprintf("%s/repos/%s/%s/actions/jobs/%s", github.APIHost, svc.Owner, svc.Repo, id)
		job, _, _, err := ep.Get(ctx, u)
		if err != nil {
			return fmt.Errorf("get job %s: %w", id, err)
		}
		fmt.Printf("%d\t%s\t%s\t%s\t%s\n",
			job.ID, job.Name, job.Status, job.Conclusion, job.RunnerName)
	}
	return nil
}

// ListJobs iterates over all jobs for the specified workflow run ID and prints
// each job to stdout.
func (c *Command) ListJobs(ctx context.Context, fv *ListJobsFlags, runID int64) error {
	opts, err := OptionsForEndpoint(c.cfg())
	if err != nil {
		return err
	}
	svc := c.cfg().Service
	scanner := github.NewJobsScanner(svc.Owner, svc.Repo, runID, fv.Filter, c.perPage(fv.PageSize), opts...)
	for scanner.Scan(ctx) {
		page := scanner.Response()
		for _, job := range page.Jobs {
			fmt.Printf("%d\t%s\t%s\t%s\t%s\n",
				job.ID, job.Name, job.Status, job.Conclusion, job.RunnerName)
		}
	}
	return scanner.Err()
}

// ListRunners iterates over all self-hosted runners for the configured repo
// and prints each runner to stdout.
func (c *Command) ListRunners(ctx context.Context, fv *ListRunnersFlags) error {
	opts, err := OptionsForEndpoint(c.cfg())
	if err != nil {
		return err
	}
	svc := c.cfg().Service
	scanner := github.NewRunnersScanner(svc.Owner, svc.Repo, c.perPage(fv.PageSize), opts...)
	for scanner.Scan(ctx) {
		page := scanner.Response()
		for _, runner := range page.Runners {
			labels := make([]string, len(runner.Labels))
			for i, l := range runner.Labels {
				labels[i] = l.Name
			}
			fmt.Printf("%d\t%s\t%s\t%s\tbusy=%v\t%s\n",
				runner.ID, runner.Name, runner.OS, runner.Status,
				runner.Busy, strings.Join(labels, ","))
		}
	}
	return scanner.Err()
}
