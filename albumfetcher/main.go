package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"millions-of-words/models"
	"millions-of-words/words"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run albumfetcher.go <bandcamp_url>")
	}
	url := os.Args[1]

	albumData, err := fetchAlbumDataFromBandcamp(url)
	if err != nil {
		log.Fatalf("Failed to fetch album data: %v", err)
	}
	err = writeAlbumsDataToJsonFile(albumData)
	if err != nil {
		log.Fatalf("Failed to write album data to JSON: %v", err)
	}
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
	doc.Find("tr.lyricsRow").Each(func(i int, s *goquery.Selection) {
		lyrics := s.Find("div").Text()
		trackTitle := doc.Find(".title-col .track-title").Eq(i).Text() // Matching track titles with lyrics

		track := models.BandcampTrackData{
			Name:             strings.TrimSpace(trackTitle),
			Lyrics:           lyrics,
			SortedWordCounts: words.CalculateAndSortWordFrequencies(lyrics),
		}

		tracklist = append(tracklist, track)
	})

	return models.BandcampAlbumData{
		ArtistName:  strings.TrimSpace(artistName),
		AlbumName:   strings.TrimSpace(albumName),
		Description: strings.TrimSpace(description),
		ImageUrl:    imageUrl,
		Tracks:      tracklist,
	}, nil
}

func writeAlbumsDataToJsonFile(album models.BandcampAlbumData) error {
	dir := "data"
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
