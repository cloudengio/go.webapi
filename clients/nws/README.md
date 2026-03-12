# Package [cloudeng.io/webapi/clients/nws](https://pkg.go.dev/cloudeng.io/webapi/clients/nws?tab=doc)

```go
import cloudeng.io/webapi/clients/nws
```

Package nws provides a client for the National Weather Service API.

Package nws provides a client for the National Weather Service API.

## Constants
### APIHost
```go
APIHost = "https://api.weather.gov"

```



## Types
### Type API
```go
type API struct {
	// contains filtered or unexported fields
}
```
API represents a client for the National Weather Service API.

### Functions

```go
func NewAPI(opts ...Option) *API
```
NewAPI creates a new instance of the National Weather Service API client.



### Methods

```go
func (a *API) GetForecasts(ctx context.Context, gp GridPoints, opts ...operations.Option) (Forecast, error)
```
GetForecasts returns the forecasts for the specified grid point. A cache
keyed by grid point is used to avoid repeated lookups. The expiration of
cache entries defaults to the ValidTimes entry in the forecast and can be
overridden with the WithForecastExpiration option.


```go
func (a *API) LookupGridPoints(ctx context.Context, lat, long float64, opts ...operations.Option) (GridPoints, error)
```
LookupGridPoints returns the grid points for the specified lat/long.
A cache keyed by exact lat/long is used to avoid repeated lookups. The
expiration of cache entries defaults to 1 week and can be overridden with
the WSithGridpointExpiration option.


```go
func (a *API) SetHost(host string)
```




### Type Forecast
```go
type Forecast struct {
	Generated time.Time
	Updated   time.Time
	ValidFrom time.Time
	ValidFor  time.Duration
	Periods   []Period
}
```
Forecast represents the forecasts for a specific grid point for a number of
periods.

### Methods

```go
func (fc Forecast) PeriodFor(when time.Time) (Period, bool)
```
PeriodFor returns the forecast for the specified time. The returned bool is
false if there is no forecast for the specified time.




### Type GridPoints
```go
type GridPoints struct {
	ID    string
	GridX int
	GridY int
}
```
GridPoints represents the grid points for a specific lat/long.


### Type OpaqueCloudCoverage
```go
type OpaqueCloudCoverage int
```
OpaqueCloudCoverage represents the cloud coverage as a fraction of the sky
as defined at https://www.weather.gov/bgm/forecast_terms.

### Constants
### UnknownOpaqueCloudCoverage, ClearSunny, MostlyClearSunny, PartlyCloudySunny, MostlyCloudy, Cloudy
```go
UnknownOpaqueCloudCoverage OpaqueCloudCoverage = iota
ClearSunny // 0 to 1/8 Opaque Cloud Coverage

MostlyClearSunny // 1/8 to 3/8

PartlyCloudySunny // 3/8 to 5/8

MostlyCloudy // 5/8 to 7/8

Cloudy // 7/8 to 8/8


```



### Functions

```go
func CloudOpacityFromShortForecast(forecast string) OpaqueCloudCoverage
```
CloudOpacityFromShortForecast returns the cloud opacity based on the short
forecast string.



### Methods

```go
func (o OpaqueCloudCoverage) String() string
```




### Type Option
```go
type Option func(o *options)
```

### Functions

```go
func WithForecastExpiration(d time.Duration) Option
```
WithForecastExpiration sets the expiration time for forecast cache entries,
if not set, the ValidTimes times entry in the forecast will be used.
Set this to a lower value to update the forecast more requgularly (eg.
to 24 hours).


```go
func WithGridpointExpiration(d time.Duration) Option
```
WithGridpointExpiration sets the expiration time for gridpoint cache
entries. The mapping from lat/long to gridpoints is largely stable so this
can be set to a long duration, the default is 1 week.




### Type Period
```go
type Period struct {
	StartTime           time.Time `json:"startTime"`
	EndTime             time.Time `json:"endTime"`
	Name                string    `json:"name"`
	ShortForecast       string    `json:"shortForecast"`
	OpaqueCloudCoverage OpaqueCloudCoverage
}
```
Period represents a forecast for a specific period of time.





