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
	Tracks               []BandcampTrackData `json:"tracks"`
	AlbumWordFrequencies []WordCount
}

type BandcampTrackData struct {
	Name             string      `json:"name"`
	Length           string      `json:"length"`
	Lyrics           string      `json:"lyrics,omitempty"`
	SortedWordCounts []WordCount `json:"sorted_word_counts,omitempty"`
}
