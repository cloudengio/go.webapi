// Copyright 2025 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package nws_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"cloudeng.io/webapi/clients/nws"
	"cloudeng.io/webapi/clients/nws/nwstestutil"
)

func TestLookup(t *testing.T) {
	ctx := context.Background()
	srv := nwstestutil.NewMockServer()
	defer srv.Close()
	url := srv.Run()

	for _, tc := range []struct {
		lat, long  float64
		gp         nws.GridPoints
		opts       []nws.Option
		pointCalls int
	}{
		{39.7456, -97.0892,
			nws.GridPoints{"TOP", 32, 81},
			nil,
			1},
		{39.7456, -97.0892,
			nws.GridPoints{"TOP", 32, 81},
			[]nws.Option{nws.WithGridpointExpiration(0)},
			3},
	} {
		api := nws.NewAPI(tc.opts...)
		api.SetHost(url)

		for i := 0; i < 3; i++ {
			gp, err := api.LookupGridPoints(ctx, tc.lat, tc.long)
			if err != nil {
				t.Fatalf("failed to get grid points: %v", err)
			}
			if got, want := gp, tc.gp; got != want {
				t.Errorf("got %v, want %v", got, want)
			}
		}
		if got, want := srv.LookupCalls(), tc.pointCalls; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		srv.ResetLookupCalls()
	}

}

func TestForecasts(t *testing.T) {
	ctx := context.Background()
	srv := nwstestutil.NewMockServer()
	srv.SetValidTimes(time.Now().In(time.UTC))
	defer srv.Close()
	url := srv.Run()

	for _, tc := range []struct {
		gp            nws.GridPoints
		opts          []nws.Option
		forecastCalls int
	}{
		{nws.GridPoints{"TOP", 32, 81}, nil, 1},
		{nws.GridPoints{"TOP", 32, 81},
			[]nws.Option{nws.WithForecastExpiration(time.Nanosecond)}, 3},
	} {

		api := nws.NewAPI(tc.opts...)
		api.SetHost(url)

		for i := 0; i < 3; i++ {
			fc, err := api.GetForecasts(ctx, tc.gp)
			if err != nil {
				t.Fatalf("failed to get forecasts: %v", err)
			}
			if got, want := fc.ValidFrom.Truncate(2*time.Second), time.Now().Truncate(2*time.Second).In(time.UTC); !got.Equal(want) {
				t.Errorf("got %v, want %v", got, want)
			}
			if got, want := fc.ValidFor, time.Hour*24*7; got != want {
				t.Errorf("got %v, want %v", got, want)
			}
			if got, want := len(fc.Periods), 14; got != want {
				t.Errorf("got %v, want %v", got, want)
			}
			start := fc.Periods[0].StartTime
			end := fc.Periods[0].EndTime
			if got, want := end.Sub(start), time.Hour*3; got != want {
				t.Errorf("got %v, want %v", got, want)
			}
			coverage := []nws.OpaqueCloudCoverage{fc.Periods[0].OpaqueCloudCoverage}
			for _, p := range fc.Periods[1:] {
				start := p.StartTime
				end := p.EndTime
				if got, want := end.Sub(start), time.Hour*12; got != want {
					t.Errorf("got %v, want %v", got, want)
				}
				coverage = append(coverage, p.OpaqueCloudCoverage)
			}

			if got, want := coverage, []nws.OpaqueCloudCoverage{
				nws.PartlyCloudySunny,
				nws.PartlyCloudySunny,
				nws.MostlyClearSunny,
				nws.PartlyCloudySunny,
				nws.ClearSunny,
				nws.MostlyClearSunny,
				nws.PartlyCloudySunny,
				nws.MostlyCloudy,
				nws.PartlyCloudySunny,
				nws.PartlyCloudySunny,
				nws.PartlyCloudySunny,
				nws.MostlyCloudy,
				nws.MostlyClearSunny,
				nws.PartlyCloudySunny,
			}; !slices.Equal(got, want) {
				t.Errorf("got %v, want %v", got, want)
			}
		}
		if got, want := srv.ForecastCalls(), tc.forecastCalls; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		srv.ResetForecastCalls()
	}

}

func TestPeriodLookup(t *testing.T) {
	ctx := context.Background()
	srv := nwstestutil.NewMockServer()
	srv.SetValidTimes(time.Now().In(time.UTC))
	defer srv.Close()
	url := srv.Run()

	gp := nws.GridPoints{"TOP", 32, 81}
	api := nws.NewAPI()
	api.SetHost(url)
	fc, err := api.GetForecasts(ctx, gp)
	if err != nil {
		t.Fatalf("failed to get forecasts: %v", err)
	}

	if got, want := srv.ForecastCalls(), 1; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	// test forecast data is in the range
	// 2025-01-06T15:00:00-06:00 to 2025-01-13T06:00:00-06:00

	first := time.Date(2025, 1, 6, 15, 0, 0, 0, time.FixedZone("CST", -6*60*60))
	last := time.Date(2025, 1, 13, 6, 0, 0, 0, time.FixedZone("CST", -6*60*60))

	if f, ok := fc.PeriodFor(first.Add(-time.Second)); ok {
		t.Errorf("expected no period: found %v...%v", f.StartTime, f.EndTime)
	}

	if f, ok := fc.PeriodFor(last); ok {
		t.Errorf("expected no period: found %v...%v", f.StartTime, f.EndTime)
	}

	frc, _ := fc.PeriodFor(first)
	if got, want := frc.Name, "This Afternoon"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	frc, _ = fc.PeriodFor(first.Add(time.Hour * 3))
	if got, want := frc.Name, "Tonight"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	frc, _ = fc.PeriodFor(last.Add(-time.Second))
	if got, want := frc.Name, "Sunday Night"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
