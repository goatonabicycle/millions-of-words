package loader

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"sort"
	"strings"

	"millions-of-words/models"
	"millions-of-words/words"

	_ "modernc.org/sqlite"
)

const dbPath = "data/db/albums.db"

func LoadAlbumsData() ([]models.BandcampAlbumData, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database at %s: %w", dbPath, err)
	}
	defer db.Close()

	albums, err := fetchAlbums(db)
	if err != nil {
		return nil, err
	}

	for i := range albums {
		if err := fetchTracks(db, &albums[i]); err != nil {
			log.Printf("Error fetching tracks for album %s: %v", albums[i].ID, err)
			continue
		}
		calculateAlbumMetrics(&albums[i])
	}

	sortAlbums(albums)
	return albums, nil
}

func fetchAlbums(db *sql.DB) ([]models.BandcampAlbumData, error) {
	query := `SELECT id, artist_name, album_name, image_url, image_data, album_color_average, 
              bandcamp_url, ampwall_url, total_length, formatted_length, date_added FROM albums`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying albums: %w", err)
	}
	defer rows.Close()

	var albums []models.BandcampAlbumData
	for rows.Next() {
		var album models.BandcampAlbumData
		var ampwallURL sql.NullString

		err := rows.Scan(
			&album.ID, &album.ArtistName, &album.AlbumName, &album.ImageUrl,
			&album.ImageData, &album.AlbumColorAverage, &album.BandcampUrl,
			&ampwallURL, &album.TotalLength, &album.FormattedLength, &album.DateAdded,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning album row: %w", err)
		}

		album.AmpwallUrl = ampwallURL.String
		album.ImageDataBase64 = base64.StdEncoding.EncodeToString(album.ImageData)
		albums = append(albums, album)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating album rows: %w", err)
	}

	return albums, nil
}

func fetchTracks(db *sql.DB, album *models.BandcampAlbumData) error {
	query := `SELECT name, total_length, formatted_length, lyrics FROM tracks WHERE album_id = ?`
	rows, err := db.Query(query, album.ID)
	if err != nil {
		return fmt.Errorf("error querying tracks: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var track models.BandcampTrackData
		err := rows.Scan(&track.Name, &track.TotalLength, &track.FormattedLength, &track.Lyrics)
		if err != nil {
			return fmt.Errorf("error scanning track row: %w", err)
		}
		album.Tracks = append(album.Tracks, track)
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating track rows: %w", err)
	}

	return nil
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
	album.AverageWordsPerTrack = calculateAverage(totalWords, len(album.Tracks))
	album.TotalUniqueWords = len(uniqueWordsMap)
	album.TotalVowelCount = totalVowelCount
	album.TotalConsonantCount = totalConsonantCount
	album.WordLengthDistribution = wordLengthDistribution
}

func calculateAverage(total, count int) int {
	if count > 0 {
		return total / count
	}
	return 0
}

func sortAlbums(albums []models.BandcampAlbumData) {
	sort.Slice(albums, func(i, j int) bool {
		artistI := strings.ToLower(albums[i].ArtistName)
		artistJ := strings.ToLower(albums[j].ArtistName)
		if artistI == artistJ {
			return strings.ToLower(albums[i].AlbumName) < strings.ToLower(albums[j].AlbumName)
		}
		return artistI < artistJ
	})
}
