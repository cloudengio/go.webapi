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

### Func VerifyWebhookSignature
```go
func VerifyWebhookSignature(secret string, body []byte, signature string) bool
```
VerifyWebhookSignature reports whether the X-Hub-Signature-256 header value
matches the HMAC-SHA256 of body computed with secret. This is the check a
relay or handler performs on receipt.



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
	KeyUser string
	KeyID   string
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




### Type CreateWebhookRequest
```go
type CreateWebhookRequest struct {
	Name   string        `json:"name"`
	Active bool          `json:"active"`
	Events []string      `json:"events"`
	Config WebhookConfig `json:"config"`
}
```
CreateWebhookRequest is the request body for creating a repository webhook.


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

### Functions

```go
func MockJob(owner, repo string) Job
```
MockJob returns a Job populated with typical values for use in tests.
Callers may overwrite any field before passing it to MockWebhook.JobRequest.




### Type JobsResponse
```go
type JobsResponse struct {
	TotalCount int   `json:"total_count"`
	Jobs       []Job `json:"jobs"`
}
```
JobsResponse is the response from the list jobs for a workflow run endpoint.


### Type Meta
```go
type Meta struct {
	VerifiablePasswordAuthentication bool               `json:"verifiable_password_authentication"`
	SSHKeyFingerprints               SSHKeyFingerprints `json:"ssh_key_fingerprints"`
	SSHKeys                          []string           `json:"ssh_keys"`
	Hooks                            []string           `json:"hooks"`
	Web                              []string           `json:"web"`
	API                              []string           `json:"api"`
	Git                              []string           `json:"git"`
	Packages                         []string           `json:"packages"`
	Pages                            []string           `json:"pages"`
	Importer                         []string           `json:"importer"`
	Actions                          []string           `json:"actions"`
	Dependabot                       []string           `json:"dependabot"`
	Domains                          MetaDomains        `json:"domains"`
}
```
Meta is the response from the GET /meta endpoint. IP ranges are in CIDR
notation.

### Functions

```go
func GetMeta(ctx context.Context, opts ...operations.Option) (Meta, error)
```
GetMeta returns GitHub's meta information including IP ranges used by GitHub
services and SSH host key fingerprints.




### Type MetaDomains
```go
type MetaDomains struct {
	Website    []string `json:"website"`
	Codespaces []string `json:"codespaces"`
	Copilot    []string `json:"copilot"`
	Packages   []string `json:"packages"`
}
```
MetaDomains holds the domain names used by various GitHub services.


### Type MockWebhook
```go
type MockWebhook struct {
	// contains filtered or unexported fields
}
```
MockWebhook creates signed HTTP POST requests that mimic GitHub webhook
deliveries. It is intended for testing webhook relays and handlers.

### Functions

```go
func NewMockWebhook(owner, repo, secret string) *MockWebhook
```
NewMockWebhook returns a MockWebhook for the given owner/repo. secret is the
webhook secret used to produce X-Hub-Signature-256 headers; pass an empty
string to skip signing.



### Methods

```go
func (m *MockWebhook) JobRequest(ctx context.Context, targetURL, action string, job Job) (*http.Request, error)
```
JobRequest returns a signed HTTP POST request to targetURL for a
workflow_job event. action must be one of "queued", "in_progress",
or "completed".


```go
func (m *MockWebhook) RunRequest(ctx context.Context, targetURL, action string, run WorkflowRun) (*http.Request, error)
```
RunRequest returns a signed HTTP POST request to targetURL for a
workflow_run event. action must be one of "requested", "in_progress",
or "completed".




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
func CreateRegistrationToken(ctx context.Context, owner, repo string, opts ...operations.Option) (RegistrationToken, error)
```
CreateRegistrationToken requests a new runner registration token for the
given owner/repo. Options (including WithAuth) follow the same pattern as
NewRunsScanner, NewRunnersScanner, and the other functions in this package.




### Type Repository
```go
type Repository struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	HTMLURL  string `json:"html_url"`
	Private  bool   `json:"private"`
	Owner    Actor  `json:"owner"`
}
```
Repository represents the repository information included in webhook
payloads.


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


### Type SSHKeyFingerprints
```go
type SSHKeyFingerprints struct {
	SHA256RSA     string `json:"SHA256_RSA"`
	SHA256DSA     string `json:"SHA256_DSA"`
	SHA256ECDSA   string `json:"SHA256_ECDSA"`
	SHA256ED25519 string `json:"SHA256_ED25519"`
}
```
SSHKeyFingerprints holds the SHA256 fingerprints for each of GitHub's host
keys.


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


### Type Webhook
```go
type Webhook struct {
	ID        int64         `json:"id"`
	Name      string        `json:"name"`
	Active    bool          `json:"active"`
	Events    []string      `json:"events"`
	Config    WebhookConfig `json:"config"`
	CreatedAt *time.Time    `json:"created_at"`
	UpdatedAt *time.Time    `json:"updated_at"`
}
```
Webhook is the response from the create/get repository webhook endpoints.

### Functions

```go
func CreateWebhook(ctx context.Context, owner, repo string, request CreateWebhookRequest, opts ...operations.Option) (Webhook, error)
```
CreateWebhook creates a new webhook for the given owner/repo. The Name field
in the request must be "web" for HTTP webhooks. GitHub returns 201 Created
on success.




### Type WebhookConfig
```go
type WebhookConfig struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type,omitempty"`
	Secret      string `json:"secret,omitempty"`
	InsecureSSL string `json:"insecure_ssl,omitempty"`
}
```
WebhookConfig holds the delivery configuration for a repository webhook.


### Type WorkflowJobEvent
```go
type WorkflowJobEvent struct {
	Action      string     `json:"action"`
	WorkflowJob Job        `json:"workflow_job"`
	Repository  Repository `json:"repository"`
	Sender      Actor      `json:"sender"`
}
```
WorkflowJobEvent is the payload delivered to a webhook for workflow_job
events. Action is one of "queued", "in_progress", or "completed".


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

### Functions

```go
func MockRun(owner, repo string) WorkflowRun
```
MockRun returns a WorkflowRun populated with typical values for
use in tests. Callers may overwrite any field before passing it to
MockWebhook.RunRequest.




### Type WorkflowRunEvent
```go
type WorkflowRunEvent struct {
	Action      string      `json:"action"`
	WorkflowRun WorkflowRun `json:"workflow_run"`
	Repository  Repository  `json:"repository"`
	Sender      Actor       `json:"sender"`
}
```
WorkflowRunEvent is the payload delivered to a webhook for workflow_run
events. Action is one of "requested", "in_progress", or "completed".


### Type WorkflowRunsResponse
```go
type WorkflowRunsResponse struct {
	TotalCount   int           `json:"total_count"`
	WorkflowRuns []WorkflowRun `json:"workflow_runs"`
}
```
WorkflowRunsResponse is the response from the list workflow runs endpoint.





