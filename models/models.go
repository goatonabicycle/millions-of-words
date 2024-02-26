package models

type AlbumData struct {
	Name        string      `json:"name"`
	ReleaseYear string      `json:"release_year"`
	Artists     []string    `json:"artists"`
	AlbumType   string      `json:"album_type"`
	Tracks      []TrackData `json:"tracks"`
}

type TrackData struct {
	Name   string `json:"name"`
	Length string `json:"length"`
}
