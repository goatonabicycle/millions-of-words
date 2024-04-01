package words

import (
	"millions-of-words/models"
	"reflect"
	"testing"
)

func TestCalculateAndSortWordFrequencies(t *testing.T) {
	tests := []struct {
		name     string
		lyrics   string
		expected []models.WordCount
	}{
		{
			name:   "Normal case",
			lyrics: "hello, world! Hello?",
			expected: []models.WordCount{
				{Word: "hello", Count: 2},
				{Word: "world", Count: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateAndSortWordFrequencies(tt.lyrics)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("CalculateAndSortWordFrequencies(%q) = %v, want %v", tt.lyrics, result, tt.expected)
			}
		})
	}
}

func TestMapToString(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]int
		expected string
	}{
		{
			name: "Normal case",
			input: map[string]int{
				"world": 2,
				"hello": 3,
			},
			expected: "hello hello hello world world",
		},
		{
			name:     "Empty map",
			input:    map[string]int{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapToString(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("MapToString() = %q, want %q", result, tt.expected)
			}
		})
	}
}
