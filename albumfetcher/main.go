package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"millions-of-words/models"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	textfileFlag := flag.String("textfile", "", "Path to a text file containing Bandcamp URLs (one per line)")
	flag.Parse()

	if *textfileFlag != "" {
		processTextFile(*textfileFlag)
	} else if len(os.Args) >= 2 {
		url := os.Args[1]
		processSingleURL(url)
	} else {
		log.Fatal("Usage: go run main.go <bandcamp_url> OR go run main.go --textfile=<file_with_links.txt>")
	}
}

func processTextFile(filePath string) {
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
		processSingleURL(url)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}
}

func processSingleURL(url string) {
	fmt.Println("=========================================")
	fmt.Printf("Processing album data for URL: %s\n", url)

	filename := filepath.Join("../data", fmt.Sprintf("%s - %s.json", sanitizeFilenameFromURL(url)))

	if _, err := os.Stat(filename); err == nil {
		fmt.Printf("File already exists for URL: %s. Skipping download.\n", url)
		fmt.Println("=========================================")
		return
	}

	albumData, err := fetchAlbumDataFromBandcamp(url)
	if err != nil {
		log.Printf("Failed to fetch album data for URL %s: %v", url, err)
		fmt.Println("=========================================")
		return
	}

	err = writeAlbumsDataToJsonFile(albumData)
	if err != nil {
		log.Printf("Failed to write album data to JSON for URL %s: %v", url, err)
	}

	fmt.Println("Finished processing.")
	fmt.Println("=========================================")
}

func fetchAlbumDataFromBandcamp(url string) (models.BandcampAlbumData, error) {
	fmt.Printf("Fetching album data from Bandcamp for URL: %s\n", url)
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
	description := doc.Find(".tralbumData.tralbum-about").Text()
	imageUrl := doc.Find("a.popupImage").AttrOr("href", "")

	var tracklist []models.BandcampTrackData
	var totalAlbumDuration time.Duration

	doc.Find("tr.track_row_view").Each(func(i int, s *goquery.Selection) {
		trackTitle := s.Find(".title-col .track-title").Text()
		trackDurationStr := strings.TrimSpace(s.Find(".title-col .time").Text())

		fmt.Printf("  - Processing track: '%s', Duration: '%s'\n", trackTitle, trackDurationStr)

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
		ID:              strings.TrimSpace(artistName) + " - " + strings.TrimSpace(albumName),
		ArtistName:      strings.TrimSpace(artistName),
		AlbumName:       strings.TrimSpace(albumName),
		Description:     strings.TrimSpace(description),
		ImageUrl:        imageUrl,
		Tracks:          tracklist,
		TotalLength:     int(totalAlbumDuration.Seconds()),
		FormattedLength: formatDuration(int(totalAlbumDuration.Seconds())),
		BandcampUrl:     url,
	}, nil
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

func writeAlbumsDataToJsonFile(album models.BandcampAlbumData) error {
	dir := "../data"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}
	}

	filename := filepath.Join(dir, fmt.Sprintf("%s - %s.json", sanitizeFilename(album.ArtistName), sanitizeFilename(album.AlbumName)))
	file, err := json.MarshalIndent(album, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling album data to JSON: %w", err)
	}

	err = ioutil.WriteFile(filename, file, 0644)
	if err != nil {
		return fmt.Errorf("error writing album data to file: %w", err)
	}
	return nil
}

func sanitizeFilename(name string) string {
	trimmedName := strings.TrimSpace(name)
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	sanitized := reg.ReplaceAllString(trimmedName, "_")

	return strings.Trim(sanitized, "_")
}

func sanitizeFilenameFromURL(url string) string {
	parts := strings.Split(url, "/")
	artist := parts[2]
	album := parts[len(parts)-1]
	return sanitizeFilename(artist + " - " + album)
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
