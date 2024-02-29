package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"millions-of-words/config"
	"millions-of-words/models"
	"millions-of-words/words"

	"github.com/PuerkitoBio/goquery"
	"github.com/zmb3/spotify"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run spotify-and-genius/main.go <artist name>")
	}

	artistName := os.Args[1]

	config.LoadEnv()
	clientID, clientSecret := config.GetSpotifyCredentials()
	client := NewSpotifyClient(clientID, clientSecret)

	fetchAlbumsForArtist(client, artistName)
}

func fetchAlbumsForArtist(client spotify.Client, artistName string) {
	fmt.Printf("Fetching albums for %s from Spotify...\n", artistName)
	albumsData := FetchAlbumsData(client, artistName)
	jsonData, err := json.MarshalIndent(albumsData, "", "    ")
	if err != nil {
		log.Fatalf("Error marshaling data to JSON: %v", err)
	}

	saveDataToFile(artistName, jsonData)
}

func saveDataToFile(artistName string, jsonData []byte) {
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

func NewSpotifyClient(clientID, clientSecret string) spotify.Client {
	config := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     spotify.TokenURL,
	}
	httpClient := &http.Client{
		Transport: &oauth2.Transport{
			Source: config.TokenSource(context.Background()),
		},
	}
	client := spotify.NewClient(httpClient)
	return client
}

func FetchAlbumsData(client spotify.Client, artistName string) []models.AlbumData {
	var albumsData []models.AlbumData

	fmt.Println("Searching for artist...")
	results, err := client.Search(artistName, spotify.SearchTypeArtist)
	if err != nil {
		log.Fatalf("Error searching for artist: %v", err)
	}

	if len(results.Artists.Artists) == 0 {
		log.Fatal("Artist not found")
	}

	artistID := results.Artists.Artists[0].ID
	albums, err := client.GetArtistAlbums(artistID)
	if err != nil {
		log.Fatalf("Error getting albums: %v", err)
	}

	for _, simpleAlbum := range albums.Albums {
		// I'll care about other types later.
		if simpleAlbum.AlbumType == "album" {
			fmt.Printf("Fetching details for album: %s\n", simpleAlbum.Name)
			fullAlbum, err := client.GetAlbum(simpleAlbum.ID)
			if err != nil {
				log.Fatalf("Error getting album details: %v", err)
				continue
			}

			var album models.AlbumData
			album.Name = fullAlbum.Name
			album.ReleaseYear = fullAlbum.ReleaseDate[0:4]
			album.AlbumType = fullAlbum.AlbumType

			for _, artist := range fullAlbum.Artists {
				album.Artists = append(album.Artists, artist.Name)
			}

			albumWordCounts := make(map[string]int)
			for _, track := range fullAlbum.Tracks.Tracks {
				durationSeconds := track.Duration / 1000
				lyricsURL, err := fetchSongURLFromGenius(artistName, track.Name)
				if err != nil {
					fmt.Printf("Error fetching lyrics URL for '%s' - '%s': %v\n", fullAlbum.Name, track.Name, err)
					continue
				}

				trackData := models.TrackData{
					Name:      track.Name,
					Length:    fmt.Sprintf("%d:%02d", durationSeconds/60, durationSeconds%60),
					LyricsURL: lyricsURL,
				}

				lyrics, err := ScrapeLyricsFromGenius(lyricsURL)
				if err != nil {
					fmt.Printf("Error scraping lyrics for track '%s': %v\n", track.Name, err)
				} else {
					trackData.Lyrics = lyrics
					trackWordCounts := words.CalculateAndSortWordFrequencies(lyrics)
					trackData.SortedWordCounts = trackWordCounts

					for _, wCount := range trackWordCounts {
						albumWordCounts[wCount.Word] += wCount.Count
					}
				}

				album.Tracks = append(album.Tracks, trackData)
			}

			// Sort the album-level word counts
			album.SortedWordCounts = words.CalculateAndSortWordFrequencies(words.MapToString(albumWordCounts))
			albumsData = append(albumsData, album)
		}
	}

	return albumsData
}

func fetchSongURLFromGenius(artistName, songTitle string) (string, error) {
	geniusAPIKey := config.GetGeniusKey()

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.genius.com/search", nil)
	if err != nil {
		return "", err
	}

	query := req.URL.Query()
	query.Add("q", artistName+" "+songTitle)
	req.URL.RawQuery = query.Encode()

	req.Header.Add("Authorization", "Bearer "+geniusAPIKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		Response struct {
			Hits []struct {
				Result struct {
					URL           string `json:"url"`
					PrimaryArtist struct {
						Name string `json:"name"`
					} `json:"primary_artist"`
					Title string `json:"title"`
				} `json:"result"`
			} `json:"hits"`
		} `json:"response"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	for _, hit := range result.Response.Hits {
		cleanArtistName := strings.ToLower(strings.TrimSpace(artistName))
		cleanSongTitle := strings.ToLower(strings.TrimSpace(songTitle))
		hitArtistName := strings.ToLower(strings.TrimSpace(hit.Result.PrimaryArtist.Name))
		hitSongTitle := strings.ToLower(strings.TrimSpace(hit.Result.Title))

		if strings.Contains(hitArtistName, cleanArtistName) && strings.Contains(hitSongTitle, cleanSongTitle) {
			return hit.Result.URL, nil
		}
	}

	return "", fmt.Errorf("no accurate match found for artist: %s, song: %s", artistName, songTitle)
}

func ScrapeLyricsFromGenius(url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("URL is empty")
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error fetching song page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error parsing HTML: %w", err)
	}

	var lyrics string
	lyricsContainer := doc.Find("div[class^='Lyrics__Container'], .lyrics")
	lyricsContainer.Each(func(i int, s *goquery.Selection) {
		s.Contents().Each(func(index int, item *goquery.Selection) {
			if item.Is("br") {
				lyrics += "\n"
			} else if goquery.NodeName(item) == "#text" {
				text := strings.TrimSpace(item.Text())
				if text != "" {
					lyrics += text + "\n"
				}
			}
		})
		lyrics += "\n"
	})

	if lyrics == "" {
		return "", fmt.Errorf("lyrics not found")
	}

	return strings.TrimSpace(lyrics), nil
}
