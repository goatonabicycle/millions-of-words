package main

import (
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	loader "millions-of-words/loaders/sqlite"
	"millions-of-words/models"
	"millions-of-words/words"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
