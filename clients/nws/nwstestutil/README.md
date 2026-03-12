# Package [cloudeng.io/webapi/clients/nws/nwstestutil](https://pkg.go.dev/cloudeng.io/webapi/clients/nws/nwstestutil?tab=doc)

```go
import cloudeng.io/webapi/clients/nws/nwstestutil
```


## Types
### Type NWSMockServer
```go
type NWSMockServer struct {
	// contains filtered or unexported fields
}
```

### Functions

```go
func NewMockServer() *NWSMockServer
```



### Methods

```go
func (ms *NWSMockServer) Close()
```


```go
func (ms *NWSMockServer) ForecastCalls() int
```


```go
func (ms *NWSMockServer) LookupCalls() int
```


```go
func (ms *NWSMockServer) ResetForecastCalls()
```


```go
func (ms *NWSMockServer) ResetLookupCalls()
```


```go
func (ms *NWSMockServer) Run() string
```


```go
func (ms *NWSMockServer) SetValidTimes(when time.Time)
```







