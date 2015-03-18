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
	if s != "" {
		t.Error("Unexpected status output.")
	}
}

func TestNewEvent(t *testing.T) {
	e := newEvent("test_toast")
	if e.Name != "test_toast" {
		t.Error("Failed to set event name.")
	}
	// This should not take even close to a second.
	if time.Since(e.Timestamp) > time.Second {
		t.Error("Failed to set event timestamp.")
	}
}

func TestNewCallback(t *testing.T) {
	c := newCallback("test_toast", "http://localhost:1339/ev/")
	if c.event != "test_toast" {
		t.Error("Failed to set callback event.")
	}
	if c.addr != "http://localhost:1339/ev/" {
		t.Error("Failed to set callback address.")
	}
}

func TestRegisterCallback(t *testing.T) {
	registerCallback("test_bread", "http://localhost:1339/ev/")
	tev := callbacks["test_bread"][0]
	if tev.event != "test_bread" {
		t.Error("Failed to set callback event.")
	}
	if tev.addr != "http://localhost:1339/ev/" {
		t.Error("Failed to set callback address.")
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
	}
	if len(callbacks["test_apiv1"]) != 1 {
		t.Error("Failed to register a callback.")
	}
	if !strings.Contains(callbacks["test_apiv1"][0].addr, "localhost") {
		t.Error("Failed to register a callback address correctly.")
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
			t.Error("Failed to parse the server's answer.")
		} else if parsed.Name != "test_apiv1" {
			t.Error("Event name mismatch.")
		} else if time.Since(parsed.Timestamp) > time.Second {
			t.Error("Event timestamp mismatch.")
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
	}

	select {
	case <-timeout:
		t.Error("Fired event did not reach test (in time).")
	case <-receivedAnswer:
		// Desired behaviour.
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
		t.Error(err.Error())
	}
	if resp.StatusCode != 200 {
		t.Error("Server says we failed to unregister a callback.")
	}
	if len(callbacks["test_apiv1"]) != 0 {
		t.Error("Failed to unregister a callback.")
	}

	// Malformed request
	tpl = []byte(`
		{
			"event":   "test_apiv1",
			"address": "http://localhost:8081/api/v1/ev/14/"
		 }
	`)

	resp, err = http.Post("http://localhost:8080/api/v1/",
	                      "application/json", bytes.NewBuffer(tpl))
	if err != nil {
		t.Error(err.Error())
	}
	if resp.StatusCode != 400 {
		t.Error("Server accepted a malformed request.")
	}
}

