package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"

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
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseFiles("templates/index.html", "templates/albums.html", "templates/album-details.html")),
	}
	e.Renderer = renderer

	e.GET("/", indexHandler)
	e.POST("/fetch-album", fetchAlbumHandler)
	e.GET("/albums", albumsHandler)
	e.GET("/album/:id", albumDetailsHandler)

	e.Logger.Fatal(e.Start(":8000"))
}

func indexHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", nil)
}

func fetchAlbumHandler(c echo.Context) error {
	url := c.FormValue("url")
	if url == "" {
		return c.String(http.StatusBadRequest, "URL parameter is missing.")
	}

	albumData := fetchAlbumDataFromBandcamp(url)
	albumData.ID = getNextAlbumID()

	albums = append(albums, albumData)
	return c.Redirect(http.StatusSeeOther, "/albums")
}

func albumsHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "albums.html", albums)
}

func albumDetailsHandler(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "Invalid album ID.")
	}

	var albumData *models.BandcampAlbumData
	for _, album := range albums {
		if album.ID == id {
			albumData = &album
			break
		}
	}

	if albumData == nil {
		return c.String(http.StatusNotFound, "Album not found.")
	}

	return c.Render(http.StatusOK, "album-details.html", albumData)
}

func getNextAlbumID() int64 {
	return atomic.AddInt64(&albumID, 1)
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
