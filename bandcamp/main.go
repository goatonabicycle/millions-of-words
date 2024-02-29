package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run bandcamp/main.go <Bandcamp URL>")
	}

	bandcampURL := os.Args[1]

	fetchLyricsFromBandcamp(bandcampURL)
}

func fetchLyricsFromBandcamp(bandcampURL string) {
	fmt.Printf("Fetching lyrics from Bandcamp for URL: %s\n", bandcampURL)
}
