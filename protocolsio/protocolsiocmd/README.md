# Package [cloudeng.io/webapi/protocolsio/protocolsiocmd](https://pkg.go.dev/cloudeng.io/webapi/protocolsio/protocolsiocmd?tab=doc)

```go
import cloudeng.io/webapi/protocolsio/protocolsiocmd
```

Package protocolsiocmd provides support for building command line tools that
access protocols.io.

Package protocolsio provides support for working with the protocols.io API.
It currently provides the ability to crawl public protocols.

## Types
### Type Auth
```go
type Auth struct {
	PublicToken  string `yaml:"public_token"`
	ClientID     string `yaml:"public_clientid"`
	ClientSecret string `yaml:"public_secret"`
}
```
Auth represents the authentication information required to access
protocols.io.


### Type Command
```go
type Command struct {
	Auth
	Config
}
```
Çommand implements the command line operations available for protocols.io.

### Functions

```go
func NewCommand(ctx context.Context, crawls apicrawlcmd.Crawls, name, authFilename string) (*Command, error)
```
NewCommand returns a new Command instance for the specified API crawl with
API authentication information read from the specified file or from the
context.



### Methods

```go
func (c *Command) Crawl(ctx context.Context, cacheRoot string, fv *CrawlFlags) error
```


```go
func (c *Command) Get(ctx context.Context, fv *GetFlags, args []string) error
```


```go
func (c *Command) ScanDownloaded(ctx context.Context, root string, fv *ScanFlags) error
```




### Type CommonFlags
```go
type CommonFlags struct {
	ProtocolsConfig string `subcmd:"protocolsio-config,$HOME/.protocolsio.yaml,'protocols.io auth config file'"`
}
```


### Type Config
```go
type Config apicrawlcmd.Crawl[Service]
```
Config represents the configuration information required to access and crawl
the protocols.io API.

### Methods

```go
func (c Config) NewProtocolCrawler(ctx context.Context, op checkpoint.Operation, fv *CrawlFlags, auth Auth) (*operations.Crawler[protocolsiosdk.ListProtocolsV3, protocolsiosdk.ProtocolPayload], error)
```
NewProtocolCrawler creates a new instance of operations.Crawler that can be
used to crawl/download protocols on protocols.io.


```go
func (c Config) OptionsForEndpoint(auth Auth) ([]operations.Option, error)
```




### Type CrawlFlags
```go
type CrawlFlags struct {
	CommonFlags
	Save             bool               `subcmd:"save,true,'save downloaded protocols to disk'"`
	IgnoreCheckpoint bool               `subcmd:"ignore-checkpoint,false,'ignore the checkpoint files'"`
	Pages            flags.IntRangeSpec `subcmd:"pages,,page range to return"`
	PageSize         int                `subcmd:"size,50,number of items in each page"`
	Key              string             `subcmd:"key,,'string may contain any characters, numbers and special symbols. System will search around protocol name, description, authors. If the search keywords are enclosed in double quotes, then result contains only the exact match of the combined term'"`
}
```


### Type GetFlags
```go
type GetFlags struct {
	CommonFlags
}
```


### Type ScanFlags
```go
type ScanFlags struct {
	CommonFlags
	Template string `subcmd:"template,'{{.ID}}',template to use for printing fields in the downloaded Protocol objects"`
}
```


### Type Service
```go
type Service struct {
	Filter         string `yaml:"filter"`
	OrderField     string `yaml:"order_field"`
	OrderDirection string `yaml:"order_direction"`
	Incremental    bool   `yaml:"incremental"`
}
```
Service represents the protocols.io specific confiugaration options.





