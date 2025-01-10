// Copyright 2025 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package nws_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"cloudeng.io/datetime"
	"cloudeng.io/webapi/clients/nws"
	"cloudeng.io/webapi/webapitestutil"
)

func writeFile(name string, w http.ResponseWriter) {
	f, err := os.Open(filepath.Join("testdata", name))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()
	w.Header().Set("Content-Type", "application/json")
	if _, err := io.Copy(w, f); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeFileModifiedExpiration(name string, w http.ResponseWriter) {
	f, err := os.Open(filepath.Join("testdata", name))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	buf := &bytes.Buffer{}
	defer f.Close()
	if _, err := io.Copy(buf, f); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	re := regexp.MustCompile(`"validTimes":\s*"(.*?)"`)
	n := fmt.Sprintf(`"validTimes": "%s/%s"`, time.Now().In(time.UTC).Format("2006-01-02T15:04:05-07:00"), datetime.AsISO8601Period(time.Hour*24*7))
	nbuf := re.ReplaceAll(buf.Bytes(), []byte(n))
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(nbuf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var (
	pointsCalled   int64
	forecastCalled int64
)

func runMock() *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "points") {
			atomic.AddInt64(&pointsCalled, 1)
			writeFile("gridpoint.json", w)
			return
		}
		if strings.Contains(r.URL.Path, "forecasts") {
			atomic.AddInt64(&forecastCalled, 1)
			writeFileModifiedExpiration("forecasts.json", w)
			return
		}
		http.Error(w, r.URL.Path, http.StatusNotFound)
	})
	return webapitestutil.NewServer(handler)
}

func TestLookup(t *testing.T) {
	ctx := context.Background()
	srv := runMock()
	defer srv.Close()

	for _, tc := range []struct {
		lat, long  float64
		gp         nws.GridPoints
		opts       []nws.Option
		pointCalls int64
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
		api.SetHost(srv.URL)

		for i := 0; i < 3; i++ {
			gp, err := api.LookupGridPoints(ctx, tc.lat, tc.long)
			if err != nil {
				t.Fatalf("failed to get grid points: %v", err)
			}
			if got, want := gp, tc.gp; got != want {
				t.Errorf("got %v, want %v", got, want)
			}
		}
		if got, want := atomic.LoadInt64(&pointsCalled), tc.pointCalls; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		atomic.StoreInt64(&pointsCalled, 0)
	}

}

func TestForecasts(t *testing.T) {
	ctx := context.Background()
	srv := runMock()
	defer srv.Close()

	for _, tc := range []struct {
		gp            nws.GridPoints
		opts          []nws.Option
		forecastCalls int64
	}{
		{nws.GridPoints{"TOP", 32, 81}, nil, 1},
		{nws.GridPoints{"TOP", 32, 81},
			[]nws.Option{nws.WithForecastExpiration(time.Nanosecond)}, 3},
	} {

		api := nws.NewAPI(tc.opts...)
		api.SetHost(srv.URL)

		for i := 0; i < 3; i++ {
			fc, err := api.GetForecasts(ctx, tc.gp)
			if err != nil {
				t.Fatalf("failed to get forecasts: %v", err)
			}
			if got, want := fc.ValidFrom, time.Now().Truncate(time.Second).In(time.UTC); !got.Equal(want) {
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
		if got, want := atomic.LoadInt64(&forecastCalled), tc.forecastCalls; got != want {
			t.Errorf("got %v, want %v", got, want)
		}
		atomic.StoreInt64(&forecastCalled, 0)
	}

}

func TestPeriodLookup(t *testing.T) {
	ctx := context.Background()
	srv := runMock()
	defer srv.Close()
	gp := nws.GridPoints{"TOP", 32, 81}
	api := nws.NewAPI()
	api.SetHost(srv.URL)
	fc, err := api.GetForecasts(ctx, gp)
	if err != nil {
		t.Fatalf("failed to get forecasts: %v", err)
	}

	if got, want := atomic.LoadInt64(&forecastCalled), int64(1); got != want {
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
