// Copyright 2024 Cosmos Nicolaou. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// Package nws provides a client for the National Weather Service API.
package nws

import (
	"sync"
	"time"

	"cloudeng.io/algo/container/list"
)

type gridpointEntry struct {
	lat, long float64
	id        string
	x, y      int
}

type gridPointsCache struct {
	mu         sync.Mutex
	lastUpdate time.Time
	entries    *list.Double[gridpointEntry]
	expiration time.Duration
}

func newGridPointsCache(expiration time.Duration) *gridPointsCache {
	return &gridPointsCache{
		entries:    list.NewDouble[gridpointEntry](),
		expiration: expiration,
	}
}

func (gpc *gridPointsCache) lookup(lat, long float64) (x, y int, id string, ok bool) {
	gpc.mu.Lock()
	defer gpc.mu.Unlock()
	if time.Since(gpc.lastUpdate) >= gpc.expiration {
		gpc.entries.Reset()
	}
	for e := range gpc.entries.Forward() {
		if e.lat == lat && e.long == long {
			return e.x, e.y, e.id, true
		}
	}
	return 0, 0, "", false
}

func (gpc *gridPointsCache) add(lat, long float64, id string, x, y int) {
	gpc.mu.Lock()
	defer gpc.mu.Unlock()
	gpc.entries.Append(gridpointEntry{lat, long, id, x, y})
	gpc.lastUpdate = time.Now()
}

type forecastEntry struct {
	gp        GridPoints
	forecasts Forecasts
}

type forecastCache struct {
	mu                 sync.Mutex
	entries            *list.Double[forecastEntry]
	forecastExpiration time.Duration
}

func newForecastCache(expiration time.Duration) *forecastCache {
	return &forecastCache{
		entries:            list.NewDouble[forecastEntry](),
		forecastExpiration: expiration,
	}
}

func (fc *forecastCache) lookup(gp GridPoints) (Forecasts, bool) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	for e := range fc.entries.Forward() {
		if e.gp == gp {
			exp := e.forecasts.ValidFor
			if fc.forecastExpiration > 0 {
				exp = min(e.forecasts.ValidFor, fc.forecastExpiration)
			}
			if time.Now().After(e.forecasts.ValidFrom.Add(exp)) {
				return Forecasts{}, false
			}
			return e.forecasts, true
		}
	}
	return Forecasts{}, false
}

func (fc *forecastCache) add(gp GridPoints, forecasts Forecasts) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.entries.Append(forecastEntry{gp: gp, forecasts: forecasts})
}
