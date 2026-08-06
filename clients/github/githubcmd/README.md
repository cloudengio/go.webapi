# Package [cloudeng.io/webapi/clients/github/githubcmd](https://pkg.go.dev/cloudeng.io/webapi/clients/github/githubcmd?tab=doc)

```go
import cloudeng.io/webapi/clients/github/githubcmd
```

Package githubcmd provides support for building command line tools that
access the GitHub Actions API.

## Constants
### DefaultPageSize
```go
DefaultPageSize = 30

```



## Functions
### Func OptionsForEndpoint
```go
func OptionsForEndpoint(cfg apicrawlcmd.Crawl[Service]) ([]operations.Option, error)
```
OptionsForEndpoint returns the operations.Option slice for making API
requests with the auth and rate-control settings from cfg.



## Types
### Type Command
```go
type Command struct {
	// contains filtered or unexported fields
}
```
Command implements the GitHub Actions API command line operations.

### Functions

```go
func NewCommand(ctx context.Context, config apicrawlcmd.Crawl[Service]) (*Command, error)
```
NewCommand returns a new Command for GitHub Actions API commands.



### Methods

```go
func (c *Command) CreateRegistrationToken(ctx context.Context) (gogithub.RegistrationToken, error)
```
CreateRegistrationToken requests a new self-hosted runner registration token
for the configured owner/repo and prints the token and its expiry to stdout.


```go
func (c *Command) CreateWebhook(ctx context.Context, secret string, fv CreateWebhookFlags) (gogithub.Hook, error)
```
CreateWebhook creates a new HTTP webhook for the configured owner/repo and
prints the resulting webhook ID, URL, active state, and events to stdout.


```go
func (c *Command) GetJobs(ctx context.Context, args []string) iter.Seq2[gogithub.WorkflowJob, error]
```
GetJobs returns an iterator over the jobs for each job ID supplied as an
argument. Each job is yielded with any error encountered while fetching it;
the caller decides whether to stop on error.


```go
func (c *Command) GetRuns(ctx context.Context, args []string) iter.Seq2[gogithub.WorkflowRun, error]
```
GetRuns returns an iterator over the workflow runs for each run ID supplied
as an argument. Each run is yielded with any error encountered while
fetching it; the caller decides whether to stop on error.


```go
func (c *Command) ListJobs(ctx context.Context, fv ListJobsFlags, runID int64) (iter.Seq[gogithub.WorkflowJob], func() ([]byte, *http.Request, error))
```
ListJobs returns an iterator over all jobs for the specified workflow
run ID and a function that, once iteration has completed, reports the
detail of the first error encountered (see operations.Scanner.ErrDetail):
the response body and request that caused it along with the error itself.
Pagination is handled transparently.


```go
func (c *Command) ListRunners(ctx context.Context, fv ListRunnersFlags) (iter.Seq[gogithub.Runner], func() ([]byte, *http.Request, error))
```
ListRunners returns an iterator over all self-hosted runners
for the configured repo and a function that, once iteration has
completed, reports the detail of the first error encountered (see
operations.Scanner.ErrDetail): the response body and request that caused it
along with the error itself. Pagination is handled transparently.


```go
func (c *Command) ListRuns(ctx context.Context, fv ListRunsFlags) (iter.Seq[gogithub.WorkflowRun], func() ([]byte, *http.Request, error))
```
ListRuns returns an iterator over all workflow runs for the configured
repo and a function that, once iteration has completed, reports the
detail of the first error encountered (see operations.Scanner.ErrDetail):
the response body and request that caused it along with the error itself.
Optional filters can be applied via ListRunsFlags and pagination is handled
transparently.




### Type CreateRegistrationTokenFlags
```go
type CreateRegistrationTokenFlags struct{}
```
CreateRegistrationTokenFlags are the flags for the CreateRegistrationToken
command.


### Type CreateWebhookFlags
```go
type CreateWebhookFlags struct {
	URL         string `subcmd:"url,,'webhook delivery URL (required)'"`
	ContentType string `subcmd:"content-type,json,'payload content type: json or form'"`
	Events      string `subcmd:"events,push,'comma-separated list of events to trigger on'"`
	Inactive    bool   `subcmd:"inactive,,'create the webhook in an inactive state'"`
}
```
CreateWebhookFlags are the flags for the CreateWebhook command.


### Type ListJobsFlags
```go
type ListJobsFlags struct {
	Filter   string `subcmd:"filter,latest,'jobs to include: latest (default) or all attempts'"`
	PageSize int    `subcmd:"size,30,'number of items per page (max 100)'"`
}
```
ListJobsFlags are the flags for the ListJobs command.


### Type ListRunnersFlags
```go
type ListRunnersFlags struct {
	PageSize int `subcmd:"size,30,'number of items per page (max 100)'"`
}
```
ListRunnersFlags are the flags for the ListRunners command.


### Type ListRunsFlags
```go
type ListRunsFlags struct {
	Branch   string `subcmd:"branch,,'filter runs by branch name'"`
	Status   string `subcmd:"status,,'filter by status: completed, in_progress, queued, etc.'"`
	Event    string `subcmd:"event,,'filter by triggering event: push, pull_request, schedule, etc.'"`
	Actor    string `subcmd:"actor,,'filter by the login of the user who triggered the run'"`
	PageSize int    `subcmd:"size,30,'number of items per page (max 100)'"`
}
```
ListRunsFlags are the flags for the ListRuns command.


### Type Service
```go
type Service struct {
	Owner   string `yaml:"owner" doc:"repository owner, the organization or user name"`
	Repo    string `yaml:"repo" doc:"repository name"`
	PerPage int    `yaml:"per_page" doc:"number of results per page (max 100, default 30)"`
}
```
Service represents the GitHub-specific configuration for API access.

### Methods

```go
func (s Service) Validate() error
```







