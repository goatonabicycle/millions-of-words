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

	"millions-of-words/models"

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
		templates: template.Must(template.ParseGlob(filepath.Join(templatesDir, "*.html"))),
	}
	e.Renderer = renderer

	loadAlbumsDataFromJsonFiles()

	e.GET("/", indexHandler)
	e.GET("/:id", albumDetailsHandler)
	e.GET("/search-albums", searchAlbumsHandler)
	e.GET("/full", fullHandler)

	e.Logger.Fatal(e.Start("0.0.0.0:8080"))
}

func indexHandler(c echo.Context) error {
	partial := c.QueryParam("partial")
	if partial == "true" {
		searchQuery := c.QueryParam("search")
		fmt.Fprintln(os.Stdout, "Search query:", searchQuery)
		filteredAlbums := filterAlbumsByQuery(searchQuery)

		// Render only the album grid part
		return c.Render(http.StatusOK, "album-grid.html", map[string]interface{}{
			"albums": filteredAlbums,
		})
	}
	// Render the full page
	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"albums": albums,
	})
}

func albumDetailsHandler(c echo.Context) error {
	id := c.Param("id")

	for _, album := range albums {
		if album.ID == id {
			album.AlbumWordFrequencies = aggregateWordFrequencies(album)
			return c.Render(http.StatusOK, "album-details.html", album)
		}
	}
	return c.String(http.StatusNotFound, "Album not found.")
}

func searchAlbumsHandler(c echo.Context) error {
	searchQuery := c.QueryParam("search")
	filteredAlbums := filterAlbumsByQuery(searchQuery)

	log.Printf("Searching for %s:", searchQuery)

	return c.Render(http.StatusOK, "album-grid.html", map[string]interface{}{
		"albums": filteredAlbums,
	})
}

func fullHandler(c echo.Context) error {

	fmt.Fprintln(os.Stdout, "You called the full handler!")
	allWords := []models.WordCount{}

	for _, album := range albums {
		for _, track := range album.Tracks {
			allWords = append(allWords, track.SortedWordCounts...)
		}
	}

	log.Printf("All words: %v", allWords)

	// Render only the album grid part
	return c.Render(http.StatusOK, "all.html", map[string]interface{}{
		"allWords": allWords,
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
		for _, wc := range track.SortedWordCounts {
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

	if len(totalWordFrequencies) > 20 {
		return totalWordFrequencies[:20]
	}

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
