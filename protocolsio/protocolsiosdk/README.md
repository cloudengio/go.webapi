# Package [cloudeng.io/webapi/protocolsio/protocolsiosdk](https://pkg.go.dev/cloudeng.io/webapi/protocolsio/protocolsiosdk?tab=doc)

```go
import cloudeng.io/webapi/protocolsio/protocolsiosdk
```

Package protocolsiosdk provides a minimal SDK for the protocols.io API.
See https://apidoc.protocols.io for details.

## Constants
### ListProtocolsV3Endpoint, GetProtocolV4Endpoint
```go
ListProtocolsV3Endpoint = "https://www.protocols.io/api/v3/protocols"
GetProtocolV4Endpoint = "https://www.protocols.io/api/v4/protocols"

```



## Functions
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


### Type ProtocolPayload
```go
type ProtocolPayload struct {
	Protocol   Protocol `json:"payload"`
	StatusCode int      `json:"status_code"`
}
```





