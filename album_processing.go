package main

import (
	"encoding/base64"
	"html/template"
	"sort"
	"strings"

	"millions-of-words/models"
	"millions-of-words/words"
)

func filterAlbumsByQuery(query string) []models.BandcampAlbumData {
	var filtered []models.BandcampAlbumData
	query = strings.ToLower(query)
	for _, album := range albums {
		if strings.Contains(strings.ToLower(album.ArtistName), query) || strings.Contains(strings.ToLower(album.AlbumName), query) {
			filtered = append(filtered, album)
		}
	}
	return filtered
}

func aggregateWordFrequencies(album models.BandcampAlbumData) []models.WordCount {
	wordFreqMap := make(map[string]int)
	for _, track := range album.Tracks {
		wordCounts, _, _, _ := words.CalculateAndSortWordFrequencies(track.Lyrics)
		for _, wc := range wordCounts {
			wordFreqMap[wc.Word] += wc.Count
		}
	}

	var totalWordFrequencies []models.WordCount
	for word, count := range wordFreqMap {
		totalWordFrequencies = append(totalWordFrequencies, models.WordCount{Word: word, Count: count})
	}

	sort.Slice(totalWordFrequencies, func(i, j int) bool {
		if totalWordFrequencies[i].Count == totalWordFrequencies[j].Count {
			return totalWordFrequencies[i].Word < totalWordFrequencies[j].Word
		}
		return totalWordFrequencies[i].Count > totalWordFrequencies[j].Count
	})

	return totalWordFrequencies
}

func prepareAlbumDetails(album models.BandcampAlbumData) map[string]interface{} {
	album.AlbumWordFrequencies = aggregateWordFrequencies(album)
	if len(album.AlbumWordFrequencies) > 20 {
		album.AlbumWordFrequencies = album.AlbumWordFrequencies[:20]
	}

	totalWords := 0
	totalVowelCount := 0
	totalConsonantCount := 0
	wordLengthDistribution := make(map[int]int)
	uniqueWordsMap := make(map[string]struct{})

	tracksWithDetails := []models.TrackWithDetails{}

	for _, track := range album.Tracks {
		track.Lyrics = removeItalics(track.Lyrics)
		trackDetails := calculateTrackDetails(track)

		totalWords += trackDetails.TotalWords
		totalVowelCount += trackDetails.VowelCount
		totalConsonantCount += trackDetails.ConsonantCount

		for length, count := range trackDetails.WordLengthDistribution {
			wordLengthDistribution[length] += count
		}

		for _, wc := range trackDetails.SortedWordCounts {
			uniqueWordsMap[wc.Word] = struct{}{}
		}

		tracksWithDetails = append(tracksWithDetails, trackDetails)
	}

	album.TotalWords = totalWords
	if len(album.Tracks) > 0 {
		album.AverageWordsPerTrack = totalWords / len(album.Tracks)
	} else {
		album.AverageWordsPerTrack = 0
	}
	album.TotalUniqueWords = len(uniqueWordsMap)
	album.TotalVowelCount = totalVowelCount
	album.TotalConsonantCount = totalConsonantCount
	album.WordLengthDistribution = wordLengthDistribution
	album.ImageDataBase64 = base64.StdEncoding.EncodeToString(album.ImageData)

	albumWPM := 0.0
	if float64(album.TotalLength)/60 > 0 {
		albumWPM = float64(totalWords) / (float64(album.TotalLength) / 60)
	}

	return map[string]interface{}{
		"Album":             album,
		"TracksWithDetails": tracksWithDetails,
		"AlbumWPM":          albumWPM,
	}
}

func calculateTrackDetails(track models.BandcampTrackData) models.TrackWithDetails {
	sortedWordCounts, vowels, consonants, wordLengths := words.CalculateAndSortWordFrequencies(track.Lyrics)
	wordCount := len(strings.Fields(track.Lyrics))
	trackUniqueWordsMap := make(map[string]struct{})

	for _, wc := range sortedWordCounts {
		trackUniqueWordsMap[wc.Word] = struct{}{}
	}

	wpm := 0.0
	if float64(track.TotalLength)/60 > 0 {
		wpm = float64(wordCount) / (float64(track.TotalLength) / 60)
	}

	lyrics := template.HTML(track.Lyrics)

	return models.TrackWithDetails{
		Track:                  track,
		FormattedLyrics:        lyrics,
		SortedWordCounts:       sortedWordCounts,
		WordsPerMinute:         wpm,
		TotalWords:             wordCount,
		UniqueWords:            len(trackUniqueWordsMap),
		VowelCount:             vowels,
		ConsonantCount:         consonants,
		WordLengthDistribution: wordLengths,
	}
}

func removeItalics(text string) string {
	replacer := strings.NewReplacer(
		"𝘢", "a", "𝘣", "b", "𝘤", "c", "𝘥", "d",
		"𝘦", "e", "𝘧", "f", "𝘨", "g", "𝘩", "h",
		"𝘪", "i", "𝘫", "j", "𝘬", "k", "𝘭", "l",
		"𝘮", "m", "𝘯", "n", "𝘰", "o", "𝘱", "p",
		"𝘲", "q", "𝘳", "r", "𝘴", "s", "𝘵", "t",
		"𝘶", "u", "𝘷", "v", "𝘸", "w", "𝘹", "x",
		"𝘺", "y", "𝘻", "z", "𝘈", "A", "𝘉", "B",
		"𝘊", "C", "𝘋", "D", "𝘌", "E", "𝘍", "F",
		"𝘎", "G", "𝘏", "H", "𝘐", "I", "𝘑", "J",
		"𝘒", "K", "𝘓", "L", "𝘔", "M", "𝘕", "N",
		"𝘖", "O", "𝘗", "P", "𝘘", "Q", "𝘙", "R",
		"𝘚", "S", "𝘛", "T", "𝘜", "U", "𝘝", "V",
		"𝘞", "W", "𝘟", "X", "𝘠", "Y", "𝘡", "Z",
	)
	return replacer.Replace(text)
}
