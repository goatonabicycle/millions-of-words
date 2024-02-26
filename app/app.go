package app

import (
	"fmt"
	"log"

	"millions-of-words/models"

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
			trackData := models.TrackData{
				Name:   track.Name,
				Length: fmt.Sprintf("%d:%02d", durationSeconds/60, durationSeconds%60),
			}
			album.Tracks = append(album.Tracks, trackData)
		}

		albumsData = append(albumsData, album)
	}

	return albumsData
}
