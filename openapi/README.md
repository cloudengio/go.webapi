# Package [cloudeng.io/webapi/openapi](https://pkg.go.dev/cloudeng.io/webapi/openapi?tab=doc)

```go
import cloudeng.io/webapi/openapi
```


## Functions
### Func AsYAML
```go
func AsYAML(indent int, doc any) (string, error)
```

### Func FormatV3
```go
func FormatV3(doc *openapi3.T, isYAML bool) ([]byte, error)
```



## Types
### Type Visitor
```go
type Visitor func(path []string, parent, node any) (ok bool, err error)
```
Visitor is called for every node in the walk. It returns true for the walk
to continue, false otherwise. The walk will stop when an error is returned.


### Type Walker
```go
type Walker interface {
	Walk(doc *openapi3.T) error
}
```
Walker represents the interface implemented by all walkers.

### Functions

```go
func NewWalker(v Visitor, opts ...WalkerOption) Walker
```
NewWalker returns a Walker that will visit every node in an openapi3
document.




### Type WalkerOption
```go
type WalkerOption func(o *walkerOptions)
```
WalkerOption represents an option for use when creating a new walker.

### Functions

```go
func WalkerFollowRefs(v bool) WalkerOption
```
WalkerFollowRefs controls wether the walker will follow $ref's and flatten
them in place.


```go
func WalkerTracePaths(v bool) WalkerOption
```


```go
func WalkerVisitPrefix(path ...string) WalkerOption
```
WalkerVisitPrefix adds a prefix that the walk should call the Visitor
function for. All other paths will be ignored.







