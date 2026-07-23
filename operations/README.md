# Package [cloudeng.io/webapi/operations](https://pkg.go.dev/cloudeng.io/webapi/operations?tab=doc)

```go
import cloudeng.io/webapi/operations
```

Package operations provides support for invoking various operations on web
APIs.

## Functions
### Func RunCrawl
```go
func RunCrawl[ScanT, EndpointT any](ctx context.Context, crawler *Crawler[ScanT, EndpointT], handler CrawlHandler[EndpointT]) error
```
RunCrawl is a convenience function that runs a crawler and calls the
supplied handler for each Object crawled.



## Types
### Type Auth
```go
type Auth interface {
	// WithAuthorization adds an authorization header or other required
	// authorization information to the provided http.Request.
	WithAuthorization(context.Context, *http.Request) error
}
```
Auth represents an authorization mechanism.


### Type CrawlHandler
```go
type CrawlHandler[EndpointT any] func(context.Context, []content.Object[EndpointT, Response]) error
```


### Type Crawler
```go
type Crawler[ScanT any, EndpointT any] struct {
	// contains filtered or unexported fields
}
```
Crawler is a generic crawler that can be used to iterate over a paginated
API endpoint (using a Scanner) that enumerates objects that can be
downloaded using a Fetcher. The Fetcher is responsible for decoding the
results of each response from the paginated API and downloading the objects.

### Functions

```go
func NewCrawler[ScanT, EndpointT any](scanner *Scanner[ScanT], fetcher Fetcher[ScanT, EndpointT]) *Crawler[ScanT, EndpointT]
```
NewCrawler creates a new crawler that scans the API using the provided
Scanner with the result of each scan being passed to the Fetcher to download
each item returned by the scan.



### Methods

```go
func (c *Crawler[ScanT, EndpointT]) Run(ctx context.Context, ch chan<- []content.Object[EndpointT, Response]) error
```
Run runs the crawler. It consists of a scan loop that calls a Fetcher for
each scan response.




### Type Encoding
```go
type Encoding int
```
Encoding represents the encoding scheme used for the response body.

### Constants
### JSONEncoding
```go
JSONEncoding Encoding = iota

```



### Methods

```go
func (e Encoding) ContentType() string
```
ContentType returns the content type associated with this encoding.




### Type Endpoint
```go
type Endpoint[T any] struct {
	// contains filtered or unexported fields
}
```
Endpoint represents an API endpoint that can be invoked using GET.
The response body is unmarshaled into the specified type T. Use PutEndpoint
for operations where both the request and response bodies can be typed.

### Functions

```go
func NewEndpoint[T any](opts ...Option) *Endpoint[T]
```
NewEndpoint returns a new endpoint for the specified type.



### Methods

```go
func (ep *Endpoint[T]) Get(ctx context.Context, url string) (T, []byte, Encoding, error)
```
Get invokes a GET request on this endpoint (without a body).


```go
func (ep *Endpoint[T]) IssueRequest(ctx context.Context, req *http.Request) (T, []byte, Encoding, *http.Response, error)
```
IssueRequest invokes an arbitrary request on this endpoint using the
supplied http.Request. The Body in the http.Response has already been read
and its contents returned as the second return value.




### Type Error
```go
type Error struct {
	Err        error
	Status     string
	StatusCode int
	Attempts   int
}
```

### Methods

```go
func (err *Error) Error() string
```




### Type FS
```go
type FS interface {
	content.FS
	filewalk.FS
}
```
FS defines a filesystem interface to be broadly used by webapi packages and
clients. It is defined in operations for convenience.


### Type Fetcher
```go
type Fetcher[ScanT, EndpointT any] interface {
	Fetch(context.Context, ScanT, chan<- []content.Object[EndpointT, Response]) error
}
```
Fetcher is a generic interface for fetching objects from an API endpoint as
part of a crawl. The Fetcher extracts/decodes items to be fetched from the
result of a Scan and then downloads each item.


### Type Marshal
```go
type Marshal func(any) ([]byte, error)
```
Marshal represents a function that can be used to marshal a request body.


### Type Option
```go
type Option func(o *options)
```
Option represents an option that can be used when creating new Endpoints and
Streams.

### Functions

```go
func WithAuth(a Auth) Option
```
WithAuth specifies the instance of Auth to use when making requests.


```go
func WithHTTPClient(client *http.Client) Option
```
WithHTTPClient specifies the http.Client to use for making requests.
If not specified, http.DefaultClient is used.


```go
func WithLogger(logger *slog.Logger) Option
```
WithLogger specifies the logger to use for logging request and response
information. If not specified, no logging is performed.


```go
func WithMarshal(marshal Marshal, e Encoding) Option
```
WithMarshal specifies a custom marshaling function to use for encoding
request bodies. The default is json.Marshal.


```go
func WithRateController(rc *ratecontrol.Controller, statusCodes ...int) Option
```
WithRateController sets the rate controller to use to enforce rate control
and backoff.


```go
func WithSigner(signer Signer) Option
```
WithSigner specifies a Signer function to use for signing requests.


```go
func WithSuccessCodes(codes ...int) Option
```
WithSuccessCodes specifies the HTTP status codes that should be considered
successful responses. If not specified, only http.StatusOK (200) is
considered a successful response for Get operations and http.StatusOK (200)
http.StatusAccepted or for Put/Post operations.


```go
func WithUnmarshal(u Unmarshal, e Encoding) Option
```
WithUnmarshal specifies a custom unmarshaling function to use for decoding
response bodies. The default is json.Unmarshal.




### Type Paginator
```go
type Paginator[T any] interface {
	// Next is called with the returned type and response for that operation.
	// The first URL to use in a scan is generated by calling Next with an empty
	// payload and nil *http.Response
	Next(ctx context.Context, t T, r *http.Response) (req *http.Request, done bool, err error)
}
```
Paginator represents the ability to generate the next request (URL with
optional body) given the response from the previous request. Paginators are
typically used with Scanners to iterate over a paginated API.


### Type PutEndpoint
```go
type PutEndpoint[RequestT, ResponseT any] struct {
	// contains filtered or unexported fields
}
```
PutEndpoint represents an API endpoint that supports PUT requests with a
request body of type RequestT and a response body of type ResponseT.

### Functions

```go
func NewPutEndpoint[RequestT, ResponseT any](opts ...Option) *PutEndpoint[RequestT, ResponseT]
```



### Methods

```go
func (ep *PutEndpoint[RequestT, ResponseT]) IssueRequest(ctx context.Context, req *http.Request, data RequestT) (ResponseT, []byte, Encoding, *http.Response, error)
```
IssueRequest invokes an arbitrary request on this endpoint using the
supplied http.Request except that the Request body is overridden with
encoding of the supplied data.


```go
func (ep *PutEndpoint[RequestT, ResponseT]) Post(ctx context.Context, url string, data RequestT) (ResponseT, []byte, Encoding, error)
```
Post invokes a POST request on this endpoint with a request of type RequestT
and a response of type ResponseT.


```go
func (ep *PutEndpoint[RequestT, ResponseT]) Put(ctx context.Context, url string, data RequestT) (ResponseT, []byte, Encoding, error)
```
Put invokes a PUT request on this endpoint with a request of type RequestT
and a response of type ResponseT.




### Type Response
```go
type Response struct {
	// The raw bytes of the response.
	Bytes []byte
	// The encoding used for Bytes.
	Encoding Encoding
	// When the response was received.
	When time.Time

	// Fields copied from the http.Response.
	Headers                http.Header
	Trailers               http.Header
	ContentLength          int64
	StatusCode             int
	ProtoMajor, ProtoMinir int
	TransferEncoding       []string

	// Any error encountered during the operation.
	Error error

	// Checkpoint is an opaque value that can be used to resume an
	// operation at a later time. This is used generally used by implementations
	// of Crawler/Fetcher.
	Checkpoint []byte

	// Current and Total, if non-zero, provide an indication of progress.
	Current int64
	Total   int64
}
```
Response contains metadata for the result of an operation and is used for
the response field of the content.Object returned by the Fetcher.

### Methods

```go
func (r *Response) FromHTTPResponse(hr *http.Response)
```




### Type Scanner
```go
type Scanner[T any] struct {
	// contains filtered or unexported fields
}
```
Scanner provides the ability to iterate over a paginated API a page at a
time.

### Functions

```go
func NewScanner[T any](paginator Paginator[T], opts ...Option) *Scanner[T]
```
NewScanner creates a new Scanner using the supplied paginator. The options
are used to create the underlying Endpoint.



### Methods

```go
func (sc *Scanner[T]) Body() []byte
```
Body returns the body for the current page.


```go
func (sc *Scanner[T]) Err() error
```
Err returns the first error encountered during scanning.


```go
func (sc *Scanner[T]) ErrDetail() (error, []byte, *http.Request)
```
ErrDetail returns the first error encountered during scanning along with
the body and request that caused the error. This can be used to provide more
context when debugging errors. The body and request will never be nil, but
may be empty if the error occurred before a request was made or a response
was received.


```go
func (sc *Scanner[T]) HTTPResponse() *http.Response
```
HTTPResponse returns the response for the current page.


```go
func (sc *Scanner[T]) Response() T
```
Response returns the response for the current page.


```go
func (sc *Scanner[T]) Scan(ctx context.Context) bool
```
Scan iterates over the paginated API. It returns true if there is another
page to scan, false otherwise.




### Type Signer
```go
type Signer func(ctx context.Context, header http.Header, payload []byte) error
```
Signer represents a function that can be used to sign requests, e.g.
by adding appropriate headers. This is used for operations that require
signing of requests. Signer is called with the payload to be signed and the
header to which signature information should be added.


### Type Unmarshal
```go
type Unmarshal func([]byte, any) error
```
Unmarshal represents a function that can be used to unmarshal a response
body.





