# Package [cloudeng.io/webapi/protocolsio](https://pkg.go.dev/cloudeng.io/webapi/protocolsio?tab=doc)

```go
import cloudeng.io/webapi/protocolsio
```


## Constants
### ContentType
```go
ContentType = "protocols.io/protocol"

```



## Functions
### Func NewFetcher
```go
func NewFetcher(fetcherOpts FetcherOptions, opts ...operations.Option) (operations.Fetcher[protocolsiosdk.ListProtocolsV3, protocolsiosdk.ProtocolPayload], error)
```
NewFetcher returns an instance of operations.Fetcher for protocols.io
'GetList' operation.

### Func NewPaginator
```go
func NewPaginator(ctx context.Context, cp Checkpoint, opts PaginatorOptions) (operations.Paginator[protocolsiosdk.ListProtocolsV3], error)
```
NewPaginator returns an instance of operations.Paginator for protocols.io
'GetList' operation.

### Func NewProtocolCrawler
```go
func NewProtocolCrawler(
	ctx context.Context, checkpoint Checkpoint,
	fopts FetcherOptions, popts PaginatorOptions, opts ...operations.Option) (*operations.Crawler[protocolsiosdk.ListProtocolsV3, protocolsiosdk.ProtocolPayload], error)
```
NewProtocolCrawler creates a new instance of operations.Crawler that can be
used to crawl/download protocols on protocols.io.



## Types
### Type Checkpoint
```go
type Checkpoint struct {
	CompletedPage int64 `json:"completed_page"`
	CurrentPage   int64 `json:"current_page"`
	TotalPages    int64 `json:"total_pages"`
}
```
Checkpoint represents the most recently completed page of the paginated
crawl.


### Type FetcherOptions
```go
type FetcherOptions struct {
	EndpointURL string
	VersionMap  map[int64]int
}
```
FetcherOptions represent the options for creating a new Fetcher, including
a VersionMap which allows for incremental downloading of Protocol objects.
The VersionMap contains the version ID of a previously downloaded instance
of that protocol, keyed by it's ID. The fetcher will only redownload a
protocol object if its version ID has channged. The VersionMap is typically
built by scanning all previously downloaded protocol objects.


### Type PaginatorOptions
```go
type PaginatorOptions struct {
	EndpointURL string
	Parameters  url.Values
	From, To    int
}
```


### Type PublicBearerToken
```go
type PublicBearerToken struct {
	Token string
}
```
PublicBearerToken is an implementation of operations.Authorizer for a
protocols.io public bearer token.

### Methods

```go
func (pbt PublicBearerToken) WithAuthorization(ctx context.Context, req *http.Request) error
```







