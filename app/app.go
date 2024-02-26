package app

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"millions-of-words/models"

	"github.com/PuerkitoBio/goquery"
	"github.com/zmb3/spotify"
)

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
	fmt.Println("Artist found, fetching albums...")

	artistID := results.Artists.Artists[0].ID
	albums, err := client.GetArtistAlbums(artistID)
	if err != nil {
		log.Fatalf("Error getting albums: %v", err)
	}

	for _, simpleAlbum := range albums.Albums {
		fmt.Printf("Fetching details for album: %s\n", simpleAlbum.Name)
		fullAlbum, err := client.GetAlbum(simpleAlbum.ID)
		if err != nil {
			log.Fatalf("Error getting album details: %v", err)
		}

		var album models.AlbumData
		album.Name = fullAlbum.Name
		album.ReleaseYear = fullAlbum.ReleaseDate[0:4]
		album.AlbumType = fullAlbum.AlbumType

		for _, artist := range fullAlbum.Artists {
			album.Artists = append(album.Artists, artist.Name)
		}

		for _, track := range fullAlbum.Tracks.Tracks {
			durationSeconds := track.Duration / 1000
			lyricsURL, err := fetchSongURLFromGenius(artistName, track.Name)
			if err != nil {
				fmt.Printf("Fetching lyrics for '%s' - %s: Not found or error occurred\n", fullAlbum.Name, track.Name)
			} else {
				fmt.Printf("Fetching lyrics for '%s' - %s: Saved\n", fullAlbum.Name, track.Name)
			}

			trackData := models.TrackData{
				Name:   track.Name,
				Length: fmt.Sprintf("%d:%02d", durationSeconds/60, durationSeconds%60),
			}
			if err == nil && lyricsURL != "" {
				trackData.LyricsURL = lyricsURL
			}

			lyrics, err := ScrapeLyrics(lyricsURL)
			if err != nil {
				fmt.Printf("Error scraping lyrics for track %s: %v\n", track.Name, err)
				continue
			}
			if err == nil && lyrics != "" {
				trackData.Lyrics = lyrics
			}

			album.Tracks = append(album.Tracks, trackData)
		}

		albumsData = append(albumsData, album)
	}

	return albumsData
}

func fetchSongURLFromGenius(artistName, songTitle string) (string, error) {
	geniusAPIKey := os.Getenv("GENIUS_API_KEY")
	if geniusAPIKey == "" {
		return "", fmt.Errorf("GENIUS_API_KEY must be set")
	}

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
					URL string `json:"url"`
				} `json:"result"`
			} `json:"hits"`
		} `json:"response"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Response.Hits) > 0 {
		return result.Response.Hits[0].Result.URL, nil
	}

	return "", fmt.Errorf("no lyrics found")
}

func ScrapeLyrics(url string) (string, error) {
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
	doc.Find(".lyrics").Each(func(i int, s *goquery.Selection) {
		lyrics = s.Text()
	})

	if lyrics == "" {
		doc.Find("div[class^=\"Lyrics__Container\"]").Each(func(i int, s *goquery.Selection) {
			lyrics += s.Text() + "\n\n"
		})
	}

	if lyrics == "" {
		return "", fmt.Errorf("lyrics not found")
	}

	return lyrics,
		nil
}
