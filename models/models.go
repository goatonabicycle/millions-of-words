package models

import "html/template"

type WordCount struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

type BandcampAlbumData struct {
	ID                      string              `json:"id"`
	Slug                    string              `json:"slug"`
	ArtistName              string              `json:"artist_name"`
	AlbumName               string              `json:"album_name"`
	ImageUrl                string              `json:"image_url"`
	ImageStoragePath        string              `json:"image_storage_path"`
	ImageData               []byte              `json:"-"`
	ImageDataBase64         string              `json:"-"`
	BandcampUrl             string              `json:"bandcamp_url"`
	AmpwallUrl              string              `json:"ampwall_url"`
	MetalArchivesURL        string              `json:"metal_archives_url"`
	AlbumColorAverage       string              `json:"album_color_average"`
	DateAdded               string              `json:"date_added"`
	ReleaseDate             string              `json:"release_date"`
	ReleaseDateDaysAgo      string              `json:"-"`
	Genre                   string              `json:"genre"`
	Country                 string              `json:"country"`
	Label                   string              `json:"label"`
	IgnoredWords            string              `json:"ignored_words"`
	Notes                   string              `json:"notes"`
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
	Enabled                 bool                `json:"enabled"`
}

type BandcampTrackData struct {
	Name                    string `json:"name"`
	TrackNumber             int    `json:"track_number"`
	TotalLength             int    `json:"total_length"`
	FormattedLength         string `json:"formatted_length"`
	Lyrics                  string `json:"lyrics,omitempty"`
	IgnoredWords            string `json:"ignored_words"`
	TotalCharacters         int    `json:"-"`
	TotalCharactersNoSpaces int    `json:"-"`
	TotalLines              int    `json:"-"`
	TotalWords              int    `json:"-"`
}

type TrackWithDetails struct {
	Track                   BandcampTrackData
	TrackNumber             int
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
	IgnoredWords            string
}

type UpdateTrackRequest struct {
	AlbumID      string `json:"albumId" form:"albumId"`
	TrackName    string `json:"trackName" form:"trackName"`
	TrackNumber  int    `json:"trackNumber" form:"trackNumber"`
	Lyrics       string `json:"lyrics" form:"lyrics"`
	IgnoredWords string `json:"ignoredWords" form:"ignoredWords"`
	AuthKey      string `json:"authKey" form:"authKey"`
}

type UpdateAlbumRequest struct {
	AlbumID          string `json:"albumId" form:"albumId"`
	MetalArchivesURL string `json:"metalArchivesUrl" form:"metalArchivesUrl"`
	ReleaseDate      string `json:"releaseDate" form:"releaseDate"`
	Genre            string `json:"genre" form:"genre"`
	Country          string `json:"country" form:"country"`
	Label            string `json:"label" form:"label"`
	IgnoredWords     string `json:"ignoredWords" form:"ignoredWords"`
	Notes            string `json:"notes" form:"notes"`
	AuthKey          string `json:"authKey" form:"authKey"`
	Enabled          string `json:"enabled" form:"enabled"`
}
