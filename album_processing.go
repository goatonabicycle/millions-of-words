package main

import (
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
	album.AverageWordsPerTrack = totalWords / len(album.Tracks)
	album.TotalUniqueWords = len(uniqueWordsMap)
	album.TotalVowelCount = totalVowelCount
	album.TotalConsonantCount = totalConsonantCount

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
