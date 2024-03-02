package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
}

func GetSpotifyCredentials() (string, string) {
	clientId := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	if clientId == "" || clientSecret == "" {
		log.Fatal("SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET must be set")
	}
	return clientId, clientSecret
}

func GetGeniusKey() string {
	geniusKey := os.Getenv("GENIUS_API_KEY")
	if geniusKey == "" {
		log.Fatal("GENIUS_API_KEY must be set")
	}
	return geniusKey
}
