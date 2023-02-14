# Package [cloudeng.io/webapi/webapitestutil](https://pkg.go.dev/cloudeng.io/webapi/webapitestutil?tab=doc)

```go
import cloudeng.io/webapi/webapitestutil
```


## Functions
### Func NewEchoHandler
```go
func NewEchoHandler(value any) http.Handler
```
NewEchoHandler returns an http.Handler that echos the json encoded value of
the supplied value.

### Func NewHeaderEchoHandler
```go
func NewHeaderEchoHandler() http.Handler
```
NewHeaderEchoHandler returns an http.Handler that returns the json encoded
value of the request headers as its response body.

### Func NewRetryHandler
```go
func NewRetryHandler(retries int) http.Handler
```
NewRetryHandler returns an http.Handler that returns an
http.StatusTooManyRequests until retries in reached in which case it returns
an http.StatusOK with a body containing the json encoded value of retries.

### Func NewServer
```go
func NewServer(handler http.Handler) *httptest.Server
```
NewServer creates a new httptest.Server using the supplied handler.



## Types
### Type Paginated
```go
type Paginated struct {
	Payload int
	Current int
	Last    int
}
```
Paginated is an example return type for a paginated API.


### Type PaginatedHandler
```go
type PaginatedHandler struct {
	Last int
	// contains filtered or unexported fields
}
```
PaginatedHandler is an example http.Handler that returns a Paginated
response, that is accepts a parameter called 'current' which is the page to
returned. Last should be initialized the number of available pages.

### Methods

```go
func (ph *PaginatedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request)
```







