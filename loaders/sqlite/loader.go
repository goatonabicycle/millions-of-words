package loader

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"strings"

	"millions-of-words/models"
	"millions-of-words/words"

	_ "modernc.org/sqlite"
)

const dbPath = "data/db/albums.db"

func LoadAlbumsData(limit ...int) ([]models.BandcampAlbumData, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database at %s: %w", dbPath, err)
	}
	defer db.Close()

	albums, err := fetchAlbums(db, limit...)
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
	return albums, nil
}

func fetchAlbums(db *sql.DB, limit ...int) ([]models.BandcampAlbumData, error) {
	query := `SELECT id, slug, artist_name, album_name, image_url, image_data, 
        album_color_average, bandcamp_url, ampwall_url, metal_archives_url, 
        total_length, formatted_length, date_added 
        FROM albums 
        ORDER BY date_added DESC`

	if len(limit) > 0 && limit[0] > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit[0])
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying albums: %w", err)
	}
	defer rows.Close()

	var albums []models.BandcampAlbumData
	for rows.Next() {
		var album models.BandcampAlbumData
		var ampwallURL sql.NullString
		var metalArchivesURL sql.NullString

		err := rows.Scan(
			&album.ID,
			&album.Slug,
			&album.ArtistName,
			&album.AlbumName,
			&album.ImageUrl,
			&album.ImageData,
			&album.AlbumColorAverage,
			&album.BandcampUrl,
			&ampwallURL,
			&metalArchivesURL,
			&album.TotalLength,
			&album.FormattedLength,
			&album.DateAdded,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning album row: %w", err)
		}

		album.AmpwallUrl = ampwallURL.String
		album.MetalArchivesURL = metalArchivesURL.String
		album.ImageDataBase64 = base64.StdEncoding.EncodeToString(album.ImageData)
		albums = append(albums, album)
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
	totalVowels := 0
	totalConsonants := 0
	totalChars := 0
	totalLines := 0
	uniqueWords := make(map[string]struct{})
	wordLengths := make(map[int]int)

	for i, track := range album.Tracks {
		wordCounts, vowels, consonants, lengths := words.CalculateAndSortWordFrequencies(track.Lyrics)
		words := len(strings.Fields(track.Lyrics))

		totalWords += words
		totalVowels += vowels
		totalConsonants += consonants
		totalChars += len(track.Lyrics)
		totalLines += len(strings.Split(track.Lyrics, "\n"))

		for l, c := range lengths {
			wordLengths[l] += c
		}
		for _, wc := range wordCounts {
			uniqueWords[wc.Word] = struct{}{}
		}

		track.TotalWords = words
		track.TotalCharacters = len(track.Lyrics)
		track.TotalCharactersNoSpaces = len(strings.ReplaceAll(strings.ReplaceAll(track.Lyrics, " ", ""), "\n", ""))
		track.TotalLines = len(strings.Split(track.Lyrics, "\n"))
		album.Tracks[i] = track
	}

	album.TotalWords = totalWords
	album.TotalCharacters = totalChars
	album.TotalLines = totalLines
	album.TotalVowelCount = totalVowels
	album.TotalConsonantCount = totalConsonants
	album.TotalUniqueWords = len(uniqueWords)
	album.WordLengthDistribution = wordLengths
	album.AverageWordsPerTrack = calculateAverage(totalWords, len(album.Tracks))
}

func calculateAverage(total, count int) int {
	if count > 0 {
		return total / count
	}
	return 0
}

func ValidateAuthKey(key string) (bool, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return false, fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM auth_keys WHERE key = ?)", key).Scan(&exists)
	if err != nil {
		log.Printf("Error checking auth key: %v", err)
		return false, fmt.Errorf("error checking auth key: %w", err)
	}

	return exists, nil
}

func UpdateTrackLyrics(req models.UpdateLyricsRequest) error {
	valid, err := ValidateAuthKey(req.AuthKey)
	if err != nil {
		return fmt.Errorf("error validating auth: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid auth key")
	}

	cleanLyrics := strings.TrimSpace(req.Lyrics)
	if strings.HasPrefix(strings.ToLower(cleanLyrics), "lyrics") {
		cleanLyrics = ""
	}

	log.Printf("Updating lyrics for album: %s, track: %s", req.AlbumID, req.TrackName)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	result, err := db.Exec("UPDATE tracks SET lyrics = ? WHERE album_id = ? AND name = ?",
		cleanLyrics, req.AlbumID, req.TrackName)
	if err != nil {
		log.Printf("Error executing update: %v", err)
		return fmt.Errorf("error updating lyrics: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error checking rows affected: %v", err)
		return fmt.Errorf("error checking update result: %w", err)
	}
	if rows == 0 {
		log.Printf("No track found for album: %s, track: %s", req.AlbumID, req.TrackName)
		return fmt.Errorf("no matching track found")
	}

	log.Printf("Successfully updated lyrics for album: %s, track: %s", req.AlbumID, req.TrackName)
	return nil
}

func GetAlbumBySlug(slug string) (models.BandcampAlbumData, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return models.BandcampAlbumData{}, fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	var album models.BandcampAlbumData
	var ampwallURL, metalArchivesURL sql.NullString

	err = db.QueryRow(`
			SELECT id, slug, artist_name, album_name, image_url, image_data, 
						 album_color_average, bandcamp_url, ampwall_url, metal_archives_url,
						 total_length, formatted_length, date_added 
			FROM albums WHERE slug = ?`, slug).Scan(
		&album.ID,
		&album.Slug,
		&album.ArtistName,
		&album.AlbumName,
		&album.ImageUrl,
		&album.ImageData,
		&album.AlbumColorAverage,
		&album.BandcampUrl,
		&ampwallURL,
		&metalArchivesURL,
		&album.TotalLength,
		&album.FormattedLength,
		&album.DateAdded)
	if err != nil {
		return models.BandcampAlbumData{}, fmt.Errorf("error fetching album: %w", err)
	}

	album.AmpwallUrl = ampwallURL.String
	album.MetalArchivesURL = metalArchivesURL.String
	album.ImageDataBase64 = base64.StdEncoding.EncodeToString(album.ImageData)

	if err := fetchTracks(db, &album); err != nil {
		return models.BandcampAlbumData{}, fmt.Errorf("error fetching tracks: %w", err)
	}

	calculateAlbumMetrics(&album)
	return album, nil
}
