# Package [cloudeng.io/webapi/openapi/transforms](https://pkg.go.dev/cloudeng.io/webapi/openapi/transforms?tab=doc)

```go
import cloudeng.io/webapi/openapi/transforms
```


## Functions
### Func List
```go
func List() []string
```
List returns a list of all available transformers.

### Func Register
```go
func Register(t T)
```
Register registers a transformer and make it available to clients of this
package.



## Types
### Type Config
```go
type Config struct {
	Configs    []yaml.Node `yaml:"configs"`
	Transforms []string
}
```
Config represents the loaded transformer configuration.

### Functions

```go
func LoadConfigFile(filename string) (Config, error)
```
LoadConfigFile loads the transform configuration from the specified YAML
file.


```go
func ParseConfig(data []byte) (Config, error)
```
ParseConfig parses the supplied YAML data to create an instance of Config.



### Methods

```go
func (c Config) ConfigureAll() error
```
ConfigureAll configures all of the transformers currently registered.




### Type Replacement
```go
type Replacement struct {
	// contains filtered or unexported fields
}
```
Replacement represents a replacement string of the form
/<match-re>/<replacement>/

### Functions

```go
func NewReplacement(s string) (Replacement, error)
```
NewReplacement accepts a string of the form /<match-re>/<replacement>/ to
create a Replacement that will apply <match-re.ReplaceAllString(<replace>).



### Methods

```go
func (sr Replacement) MatchString(input string) bool
```
Match applies regexp.MatchString.


```go
func (sr Replacement) ReplaceAllString(input string) string
```
ReplaceAllString(input string) applies regexp.ReplaceAllString.




### Type T
```go
type T interface {
	Name() string
	Describe(node yaml.Node) string
	Configure(node yaml.Node) error
	Transform(*openapi3.T) (*openapi3.T, error)
}
```
T represents a 'Transformer' that can be used to perform structured
edits/transforms on an openapi 3 specification.

### Functions

```go
func Get(name string) T
```
Get returns the transformer, if any, for the requested name. It returns nil
if no transformer with that name has been registered.







