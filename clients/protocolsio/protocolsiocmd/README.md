# Package [cloudeng.io/webapi/clients/protocolsio/protocolsiocmd](https://pkg.go.dev/cloudeng.io/webapi/clients/protocolsio/protocolsiocmd?tab=doc)

```go
import cloudeng.io/webapi/clients/protocolsio/protocolsiocmd
```

Package protocolsio provides support for working with the protocols.io API.
It currently provides the ability to crawl public protocols.

Package protocolsiocmd provides support for building command line tools that
access protocols.io.

## Functions
### Func OptionsForEndpoint
```go
func OptionsForEndpoint(cfg apicrawlcmd.Crawl[Service]) ([]operations.Option, error)
```



## Types
### Type Command
```go
type Command struct {
	// contains filtered or unexported fields
}
```
Çommand implements the command line operations available for protocols.io.

### Functions

```go
func NewCommand(ctx context.Context, config apicrawlcmd.Crawl[yaml.Node], resources apicrawlcmd.Resources) (*Command, error)
```
NewCommand returns a new Command instance for protocols.io API related
commands.



### Methods

```go
func (c *Command) Crawl(ctx context.Context, fv *CrawlFlags) error
```


```go
func (c *Command) Get(ctx context.Context, _ *GetFlags, args []string) error
```


```go
func (c *Command) ScanDownloaded(ctx context.Context, fv *ScanFlags) error
```




### Type CrawlFlags
```go
type CrawlFlags struct {
	Save             bool               `subcmd:"save,true,'save downloaded protocols to disk'"`
	IgnoreCheckpoint bool               `subcmd:"ignore-checkpoint,false,'ignore the checkpoint files'"`
	Pages            flags.IntRangeSpec `subcmd:"pages,,page range to return"`
	PageSize         int                `subcmd:"size,50,number of items in each page"`
	Key              string             `subcmd:"key,,'string may contain any characters, numbers and special symbols. System will search around protocol name, description, authors. If the search keywords are enclosed in double quotes, then result contains only the exact match of the combined term'"`
}
```


### Type GetFlags
```go
type GetFlags struct{}
```


### Type ScanFlags
```go
type ScanFlags struct {
	Template string `subcmd:"template,'{{.ID}}',template to use for printing fields in the downloaded Protocol objects"`
}
```


### Type Service
```go
type Service struct {
	Filter         string `yaml:"filter" cmd:"filter to apply to protocols.io API calls, typically public"`
	OrderField     string `yaml:"order_field" cmd:"field used to order API responses, typically id"`
	OrderDirection string `yaml:"order_direction" cmd:"order direction to apply to protocols.io API calls, typically asc"`
	Incremental    bool   `yaml:"incremental" cmd:"if true, only download new or updated protocols"`
}
```
Service represents the protocols.io specific confiugaration options.





