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
		{
			name:   "Case with numbers and punctuation",
			lyrics: "123, hello! 123... Hello, world?",
			expected: []models.WordCount{
				{Word: "123", Count: 2},
				{Word: "hello", Count: 2},
				{Word: "world", Count: 1},
			},
		},
		{
			name:     "Empty string",
			lyrics:   "",
			expected: nil,
		},
		{
			name:   "Contractions and hyphens",
			lyrics: "it's a well-known fact. I'm happy, you're happy, and they're excited!",
			expected: []models.WordCount{
				{Word: "happy", Count: 2},
				{Word: "i'm", Count: 1},
				{Word: "it's", Count: 1},
				{Word: "known", Count: 1},
				{Word: "they're", Count: 1},
				{Word: "well-known", Count: 1},
				{Word: "a", Count: 1},
				{Word: "fact", Count: 1},
				{Word: "and", Count: 1},
				{Word: "you're", Count: 1},
				{Word: "excited", Count: 1},
			},
		},
		{
			name:   "Words with single letters and punctuation",
			lyrics: "re, et, it's, I'm, test.",
			expected: []models.WordCount{
				{Word: "it's", Count: 1},
				{Word: "i'm", Count: 1},
				{Word: "test", Count: 1},
				{Word: "re", Count: 1},
				{Word: "et", Count: 1},
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
