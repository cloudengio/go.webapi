# Package [cloudeng.io/webapi/protocolsio/protocolsiosdk](https://pkg.go.dev/cloudeng.io/webapi/protocolsio/protocolsiosdk?tab=doc)

```go
import cloudeng.io/webapi/protocolsio/protocolsiosdk
```


## Variables
### ErrTooManyRequests
```go
ErrTooManyRequests = errors.New("too many requests")

```



## Functions
### Func Get
```go
func Get[T any](ctx context.Context, url string) (T, []byte, error)
```

### Func ParsePayload
```go
func ParsePayload[T any](buf []byte) (T, error)
```

### Func WithPublicToken
```go
func WithPublicToken(ctx context.Context, token string) context.Context
```



## Types
### Type Creator
```go
type Creator struct {
	Name       string
	Username   string
	Affilation string
}
```


### Type ListProtocolsV3
```go
type ListProtocolsV3 struct {
	Extras       json.RawMessage
	Items        []json.RawMessage `json:"items"`
	Pagination   Pagination        `json:"pagination"`
	Total        int64             `json:"total"`
	TotalPages   int64             `json:"total_pages"`
	TotalResults int64             `json:"total_results"`
}
```


### Type Pagination
```go
type Pagination struct {
	CurrentPage  int64       `json:"current_page"`
	TotalPages   int64       `json:"total_pages"`
	TotalResults int64       `json:"total_results"`
	NextPage     string      `json:"next_page"`
	PrevPage     interface{} `json:"prev_page"`
	PageSize     int64       `json:"page_size"`
	First        int64       `json:"first"`
	Last         int64       `json:"last"`
	ChangedOn    interface{} `json:"changed_on"`
}
```

### Methods

```go
func (p Pagination) Done() bool
```


```go
func (p Pagination) PageInfo() (next, total int, done bool, err error)
```




### Type Payload
```go
type Payload struct {
	Payload    json.RawMessage `json:"payload"`
	StatusCode int             `json:"status_code"`
}
```


### Type Protocol
```go
type Protocol struct {
	ID          int64  `json:"id"`
	URI         string `json:"uri"`
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	VersionID   int    `json:"version_id"`
	CreatedOn   int    `json:"created_on"`
	Creator     Creator
}
```





