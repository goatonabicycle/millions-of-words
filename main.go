package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

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

	loadAlbumsDataFromJsonFiles()

	e.GET("/", indexHandler)
	e.GET("/all-words", allWordsHandler)
	e.GET("/album/:id", albumDetailsHandler)
	e.GET("/search-albums", searchAlbumsHandler)

	e.Logger.Fatal(e.Start("0.0.0.0:8080"))
}

func indexHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
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
			uniqueWordsMap := make(map[string]struct{})

			type TrackWithDetails struct {
				Track            models.BandcampTrackData
				FormattedLyrics  template.HTML
				SortedWordCounts []models.WordCount
				WordsPerMinute   float64
				TotalWords       int
				UniqueWords      int
			}
			tracksWithDetails := []TrackWithDetails{}

			var totalAlbumDuration time.Duration

			for _, track := range album.Tracks {
				sortedWordCounts := words.CalculateAndSortWordFrequencies(track.Lyrics)
				wordCount := len(strings.Fields(track.Lyrics))
				totalWords += wordCount
				trackUniqueWordsMap := make(map[string]struct{})

				for _, wc := range sortedWordCounts {
					uniqueWordsMap[wc.Word] = struct{}{}
					trackUniqueWordsMap[wc.Word] = struct{}{}
				}

				trackDuration, err := time.ParseDuration(track.Duration)
				if err != nil {
					log.Printf("Error parsing track duration for track %s: %v", track.Name, err)
				}
				totalAlbumDuration += trackDuration
				wpm := 0.0
				if trackDuration.Minutes() > 0 {
					wpm = float64(wordCount) / trackDuration.Minutes()
				}

				lyrics := template.HTML(track.Lyrics)

				tracksWithDetails = append(tracksWithDetails, TrackWithDetails{
					Track:            track,
					FormattedLyrics:  lyrics,
					SortedWordCounts: sortedWordCounts,
					WordsPerMinute:   wpm,
					TotalWords:       wordCount,
					UniqueWords:      len(trackUniqueWordsMap),
				})
			}

			album.TotalWords = totalWords
			album.AverageWordsPerTrack = totalWords / len(album.Tracks)
			album.TotalUniqueWords = len(uniqueWordsMap)
			album.TotalAlbumLength = totalAlbumDuration.String()

			albumWPM := 0.0
			if totalAlbumDuration.Minutes() > 0 {
				albumWPM = float64(totalWords) / totalAlbumDuration.Minutes()
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
	wordAlbumMap := make(map[string]map[string]struct{})

	for _, album := range albums {
		fmt.Println("Processing album:", album.AlbumName)

		for _, track := range album.Tracks {
			fmt.Println("Processing track:", track.Name)

			lyrics := track.Lyrics
			wordCounts := words.CalculateAndSortWordFrequencies(lyrics)
			fmt.Println("Word counts for track:", wordCounts)

			for _, wc := range wordCounts {
				wordFrequencyMap[wc.Word] += wc.Count

				if wordAlbumMap[wc.Word] == nil {
					wordAlbumMap[wc.Word] = make(map[string]struct{})
				}
				wordAlbumMap[wc.Word][album.AlbumName] = struct{}{}
				fmt.Printf("Added word '%s' to album '%s'\n", wc.Word, album.AlbumName)
			}
		}
	}

	fmt.Println("Word Frequency Map:", wordFrequencyMap)
	fmt.Println("Word Album Map:", wordAlbumMap)

	wordFrequencies := words.MapToSortedList(wordFrequencyMap)
	wordAlbums := make(map[string][]string)
	for word, albums := range wordAlbumMap {
		for album := range albums {
			wordAlbums[word] = append(wordAlbums[word], album)
		}
	}

	fmt.Println("Word Albums after processing:", wordAlbums)

	for _, wc := range wordFrequencies {
		if _, exists := wordAlbums[wc.Word]; !exists {
			wordAlbums[wc.Word] = []string{"No data"}
		}
	}

	wordFrequenciesJSON, err := json.Marshal(wordFrequencies)
	if err != nil {
		return err
	}
	wordAlbumsJSON, err := json.Marshal(wordAlbums)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "all-words.html", map[string]interface{}{
		"wordFrequencies":     wordFrequencies,
		"wordFrequenciesJSON": template.JS(wordFrequenciesJSON),
		"wordAlbumsJSON":      template.JS(wordAlbumsJSON),
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
		wordCounts := words.CalculateAndSortWordFrequencies(track.Lyrics)
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

func loadAlbumsDataFromJsonFiles() {
	albums = []models.BandcampAlbumData{}
	files, err := os.ReadDir(dataDir)
	if err != nil {
		log.Fatalf("Error reading album data directory: %v", err)
	}

	for _, f := range files {
		filePath := filepath.Join(dataDir, f.Name())
		file, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Error reading file %s: %v", filePath, err)
			continue
		}

		var album models.BandcampAlbumData
		if err := json.Unmarshal(file, &album); err != nil {
			log.Printf("Error unmarshalling file %s: %v", filePath, err)
			continue
		}

		albums = append(albums, album)
	}
}
