# Package [cloudeng.io/webapi/clients/benchling/benchlingcmd](https://pkg.go.dev/cloudeng.io/webapi/clients/benchling/benchlingcmd?tab=doc)

```go
import cloudeng.io/webapi/clients/benchling/benchlingcmd
```

Package benchlingcmd provides support for building command line tools that
access benchling.com

## Functions
### Func OptionsForEndpoint
```go
func OptionsForEndpoint(cfg apicrawlcmd.Crawl[Service]) ([]operations.Option, error)
```



## Types
### Type Checkpoint
```go
type Checkpoint struct {
	// Dates in rfc.3339 format.
	UsersDate   string `json:"users_date"`
	EntriesDate string `json:"entries_date"`
}
```


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
NewCommand returns a new Command instance for benchling API related
commands.



### Methods

```go
func (c *Command) Crawl(ctx context.Context, _ CrawlFlags, entities ...string) error
```


```go
func (c *Command) CreateIndexableDocuments(ctx context.Context, _ IndexFlags) error
```
CreateIndexableDocuments constructs the documents to be indexed from the
various objects crawled from the benchling.com API.




### Type Config
```go
type Config apicrawlcmd.Crawl[Service]
```


### Type CrawlFlags
```go
type CrawlFlags struct{}
```


### Type GetFlags
```go
type GetFlags struct{}
```


### Type IndexFlags
```go
type IndexFlags struct{}
```


### Type Service
```go
type Service struct {
	ServiceURL       string `yaml:"service_url" cmd:"benchling service URL, typically https://altoslabs.benchling.com/api/v2/"`
	UsersPageSize    int    `yaml:"users_page_size" cmd:"number of users in each page of results, typically 50"`
	EntriesPageSize  int    `yaml:"entries_page_size" cmd:"number of entries in each page of results, typically 50"`
	FoldersPageSize  int    `yaml:"folders_page_size" cmd:"number of folders in each page of results, typically 50"`
	ProjectsPageSize int    `yaml:"projects_page_size" cmd:"number of projects in each page of results, typically 50"`
}
```

### Methods

```go
func (s Service) ListEntriesConfig() *benchlingsdk.ListEntriesParams
```


```go
func (s Service) ListFoldersConfig() *benchlingsdk.ListFoldersParams
```


```go
func (s Service) ListProjectsConfig() *benchlingsdk.ListProjectsParams
```


```go
func (s Service) ListUsersConfig() *benchlingsdk.ListUsersParams
```







