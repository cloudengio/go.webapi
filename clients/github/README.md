# Package [cloudeng.io/webapi/clients/github](https://pkg.go.dev/cloudeng.io/webapi/clients/github?tab=doc)

```go
import cloudeng.io/webapi/clients/github
```

Package github provides a client for the GitHub REST API, with a focus on
the Actions endpoints (runs, jobs, runners).

## Constants
### APIHost
```go
APIHost = "https://api.github.com"

```



## Functions
### Func NewJobsScanner
```go
func NewJobsScanner(owner, repo string, runID int64, filter string, perPage int, opts ...operations.Option) *operations.Scanner[JobsResponse]
```
NewJobsScanner returns an operations.Scanner that iterates over jobs for
the specified workflow run. filter may be "latest" (the default) or "all" to
include jobs from all prior run attempts.

### Func NewRunnersScanner
```go
func NewRunnersScanner(owner, repo string, perPage int, opts ...operations.Option) *operations.Scanner[RunnersResponse]
```
NewRunnersScanner returns an operations.Scanner that iterates over
self-hosted runners registered for the specified owner and repo.

### Func NewRunsScanner
```go
func NewRunsScanner(owner, repo string, perPage int, filter RunsFilter, opts ...operations.Option) *operations.Scanner[WorkflowRunsResponse]
```
NewRunsScanner returns an operations.Scanner that iterates over workflow
runs for the specified owner and repo, one page at a time. Non-empty
fields in filter are sent as query parameters so the GitHub API performs
server-side filtering before any results are returned.



## Types
### Type Actor
```go
type Actor struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
	Type  string `json:"type"`
}
```
Actor represents a GitHub user or app that triggered a workflow run.


### Type BearerToken
```go
type BearerToken struct {
	KeyID string
}
```
BearerToken implements operations.Auth for GitHub personal access tokens and
GitHub Apps installation tokens. The token is retrieved from the context via
the apitokens package using the configured KeyID.

### Methods

```go
func (bt BearerToken) WithAuthorization(ctx context.Context, req *http.Request) error
```
WithAuthorization implements operations.Auth. It sets the Authorization
header and the required GitHub API headers on the request.




### Type HeadCommit
```go
type HeadCommit struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Author    struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"author"`
}
```
HeadCommit contains information about the commit that triggered a run.


### Type Job
```go
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
```
Job represents a single GitHub Actions job within a workflow run.


### Type JobsResponse
```go
type JobsResponse struct {
	TotalCount int   `json:"total_count"`
	Jobs       []Job `json:"jobs"`
}
```
JobsResponse is the response from the list jobs for a workflow run endpoint.


### Type RegistrationToken
```go
type RegistrationToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}
```
RegistrationToken is the response from the runner registration-token
endpoint.

### Functions

```go
func CreateRegistrationToken(ctx context.Context, owner, repo string, auth operations.Auth) (RegistrationToken, error)
```
CreateRegistrationToken requests a new runner registration token for the
given owner/repo. Auth must be set on ctx via the same mechanism used by the
other functions in this package (e.g. apitokens.ContextWithKey).




### Type Runner
```go
type Runner struct {
	ID     int64         `json:"id"`
	Name   string        `json:"name"`
	OS     string        `json:"os"`
	Status string        `json:"status"`
	Busy   bool          `json:"busy"`
	Labels []RunnerLabel `json:"labels"`
}
```
Runner represents a GitHub Actions self-hosted runner.


### Type RunnerLabel
```go
type RunnerLabel struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}
```
RunnerLabel represents a label assigned to a self-hosted runner.


### Type RunnersResponse
```go
type RunnersResponse struct {
	TotalCount int      `json:"total_count"`
	Runners    []Runner `json:"runners"`
}
```
RunnersResponse is the response from the list runners endpoint.


### Type RunsFilter
```go
type RunsFilter struct {
	Actor  string
	Branch string
	Event  string
	Status string
}
```
RunsFilter holds optional server-side filter parameters for listing workflow
runs.


### Type Step
```go
type Step struct {
	Name        string     `json:"name"`
	Status      string     `json:"status"`
	Conclusion  string     `json:"conclusion"`
	Number      int        `json:"number"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}
```
Step represents a single step within a GitHub Actions job.


### Type WorkflowRun
```go
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
```
WorkflowRun represents a single GitHub Actions workflow run.


### Type WorkflowRunsResponse
```go
type WorkflowRunsResponse struct {
	TotalCount   int           `json:"total_count"`
	WorkflowRuns []WorkflowRun `json:"workflow_runs"`
}
```
WorkflowRunsResponse is the response from the list workflow runs endpoint.





