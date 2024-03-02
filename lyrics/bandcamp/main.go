package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"millions-of-words/files"
	"millions-of-words/models"
	"millions-of-words/words"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run bandcamp/main.go <Bandcamp URL>")
	}

	bandcampURL := os.Args[1]
	albumData := fetchAlbumDataFromBandcamp(bandcampURL)
	files.SaveAlbumDataToFile(albumData.ArtistName+"_bandcamp_album", albumData) // Consider a dynamic name based on the album
}

func fetchAlbumDataFromBandcamp(url string) models.BandcampAlbumData {
	fmt.Printf("Fetching album data from Bandcamp for URL: %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error fetching Bandcamp page: %v", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalf("Error parsing HTML: %v", err)
	}

	artistName := doc.Find("#name-section h3 span a").Text()
	albumName := doc.Find(".trackTitle").First().Text()
	description := doc.Find(".tralbumData.tralbum-about").Text()
	imageUrl := doc.Find("a.popupImage").AttrOr("href", "")

	tags := make([]string, 0)
	doc.Find(".tralbumData.tralbum-tags a.tag").Each(func(i int, s *goquery.Selection) {
		tags = append(tags, s.Text())
	})

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
		Tags:        tags,
		Tracks:      tracklist,
	}
}
