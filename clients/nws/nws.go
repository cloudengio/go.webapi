// Copyright 2025 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package nws provides a client for the National Weather Service API.
package nws

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"cloudeng.io/datetime"
	"cloudeng.io/webapi/operations"
)

// OpaqueCloudCoverage represents the cloud coverage as a fraction of the sky
// as defined at https://www.weather.gov/bgm/forecast_terms.
type OpaqueCloudCoverage int

const (
	UnknownOpaqueCloudCoverage OpaqueCloudCoverage = iota
	ClearSunny                                     // 0 to 1/8 Opaque Cloud Coverage
	MostlyClearSunny                               // 1/8 to 3/8
	PartlyCloudySunny                              // 3/8 to 5/8
	MostlyCloudy                                   // 5/8 to 7/8
	Cloudy                                         // 7/8 to 8/8
)

const (
	APIHost = "https://api.weather.gov"
)

type gridPointResponse struct {
	Properties struct {
		X        int    `json:"gridX"`
		Y        int    `json:"gridY"`
		ID       string `json:"gridId"`
		Forecast string `json:"forecast"`
	}
}

type forecastResponse struct {
	Properties struct {
		Generated  time.Time `json:"generatedAt"`
		Updated    time.Time `json:"updateTime"`
		ValidTimes string    `json:"validTimes"`
		Periods    []Period  `json:"periods"`
	}
}

// API represents a client for the National Weather Service API.
type API struct {
	host          string
	pointsCache   *gridPointsCache
	forecastCache *forecastCache
}

type Option func(o *options)

// WithGridpointExpiration sets the expiration time for gridpoint cache entries.
// The mapping from lat/long to gridpoints is largely stable so this can be
// set to a long duration, the default is 1 week.
func WithGridpointExpiration(d time.Duration) Option {
	return func(o *options) {
		o.gridpointExpiration = d
	}
}

// WithForecastExpiration sets the expiration time for forecast cache entries,
// if not set, the ValidTimes times entry in the forecast will be used. Set
// this to a lower value to update the forecast more requgularly (eg. to 24 hours).
func WithForecastExpiration(d time.Duration) Option {
	return func(o *options) {
		o.forecastExpiration = d
	}
}

type options struct {
	gridpointExpiration time.Duration
	forecastExpiration  time.Duration
}

// NewAPI creates a new instance of the National Weather Service API client.
func NewAPI(opts ...Option) *API {
	var o options
	o.gridpointExpiration = time.Hour * 24 * 7
	for _, fn := range opts {
		fn(&o)
	}
	api := &API{
		host:          APIHost,
		pointsCache:   newGridPointsCache(o.gridpointExpiration),
		forecastCache: newForecastCache(o.forecastExpiration),
	}

	return api
}
func (a *API) SetHost(host string) {
	a.host = host
}

// GridPoints represents the grid points for a specific lat/long.
type GridPoints struct {
	ID    string
	GridX int
	GridY int
}

// LookupGridPoints returns the grid points for the specified lat/long. A
// cache keyed by exact lat/long is used to avoid repeated lookups. The
// expiration of cache entries defaults to 1 week and can be overridden
// with the WSithGridpointExpiration option.
func (a *API) LookupGridPoints(ctx context.Context, lat, long float64, opts ...operations.Option) (GridPoints, error) {
	if x, y, f, ok := a.pointsCache.lookup(lat, long); ok {
		return GridPoints{
			ID:    f,
			GridX: x,
			GridY: y,
		}, nil
	}
	ustr := fmt.Sprintf("%s/points/%f,%f", a.host, lat, long)
	u, err := url.Parse(ustr)
	if err != nil {
		return GridPoints{}, fmt.Errorf("failed to parse URL: %v: %w", ustr, err)
	}

	ep := operations.NewEndpoint[gridPointResponse](opts...)
	gpr, _, _, err := ep.Get(ctx, u.String())
	if err != nil {
		return GridPoints{}, fmt.Errorf("%v: grid point lookup failed: %w", u.String(), err)
	}
	gr := GridPoints{
		ID:    gpr.Properties.ID,
		GridX: gpr.Properties.X,
		GridY: gpr.Properties.Y,
	}
	a.pointsCache.add(lat, long, gr.ID, gr.GridX, gr.GridY)
	return gr, nil
}

// Period represents a forecast for a specific period of time.
type Period struct {
	StartTime           time.Time `json:"startTime"`
	EndTime             time.Time `json:"endTime"`
	Name                string    `json:"name"`
	ShortForecast       string    `json:"shortForecast"`
	OpaqueCloudCoverage OpaqueCloudCoverage
}

// Forecast represents the forecasts for a specific grid point for a number
// of periods.
type Forecast struct {
	Generated time.Time
	Updated   time.Time
	ValidFrom time.Time
	ValidFor  time.Duration
	Periods   []Period
}

// GetForecasts returns the forecasts for the specified grid point. A cache
// keyed by grid point is used to avoid repeated lookups. The expiration of
// cache entries defaults to the ValidTimes entry in the forecast and can be
// overridden with the WithForecastExpiration option.
func (a *API) GetForecasts(ctx context.Context, gp GridPoints, opts ...operations.Option) (Forecast, error) {
	if fc, ok := a.forecastCache.lookup(gp); ok {
		return fc, nil
	}
	ustr := fmt.Sprintf("%s/forecasts/%s/%d,%d", a.host, gp.ID, gp.GridX, gp.GridY)
	u, err := url.Parse(ustr)
	if err != nil {
		return Forecast{}, fmt.Errorf("failed to parse URL: %v: %w", ustr, err)
	}
	ep := operations.NewEndpoint[forecastResponse](opts...)
	frc, _, _, err := ep.Get(ctx, u.String())
	if err != nil {
		return Forecast{}, fmt.Errorf("%v: forecast download failed: %w", u.String(), err)
	}
	valid, dur, err := parseValidTimes(frc.Properties.ValidTimes)
	if err != nil {
		return Forecast{}, fmt.Errorf("failed to parse valid times: %q: %w", frc.Properties.ValidTimes, err)
	}
	fc := Forecast{
		Generated: frc.Properties.Generated,
		Updated:   frc.Properties.Updated,
		ValidFrom: valid,
		ValidFor:  dur,
		Periods:   make([]Period, len(frc.Properties.Periods)),
	}
	fc.Periods = make([]Period, len(frc.Properties.Periods))
	copy(fc.Periods, frc.Properties.Periods)
	for i, p := range fc.Periods {
		fc.Periods[i].OpaqueCloudCoverage = CloudOpacityFromShortForecast(p.ShortForecast)
	}
	a.forecastCache.add(gp, fc)
	return fc, nil
}

func parseValidTimes(val string) (time.Time, time.Duration, error) {
	parts := strings.Split(val, "/")
	if len(parts) != 2 {
		return time.Time{}, 0, fmt.Errorf("expected two parts separated by /")
	}
	start, err := time.Parse(time.RFC3339, parts[0])
	if err != nil {
		return time.Time{}, 0, fmt.Errorf("failed to parse start time: %w", err)
	}
	period, err := datetime.ParseISO8601Period(parts[1])
	if err != nil {
		return time.Time{}, 0, fmt.Errorf("failed to parse period: %w", err)
	}
	return start, period, nil
}

// PeriodFor returns the forecast for the specified time. The returned bool
// is false if there is no forecast for the specified time.
func (fc Forecast) PeriodFor(when time.Time) (Period, bool) {
	for _, f := range fc.Periods {
		if !f.StartTime.After(when) && f.EndTime.After(when) {
			return f, true
		}
	}
	return Period{}, false
}

// CloudOpacityFromShortForecast returns the cloud opacity based on the short
// forecast string.
func CloudOpacityFromShortForecast(forecast string) OpaqueCloudCoverage {
	forecast = strings.ToLower(forecast)
	switch {
	case strings.HasPrefix(forecast, "clear") || strings.HasPrefix(forecast, "sunny"):
		return ClearSunny
	case strings.HasPrefix(forecast, "mostly clear") || strings.HasPrefix(forecast, "mostly sunny"):
		return MostlyClearSunny
	case strings.HasPrefix(forecast, "partly cloudy") || strings.HasPrefix(forecast, "partly sunny"):
		return PartlyCloudySunny
	case strings.HasPrefix(forecast, "mostly cloudy"):
		return MostlyCloudy
	case strings.HasPrefix(forecast, "cloudy"):
		return Cloudy
	}
	return UnknownOpaqueCloudCoverage
}
