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

	"millions-of-words/models"
	"millions-of-words/words"

	"github.com/PuerkitoBio/goquery"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	albums  []models.BandcampAlbumData
	albumID int64
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
	e.POST("/fetch-album", fetchAlbumHandler)
	e.GET("/:id", albumDetailsHandler)

	e.Logger.Fatal(e.Start("0.0.0.0:8080"))
}

func indexHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", map[string]interface{}{
		"albums": albums,
	})
}

func fetchAlbumHandler(c echo.Context) error {
	url := c.FormValue("url")
	if url == "" {
		return c.String(http.StatusBadRequest, "URL parameter is missing.")
	}

	albumData := fetchAlbumDataFromBandcamp(url)
	albumData.ID = fmt.Sprintf("%d", albumID) // Use a unique identifier
	albumID++

	albums = append(albums, albumData)
	writeAlbumsDataToJsonFile(albumData)

	return c.Redirect(http.StatusSeeOther, "/")
}

func albumDetailsHandler(c echo.Context) error {
	id := c.Param("id")

	for _, album := range albums {
		if album.ID == id {
			return c.Render(http.StatusOK, "album-details.html", album)
		}
	}
	return c.String(http.StatusNotFound, "Album not found.")
}

func fetchAlbumDataFromBandcamp(url string) models.BandcampAlbumData {
	fmt.Printf("Fetching album data from Bandcamp for URL: %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error fetching Bandcamp page: %v", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalf("Error parsing HTML: %v", err)
	}

	artistName := doc.Find("#name-section h3 span a").Text()
	albumName := doc.Find(".trackTitle").First().Text()
	description := doc.Find(".tralbumData.tralbum-about").Text()
	imageUrl := doc.Find("a.popupImage").AttrOr("href", "")

	var tracklist []models.BandcampTrackData
	doc.Find("tr.lyricsRow").Each(func(i int, s *goquery.Selection) {
		lyrics := s.Find("div").Text()
		trackTitle := doc.Find(".title-col .track-title").Eq(i).Text() // Matching track titles with lyrics

		track := models.BandcampTrackData{
			Name:             strings.TrimSpace(trackTitle),
			Lyrics:           lyrics,
			SortedWordCounts: words.CalculateAndSortWordFrequencies(lyrics),
		}

		tracklist = append(tracklist, track)
	})

	return models.BandcampAlbumData{
		ArtistName:  strings.TrimSpace(artistName),
		AlbumName:   strings.TrimSpace(albumName),
		Description: strings.TrimSpace(description),
		ImageUrl:    imageUrl,
		Tracks:      tracklist,
	}
}

func writeAlbumsDataToJsonFile(album models.BandcampAlbumData) {
	dir := "data"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Error creating directory: %v", err)
		}
	}

	filename := filepath.Join(dir, fmt.Sprintf("%s - %s.json", sanitizeFilename(album.ArtistName), sanitizeFilename(album.AlbumName)))
	file, err := json.MarshalIndent(album, "", " ")
	if err != nil {
		log.Fatalf("Error marshalling album data to JSON: %v", err)
	}

	err = os.WriteFile(filename, file, 0644)
	if err != nil {
		log.Fatalf("Error writing album data to file: %v", err)
	}
}

func sanitizeFilename(name string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(name)), "_")
}

func loadAlbumsDataFromJsonFiles() {
	albums = []models.BandcampAlbumData{}
	files, err := os.ReadDir("data")
	if err != nil {
		log.Fatalf("Error reading album data directory: %v", err)
	}

	for _, f := range files {
		file, err := os.ReadFile(filepath.Join("data", f.Name()))
		if err != nil {
			log.Printf("Error reading file %s: %v", f.Name(), err)
			continue
		}

		var album models.BandcampAlbumData
		if err := json.Unmarshal(file, &album); err != nil {
			log.Printf("Error unmarshalling file %s: %v", f.Name(), err)
			continue
		}

		albums = append(albums, album)
		// Log albums:
		fmt.Printf("Loaded album data from JSON file: %s\n", f.Name())
	}
}
