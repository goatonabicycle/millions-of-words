package words

import (
	"millions-of-words/models"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

func CalculateAndSortWordFrequencies(lyrics string) []models.WordCount {
	if lyrics == "" {
		return nil
	}

	wordCounts := make(map[string]int)

	words := strings.FieldsFunc(strings.ToLower(lyrics), func(r rune) bool {
		return unicode.IsSpace(r) || (unicode.IsPunct(r) && r != '\'')
	})

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

func cleanWord(word string) string {
	for len(word) > 0 {
		r, size := utf8.DecodeRuneInString(word)
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '\'' {
			break
		}
		word = word[size:]
	}

	for len(word) > 0 {
		r, size := utf8.DecodeLastRuneInString(word)
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '\'' {
			break
		}
		word = word[:len(word)-size]
	}

	word = strings.ReplaceAll(word, "--", "")
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
