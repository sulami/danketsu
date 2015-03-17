package main

import (
	"testing"
)

func TestSanity(t *testing.T) {
	if 1 != 1 {
		t.Error("Failed sanity check.")
	}
}

