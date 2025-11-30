# Package [cloudeng.io/webapi/operations/apitokens](https://pkg.go.dev/cloudeng.io/webapi/operations/apitokens?tab=doc)

```go
import cloudeng.io/webapi/operations/apitokens
```

Package apitokens provides types and functions for managing API tokens and
is built on top of the cmdutil/keys package and its InmemoryKeyStore.

## Functions
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

### Func OAuthFromContext
```go
func OAuthFromContext(ctx context.Context, id string) oauth2.TokenSource
```
OAuthFromContext returns the TokenSource for the specified name, if any,
that are stored in the context.




