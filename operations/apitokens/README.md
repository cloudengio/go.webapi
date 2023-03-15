# Package [cloudeng.io/webapi/operations/apitokens](https://pkg.go.dev/cloudeng.io/webapi/operations/apitokens?tab=doc)

```go
import cloudeng.io/webapi/operations/apitokens
```


## Functions
### Func ContextWithTokens
```go
func ContextWithTokens(ctx context.Context, name string, tokens []byte) context.Context
```
ContextWithTokens returns a new context that contains the provided named
tokens in addition to any existing tokens. The tokens are typically encoded
as JSON or YAML.

### Func ParseTokensYAML
```go
func ParseTokensYAML(ctx context.Context, name string, cfg any) (bool, error)
```
ParseTokensYAML parses the tokens stored in the context for the specified
name as JSON. It will return false if there are no tockens stored, true
otherwise and an error if the unmsrshal fails.

### Func TokensFromContext
```go
func TokensFromContext(ctx context.Context, name string) ([]byte, bool)
```
TokensFromContext returns the tokens for the specified name, if any,
that are stored in the context.




