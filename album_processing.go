package main

import (
	"encoding/base64"
	"html/template"
	"strings"
	"sync"

	"millions-of-words/models"
	"millions-of-words/words"
)

const (
	maxTopWords = 20
)

var (
	albumDetailsCache sync.Map
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

func prepareAlbumDetails(album models.BandcampAlbumData) map[string]interface{} {
	if cachedDetails, ok := albumDetailsCache.Load(album.ID); ok {
		return cachedDetails.(map[string]interface{})
	}

	album.AlbumWordFrequencies = words.AggregateWordFrequencies(album)
	if len(album.AlbumWordFrequencies) > maxTopWords {
		album.AlbumWordFrequencies = album.AlbumWordFrequencies[:maxTopWords]
	}

	tracksWithDetails := make([]models.TrackWithDetails, 0, len(album.Tracks))
	var totalWords, totalCharacters, totalCharactersNoSpaces, totalLines int
	var totalVowelCount, totalConsonantCount int
	uniqueWordsMap := make(map[string]struct{})
	wordLengthDistribution := make(map[int]int)

	for _, track := range album.Tracks {
		trackDetails := calculateTrackDetails(track)
		tracksWithDetails = append(tracksWithDetails, trackDetails)

		totalWords += trackDetails.TotalWords
		totalVowelCount += trackDetails.VowelCount
		totalConsonantCount += trackDetails.ConsonantCount
		totalCharacters += trackDetails.TotalCharacters
		totalCharactersNoSpaces += trackDetails.TotalCharactersNoSpaces
		totalLines += trackDetails.TotalLines

		for length, count := range trackDetails.WordLengthDistribution {
			wordLengthDistribution[length] += count
		}

		for _, wc := range trackDetails.SortedWordCounts {
			uniqueWordsMap[wc.Word] = struct{}{}
		}
	}

	album.TotalWords = totalWords
	album.TotalCharacters = totalCharacters
	album.TotalCharactersNoSpaces = totalCharactersNoSpaces
	album.TotalLines = totalLines
	album.AverageWordsPerTrack = calculateAverage(totalWords, len(album.Tracks))
	album.TotalUniqueWords = len(uniqueWordsMap)
	album.TotalVowelCount = totalVowelCount
	album.TotalConsonantCount = totalConsonantCount
	album.WordLengthDistribution = wordLengthDistribution
	album.ImageDataBase64 = base64.StdEncoding.EncodeToString(album.ImageData)

	albumWPM := calculateWPM(float64(totalWords), float64(album.TotalLength))

	result := map[string]interface{}{
		"Album":             album,
		"TracksWithDetails": tracksWithDetails,
		"AlbumWPM":          albumWPM,
	}

	albumDetailsCache.Store(album.ID, result)
	return result
}

func calculateTrackDetails(track models.BandcampTrackData) models.TrackWithDetails {
	sortedWordCounts, vowels, consonants, wordLengths := words.CalculateAndSortWordFrequencies(track.Lyrics)
	wordCount := len(strings.Fields(track.Lyrics))

	trackUniqueWordsMap := make(map[string]struct{}, len(sortedWordCounts))
	for _, wc := range sortedWordCounts {
		trackUniqueWordsMap[wc.Word] = struct{}{}
	}

	wpm := calculateWPM(float64(wordCount), float64(track.TotalLength))

	lyrics := template.HTML(track.Lyrics)
	totalCharacters := len(track.Lyrics)
	totalCharactersNoSpaces := len(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(track.Lyrics, " ", ""), "\n", ""), "\r", ""))
	totalLines := len(strings.Split(strings.ReplaceAll(track.Lyrics, "\r\n", "\n"), "\n"))

	return models.TrackWithDetails{
		Track:                   track,
		FormattedLyrics:         lyrics,
		SortedWordCounts:        sortedWordCounts,
		WordsPerMinute:          wpm,
		TotalWords:              wordCount,
		UniqueWords:             len(trackUniqueWordsMap),
		VowelCount:              vowels,
		ConsonantCount:          consonants,
		WordLengthDistribution:  wordLengths,
		TotalCharacters:         totalCharacters,
		TotalCharactersNoSpaces: totalCharactersNoSpaces,
		TotalLines:              totalLines,
	}
}

func calculateAverage(total, count int) int {
	if count > 0 {
		return total / count
	}
	return 0
}

func calculateWPM(words, minutes float64) float64 {
	if minutes > 0 {
		return words / (minutes / 60)
	}
	return 0
}
