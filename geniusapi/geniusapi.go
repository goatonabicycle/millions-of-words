package geniusapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type GeniusAPI struct {
	AccessToken string
}

func NewGeniusAPI(accessToken string) *GeniusAPI {
	return &GeniusAPI{AccessToken: accessToken}
}

type SearchArtistsResponse struct {
	Response struct {
		Hits []struct {
			Type string `json:"type"`
			Result struct {
				PrimaryArtist struct {
					ID   int    `json:"id"`
					Name string `json:"name"`
				} `json:"primary_artist"`
			} `json:"result"`
		} `json:"hits"`
	} `json:"response"`
}

type SongsByArtistResponse struct {
	Response struct {
		Songs []struct {
			ID     int    `json:"id"`
			Title  string `json:"title"`
			Artist struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"primary_artist"`
		} `json:"songs"`
	} `json:"response"`
}

func (api *GeniusAPI) makeRequest(ctx context.Context, endpoint string) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.genius.com/"+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+api.AccessToken)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (api *GeniusAPI) SearchArtists(ctx context.Context, artistName string) (int, error) {
	endpoint := fmt.Sprintf("search?q=%s", strings.ReplaceAll(artistName, " ", "%20"))
	body, err := api.makeRequest(ctx, endpoint)
	if err != nil {
		return 0, err
	}

	var response SearchArtistsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, err
	}

	for _, hit := range response.Response.Hits {
		if hit.Type == "song" && strings.ToLower(hit.Result.PrimaryArtist.Name) == strings.ToLower(artistName) {
			return hit.Result.PrimaryArtist.ID, nil
		}
	}

	return 0, fmt.Errorf("artist %s not found", artistName)
}

func (api *GeniusAPI) GetSongsByArtist(ctx context.Context, artistID int) ([]string, error) {
    var allSongs []string
    page := 1
    perPage := 10

    for {
        fmt.Printf("Fetching page %d for artist ID %d...\n", page, artistID)
        endpoint := fmt.Sprintf("artists/%d/songs?sort=title&per_page=%d&page=%d", artistID, perPage, page)
        fmt.Printf("Constructed endpoint: %s\n", endpoint)
        
        body, err := api.makeRequest(ctx, endpoint)
        if err != nil {
            fmt.Printf("Error making request to endpoint: %s, error: %v\n", endpoint, err)
            return nil, fmt.Errorf("failed to make request: %w", err)
        }
        
        if len(body) == 0 {
            fmt.Println("Received empty body from API request. Possible issue with request or API.")
        } 

        var response SongsByArtistResponse
        err = json.Unmarshal(body, &response)
        if err != nil {
            fmt.Printf("Failed to unmarshal JSON response: %v\n", err)
            return nil, fmt.Errorf("failed to unmarshal response: %w", err)
        }

        fetchedSongsCount := len(response.Response.Songs)
        if fetchedSongsCount == 0 {
            fmt.Printf("No songs were returned for page %d. Ending pagination.\n", page)
            break
        }

        for _, song := range response.Response.Songs {
            allSongs = append(allSongs, song.Title)
        }

        fmt.Printf("Fetched %d songs from page %d. Total songs collected so far: %d.\n", fetchedSongsCount, page, len(allSongs))

        page++ 
        fmt.Printf("Moving to page %d.\n", page)

		// Let's not kill the API.
		if len(allSongs) > 30 {
			break
		}
    }

    fmt.Printf("Finished fetching songs. Total number of songs fetched: %d\n", len(allSongs))
    return allSongs, nil
}

