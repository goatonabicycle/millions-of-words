package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"

	loader "millions-of-words/loaders/sqlite"
	"millions-of-words/models"
	"millions-of-words/words"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	defaultPort         = "8080"
	defaultTemplatesDir = "./templates"
)

var (
	albums []models.BandcampAlbumData
)

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	templatesDir := getEnv("TEMPLATES_DIR", defaultTemplatesDir)
	renderer := &TemplateRenderer{
		templates: template.Must(template.New("").ParseGlob(filepath.Join(templatesDir, "*.html"))),
	}
	e.Renderer = renderer

	if err := loadAlbums(); err != nil {
		e.Logger.Fatal(err)
	}

	setupRoutes(e)

	port := getEnv("PORT", defaultPort)
	e.Logger.Fatal(e.Start(":" + port))
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func loadAlbums() error {
	var err error
	albums, err = loader.LoadAlbumsData()
	if err != nil {
		return fmt.Errorf("failed to load album data: %w", err)
	}
	return nil
}

func setupRoutes(e *echo.Echo) {
	e.GET("/", indexHandler)
	e.GET("/all-words", allWordsHandler)
	e.GET("/all-albums", allAlbumsHandler)
	e.GET("/album/:id", albumDetailsHandler)
	e.GET("/search-albums", searchAlbumsHandler)
}

func renderTemplate(c echo.Context, name string, data map[string]interface{}) error {
	return c.Render(http.StatusOK, name, data)
}

func indexHandler(c echo.Context) error {
	return renderTemplate(c, "index.html", map[string]interface{}{
		"albums": albums,
	})
}

func allAlbumsHandler(c echo.Context) error {
	return renderTemplate(c, "all-albums.html", map[string]interface{}{
		"albums": albums,
	})
}

func albumDetailsHandler(c echo.Context) error {
	id := c.Param("id")

	for _, album := range albums {
		if album.ID == id {
			data := prepareAlbumDetails(album)
			return renderTemplate(c, "album-details.html", data)
		}
	}
	return echo.NewHTTPError(http.StatusNotFound, "Album not found")
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
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to marshal word frequencies")
	}

	return c.Render(http.StatusOK, "all-words.html", map[string]interface{}{
		"wordFrequencies":     wordFrequencies,
		"wordFrequenciesJSON": template.JS(wordFrequenciesJSON),
	})
}
