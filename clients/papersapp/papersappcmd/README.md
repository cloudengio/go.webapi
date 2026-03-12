# Package [cloudeng.io/webapi/clients/papersapp/papersappcmd](https://pkg.go.dev/cloudeng.io/webapi/clients/papersapp/papersappcmd?tab=doc)

```go
import cloudeng.io/webapi/clients/papersapp/papersappcmd
```


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
Çommand implements the command line operations available for papersapp.com.

### Functions

```go
func NewCommand(ctx context.Context, config apicrawlcmd.Crawl[yaml.Node], resources apicrawlcmd.Resources) (*Command, error)
```
NewCommand returns a new Command instance for readcube/papersapp API related
commands.



### Methods

```go
func (c *Command) Crawl(ctx context.Context, _ *CrawlFlags) error
```


```go
func (c *Command) ScanDownloaded(ctx context.Context, _ *ScanFlags) error
```




### Type CrawlFlags
```go
type CrawlFlags struct{}
```


### Type ScanFlags
```go
type ScanFlags struct{}
```


### Type Service
```go
type Service struct {
	ServiceURL        string `yaml:"service_url" cmd:"papersapp service URL, typically https://api.papers.ai"`
	RefreshTokenURL   string `yaml:"refresh_token_url" cmd:"papersapp refresh token URL, typically https://api.papers.ai/oauth/token"`
	ListItemsPageSize int    `yaml:"list_items_page_size" cmd:"number of items in each page of results, typically 50"`
}
```





