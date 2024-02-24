package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"millions-of-words/geniusapi"

	"github.com/joho/godotenv"
)

func main() {	
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("GENIUS_API_KEY")
	if apiKey == "" {
		log.Fatal("GENIUS_API_KEY must be set")
	}
	
	api := geniusapi.NewGeniusAPI(apiKey)

	ctx := context.Background()
	artistName := "Aesop Rock" 
	
	artistID, err := api.SearchArtists(ctx, artistName)
	if err != nil {
		log.Fatalf("Failed to search for artist %s: %v", artistName, err)
	}
	
	songs, err := api.GetSongsByArtist(ctx, artistID)
	if err != nil {
		log.Fatalf("Failed to get songs for artist ID %d: %v", artistID, err)
	}

	fmt.Printf("Songs by %s:\n", artistName)
	for _, song := range songs {
		fmt.Println(song)
	}
}
