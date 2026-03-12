# Package [cloudeng.io/webapi/clients/biorxiv/biorxivcmd](https://pkg.go.dev/cloudeng.io/webapi/clients/biorxiv/biorxivcmd?tab=doc)

```go
import cloudeng.io/webapi/clients/biorxiv/biorxivcmd
```

Package biorxivcmd provides support for building command line tools that
access api.biorxiv.com

## Functions
### Func OptionsForEndpoint
```go
func OptionsForEndpoint(cfg apicrawlcmd.Crawl[Service]) ([]operations.Option, error)
```
OptionsForEndpoint returns the operations.Option's derived from the
apicrawlcmd configuration.



## Types
### Type Command
```go
type Command struct {
	// contains filtered or unexported fields
}
```
Çommand implements the command line operations available for
api.biorxiv.org.

### Functions

```go
func NewCommand(ctx context.Context, config apicrawlcmd.Crawl[yaml.Node], resources apicrawlcmd.Resources) (*Command, error)
```
NewCommand returns a new Command instance for biorxiv API related commands.



### Methods

```go
func (c *Command) Crawl(ctx context.Context, flags CrawlFlags) error
```
Crawl implements the crawl command. The crawl is incremental and utilizes
an internal state file to track progress and restart from that point in a
subsequent crawl. This makes it possible to have a start date that predates
the creation of biorxiv and an end date of 'now' with each incremental crawl
picking up where the previous one left off assuming that biorxiv doesn't add
new preprints with dates that predate the current one.


```go
func (c *Command) LookupDownloaded(ctx context.Context, fv *LookupFlags, dois ...string) error
```
LookupDownloaded looks up the specified preprints via their 'PreprintDOI'
printing out fields using the specified template.


```go
func (c *Command) ScanDownloaded(ctx context.Context, fv *ScanFlags) error
```
ScanDownloaded scans downloaded preprints printing out fields using the
specified template.




### Type CrawlFlags
```go
type CrawlFlags struct {
	Restart bool `subcmd:"restart,false,'restart the crawl, ignoring the saved checkpoint'"`
}
```


### Type GetFlags
```go
type GetFlags struct{}
```


### Type IndexFlags
```go
type IndexFlags struct{}
```


### Type LookupFlags
```go
type LookupFlags struct {
	Template string `subcmd:"template,'{{.}}',template to use for printing fields in the downloaded Preprint objects"`
}
```


### Type ScanFlags
```go
type ScanFlags struct {
	Template string `subcmd:"template,'{{.PreprintDOI}} {{.PreprintTitle}}',template to use for printing fields in the downloaded Preprint objects"`
}
```


### Type Service
```go
type Service struct {
	ServiceURL string           `yaml:"service_url" cmd:"rxiv service URL, eg. https://api.biorxiv.org/pubs/biorxiv for biorxiv"`
	StartDate  cmdyaml.FlexTime `yaml:"start_date" cmd:"start date for crawl, eg. 2020-01-01"`
	EndDate    cmdyaml.FlexTime `yaml:"end_date" cmd:"end date for crawl, eg. 2020-12-01"`
}
```
Service represents biorxiv specific configuration parameters.





