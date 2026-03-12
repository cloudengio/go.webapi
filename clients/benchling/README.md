# Package [cloudeng.io/webapi/clients/benchling](https://pkg.go.dev/cloudeng.io/webapi/clients/benchling?tab=doc)

```go
import cloudeng.io/webapi/clients/benchling
```

Package benchling provides support for crawling and indexing benchling.com.

## Constants
### DocumentType, EntryType, ProjectType, FolderType, UserType
```go
DocumentType = content.Type("benchling.com/document")
EntryType = content.Type("benchling.com/entry")
ProjectType = content.Type("benchling.com/project")
FolderType = content.Type("benchling.com/folder")
UserType = content.Type("benchling.com/user")

```



## Functions
### Func ContentType
```go
func ContentType[ObjectT Objects](obj ObjectT) content.Type
```

### Func NewScanner
```go
func NewScanner[ScannerT Scanners, ParamsT Params](ctx context.Context, serviceURL string, params ParamsT, opts ...operations.Option) *operations.Scanner[ScannerT]
```

### Func ObjectID
```go
func ObjectID[ObjectT Objects](obj ObjectT) string
```



## Types
### Type APIToken
```go
type APIToken struct {
	TokenID string
}
```
APIToken is an implementation of operations.Authorizer for a benchling API
token.

### Methods

```go
func (pbt APIToken) WithAuthorization(ctx context.Context, req *http.Request) error
```
WithAuthorization implements operations.Authorizer.




### Type Backoff
```go
type Backoff struct {
	// contains filtered or unexported fields
}
```
Backoff implements a backoff strategy that first looks for the specific
backoff period specified in the x-rate-limit-reset header in benchling.com's
http response when a rate limit is reached. If no such header is found then
exponential backoff is used.

### Functions

```go
func NewBackoff(initial time.Duration, steps int) *Backoff
```



### Methods

```go
func (bb *Backoff) Retries() int
```


```go
func (bb *Backoff) Wait(ctx context.Context, r any) (bool, error)
```
Wait implements Backoff.




### Type Document
```go
type Document struct {
	Entry    benchlingsdk.Entry   // An actual data entry.
	Folder   benchlingsdk.Folder  // The folder containing the entry.
	Project  benchlingsdk.Project // The project containing the folder.
	DayNotes string
	Parents  []string                     // The parent folders of the folder containing the entry.
	Users    map[string]benchlingsdk.User // All users referenced in the entry, keyed by their userid.
}
```
Document represents the structure of information within benchling in terms
of an a single indexable document.


### Type DocumentIndexer
```go
type DocumentIndexer struct {
	// contains filtered or unexported fields
}
```

### Functions

```go
func NewDocumentIndexer(fs operations.FS, downloads string, sharder path.Sharder, concurrency int) *DocumentIndexer
```



### Methods

```go
func (di *DocumentIndexer) Index(ctx context.Context) error
```




### Type Entries
```go
type Entries struct {
	NextToken *string
	Entries   []benchlingsdk.Entry
}
```


### Type Folders
```go
type Folders struct {
	NextToken *string
	Folders   []benchlingsdk.Folder
}
```


### Type Objects
```go
type Objects interface {
	benchlingsdk.Entry | benchlingsdk.User | benchlingsdk.Folder | benchlingsdk.Project | Document
}
```


### Type Params
```go
type Params interface {
	*benchlingsdk.ListEntriesParams | *benchlingsdk.ListUsersParams | *benchlingsdk.ListFoldersParams | *benchlingsdk.ListProjectsParams
}
```


### Type Projects
```go
type Projects struct {
	NextToken *string
	Projects  []benchlingsdk.Project
}
```


### Type ScanPayload
```go
type ScanPayload[T any] struct {
	NextToken *string
	Payload   []T
}
```


### Type Scanners
```go
type Scanners interface {
	Entries | Users | Folders | Projects
}
```


### Type Users
```go
type Users struct {
	NextToken *string
	Users     []benchlingsdk.User
}
```





