package words

import (
	"millions-of-words/models"
	"sort"
	"strings"
	"unicode"
)

func CalculateAndSortWordFrequencies(lyrics string) ([]models.WordCount, int, int, map[int]int) {
	if lyrics == "" {
		return nil, 0, 0, nil
	}

	lyrics = removeItalics(lyrics)

	wordCounts := make(map[string]int)
	vowelCount := 0
	consonantCount := 0
	wordLengthDistribution := make(map[int]int)

	words := splitLyricsIntoWords(lyrics)

	for _, word := range words {
		cleanedWord := cleanWord(word)
		if cleanedWord != "" {
			wordCounts[cleanedWord]++
			vowels, consonants := countVowelsAndConsonants(cleanedWord)
			vowelCount += vowels
			consonantCount += consonants
			wordLengthDistribution[len(cleanedWord)]++
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

	return sortedWordCounts, vowelCount, consonantCount, wordLengthDistribution
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

	if word == "-" || word == "--" {
		return ""
	}

	if strings.ContainsRune(word, '\'') || strings.ContainsRune(word, '’') || strings.ContainsRune(word, '-') {
		return word
	}
	return word
}

func countVowelsAndConsonants(word string) (int, int) {
	vowels := 0
	consonants := 0
	for _, r := range word {
		switch r {
		case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
			vowels++
		case 'b', 'c', 'd', 'f', 'g', 'h', 'j', 'k', 'l', 'm', 'n', 'p', 'q', 'r', 's', 't', 'v', 'w', 'x', 'y', 'z',
			'B', 'C', 'D', 'F', 'G', 'H', 'J', 'K', 'L', 'M', 'N', 'P', 'Q', 'R', 'S', 'T', 'V', 'W', 'X', 'Y', 'Z':
			consonants++
		}
	}
	return vowels, consonants
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

func removeItalics(text string) string {
	replacer := strings.NewReplacer(
		"𝘢", "a", "𝘣", "b", "𝘤", "c", "𝘥", "d", "𝘦", "e", "𝘧", "f", "𝘨", "g", "𝘩", "h",
		"𝘪", "i", "𝘫", "j", "𝘬", "k", "𝘭", "l", "𝘮", "m", "𝘯", "n", "𝘰", "o", "𝘱", "p",
		"𝘲", "q", "𝘳", "r", "𝘴", "s", "𝘵", "t", "𝘶", "u", "𝘷", "v", "𝘸", "w", "𝘹", "x",
		"𝘺", "y", "𝘻", "z", "𝘈", "A", "𝘉", "B", "𝘊", "C", "𝘋", "D", "𝘌", "E", "𝘍", "F",
		"𝘎", "G", "𝘏", "H", "𝘐", "I", "𝘑", "J", "𝘒", "K", "𝘓", "L", "𝘔", "M", "𝘕", "N",
		"𝘖", "O", "𝘗", "P", "𝘘", "Q", "𝘙", "R", "𝘚", "S", "𝘛", "T", "𝘜", "U", "𝘝", "V",
		"𝘞", "W", "𝘟", "X", "𝘠", "Y", "𝘡", "Z",
	)
	return replacer.Replace(text)
}
