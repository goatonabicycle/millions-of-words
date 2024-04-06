package words

import (
	"millions-of-words/models"
	"sort"
	"strings"
	"unicode"
)

func CalculateAndSortWordFrequencies(lyrics string) []models.WordCount {
	if lyrics == "" {
		return nil
	}

	wordCounts := make(map[string]int)
	words := strings.FieldsFunc(strings.ToLower(lyrics), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	for _, word := range words {
		cleanedWord := strings.Trim(word, ",.!?\"'")
		if cleanedWord != "" {
			wordCounts[cleanedWord]++
		}
	}

	var sortedWordCounts []models.WordCount
	for word, count := range wordCounts {
		sortedWordCounts = append(sortedWordCounts, models.WordCount{Word: word, Count: count})
	}
	sort.Slice(sortedWordCounts, func(i, j int) bool {
		return sortedWordCounts[i].Count > sortedWordCounts[j].Count ||
			(sortedWordCounts[i].Count == sortedWordCounts[j].Count &&
				sortedWordCounts[i].Word < sortedWordCounts[j].Word)
	})

	return sortedWordCounts
}

func MapToString(wordCounts map[string]int) string {
	var words []string
	for word := range wordCounts {
		words = append(words, word)
	}
	sort.Strings(words)

	var lyricsBuilder strings.Builder
	for _, word := range words {
		for i := 0; i < wordCounts[word]; i++ {
			lyricsBuilder.WriteString(word + " ")
		}
	}
	return strings.TrimSpace(lyricsBuilder.String())
}
