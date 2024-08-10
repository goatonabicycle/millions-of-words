package models

type WordCount struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

type BandcampAlbumData struct {
	ID                   string              `json:"id"`
	ArtistName           string              `json:"artist_name"`
	AlbumName            string              `json:"album_name"`
	Description          string              `json:"description"`
	ImageUrl             string              `json:"image_url"`
	BandcampUrl          string              `json:"bandcamp_url"`
	Tracks               []BandcampTrackData `json:"tracks"`
	TotalAlbumLength     string              `json:"total_album_length"`
	AlbumWordFrequencies []WordCount         `json:"-"`
	TotalWords           int                 `json:"-"`
	AverageWordsPerTrack int                 `json:"-"`
	TotalUniqueWords     int                 `json:"-"`
}

type BandcampTrackData struct {
	Name     string `json:"name"`
	Duration string `json:"duration"`
	Lyrics   string `json:"lyrics,omitempty"`
}
