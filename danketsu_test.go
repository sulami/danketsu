package main

import (
	"bytes"
	"strings"
	"testing"
	"time"
	"encoding/json"
	"net/http"
)

func TestSanity(t *testing.T) {
	if 1 != 1 {
		t.Error("Failed sanity check.")
	}
}

func TestStatus(t *testing.T) {
	s := status()
	if !strings.Contains(string(s[:]), "\"Events\": 0") ||
	   !strings.Contains(string(s[:]), "\"Callbacks\": 0") {
		t.Error("Unexpected status output.")
		t.Error(string(s[:]))
	}
}

func TestNewEvent(t *testing.T) {
	e := newEvent("test_toast")
	if e.Name != "test_toast" {
		t.Error("Failed to set event name.")
		t.Errorf("test_toast != %v", e.Name)
	}
	// This should not take even close to a second.
	if time.Since(e.Timestamp) > time.Second {
		t.Error("Failed to set event timestamp.")
		t.Error("Timestamp: %v", e.Timestamp)
	}
}

func TestNewCallback(t *testing.T) {
	c := newCallback("test_toast", "http://localhost:1339/ev/")
	if c.event != "test_toast" {
		t.Error("Failed to set callback event.")
		t.Errorf("test_toast != %v", c.event)
	}
	if c.addr != "http://localhost:1339/ev/" {
		t.Error("Failed to set callback address.")
		t.Errorf("http://localhost:1339/ev/ != %v", c.addr)
	}
}

func TestRegisterCallback(t *testing.T) {
	registerCallback("test_bread", "http://localhost:1339/ev/")
	tev := state.callbacks["test_bread"][0]
	if tev.event != "test_bread" {
		t.Error("Failed to set callback event.")
		t.Errorf("test_bread != %v", tev.event)
	}
	if tev.addr != "http://localhost:1339/ev/" {
		t.Error("Failed to set callback address.")
		t.Errorf("http://localhost:1339/ev/ != %v", tev.addr)
	}
}

func TestApiV1Access(t *testing.T) {
	go main() // Start the actual webserver

	// Check general connectivity.
	_, err := http.Get("http://localhost:8080/api/v1/")
	if err != nil {
		t.Error(err.Error())
	}

	// Register some callback.
	var tpl = []byte(`
		{
			"action":  "register",
			"event":   "test_apiv1",
			"address": "http://localhost:8081/"
		 }
	`)

	resp, err := http.Post("http://localhost:8080/api/v1/",
	                       "application/json", bytes.NewBuffer(tpl))
	if err != nil {
		t.Error(err.Error())
	}
	if resp.StatusCode != 200 {
		t.Error("Server says we failed to register a callback.")
		t.Errorf("Status code: %v", resp.StatusCode)
	}
	if len(state.callbacks["test_apiv1"]) != 1 {
		t.Error("Failed to register a callback.")
		t.Errorf("1 != %v", len(state.callbacks["test_apiv1"]))
	}
	if state.callbacks["test_apiv1"][0].addr != "http://localhost:8081/" {
		t.Error("Failed to register a callback address correctly.")
		t.Errorf("http://localhost:8081/ != %v",
		         state.callbacks["test_apiv1"][0].addr)
	}

	// Fire an event and listen for an answer.
	receivedAnswer := make(chan bool, 1)
	timeout := time.After(time.Second / 10) // Should be enough.

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		type Answer struct {
			Name      string
			Timestamp time.Time
		}
		ldec := json.NewDecoder(r.Body)
		var parsed Answer
		err := ldec.Decode(&parsed)
		if err != nil {
			t.Error(err)
		} else if parsed.Name != "test_apiv1" {
			t.Error("Event name mismatch.")
			t.Errorf("test_apiv1 != %v", parsed.Name)
		} else if time.Since(parsed.Timestamp) > time.Second {
			t.Error("Event timestamp mismatch.")
			t.Errorf("Timestamp: %v", parsed.Timestamp)
		} else {
			receivedAnswer <- true
		}
	})
	go http.ListenAndServe(":8081", nil)

	tpl = []byte(`
		{
			"action":  "fire",
			"event":   "test_apiv1",
			"address": "http://localhost:8081/"
		 }
	`)

	resp, err = http.Post("http://localhost:8080/api/v1/",
	                      "application/json", bytes.NewBuffer(tpl))
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != 200 {
		t.Error("Server says we failed to fire an event.")
		t.Errorf("HTTP status code %v != 200", resp.StatusCode)
	}

	select {
	case <-timeout:
		t.Error("Fired event did not reach test (in time).")
	case <-receivedAnswer:
		// Desired behaviour.
	}

	if len(state.events) != 1 || state.events[0].Name != "test_apiv1" {
		t.Error("In-memory event log has not been updated properly.")
		t.Errorf("Number of events: %v", len(state.events))
		t.Errorf("test_apiv1 != %v", state.events[0].Name)
	}

	// Unregister the same callback.
	tpl = []byte(`
		{
			"action":  "unregister",
			"event":   "test_apiv1",
			"address": "http://localhost:8081/"
		 }
	`)

	resp, err = http.Post("http://localhost:8080/api/v1/",
	                      "application/json", bytes.NewBuffer(tpl))
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != 200 {
		t.Error("Server says we failed to unregister a callback.")
		t.Errorf("HTTP status code %v != 200", resp.StatusCode)
	}
	if len(state.callbacks["test_apiv1"]) != 0 {
		t.Error("Failed to unregister a callback.")
		t.Errorf("0 != %v", len(state.callbacks["test_apiv1"]))
	}

	// Malformed request
	tpl = []byte(`
		{
			"event":   "test_apiv1"
			"address": "http://localhost:8081/api/v1/ev/14/"
		 }
	`)

	resp, err = http.Post("http://localhost:8080/api/v1/",
	                      "application/json", bytes.NewBuffer(tpl))
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != 400 {
		t.Error("Server accepted a malformed request.")
		t.Errorf("HTTP status code %v != 400", resp.StatusCode)
	}

	s := status()
	if !strings.Contains(string(s[:]), "\"Events\": 1") ||
	   !strings.Contains(string(s[:]), "\"Callbacks\": 2") {
		t.Error("Unexpected status output.")
		t.Error(string(s[:]))
	}
}

