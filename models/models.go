package models

type WordCount struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

type BandcampAlbumData struct {
	ID                     string              `json:"id"`
	ArtistName             string              `json:"artist_name"`
	AlbumName              string              `json:"album_name"`
	Description            string              `json:"description"`
	ImageUrl               string              `json:"image_url"`
	BandcampUrl            string              `json:"bandcamp_url"`
	AmpwallUrl             string              `json:"ampwall_url"`
	Tracks                 []BandcampTrackData `json:"tracks"`
	TotalLength            int                 `json:"total_length"`
	FormattedLength        string              `json:"formatted_length"`
	AlbumWordFrequencies   []WordCount         `json:"-"`
	TotalWords             int                 `json:"-"`
	AverageWordsPerTrack   int                 `json:"-"`
	TotalUniqueWords       int                 `json:"-"`
	TotalVowelCount        int                 `json:"-"`
	TotalConsonantCount    int                 `json:"-"`
	WordLengthDistribution map[int]int         `json:"-"`
}

type BandcampTrackData struct {
	Name            string `json:"name"`
	TotalLength     int    `json:"total_length"`
	FormattedLength string `json:"formatted_length"`
	Lyrics          string `json:"lyrics,omitempty"`
}
