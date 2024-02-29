package words

import (
	"millions-of-words/models"
	"sort"
	"strings"
)

func CalculateAndSortWordFrequencies(lyrics string) []models.WordCount {
	wordCounts := make(map[string]int)
	words := strings.Fields(strings.ToLower(lyrics))
	for _, word := range words {
		cleanedWord := strings.Trim(word, ",.!?\"'")
		wordCounts[cleanedWord]++
	}

	var sortedWordCounts []models.WordCount
	for word, count := range wordCounts {
		sortedWordCounts = append(sortedWordCounts, models.WordCount{Word: word, Count: count})
	}
	sort.Slice(sortedWordCounts, func(i, j int) bool {
		return sortedWordCounts[i].Count > sortedWordCounts[j].Count
	})

	return sortedWordCounts
}

func MapToString(wordCounts map[string]int) string {
	var lyricsBuilder strings.Builder
	for word, count := range wordCounts {
		for i := 0; i < count; i++ {
			lyricsBuilder.WriteString(word + " ")
		}
	}
	return lyricsBuilder.String()
}
