# Package [cloudeng.io/webapi/clients/biorxiv](https://pkg.go.dev/cloudeng.io/webapi/clients/biorxiv?tab=doc)

```go
import cloudeng.io/webapi/clients/biorxiv
```

Package biorxiv provides support for crawling and indexing biorxiv.org and
medrxiv.org via api.biorxiv.org.

## Constants
### PreprintType
```go
// PreprintType is the content type for Preprints.
PreprintType = content.Type("api.biorxiv.org/article")

```



## Functions
### Func NewScanner
```go
func NewScanner(serviceURL string, from, to time.Time, cursor int64, opts ...operations.Option) *operations.Scanner[Response]
```
NewScanner returns an instance of operations.Scanner for scanning biorxiv or
medrxiv via api.biorxiv.org. The from, to and cursor values corresponding to
URL path components as documented at: https://api.biorxiv.org/



## Types
### Type Message
```go
type Message struct {
	Status   string `json:"status"`
	Interval string `json:"interval"`
	Cursor   any    `json:"cursor"` // Can be an int or a string!
	Count    int64  `json:"count"`
	Total    any    `json:"total"` // Can be an int or a string!
}
```
Message represents the detailed response from api.biorxiv.org for an API
request. The status will 'ok' for successful requests. Each request will
return a Collection containing at most 100 Preprints. Pagination is achieved
using the Cursor value.


### Type PreprintDetail
```go
type PreprintDetail struct {
	PreprintDOI                            string `json:"preprint_doi"`
	PublishedDOI                           string `json:"published_doi"`
	PublishedJournal                       string `json:"published_journal"`
	PreprintPlatform                       string `json:"preprint_platform"`
	PreprintTitle                          string `json:"preprint_title"`
	PreprintAuthors                        string `json:"preprint_authors"`
	PreprintCategory                       string `json:"preprint_category"`
	PreprintDate                           string `json:"preprint_date"`
	PublishedDate                          string `json:"published_date"`
	PreprintAbstract                       string `json:"preprint_abstract"`
	PreprintAuthorCorresponding            string `json:"preprint_author_corresponding"`
	PreprintAuthorCoresspondingInstitution string `json:"preprint_author_corresponding_institution"`
}
```
PreprintDetail represents the details of a single preprint.


### Type Response
```go
type Response struct {
	Messages   []Message        `json:"messages"`
	Collection []PreprintDetail `json:"collection"`
}
```
Response represents the response from api.biorxiv.org for an API request and
is wrapper for the actual Message response and Collection.





