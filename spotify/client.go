package spotifyclient

import (
	"context"
	"net/http"

	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

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
