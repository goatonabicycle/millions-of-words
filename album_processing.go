package main

import (
	"encoding/base64"
	"html/template"
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

func prepareAlbumDetails(album models.BandcampAlbumData) map[string]interface{} {
	album.AlbumWordFrequencies = words.AggregateWordFrequencies(album)
	if len(album.AlbumWordFrequencies) > 20 {
		album.AlbumWordFrequencies = album.AlbumWordFrequencies[:20]
	}

	totalWords := 0
	totalVowelCount := 0
	totalConsonantCount := 0
	totalCharacters := 0
	totalCharactersNoSpaces := 0
	totalLines := 0
	wordLengthDistribution := make(map[int]int)
	uniqueWordsMap := make(map[string]struct{})

	tracksWithDetails := []models.TrackWithDetails{}

	for _, track := range album.Tracks {
		trackDetails := calculateTrackDetails(track)

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

		tracksWithDetails = append(tracksWithDetails, trackDetails)
	}

	album.TotalWords = totalWords
	album.TotalCharacters = totalCharacters
	album.TotalCharactersNoSpaces = totalCharactersNoSpaces
	album.TotalLines = totalLines

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

	posCategorization := words.CategorizeWordsByPOS(track.Lyrics)

	broadCategoryCounts := make(map[string]int)
	for pos, wordsList := range posCategorization {
		if category, exists := words.PosTagToCategory()[pos]; exists {
			broadCategoryCounts[category] += len(wordsList)
		}
	}

	posCategorizationByWord := make(map[string]string)
	for pos, wordsList := range posCategorization {
		if category, exists := words.PosTagToCategory()[pos]; exists {
			for _, word := range wordsList {
				posCategorizationByWord[word] = category
			}
		}
	}

	wpm := 0.0
	if float64(track.TotalLength)/60 > 0 {
		wpm = float64(wordCount) / (float64(track.TotalLength) / 60)
	}

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
		POSCategorizationCounts: broadCategoryCounts,
		POSCategorization:       posCategorizationByWord,
		TotalCharacters:         totalCharacters,
		TotalCharactersNoSpaces: totalCharactersNoSpaces,
		TotalLines:              totalLines,
	}
}
