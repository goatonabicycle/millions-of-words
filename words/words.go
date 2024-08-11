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

	words := splitLyricsIntoWords(lyrics)

	for _, word := range words {
		cleanedWord := cleanWord(word)
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

func splitLyricsIntoWords(lyrics string) []string {
	words := strings.FieldsFunc(lyrics, func(r rune) bool {
		return unicode.IsSpace(r) || (unicode.IsPunct(r) && r != '\'' && r != '’' && r != '-')
	})
	return words
}

func cleanWord(word string) string {
	word = strings.TrimFunc(word, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '\'' && r != '’' && r != '-'
	})

	word = strings.ToLower(word)

	if strings.ContainsRune(word, '\'') || strings.ContainsRune(word, '’') || strings.ContainsRune(word, '-') {
		return word
	}

	return word
}

func MapToSortedList(wordCounts map[string]int) []models.WordCount {
	var wordFrequencies []models.WordCount
	for word, count := range wordCounts {
		wordFrequencies = append(wordFrequencies, models.WordCount{Word: word, Count: count})
	}
	sort.Slice(wordFrequencies, func(i, j int) bool {
		return wordFrequencies[i].Count > wordFrequencies[j].Count ||
			(wordFrequencies[i].Count == wordFrequencies[j].Count &&
				wordFrequencies[i].Word < wordFrequencies[j].Word)
	})
	return wordFrequencies
}
