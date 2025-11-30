# Package [cloudeng.io/webapi/operations/apitokens](https://pkg.go.dev/cloudeng.io/webapi/operations/apitokens?tab=doc)

```go
import cloudeng.io/webapi/operations/apitokens
```

Package apitokens provides types and functions for managing API tokens and
is built on top of the cmdutil/keys package and its InmemoryKeyStore.

## Functions
### Func ClearToken
```go
func ClearToken(token []byte)
```
ClearToken overwrites the contents of the provided token byte slice with
zeros.

### Func ContextWithKey
```go
func ContextWithKey(ctx context.Context, ki keys.Info) context.Context
```
ContextWithKey returns a new context that contains the provided named
key.Info in addition to any existing keys. It wraps keys.ContextWithKey.

### Func ContextWithOAuth
```go
func ContextWithOAuth(ctx context.Context, id, user string, source oauth2.TokenSource) context.Context
```
ContextWithOauth returns a new context that contains the provided named
oauth2.TokenSource in addition to any existing TokenSources.

### Func KeyFromContext
```go
func KeyFromContext(ctx context.Context, id string) (keys.Info, bool)
```
KeyFromContext retrieves the key.Info for the specified id from the context.
It wraps keys.KeyInfoFromContextForID.

### Func NewErrNotFound
```go
func NewErrNotFound(keyID, service string) error
```
NewErrNotFound creates a new Error indicating that the specified token was
not found. Errors.Is(err, fs.ErrNotExist) can be used to check for this
condition.

### Func OAuthFromContext
```go
func OAuthFromContext(ctx context.Context, id string) oauth2.TokenSource
```
OAuthFromContext returns the TokenSource for the specified name, if any,
that are stored in the context.

### Func TokenFromContext
```go
func TokenFromContext(ctx context.Context, id string) (*keys.Token, bool)
```
TokenFromContext retrieves the token value for the specified id from the
context. It returns the token value as a string and a boolean indicating
whether the token was found.



## Types
### Type Error
```go
type Error struct {
	KeyID   string
	Service string
	Err     error
}
```
Error represents an error related to API tokens.

### Methods

```go
func (e Error) Error() string
```
Error implements the error interface.


```go
func (e Error) Unwrap() error
```







