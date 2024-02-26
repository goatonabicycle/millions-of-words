package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"millions-of-words/app"
	"millions-of-words/config"
	spotifyclient "millions-of-words/spotify"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <artist name>")
	}
	artistName := os.Args[1]

	config.LoadEnv()
	clientID, clientSecret := config.GetSpotifyCredentials()
	client := spotifyclient.NewSpotifyClient(clientID, clientSecret)

	fmt.Printf("Fetching albums for %s...\n", artistName)
	albumsData := app.FetchAlbumsData(client, artistName)
	jsonData, err := json.MarshalIndent(albumsData, "", "    ")
	if err != nil {
		log.Fatalf("Error marshaling data to JSON: %v", err)
	}

	artistDir := filepath.Join("data", artistName)
	if err := os.MkdirAll(artistDir, os.ModePerm); err != nil {
		log.Fatalf("Error creating directory: %v", err)
	}

	fileName := fmt.Sprintf("%s_albums_data.json", artistName)
	filePath := filepath.Join(artistDir, fileName)
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		log.Fatalf("Error writing JSON data to file: %v", err)
	}

	fmt.Printf("Album data for %s successfully written to %s\n", artistName, filePath)
}
