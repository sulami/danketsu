package main

import (
	"bytes"
	"strings"
	"testing"
	"time"
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
	if e.name != "test_toast" {
		t.Error("Failed to set event name.")
	}
	// This should not take even close to a second.
	if time.Since(e.timestamp) > time.Second {
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
	var tpl1 = []byte(`
		{"action": "register", "event": "test_apiv1"}
	`)

	resp, err := http.Post("http://localhost:8080/api/v1/",
	                       "application/json", bytes.NewBuffer(tpl1))
	if err != nil {
		t.Error(err.Error())
	}
	if resp.StatusCode != 200 {
		t.Error("Server says we failed to register a callback.")
	}
	if len(callbacks["test_apiv1"]) != 1 {
		t.Error("Failed to register a callback.")
	}
	if !strings.Contains(callbacks["test_apiv1"][0].addr, "127.0.0.1:") {
		t.Error("Failed to register a callback address correctly.")
	}

	// Unregister the same callback.
	var tpl2 = []byte(`
		{"action": "unregister", "event": "test_apiv1"}
	`)

	resp, err = http.Post("http://localhost:8080/api/v1/",
	                      "application/json", bytes.NewBuffer(tpl2))
	if err != nil {
		t.Error(err.Error())
	}
	if resp.StatusCode != 200 {
		t.Error("Server says we failed to unregister a callback.")
	}
	if len(callbacks["test_apiv1"]) != 0 {
		t.Error("Failed to unregister a callback.")
	}
}

