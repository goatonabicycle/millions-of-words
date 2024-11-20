package models

import "html/template"

type WordCount struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

type BandcampAlbumData struct {
	ID                      string              `json:"id"`
	ArtistName              string              `json:"artist_name"`
	AlbumName               string              `json:"album_name"`
	ImageUrl                string              `json:"image_url"`
	ImageData               []byte              `json:"-"`
	ImageDataBase64         string              `json:"-"`
	BandcampUrl             string              `json:"bandcamp_url"`
	AmpwallUrl              string              `json:"ampwall_url"`
	AlbumColorAverage       string              `json:"album_color_average"`
	DateAdded               string              `json:"date_added"`
	Tracks                  []BandcampTrackData `json:"tracks"`
	TotalLength             int                 `json:"total_length"`
	FormattedLength         string              `json:"formatted_length"`
	AlbumWordFrequencies    []WordCount         `json:"-"`
	TotalWords              int                 `json:"-"`
	AverageWordsPerTrack    int                 `json:"-"`
	TotalUniqueWords        int                 `json:"-"`
	TotalVowelCount         int                 `json:"-"`
	TotalConsonantCount     int                 `json:"-"`
	WordLengthDistribution  map[int]int         `json:"-"`
	TotalCharacters         int                 `json:"-"`
	TotalCharactersNoSpaces int                 `json:"-"`
	TotalLines              int                 `json:"-"`
}

type BandcampTrackData struct {
	Name                    string `json:"name"`
	TotalLength             int    `json:"total_length"`
	FormattedLength         string `json:"formatted_length"`
	Lyrics                  string `json:"lyrics,omitempty"`
	TotalCharacters         int    `json:"-"`
	TotalCharactersNoSpaces int    `json:"-"`
	TotalLines              int    `json:"-"`
}

type TrackWithDetails struct {
	Track                   BandcampTrackData
	FormattedLyrics         template.HTML
	SortedWordCounts        []WordCount
	WordsPerMinute          float64
	TotalWords              int
	UniqueWords             int
	VowelCount              int
	ConsonantCount          int
	WordLengthDistribution  map[int]int
	TotalCharacters         int
	TotalCharactersNoSpaces int
	TotalLines              int
}
