package loader

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"millions-of-words/models"
	"millions-of-words/words"
)

func LoadAlbumsData() ([]models.BandcampAlbumData, error) {
	var albums []models.BandcampAlbumData

	dataDir := "data"

	files, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		filePath := filepath.Join(dataDir, f.Name())
		file, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		var album models.BandcampAlbumData
		if err := json.Unmarshal(file, &album); err != nil {
			continue
		}

		calculateAlbumMetrics(&album)
		albums = append(albums, album)
	}

	sortAlbums(albums)
	return albums, nil
}

func calculateAlbumMetrics(album *models.BandcampAlbumData) {
	totalWords := 0
	totalVowelCount := 0
	totalConsonantCount := 0
	wordLengthDistribution := make(map[int]int)
	uniqueWordsMap := make(map[string]struct{})

	for _, track := range album.Tracks {
		sortedWordCounts, vowels, consonants, wordLengths := words.CalculateAndSortWordFrequencies(track.Lyrics)
		wordCount := len(strings.Fields(track.Lyrics))
		totalWords += wordCount
		totalVowelCount += vowels
		totalConsonantCount += consonants

		for length, count := range wordLengths {
			wordLengthDistribution[length] += count
		}

		for _, wc := range sortedWordCounts {
			uniqueWordsMap[wc.Word] = struct{}{}
		}
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
}

func sortAlbums(albums []models.BandcampAlbumData) {
	sort.Slice(albums, func(i, j int) bool {
		if strings.ToLower(albums[i].ArtistName) == strings.ToLower(albums[j].ArtistName) {
			return strings.ToLower(albums[i].AlbumName) < strings.ToLower(albums[j].AlbumName)
		}
		return strings.ToLower(albums[i].ArtistName) < strings.ToLower(albums[j].ArtistName)
	})
}
