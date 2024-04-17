package main

import (
	"testing"
)

func TestSanitizeFilename(t *testing.T) {

	tests := []struct {
		name     string // Test case name
		input    string // Input string to sanitize
		expected string // Expected result after sanitization
	}{
		{"Normal String", "The Sway of Mountains", "The_Sway_of_Mountains"},
		{"Extra Spaces", "   The    Sway      of       Mountains", "The_Sway_of_Mountains"},
		{"Special Characters", "My/Album:Name*", "My_Album_Name"},
		{"Leading And Trailing Spaces", "  My Album  ", "My_Album"},
		{"Empty String", "", ""},
		{"Newlines And Tabs", "\nMy\tAlbum\n", "My_Album"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeFilename(tc.input)
			if got != tc.expected {
				t.Errorf("sanitizeFilename(%q) = %q; want %q", tc.input, got, tc.expected)
			}
		})
	}
}
