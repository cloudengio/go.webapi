// Copyright 2026 cloudeng llc. All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

package operations_test

import (
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"cloudeng.io/webapi/operations"
)

// SensorReading represents a single typed record in a homogeneous stream.
type SensorReading struct {
	SensorID string  `json:"sensor_id"`
	Value    float64 `json:"value"`
}

// UnmarshalJSONFrom implements operations.Decodable (json.UnmarshalerFrom).
func (s *SensorReading) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	type plain SensorReading
	return json.UnmarshalDecode(dec, (*plain)(s))
}

// ExampleStreamEndpoint_singleType demonstrates reading a stream of homogeneous items.
func ExampleStreamEndpoint_singleType() {
	// Mock server that streams three newline-delimited SensorReading objects.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		readings := []SensorReading{
			{SensorID: "temp-1", Value: 21.5},
			{SensorID: "temp-2", Value: 22.0},
			{SensorID: "temp-3", Value: 21.8},
		}
		for _, r := range readings {
			b, err := json.Marshal(r)
			if err != nil {
				return
			}
			fmt.Fprintf(w, "%s\n", b)
			w.(http.Flusher).Flush()
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	ep := operations.NewStreamEndpoint[*SensorReading]()

	stream, err := ep.Get(ctx, srv.URL)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer stream.Close()

	// Reuse a single struct instance across iterations for zero allocations.
	var reading SensorReading
	for {
		err := stream.Next(&reading)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		fmt.Printf("%s: %.1f\n", reading.SensorID, reading.Value)
	}

	// Output:
	// temp-1: 21.5
	// temp-2: 22.0
	// temp-3: 21.8
}

// Concrete event types for the heterogeneous stream example.
type MessageEvent struct {
	Text string `json:"text"`
}

type ProgressEvent struct {
	Percent int `json:"percent"`
}

type DoneEvent struct {
	Status string `json:"status"`
}

// StreamEvent is a union/envelope type that can unmarshal any of the supported event variants.
type StreamEvent struct {
	Type     string
	Message  *MessageEvent
	Progress *ProgressEvent
	Done     *DoneEvent
}

// UnmarshalJSONFrom implements operations.Decodable (json.UnmarshalerFrom).
// It reads the raw JSON value, extracts the discriminator ("type"),
// and unmarshals into the corresponding variant.
func (e *StreamEvent) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	val, err := dec.ReadValue()
	if err != nil {
		return err
	}

	var disc struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(val, &disc); err != nil {
		return err
	}

	*e = StreamEvent{Type: disc.Type}
	switch disc.Type {
	case "message":
		e.Message = new(MessageEvent)
		return json.Unmarshal(val, e.Message)
	case "progress":
		e.Progress = new(ProgressEvent)
		return json.Unmarshal(val, e.Progress)
	case "done":
		e.Done = new(DoneEvent)
		return json.Unmarshal(val, e.Done)
	default:
		return fmt.Errorf("unknown event type: %q", disc.Type)
	}
}

// ExampleStreamEndpoint_unionType demonstrates reading a stream of heterogeneous items
// using a tagged union / envelope type.
func ExampleStreamEndpoint_unionType() {
	// Mock server streaming different event types in a single stream.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		events := []string{
			`{"type":"message","text":"task started"}`,
			`{"type":"progress","percent":50}`,
			`{"type":"progress","percent":100}`,
			`{"type":"done","status":"success"}`,
		}
		for _, ev := range events {
			fmt.Fprintln(w, ev)
			w.(http.Flusher).Flush()
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	ep := operations.NewStreamEndpoint[*StreamEvent]()

	stream, err := ep.Get(ctx, srv.URL)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer stream.Close()

	// Reuse a single envelope instance across iterations.
	var ev StreamEvent
	for {
		err := stream.Next(&ev)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			fmt.Println("error:", err)
			return
		}

		switch ev.Type {
		case "message":
			fmt.Printf("message: %s\n", ev.Message.Text)
		case "progress":
			fmt.Printf("progress: %d%%\n", ev.Progress.Percent)
		case "done":
			fmt.Printf("done: %s\n", ev.Done.Status)
		}
	}

	// Output:
	// message: task started
	// progress: 50%
	// progress: 100%
	// done: success
}
