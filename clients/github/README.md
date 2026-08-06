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



## Variables
### ErrRelayClosed
```go
ErrRelayClosed = errors.New("relay closed the connection without delivering an event")

```
ErrRelayClosed is returned by ReadWebhookEvent when the relay closes the
long-poll connection without delivering an event (for example when the
relay is shutting down). Callers looping over ReadWebhookEvent can use it to
distinguish a clean relay shutdown from a transport or decode error.

### ErrUnexpectedEvent
```go
ErrUnexpectedEvent = errors.New("unexpected webhook event type")

```
ErrUnexpectedEvent is returned by ReadWorkflowRunEvent and
ReadWorkflowJobEvent when the relay delivers an event whose type does not
match the one the caller asked for. It is wrapped with the delivered event
type so callers can match it with errors.Is and skip the delivery.



## Functions
### Func CreateRegistrationToken
```go
func CreateRegistrationToken(ctx context.Context, owner, repo string, opts ...operations.Option) (gogithub.RegistrationToken, error)
```
CreateRegistrationToken requests a new runner registration token for the
given owner/repo. Options (including WithAuth) follow the same pattern as
NewRunsScanner, NewRunnersScanner, and the other functions in this package.

### Func CreateWebhook
```go
func CreateWebhook(ctx context.Context, owner, repo string, request *gogithub.Hook, opts ...operations.Option) (gogithub.Hook, error)
```
CreateWebhook creates a new webhook for the given owner/repo. The Name field
in the request must be "web" for HTTP webhooks. GitHub returns 201 Created
on success.

### Func GetWorkflowJob
```go
func GetWorkflowJob(ctx context.Context, owner, repo string, jobID int64, opts ...operations.Option) (gogithub.WorkflowJob, error)
```
GetWorkflowJob returns the job with the specified ID via the
/repos/{owner}/{repo}/actions/jobs/{job_id} endpoint. Unlike the jobs
returned by NewJobsScanner, which are scoped to a single workflow run,
a job is addressed here by its own ID, as reported by, for example,
the workflow_job webhook event.

### Func MockJob
```go
func MockJob(owner, repo string) *gogithub.WorkflowJob
```
MockJob returns a WorkflowJob populated with typical values for
use in tests. Callers may overwrite any field before passing it to
MockWebhook.JobRequest.

### Func MockRun
```go
func MockRun(owner, repo string) *gogithub.WorkflowRun
```
MockRun returns a WorkflowRun populated with typical values for
use in tests. Callers may overwrite any field before passing it to
MockWebhook.RunRequest.

### Func NewJobsScanner
```go
func NewJobsScanner(owner, repo string, runID int64, filter string, perPage int, opts ...operations.Option) *operations.Scanner[gogithub.Jobs]
```
NewJobsScanner returns an operations.Scanner that iterates over jobs for
the specified workflow run. filter may be "latest" (the default) or "all" to
include jobs from all prior run attempts.

### Func NewRunnersScanner
```go
func NewRunnersScanner(owner, repo string, perPage int, opts ...operations.Option) *operations.Scanner[gogithub.Runners]
```
NewRunnersScanner returns an operations.Scanner that iterates over
self-hosted runners registered for the specified owner and repo.

### Func NewRunsScanner
```go
func NewRunsScanner(owner, repo string, perPage int, filter RunsFilter, opts ...operations.Option) *operations.Scanner[gogithub.WorkflowRuns]
```
NewRunsScanner returns an operations.Scanner that iterates over workflow
runs for the specified owner and repo, one page at a time. Non-empty
fields in filter are sent as query parameters so the GitHub API performs
server-side filtering before any results are returned.

### Func ReadWebhookEvent
```go
func ReadWebhookEvent(ctx context.Context, relayURL string, opts ...operations.Option) (eventType string, payload json.RawMessage, err error)
```
ReadWebhookEvent performs a single long-poll ("hanging read") GET against
a webhooks.Relay polling endpoint at relayURL and returns the GitHub event
type together with the raw, already-verified webhook payload. The relay
validates the webhook signature before queuing deliveries, so the returned
payload is already authenticated.

eventType is taken from the X-GitHub-Event header the relay forwards;
it is empty if the relay is not configured to forward that header. Callers
can switch on eventType to demultiplex different event types (for example
workflow_run vs workflow_job) from a single relay.

The request is issued via an operations.Endpoint, so it shares the same
rate-control, retry/backoff, auth, HTTP client, and logging behaviour
as the rest of the package; configure them via opts (for example
operations.WithHTTPClient or operations.WithRateController). Because a
hanging read can block indefinitely, any configured HTTP client should not
impose a request timeout — use ctx to bound or cancel the call.

It blocks until the relay delivers a payload, ctx is cancelled, or the
request fails, and is intended to be called in a loop to consume a stream of
events. ErrRelayClosed is returned if the relay responds with an empty body,
which indicates a clean shutdown.

### Func ReadWorkflowJobEvent
```go
func ReadWorkflowJobEvent(ctx context.Context, relayURL string, opts ...operations.Option) (gogithub.WorkflowJobEvent, error)
```
ReadWorkflowJobEvent performs a single long-poll ("hanging read") GET
against a webhooks.Relay polling endpoint at relayURL and decodes the
delivered payload as a workflow_job webhook event. It is a convenience
wrapper around ReadWebhookEvent for callers that consume only workflow_job
events; see that function for the blocking, options, and shutdown semantics.

If the relay reports an event type (via the forwarded X-GitHub-Event header)
that is not workflow_job, it returns ErrUnexpectedEvent wrapped with the
delivered type so the caller can skip it. When the relay does not forward
the event type the payload is decoded as workflow_job unconditionally.

### Func ReadWorkflowRunEvent
```go
func ReadWorkflowRunEvent(ctx context.Context, relayURL string, opts ...operations.Option) (gogithub.WorkflowRunEvent, error)
```
ReadWorkflowRunEvent performs a single long-poll ("hanging read") GET
against a webhooks.Relay polling endpoint at relayURL and decodes the
delivered payload as a workflow_run webhook event. It is a convenience
wrapper around ReadWebhookEvent for callers that consume only workflow_run
events; see that function for the blocking, options, and shutdown semantics.

If the relay reports an event type (via the forwarded X-GitHub-Event header)
that is not workflow_run, it returns ErrUnexpectedEvent wrapped with the
delivered type so the caller can skip it. When the relay does not forward
the event type the payload is decoded as workflow_run unconditionally.

### Func RerunWorkflowJob
```go
func RerunWorkflowJob(ctx context.Context, owner, repo string, jobID int64, opts ...operations.Option) error
```
RerunWorkflowJob requests that the job with the specified ID be rerun,
via the /repos/{owner}/{repo}/actions/jobs/{job_id}/rerun endpoint. GitHub
reruns the job along with any jobs in the same workflow run that depend
on it, and responds 201 Created with an empty body, so there is nothing to
return beyond any error. The rerun is queued asynchronously: a successful
return means GitHub accepted the request, not that the job has started.

### Func VerifyWebhookSignature
```go
func VerifyWebhookSignature(secret string, body []byte, signature string) bool
```
VerifyWebhookSignature reports whether the X-Hub-Signature-256 header value
matches the HMAC-SHA256 of body computed with secret. This is the check a
relay or handler performs on receipt.



## Types
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
func (m *MockWebhook) JobRequest(ctx context.Context, targetURL, action string, job *gogithub.WorkflowJob) (*http.Request, error)
```
JobRequest returns a signed HTTP POST request to targetURL for a
workflow_job event. action must be one of "queued", "in_progress",
or "completed".


```go
func (m *MockWebhook) RunRequest(ctx context.Context, targetURL, action string, run *gogithub.WorkflowRun) (*http.Request, error)
```
RunRequest returns a signed HTTP POST request to targetURL for a
workflow_run event. action must be one of "requested", "in_progress",
or "completed".




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





