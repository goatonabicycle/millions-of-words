package main

import (
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"millions-of-words/loader"
	"millions-of-words/models"
	"millions-of-words/words"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	albums  []models.BandcampAlbumData
	dataDir = "data"
)

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	templatesDir := os.Getenv("TEMPLATES_DIR")
	if templatesDir == "" {
		templatesDir = "./templates"
	}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	renderer := &TemplateRenderer{
		templates: template.Must(template.New("").ParseGlob(filepath.Join(templatesDir, "*.html"))),
	}
	e.Renderer = renderer

	var err error
	albums, err = loader.LoadAlbumsData()
	if err != nil {
		log.Fatalf("Failed to load album data: %v", err)
	}

	e.GET("/", indexHandler)
	e.GET("/all-words", allWordsHandler)
	e.GET("/all-albums", allAlbumsHandler)
	e.GET("/album/:id", albumDetailsHandler)
	e.GET("/search-albums", searchAlbumsHandler)

	e.Logger.Fatal(e.Start("0.0.0.0:8080"))
}

func indexHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"albums": albums,
	})
}

func allAlbumsHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "all-albums.html", map[string]interface{}{
		"albums": albums,
	})
}

func albumDetailsHandler(c echo.Context) error {
	id := c.Param("id")

	for _, album := range albums {
		if album.ID == id {
			album.AlbumWordFrequencies = aggregateWordFrequencies(album)
			if len(album.AlbumWordFrequencies) > 20 {
				album.AlbumWordFrequencies = album.AlbumWordFrequencies[:20]
			}

			totalWords := 0
			totalVowelCount := 0
			totalConsonantCount := 0
			wordLengthDistribution := make(map[int]int)
			uniqueWordsMap := make(map[string]struct{})

			type TrackWithDetails struct {
				Track                  models.BandcampTrackData
				FormattedLyrics        template.HTML
				SortedWordCounts       []models.WordCount
				WordsPerMinute         float64
				TotalWords             int
				UniqueWords            int
				VowelCount             int
				ConsonantCount         int
				WordLengthDistribution map[int]int
			}
			tracksWithDetails := []TrackWithDetails{}

			for _, track := range album.Tracks {
				sortedWordCounts, vowels, consonants, wordLengths := words.CalculateAndSortWordFrequencies(track.Lyrics)
				wordCount := len(strings.Fields(track.Lyrics))
				totalWords += wordCount
				totalVowelCount += vowels
				totalConsonantCount += consonants
				trackUniqueWordsMap := make(map[string]struct{})

				for length, count := range wordLengths {
					wordLengthDistribution[length] += count
				}

				for _, wc := range sortedWordCounts {
					uniqueWordsMap[wc.Word] = struct{}{}
					trackUniqueWordsMap[wc.Word] = struct{}{}
				}

				wpm := 0.0
				if float64(track.TotalLength)/60 > 0 {
					wpm = float64(wordCount) / (float64(track.TotalLength) / 60)
				}

				lyrics := template.HTML(track.Lyrics)

				tracksWithDetails = append(tracksWithDetails, TrackWithDetails{
					Track:                  track,
					FormattedLyrics:        lyrics,
					SortedWordCounts:       sortedWordCounts,
					WordsPerMinute:         wpm,
					TotalWords:             wordCount,
					UniqueWords:            len(trackUniqueWordsMap),
					VowelCount:             vowels,
					ConsonantCount:         consonants,
					WordLengthDistribution: wordLengths,
				})
			}

			album.TotalWords = totalWords
			album.AverageWordsPerTrack = totalWords / len(album.Tracks)
			album.TotalUniqueWords = len(uniqueWordsMap)
			album.TotalVowelCount = totalVowelCount
			album.TotalConsonantCount = totalConsonantCount

			albumWPM := 0.0
			if float64(album.TotalLength)/60 > 0 {
				albumWPM = float64(totalWords) / (float64(album.TotalLength) / 60)
			}

			data := struct {
				Album             models.BandcampAlbumData
				TracksWithDetails []TrackWithDetails
				AlbumWPM          float64
			}{
				Album:             album,
				TracksWithDetails: tracksWithDetails,
				AlbumWPM:          albumWPM,
			}

			return c.Render(http.StatusOK, "album-details.html", data)
		}
	}
	return c.String(http.StatusNotFound, "Album not found.")
}

func searchAlbumsHandler(c echo.Context) error {
	searchQuery := c.QueryParam("search")
	filteredAlbums := filterAlbumsByQuery(searchQuery)

	return c.Render(http.StatusOK, "album-grid.html", map[string]interface{}{
		"albums": filteredAlbums,
	})
}

func allWordsHandler(c echo.Context) error {
	wordFrequencyMap := make(map[string]int)

	for _, album := range albums {
		for _, track := range album.Tracks {
			wordCounts, _, _, _ := words.CalculateAndSortWordFrequencies(track.Lyrics)
			for _, wc := range wordCounts {
				wordFrequencyMap[wc.Word] += wc.Count
			}
		}
	}

	wordFrequencies := words.MapToSortedList(wordFrequencyMap)
	wordFrequenciesJSON, err := json.Marshal(wordFrequencies)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "all-words.html", map[string]interface{}{
		"wordFrequencies":     wordFrequencies,
		"wordFrequenciesJSON": template.JS(wordFrequenciesJSON),
	})
}

func filterAlbumsByQuery(query string) []models.BandcampAlbumData {
	var filtered []models.BandcampAlbumData
	query = strings.ToLower(query)
	for _, album := range albums {
		if strings.Contains(strings.ToLower(album.ArtistName), query) || strings.Contains(strings.ToLower(album.AlbumName), query) {
			filtered = append(filtered, album)
		}
	}
	return filtered
}

func aggregateWordFrequencies(album models.BandcampAlbumData) []models.WordCount {
	wordFreqMap := make(map[string]int)
	for _, track := range album.Tracks {
		wordCounts, _, _, _ := words.CalculateAndSortWordFrequencies(track.Lyrics)
		for _, wc := range wordCounts {
			wordFreqMap[wc.Word] += wc.Count
		}
	}

	var totalWordFrequencies []models.WordCount
	for word, count := range wordFreqMap {
		totalWordFrequencies = append(totalWordFrequencies, models.WordCount{Word: word, Count: count})
	}

	sort.Slice(totalWordFrequencies, func(i, j int) bool {
		if totalWordFrequencies[i].Count == totalWordFrequencies[j].Count {
			return totalWordFrequencies[i].Word < totalWordFrequencies[j].Word
		}
		return totalWordFrequencies[i].Count > totalWordFrequencies[j].Count
	})

	return totalWordFrequencies
}
