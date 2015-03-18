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
	e := NewEvent("test_toast")
	// This should not take even close to a second.
	if e.name != "test_toast" {
		t.Error("Failed to set event name.")
	}
	if time.Since(e.timestamp) > time.Second {
		t.Error("Failed to set event timestamp.")
	}
}

