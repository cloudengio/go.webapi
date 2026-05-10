# Package [cloudeng.io/webapi/clients/papersapp](https://pkg.go.dev/cloudeng.io/webapi/clients/papersapp?tab=doc)

```go
import cloudeng.io/webapi/clients/papersapp
```


## Constants
### CollectionType, ItemType
```go
CollectionType = "papersapp.com/collection"
ItemType = "papersapp.com/item"

```



## Functions
### Func ListCollections
```go
func ListCollections(ctx context.Context, serviceURL string, opts ...operations.Option) ([]*papersappsdk.Collection, error)
```

### Func NewItemPaginator
```go
func NewItemPaginator(opts ItemPaginatorOptions) operations.Paginator[papersappsdk.Items]
```



## Types
### Type APIToken
```go
type APIToken struct {
	// contains filtered or unexported fields
}
```
APIToken is an implementation of operations.Authorizer for a papersapp API
token.

### Functions

```go
func NewAPIToken(refreshUserID, refreshKeyID, refreshURL string) *APIToken
```



### Methods

```go
func (pbt *APIToken) Refresh(ctx context.Context) (papersappsdk.Token, error)
```


```go
func (pbt *APIToken) WithAuthorization(ctx context.Context, req *http.Request) error
```




### Type Item
```go
type Item struct {
	Item       *papersappsdk.Item
	Collection *papersappsdk.Collection
}
```
Item represents a single item in a collection and also includes the
collection to which it belongs.


### Type ItemPaginatorOptions
```go
type ItemPaginatorOptions struct {
	EndpointURL string
	Parameters  url.Values
}
```





