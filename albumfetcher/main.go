package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"millions-of-words/models"

	"github.com/PuerkitoBio/goquery"
	_ "modernc.org/sqlite"
)

func main() {
	dbPath := "../data/db/albums.db"
	urlFilePath := "bandcamp_urls.txt"

	ensureDatabaseExists(dbPath)
	processTextFile(urlFilePath, dbPath)
}

func ensureDatabaseExists(dbPath string) {
	dbDir := "../data/db"
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			log.Fatalf("Failed to create directory for SQLite database: %v", err)
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	if err := createTablesIfNotExist(db); err != nil {
		log.Fatalf("Error creating tables in database: %v", err)
	}
}

func processTextFile(filePath, dbPath string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url == "" {
			continue
		}
		processSingleURL(url, dbPath)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}
}

func processSingleURL(url, dbPath string) {
	log.Println("=========================================")
	log.Printf("Processing album data for URL: %s\n", url)

	if urlExistsInDatabase(url, dbPath) {
		log.Printf("Album with URL %s already exists in the database. Skipping.\n", url)
		log.Println("=========================================")
		return
	}

	albumData, err := fetchAlbumDataFromBandcamp(url)
	if err != nil {
		log.Printf("Failed to fetch album data for URL %s: %v", url, err)
		log.Println("=========================================")
		return
	}

	albumData.DateAdded = time.Now().Format("2006-01-02 15:04:05")

	albumData.AlbumColorAverage, err = calculateAverageColor(albumData.ImageData)
	if err != nil {
		log.Printf("Failed to calculate average color for URL %s: %v", url, err)
		albumData.AlbumColorAverage = "#000000"
	}

	err = insertAlbumDataIntoSQLite(albumData, dbPath)
	if err != nil {
		log.Printf("Failed to write album data to SQLite for URL %s: %v", url, err)
	}

	log.Println("Finished processing.")
	log.Println("=========================================")
}

func urlExistsInDatabase(url, dbPath string) bool {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM albums WHERE bandcamp_url=? LIMIT 1)"
	err = db.QueryRow(query, url).Scan(&exists)
	if err != nil {
		log.Fatalf("Error checking if URL exists in database: %v", err)
	}
	return exists
}

func fetchAlbumDataFromBandcamp(url string) (models.BandcampAlbumData, error) {
	log.Printf("Fetching album data from Bandcamp for URL: %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return models.BandcampAlbumData{}, fmt.Errorf("error fetching Bandcamp page: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return models.BandcampAlbumData{}, fmt.Errorf("error parsing HTML: %w", err)
	}

	artistName := doc.Find("#name-section h3 span a").Text()
	albumName := doc.Find(".trackTitle").First().Text()
	imageUrl := doc.Find("a.popupImage").AttrOr("href", "")

	imageData, err := fetchImageData(imageUrl)
	if err != nil {
		log.Printf("Failed to fetch album image for URL %s: %v", url, err)
		imageData = nil
	}

	var tracklist []models.BandcampTrackData
	var totalAlbumDuration time.Duration

	doc.Find("tr.track_row_view").Each(func(i int, s *goquery.Selection) {
		trackTitle := s.Find(".title-col .track-title").Text()
		trackDurationStr := strings.TrimSpace(s.Find(".title-col .time").Text())

		log.Printf("  - Processing track: '%s', Duration: '%s'\n", trackTitle, trackDurationStr)

		if strings.TrimSpace(trackTitle) == "" || trackDurationStr == "" {
			log.Printf("Warning: Missing essential data for track %d (%s). Skipping this track.", i+1, trackTitle)
			return
		}

		trackDuration, err := parseTrackDuration(trackDurationStr)
		if err != nil {
			log.Printf("Error parsing track duration for track %s: %v", trackTitle, err)
			return
		}

		totalAlbumDuration += trackDuration

		lyrics := strings.TrimSpace(s.Next().Find("div").Text())
		if strings.HasPrefix(lyrics, "lyrics") || strings.Contains(lyrics, "buy track") {
			lyrics = ""
		}

		track := models.BandcampTrackData{
			Name:            strings.TrimSpace(trackTitle),
			TotalLength:     int(trackDuration.Seconds()),
			FormattedLength: formatDuration(int(trackDuration.Seconds())),
			Lyrics:          lyrics,
		}

		tracklist = append(tracklist, track)
	})

	return models.BandcampAlbumData{
		ID:                strings.TrimSpace(artistName) + " - " + strings.TrimSpace(albumName),
		ArtistName:        strings.TrimSpace(artistName),
		AlbumName:         strings.TrimSpace(albumName),
		ImageUrl:          imageUrl,
		ImageData:         imageData,
		AlbumColorAverage: "",
		Tracks:            tracklist,
		TotalLength:       int(totalAlbumDuration.Seconds()),
		FormattedLength:   formatDuration(int(totalAlbumDuration.Seconds())),
		BandcampUrl:       url,
		AmpwallUrl:        "",
	}, nil
}

func fetchImageData(imageUrl string) ([]byte, error) {
	if imageUrl == "" {
		return nil, fmt.Errorf("no image URL provided")
	}

	resp, err := http.Get(imageUrl)
	if err != nil {
		return nil, fmt.Errorf("error fetching image: %w", err)
	}
	defer resp.Body.Close()

	imageData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading image data: %w", err)
	}

	return imageData, nil
}

func calculateAverageColor(imageData []byte) (string, error) {
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return "", fmt.Errorf("error decoding image: %w", err)
	}

	var rSum, gSum, bSum, pixelCount uint64

	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			rSum += uint64(r)
			gSum += uint64(g)
			bSum += uint64(b)
			pixelCount++
		}
	}

	avgR := uint8(rSum / pixelCount / 256)
	avgG := uint8(gSum / pixelCount / 256)
	avgB := uint8(bSum / pixelCount / 256)

	return fmt.Sprintf("#%02x%02x%02x", avgR, avgG, avgB), nil
}

func parseTrackDuration(durationStr string) (time.Duration, error) {
	parts := strings.Split(durationStr, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid track duration format: %s", durationStr)
	}

	minutes := strings.TrimSpace(parts[0])
	seconds := strings.TrimSpace(parts[1])

	return time.ParseDuration(fmt.Sprintf("%sm%ss", minutes, seconds))
}

func insertAlbumDataIntoSQLite(album models.BandcampAlbumData, dbPath string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	if err := createTablesIfNotExist(db); err != nil {
		return fmt.Errorf("error creating tables: %w", err)
	}

	_, err = db.Exec(`INSERT INTO albums 
		(id, artist_name, album_name, image_url, image_data, bandcamp_url, ampwall_url, album_color_average, total_length, formatted_length, date_added) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		album.ID, album.ArtistName, album.AlbumName, album.ImageUrl, album.ImageData, album.BandcampUrl, album.AmpwallUrl, album.AlbumColorAverage, album.TotalLength, album.FormattedLength, album.DateAdded)
	if err != nil {
		return fmt.Errorf("error inserting album data: %w", err)
	}

	for _, track := range album.Tracks {
		_, err = db.Exec(`INSERT INTO tracks (album_id, name, total_length, formatted_length, lyrics) VALUES (?, ?, ?, ?, ?)`,
			album.ID, track.Name, track.TotalLength, track.FormattedLength, track.Lyrics)
		if err != nil {
			return fmt.Errorf("error inserting track data: %w", err)
		}
	}

	return nil
}

func createTablesIfNotExist(db *sql.DB) error {
	albumTable := `
	CREATE TABLE IF NOT EXISTS albums (
		id TEXT PRIMARY KEY,
		artist_name TEXT,
		album_name TEXT,
		image_url TEXT,
		image_data BLOB,
		bandcamp_url TEXT,
		ampwall_url TEXT,
		album_color_average TEXT,
		total_length INTEGER,
		formatted_length TEXT,
		date_added DATETIME
	);
	`
	trackTable := `
	CREATE TABLE IF NOT EXISTS tracks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		album_id TEXT,
		name TEXT,
		total_length INTEGER,
		formatted_length TEXT,
		lyrics TEXT,
		FOREIGN KEY(album_id) REFERENCES albums(id)
	);
	`
	_, err := db.Exec(albumTable)
	if err != nil {
		return err
	}

	_, err = db.Exec(trackTable)
	if err != nil {
		return err
	}

	return nil
}

func formatDuration(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	seconds = seconds % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
