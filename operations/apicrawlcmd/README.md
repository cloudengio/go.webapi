# Package [cloudeng.io/webapi/operations/apicrawlcmd](https://pkg.go.dev/cloudeng.io/webapi/operations/apicrawlcmd?tab=doc)

```go
import cloudeng.io/webapi/operations/apicrawlcmd
```

Package apicrawlcmd provides support for building command line tools that
implement API crawls.

## Functions
### Func CachePaths
```go
func CachePaths(crawls Crawls) []string
```
CachePaths returns the paths of all cache directories.

### Func CheckpointPaths
```go
func CheckpointPaths(crawls Crawls) []string
```
CheckpointPaths returns the paths of all checkpoint directories.

### Func ParseCrawlConfig
```go
func ParseCrawlConfig[T any](crawls Crawls, name string, crawlConfig *Crawl[T]) (bool, error)
```
ParseCrawlConfig parses an API specific crawl config of the specified name.



## Types
### Type Crawl
```go
type Crawl[T any] struct {
	RateControl crawlcmd.RateControl      `yaml:",inline"`
	Cache       crawlcmd.CrawlCacheConfig `yaml:",inline"`
	Service     T                         `yaml:",inline"`
}
```
Crawl is a generic type that defines common crawl configuration options as
well as allowing for service specific ones.


### Type Crawls
```go
type Crawls map[string]struct {
	RateControl crawlcmd.RateControl      `yaml:",inline"`
	Cache       crawlcmd.CrawlCacheConfig `yaml:",inline"`
	Service     yaml.Node                 `yaml:"service"`
}
```
Crawls represents the configuration of multiple API crawls.





