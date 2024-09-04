package main

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0s"},
		{30, "30s"},
		{60, "1m 0s"},
		{90, "1m 30s"},
		{3600, "1h 0m 0s"},
		{3661, "1h 1m 1s"},
	}

	for _, test := range tests {
		result := formatDuration(test.input)
		if result != test.expected {
			t.Errorf("formatDuration(%d) = %s; want %s", test.input, result, test.expected)
		}
	}
}

func TestParseTrackDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		hasError bool
	}{
		{"3:30", 3*time.Minute + 30*time.Second, false},
		{"1:05", 1*time.Minute + 5*time.Second, false},
		{"0:45", 45 * time.Second, false},
		{"invalid", 0, true},
		{"5:", 0, true},
		{":30", 0, true},
		{"5:60", 0, true},
	}

	for _, test := range tests {
		result, err := parseTrackDuration(test.input)
		if test.hasError {
			if err == nil {
				t.Errorf("parseTrackDuration(%s) expected an error, but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("parseTrackDuration(%s) unexpected error: %v", test.input, err)
			}
			if result != test.expected {
				t.Errorf("parseTrackDuration(%s) = %v; want %v", test.input, result, test.expected)
			}
		}
	}
}
