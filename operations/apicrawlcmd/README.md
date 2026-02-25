# Package [cloudeng.io/webapi/operations/apicrawlcmd](https://pkg.go.dev/cloudeng.io/webapi/operations/apicrawlcmd?tab=doc)

```go
import cloudeng.io/webapi/operations/apicrawlcmd
```

Package apicrawlcmd provides support for building command line tools that
implement API crawls.

## Functions
### Func ParseCrawlConfig
```go
func ParseCrawlConfig[T any](cfg Crawl[yaml.Node], service *Crawl[T]) error
```
ParseCrawlConfig parses an API specific crawl config, it's parametized by
the types of the service specific and crawl cache specific data types.



## Types
### Type Crawl
```go
type Crawl[T any] struct {
	RateControl crawlcmd.RateControl      `yaml:",inline"`
	Cache       crawlcmd.CrawlCacheConfig `yaml:"cache"`
	KeyID       string                    `yaml:"key_id" cmd:"identifier of the API key to use for this crawl"`
	Service     T                         `yaml:"service_config" cmd:"service specific configuration"`
}
```
Crawl is a generic type that defines common crawl configuration options as
well as allowing for service specific ones. The type of the service specific
configuration is generally determined by the API being crawled.


### Type Crawls
```go
type Crawls map[string]Crawl[yaml.Node]
```
Crawls represents the configuration of multiple API crawls.


### Type Resources
```go
type Resources struct {
	NewOperationsFS func(ctx context.Context, cfg crawlcmd.CrawlCacheConfig) (operations.FS, error)

	NewCheckpointOp func(ctx context.Context, cfg crawlcmd.CrawlCacheConfig) (checkpoint.Operation, error)
}
```
Resources represents the resources typically required to perform an API
crawl.

### Methods

```go
func (r Resources) CreateResources(ctx context.Context, cfg crawlcmd.CrawlCacheConfig) (store operations.FS, chkpt checkpoint.Operation, err error)
```




### Type State
```go
type State[T any] struct {
	Config     Crawl[T]
	Store      operations.FS
	Checkpoint checkpoint.Operation
}
```

### Functions

```go
func NewState[T any](ctx context.Context, config Crawl[yaml.Node], resources Resources) (State[T], error)
```







