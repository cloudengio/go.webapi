# Package [cloudeng.io/webapi/clients/github/githubcmd](https://pkg.go.dev/cloudeng.io/webapi/clients/github/githubcmd?tab=doc)

```go
import cloudeng.io/webapi/clients/github/githubcmd
```

Package githubcmd provides support for building command line tools that
access the GitHub Actions API.

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
func NewCommand(ctx context.Context, config apicrawlcmd.Crawl[yaml.Node], resources apicrawlcmd.Resources) (*Command, error)
```
NewCommand returns a new Command for GitHub Actions API commands.



### Methods

```go
func (c *Command) GetJob(ctx context.Context, _ *GetJobFlags, args []string) error
```
GetJob retrieves the job for each job ID supplied as an argument and prints
it to stdout.


```go
func (c *Command) GetRun(ctx context.Context, _ *GetRunFlags, args []string) error
```
GetRun retrieves the workflow run for each run ID supplied as an argument
and prints it to stdout.


```go
func (c *Command) ListJobs(ctx context.Context, fv *ListJobsFlags, runID int64) error
```
ListJobs iterates over all jobs for the specified workflow run ID and prints
each job to stdout.


```go
func (c *Command) ListRunners(ctx context.Context, fv *ListRunnersFlags) error
```
ListRunners iterates over all self-hosted runners for the configured repo
and prints each runner to stdout.


```go
func (c *Command) ListRuns(ctx context.Context, fv *ListRunsFlags) error
```
ListRuns iterates over all workflow runs for the configured repo and prints
each run to stdout. Runs are retrieved using the scanner/paginator pattern
and optional filters can be applied via ListRunsFlags.




### Type GetJobFlags
```go
type GetJobFlags struct{}
```
GetJobFlags are the flags for the GetJob command.


### Type GetRunFlags
```go
type GetRunFlags struct{}
```
GetRunFlags are the flags for the GetRun command.


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
	Owner   string `yaml:"owner" cmd:"repository owner (user or organization)"`
	Repo    string `yaml:"repo" cmd:"repository name"`
	PerPage int    `yaml:"per_page" cmd:"number of results per page (max 100, default 30)"`
}
```
Service represents the GitHub-specific configuration for API access.





