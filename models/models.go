package models

type WordCount struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

type TrackData struct {
	Name             string      `json:"name"`
	Length           string      `json:"length"`
	LyricsURL        string      `json:"lyrics_url,omitempty"`
	Lyrics           string      `json:"lyrics,omitempty"`
	SortedWordCounts []WordCount `json:"sorted_word_counts,omitempty"`
}

type AlbumData struct {
	Name             string      `json:"name"`
	ReleaseYear      string      `json:"release_year"`
	Artists          []string    `json:"artists"`
	AlbumType        string      `json:"album_type"`
	Tracks           []TrackData `json:"tracks"`
	SortedWordCounts []WordCount `json:"sorted_album_word_counts,omitempty"`
}
