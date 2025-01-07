package main

import (
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
		if strings.Contains(strings.ToLower(album.ArtistName), query) ||
			strings.Contains(strings.ToLower(album.AlbumName), query) {
			filtered = append(filtered, album)
		}
	}
	return filtered
}

func prepareAlbumDetails(album models.BandcampAlbumData) map[string]interface{} {
	if cachedDetails, ok := albumDetailsCache.Load(album.ID); ok {
		return cachedDetails.(map[string]interface{})
	}

	for i := range album.Tracks {
		album.Tracks[i].Lyrics = words.NormalizeText(album.Tracks[i].Lyrics)
	}

	album.AlbumWordFrequencies = words.AggregateWordFrequencies(album)
	if len(album.AlbumWordFrequencies) > maxTopWords {
		album.AlbumWordFrequencies = album.AlbumWordFrequencies[:maxTopWords]
	}

	displayTitle := album.ArtistName + " - " + album.AlbumName
	if len(album.ReleaseDate) >= 4 {
		displayTitle += " (" + album.ReleaseDate[:4] + ")"
	}

	tracksWithDetails := make([]models.TrackWithDetails, 0, len(album.Tracks))
	for i, track := range album.Tracks {
		trackDetails := calculateTrackDetails(track)
		trackDetails.TrackNumber = i + 1
		tracksWithDetails = append(tracksWithDetails, trackDetails)
	}

	result := map[string]interface{}{
		"Album":             album,
		"DisplayTitle":      displayTitle,
		"TracksWithDetails": tracksWithDetails,
		"AlbumWPM":          calculateWPM(float64(album.TotalWords), float64(album.TotalLength)),
	}

	albumDetailsCache.Store(album.ID, result)
	return result
}

func calculateTrackDetails(track models.BandcampTrackData) models.TrackWithDetails {
	sortedWordCounts, vowels, consonants, wordLengths := words.CalculateAndSortWordFrequencies(track.Lyrics, track.IgnoredWords)

	wordCount := 0
	for _, wc := range sortedWordCounts {
		wordCount += wc.Count
	}

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

func calculateWPM(words, minutes float64) float64 {
	if minutes > 0 {
		return words / (minutes / 60)
	}
	return 0
}
