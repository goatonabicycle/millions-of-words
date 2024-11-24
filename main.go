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
	"strings"

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
	setupMiddleware(e)

	templates, err := parseTemplates(getEnv("TEMPLATES_DIR", defaultTemplatesDir))
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}

	e.Renderer = &TemplateRenderer{templates: templates}

	if err := loadAlbums(); err != nil {
		e.Logger.Fatal(err)
	}

	setupRoutes(e)

	port := getEnv("PORT", defaultPort)
	e.Logger.Fatal(e.Start(":" + port))
}

func setupMiddleware(e *echo.Echo) {
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/static", "static")
}

func parseTemplates(templatesDir string) (*template.Template, error) {
	templates := template.New("")

	err := filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".html") {
			relPath, err := filepath.Rel(templatesDir, path)
			if err != nil {
				return err
			}
			relPath = filepath.ToSlash(relPath)

			_, err = templates.New(relPath).ParseFiles(path)
			if err != nil {
				return err
			}
			fmt.Printf("Parsed template: %s\n", relPath)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return templates, nil
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
	e.GET("/album/:slug", albumDetailsHandler)
	e.GET("/search-albums", searchAlbumsHandler)
	e.GET("/edit-albums", editAlbumsHandler)
	e.POST("/edit-albums/verify-auth", verifyAuthHandler)
	e.GET("/edit-albums/tracks", editAlbumTracksHandler)
	e.POST("/edit-albums/update-track", updateTrackHandler)
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
		"albums":      albums,
		"Title":       "All Albums - Millions of Words",
		"IsAllAlbums": true,
	})
}

func albumDetailsHandler(c echo.Context) error {
	slug := c.Param("slug")

	album, err := loader.GetAlbumBySlug(slug)
	if err != nil {
		log.Printf("Error loading album with slug %s: %v", slug, err)
		return echo.NewHTTPError(http.StatusNotFound, "Album not found")
	}

	data := prepareAlbumDetails(album)
	return renderTemplate(c, "album-details.html", data)
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

	return renderTemplate(c, "all-words.html", map[string]interface{}{
		"wordFrequencies":     wordFrequencies,
		"wordFrequenciesJSON": template.JS(wordFrequenciesJSON),
		"Title":               "Word Frequencies - Millions of Words",
		"IsAllWords":          true,
	})
}

func editAlbumsHandler(c echo.Context) error {
	log.Printf("Loading edit albums page")
	return renderTemplate(c, "edit-albums.html", map[string]interface{}{
		"albums": albums,
		"Title":  "Edit Albums - Millions of Words",
	})
}

func verifyAuthHandler(c echo.Context) error {
	key := c.FormValue("authKey")

	valid, err := loader.ValidateAuthKey(key)
	if err != nil {
		log.Printf("Error validating auth key: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	if !valid {
		return c.NoContent(http.StatusUnauthorized)
	}

	return renderTemplate(c, "editor-section.html", map[string]interface{}{
		"albums": albums,
	})
}

func editAlbumTracksHandler(c echo.Context) error {
	key := c.FormValue("authKey")
	valid, err := loader.ValidateAuthKey(key)
	if err != nil || !valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid auth key")
	}

	id := c.FormValue("albumId")
	log.Printf("Loading tracks for album: %s", id)

	if id == "" {
		log.Printf("No album ID provided")
		return echo.NewHTTPError(http.StatusBadRequest, "No album ID provided")
	}

	for _, album := range albums {
		if album.ID == id {
			return renderTemplate(c, "edit-album-tracks.html", map[string]interface{}{
				"Album": album,
			})
		}
	}
	return echo.NewHTTPError(http.StatusNotFound, "Album not found")
}

func updateTrackHandler(c echo.Context) error {
	var req models.UpdateLyricsRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Error binding request: %v", err)
		return c.HTML(http.StatusOK, `<div class="text-red-500">Error: Failed to process request</div>`)
	}

	if err := loader.UpdateTrackLyrics(req); err != nil {
		log.Printf("Error updating lyrics: %v", err)
		return c.HTML(http.StatusOK, `<div class="text-red-500">Error: Failed to update lyrics</div>`)
	}

	if err := loadAlbums(); err != nil {
		log.Printf("Error reloading albums: %v", err)
		return c.HTML(http.StatusOK, `<div class="text-yellow-500">Saved but failed to refresh</div>`)
	}

	albumDetailsCache.Delete(req.AlbumID)

	return c.HTML(http.StatusOK, `<div class="text-green-500">Updated successfully</div>`)
}
