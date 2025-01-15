// Copyright 2025 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package nwstestutil

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"time"

	"cloudeng.io/datetime"
	"cloudeng.io/webapi/webapitestutil"
)

//go:embed forecasts.json gridpoint.json
var cannedData embed.FS

type NWSMockServer struct {
	mu            sync.Mutex
	srv           *httptest.Server
	lookupCalls   int64
	forecastCalls int64
	validTimes    string
}

func NewMockServer() *NWSMockServer {
	return &NWSMockServer{}
}

func (ms *NWSMockServer) LookupCalls() int64 {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.lookupCalls
}

func (ms *NWSMockServer) ForecastCalls() int64 {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.forecastCalls
}

func (ms *NWSMockServer) ResetLookupCalls() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.lookupCalls = 0
}

func (ms *NWSMockServer) ResetForecastCalls() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.forecastCalls = 0
}

func (ms *NWSMockServer) SetValidTimes(when time.Time) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.validTimes = fmt.Sprintf(`"validTimes": "%s/%s"`, when.Format("2006-01-02T15:04:05-07:00"), datetime.AsISO8601Period(time.Hour*24*7))
}

func (ms *NWSMockServer) writeResponse(w http.ResponseWriter, err error, buf []byte) {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(w, bytes.NewBuffer(buf)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

var validTimesRE = regexp.MustCompile(`"validTimes":\s*"(.*?)"`)

func (ms *NWSMockServer) Run() string {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ms.mu.Lock()
		defer ms.mu.Unlock()

		if strings.Contains(r.URL.Path, "points") {
			ms.lookupCalls++
			data, err := cannedData.ReadFile("gridpoint.json")
			ms.writeResponse(w, err, data)
			return
		}
		if strings.Contains(r.URL.Path, "forecasts") {
			ms.forecastCalls++
			data, err := cannedData.ReadFile("forecasts.json")
			buf := validTimesRE.ReplaceAll(data, []byte(ms.validTimes))
			ms.writeResponse(w, err, buf)
			return
		}
		http.Error(w, r.URL.Path, http.StatusNotFound)
	})
	ms.srv = webapitestutil.NewServer(handler)
	return ms.srv.URL
}

func (ms *NWSMockServer) Close() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if ms.srv != nil {
		ms.srv.Close()
	}
}
