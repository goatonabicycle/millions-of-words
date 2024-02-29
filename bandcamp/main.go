package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

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
	files.SaveAlbumDataToFile("bandcamp_Woe", albumData)

}

func fetchAlbumDataFromBandcamp(url string) models.BandcampAlbumData {
	fmt.Printf("Fetching album data from Bandcamp for URL: %s\n", url)

	album := models.BandcampAlbumData{
		// There's lots more to scrape here.
		// Artist name, album name, whatever else could be cool.
	}

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error fetching Bandcamp page: %v", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalf("Error parsing HTML: %v", err)
	}

	doc.Find("tr.lyricsRow").Each(func(i int, s *goquery.Selection) {
		lyrics := s.Find("div").Text()
		track := models.BandcampTrackData{
			Name:             fmt.Sprintf("Track %d", i+1), // Do this properly!
			Lyrics:           lyrics,
			SortedWordCounts: words.CalculateAndSortWordFrequencies(lyrics),
		}

		album.Tracks = append(album.Tracks, track)
	})

	return album
}
