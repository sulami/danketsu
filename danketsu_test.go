package main

import (
	"testing"
	"time"
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

