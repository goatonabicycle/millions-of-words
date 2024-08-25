package loader

import (
	"database/sql"
	"sort"
	"strings"

	"millions-of-words/models"
	"millions-of-words/words"

	_ "modernc.org/sqlite"
)

func LoadAlbumsData() ([]models.BandcampAlbumData, error) {
	dbPath := "data/db/albums.db"
	var albums []models.BandcampAlbumData

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	albumRows, err := db.Query("SELECT id, artist_name, album_name, description, image_url, bandcamp_url, ampwall_url, total_length, formatted_length FROM albums")
	if err != nil {
		return nil, err
	}
	defer albumRows.Close()

	for albumRows.Next() {
		var album models.BandcampAlbumData
		var ampwallURL sql.NullString

		err := albumRows.Scan(
			&album.ID,
			&album.ArtistName,
			&album.AlbumName,
			&album.Description,
			&album.ImageUrl,
			&album.BandcampUrl,
			&ampwallURL,
			&album.TotalLength,
			&album.FormattedLength,
		)
		if err != nil {
			return nil, err
		}

		// Convert sql.NullString to string
		if ampwallURL.Valid {
			album.AmpwallUrl = ampwallURL.String
		} else {
			album.AmpwallUrl = ""
		}

		trackRows, err := db.Query("SELECT name, total_length, formatted_length, lyrics FROM tracks WHERE album_id = ?", album.ID)
		if err != nil {
			return nil, err
		}

		var tracks []models.BandcampTrackData
		for trackRows.Next() {
			var track models.BandcampTrackData
			err := trackRows.Scan(&track.Name, &track.TotalLength, &track.FormattedLength, &track.Lyrics)
			if err != nil {
				return nil, err
			}
			tracks = append(tracks, track)
		}
		trackRows.Close()

		album.Tracks = tracks
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
