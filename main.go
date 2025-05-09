package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"millions-of-words/fetch"
	"millions-of-words/internal/admin"
	loader "millions-of-words/loaders/supabase"
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
	// Add a cache for home page stats
	homePageCache struct {
		Albums             []models.BandcampAlbumData
		TotalAlbums        int
		TotalSongs         int
		TotalWords         int
		TotalChars         int
		TotalCharsNoSpaces int
		TotalVowels        int
		TotalConsonants    int
		TotalLines         int
		TotalDuration      int
		AvgWordsPerAlbum   int
		AvgCharsPerAlbum   int
		AvgWordLength      float64
		AvgSongsPerAlbum   float64
		WPM                int
		ProjectedAlbums    float64
		FuckCount          int
		DisplayAlbums      []models.BandcampAlbumData
	}
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

	renderer := &TemplateRenderer{templates: templates}
	e.Renderer = renderer

	if err := loadAlbums(); err != nil {
		e.Logger.Fatal(err)
	}
	if err := refreshHomePageCache(); err != nil {
		e.Logger.Fatal(err)
	}

	setupRoutes(e)
	setupAdminRoutes(e, renderer)

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

	e.GET("/pos", partsOfSpeechHandler)

	e.GET("/about", aboutHandler)
	e.GET("/all-words", allWordsHandler)
	e.GET("/all-albums", allAlbumsHandler)
	e.GET("/album/:slug", albumDetailsHandler)
	e.GET("/search-albums", searchAlbumsHandler)
	e.GET("/all-albums/sort", sortAlbumsHandler)
	e.GET("/all-albums/filter", filterAlbumsHandler)
}

func setupAdminRoutes(e *echo.Echo, renderer *TemplateRenderer) {
	adminHandler := admin.NewHandler(renderer)
	admin.SetupRoutes(e, adminHandler)
}

func renderTemplate(c echo.Context, name string, data map[string]interface{}) error {
	return c.Render(http.StatusOK, name, data)
}

func aboutHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "about.html", nil)
}

func countWordOccurrences(albums []models.BandcampAlbumData, word string) int {
	count := 0
	word = strings.ToLower(word)
	for _, album := range albums {
		for _, track := range album.Tracks {
			lyrics := strings.ToLower(track.Lyrics)
			count += strings.Count(lyrics, word)
		}
	}
	return count
}

func refreshHomePageCache() error {
	allAlbums, err := loader.LoadAlbumsData()
	if err != nil {
		return err
	}

	displayAlbums := allAlbums
	if len(allAlbums) > 18 {
		displayAlbums = allAlbums[:18]
	}

	totalSongs := 0
	totalWords := 0
	totalChars := 0
	totalCharsNoSpaces := 0
	totalVowels := 0
	totalConsonants := 0
	totalLines := 0
	totalDuration := 0

	for _, album := range allAlbums {
		totalSongs += len(album.Tracks)
		totalWords += album.TotalWords
		totalChars += album.TotalCharacters
		totalCharsNoSpaces += album.TotalCharactersNoSpaces
		totalVowels += album.TotalVowelCount
		totalConsonants += album.TotalConsonantCount
		totalLines += album.TotalLines
		totalDuration += album.TotalLength
	}

	albumCount := len(allAlbums)
	var avgWordsPerAlbum int
	var avgSongsPerAlbum float64
	var avgWordLength float64
	var projectedAlbums float64
	var wpm int

	if albumCount > 0 {
		avgWordsPerAlbum = totalWords / albumCount
		avgSongsPerAlbum = math.Round(float64(totalSongs)/float64(albumCount)*100) / 100

		if totalWords > 0 {
			avgWordLength = math.Round(float64(totalChars)/float64(totalWords)*100) / 100
		}

		if avgWordsPerAlbum > 0 {
			projectedAlbums = float64(((1000000 - totalWords) / avgWordsPerAlbum)) + float64(albumCount)
		}

		if totalDuration > 0 {
			wpm = totalWords / (totalDuration / 60)
		}
	}

	fuckCount := countWordOccurrences(allAlbums, "fuck")

	homePageCache = struct {
		Albums             []models.BandcampAlbumData
		TotalAlbums        int
		TotalSongs         int
		TotalWords         int
		TotalChars         int
		TotalCharsNoSpaces int
		TotalVowels        int
		TotalConsonants    int
		TotalLines         int
		TotalDuration      int
		AvgWordsPerAlbum   int
		AvgCharsPerAlbum   int
		AvgWordLength      float64
		AvgSongsPerAlbum   float64
		WPM                int
		ProjectedAlbums    float64
		FuckCount          int
		DisplayAlbums      []models.BandcampAlbumData
	}{
		Albums:             allAlbums,
		TotalAlbums:        albumCount,
		TotalSongs:         totalSongs,
		TotalWords:         totalWords,
		TotalChars:         totalChars,
		TotalCharsNoSpaces: totalCharsNoSpaces,
		TotalVowels:        totalVowels,
		TotalConsonants:    totalConsonants,
		TotalLines:         totalLines,
		TotalDuration:      totalDuration / 60,
		AvgWordsPerAlbum:   avgWordsPerAlbum,
		AvgCharsPerAlbum:   totalChars / max(albumCount, 1),
		AvgWordLength:      avgWordLength,
		AvgSongsPerAlbum:   avgSongsPerAlbum,
		WPM:                wpm,
		ProjectedAlbums:    projectedAlbums,
		FuckCount:          fuckCount,
		DisplayAlbums:      displayAlbums,
	}
	return nil
}

func indexHandler(c echo.Context) error {
	return renderTemplate(c, "index.html", map[string]interface{}{
		"albums":             homePageCache.DisplayAlbums,
		"TotalAlbums":        homePageCache.TotalAlbums,
		"TotalSongs":         homePageCache.TotalSongs,
		"TotalWords":         homePageCache.TotalWords,
		"TotalChars":         homePageCache.TotalChars,
		"TotalCharsNoSpaces": homePageCache.TotalCharsNoSpaces,
		"TotalVowels":        homePageCache.TotalVowels,
		"TotalConsonants":    homePageCache.TotalConsonants,
		"TotalLines":         homePageCache.TotalLines,
		"TotalDuration":      homePageCache.TotalDuration,
		"AvgWordsPerAlbum":   homePageCache.AvgWordsPerAlbum,
		"AvgCharsPerAlbum":   homePageCache.AvgCharsPerAlbum,
		"AvgWordLength":      homePageCache.AvgWordLength,
		"AvgSongsPerAlbum":   homePageCache.AvgSongsPerAlbum,
		"WPM":                homePageCache.WPM,
		"ProjectedAlbums":    homePageCache.ProjectedAlbums,
		"FuckCount":          homePageCache.FuckCount,
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
			wordCounts, _, _, _ := words.CalculateAndSortWordFrequencies(track.Lyrics, track.IgnoredWords)
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

func updateTrackHandler(c echo.Context) error {
	var req models.UpdateTrackRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Error binding request: %v", err)
		return c.HTML(http.StatusOK, `<div class="text-red-500">Error: Failed to process request</div>`)
	}

	if err := loader.UpdateTrack(req); err != nil {
		log.Printf("Error updating track: %v", err)
		return c.HTML(http.StatusOK, `<div class="text-red-500">Error: Failed to update track</div>`)
	}

	if err := loadAlbums(); err != nil {
		log.Printf("Error reloading albums: %v", err)
		return c.HTML(http.StatusOK, `<div class="text-yellow-500">Saved but failed to refresh</div>`)
	}

	albumDetailsCache.Delete(req.AlbumID)

	return c.HTML(http.StatusOK, `<div class="text-green-500">Updated successfully</div>`)
}

func allAlbumsHandler(c echo.Context) error {
	albums, err := loader.LoadAlbumsData()
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "all-albums.html", map[string]interface{}{
		"Title":       "All Albums - Millions of Words",
		"IsAllAlbums": true,
		"Albums":      albums,
		"Sort":        "date_added",
		"SortDir":     "desc",
		"NextSortDir": "asc",
	})
}

func sortAlbumsHandler(c echo.Context) error {
	log.Printf("Sorting albums...")
	allAlbums, err := loader.LoadAlbumsData()
	if err != nil {
		log.Printf("Error loading albums: %v", err)
		return err
	}

	sort := c.QueryParam("sort")
	dir := c.QueryParam("dir")
	log.Printf("Sort: %s, Dir: %s", sort, dir)

	nextSortDir := "asc"
	if dir == "asc" {
		nextSortDir = "desc"
	}

	sortAlbums(allAlbums, sort, dir)

	return c.Render(http.StatusOK, "album-table.html", map[string]interface{}{
		"Albums":      allAlbums,
		"Sort":        sort,
		"SortDir":     dir,
		"NextSortDir": nextSortDir,
	})
}

func filterAlbumsHandler(c echo.Context) error {
	allAlbums, err := loader.LoadAlbumsData()
	if err != nil {
		return err
	}

	search := c.QueryParam("search")
	filtered := filterAlbums(allAlbums, search)

	return c.Render(http.StatusOK, "album-table.html", map[string]interface{}{
		"Albums":      filtered,
		"Sort":        "date_added",
		"SortDir":     "desc",
		"NextSortDir": "asc",
	})
}

type AlbumTableData struct {
	Albums      []models.BandcampAlbumData
	Sort        string
	SortDir     string
	NextSortDir string
}

func renderAlbumsTable(c echo.Context, sortField, sortDir, search string) error {
	log.Printf("renderAlbumsTable: sortField=%s, sortDir=%s, search=%s", sortField, sortDir, search)

	allAlbums, err := loader.LoadAllAlbumsData()
	if err != nil {
		log.Printf("Error loading albums: %v", err)
		return err
	}
	log.Printf("Loaded %d albums total", len(allAlbums))

	nextSortDir := "asc"
	if sortDir == "asc" {
		nextSortDir = "desc"
	}

	filtered := filterAlbums(allAlbums, search)
	log.Printf("After filtering: %d albums", len(filtered))

	sortAlbums(filtered, sortField, sortDir)
	log.Printf("After sorting by %s %s", sortField, sortDir)

	data := AlbumTableData{
		Albums:      filtered,
		Sort:        sortField,
		SortDir:     sortDir,
		NextSortDir: nextSortDir,
	}

	return c.Render(http.StatusOK, "album-table.html", data)
}

func filterAlbums(albums []models.BandcampAlbumData, search string) []models.BandcampAlbumData {
	if search == "" {
		log.Printf("No search term, returning all %d albums", len(albums))
		return albums
	}

	search = strings.ToLower(search)
	var filtered []models.BandcampAlbumData

	for _, album := range albums {
		if strings.Contains(strings.ToLower(album.ArtistName), search) ||
			strings.Contains(strings.ToLower(album.AlbumName), search) {
			filtered = append(filtered, album)
		}
	}

	log.Printf("Filtered albums by '%s': %d results", search, len(filtered))
	return filtered
}

func sortAlbums(albums []models.BandcampAlbumData, sortField, sortDir string) {
	sort.Slice(albums, func(i, j int) bool {
		var result bool

		switch sortField {
		case "name":
			nameI := strings.ToLower(albums[i].ArtistName + " " + albums[i].AlbumName)
			nameJ := strings.ToLower(albums[j].ArtistName + " " + albums[j].AlbumName)
			result = nameI < nameJ
		case "words":
			result = albums[i].TotalWords < albums[j].TotalWords
		case "unique":
			result = albums[i].TotalUniqueWords < albums[j].TotalUniqueWords
		case "length":
			result = albums[i].TotalLength < albums[j].TotalLength
		case "wpt":
			result = albums[i].AverageWordsPerTrack < albums[j].AverageWordsPerTrack
		default:
			result = albums[i].DateAdded < albums[j].DateAdded
		}

		if sortDir == "desc" {
			result = !result
		}
		return result
	})
}

func importAlbumHandler(c echo.Context) error {
	user := c.Get("user").(*loader.User)
	if user == nil {
		return c.HTML(http.StatusUnauthorized, `
			<div class="bg-red-500/10 border border-red-500 text-red-500 p-4 rounded">
				Invalid authentication
			</div>
		`)
	}

	urls := strings.Split(c.FormValue("bandcampUrls"), "\n")
	var results []string
	total := len(urls)

	log.Printf("Starting import of %d URLs", total)

	for i, url := range urls {
		url = strings.TrimSpace(url)
		if url == "" {
			continue
		}

		log.Printf("Processing URL %d/%d: %s", i+1, total, url)

		if !strings.Contains(url, "bandcamp.com") {
			results = append(results, fmt.Sprintf(`
				<div class="bg-red-500/10 border border-red-500 text-red-500 p-2 rounded text-sm">
					Invalid Bandcamp URL: %s
				</div>`, url))
			continue
		}

		exists, err := loader.AlbumUrlExists(url)
		if err != nil {
			results = append(results, fmt.Sprintf(`
				<div class="bg-red-500/10 border border-red-500 text-red-500 p-2 rounded text-sm">
					Error checking database for %s: %v
				</div>`, url, err))
			continue
		}

		if exists {
			results = append(results, fmt.Sprintf(`
				<div class="bg-yellow-500/10 border border-yellow-500 text-yellow-500 p-2 rounded text-sm">
					%s has already been imported
				</div>`, url))
			continue
		}

		albumData, err := fetch.FetchFromBandcamp(url)
		if err != nil {
			results = append(results, fmt.Sprintf(`
				<div class="bg-red-500/10 border border-red-500 text-red-500 p-2 rounded text-sm">
					Error fetching %s: %v
				</div>`, url, err))
			continue
		}

		if err := loader.SaveAlbum(albumData); err != nil {
			results = append(results, fmt.Sprintf(`
				<div class="bg-red-500/10 border border-red-500 text-red-500 p-2 rounded text-sm">
					Error saving %s: %v
				</div>`, url, err))
			continue
		}

		results = append(results, fmt.Sprintf(`
			<div class="bg-green-500/10 border border-green-500 text-green-500 p-2 rounded text-sm">
				Successfully imported %s - %s
			</div>`, albumData.ArtistName, albumData.AlbumName))

		progressMsg := fmt.Sprintf(`
			<div class="text-gray-400 text-sm text-right">
				Processed %d of %d
			</div>
		`, i+1, total)

		log.Printf("Writing progress update to response")
		_, err = c.Response().Write([]byte(strings.Join(results, "\n") + progressMsg))
		if err != nil {
			log.Printf("Error writing progress: %v", err)
		}

		log.Printf("Flushing response")
		c.Response().Flush()
	}

	if err := loadAlbums(); err != nil {
		log.Printf("Error reloading albums after import: %v", err)
	}
	if err := refreshHomePageCache(); err != nil {
		log.Printf("Error refreshing home page cache after import: %v", err)
	}

	log.Printf("Import complete")
	return c.HTML(http.StatusOK, strings.Join(results, "\n"))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func updateAlbumHandler(c echo.Context) error {
	var req models.UpdateAlbumRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Error binding request: %v", err)
		return c.HTML(http.StatusOK, `<div class="text-red-500">Error: Failed to process request</div>`)
	}

	if err := loader.UpdateAlbum(req); err != nil {
		log.Printf("Error updating album: %v", err)
		return c.HTML(http.StatusOK, `<div class="text-red-500">Error: Failed to update album</div>`)
	}

	if err := loadAlbums(); err != nil {
		log.Printf("Error reloading albums: %v", err)
		return c.HTML(http.StatusOK, `<div class="text-yellow-500">Saved but failed to refresh</div>`)
	}

	return c.HTML(http.StatusOK, `<div class="text-green-500">Album updated successfully</div>`)
}

func partsOfSpeechHandler(c echo.Context) error {
	return renderTemplate(c, "pos.html", map[string]interface{}{
		"Title":         "Parts of Speech Analyzer",
		"IsPOSAnalyzer": true,
	})
}

func adminAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := c.Cookie("session")
		if err != nil || session.Value == "" {
			return c.Redirect(http.StatusSeeOther, "/admin/login")
		}

		// Validate session with Supabase
		user, err := loader.ValidateSession(session.Value)
		if err != nil {
			c.SetCookie(&http.Cookie{
				Name:     "session",
				Value:    "",
				Path:     "/",
				Expires:  time.Now().Add(-1 * time.Hour),
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
			})
			return c.Redirect(http.StatusSeeOther, "/admin/login")
		}

		c.Set("user", user)
		return next(c)
	}
}

func handleAdminAuth(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	user, err := loader.SignInWithEmail(email, password)
	if err != nil {
		return c.HTML(http.StatusUnauthorized, `<div class="text-red-500">Invalid credentials</div>`)
	}

	c.SetCookie(&http.Cookie{
		Name:     "session",
		Value:    user.AccessToken,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	return c.HTML(http.StatusOK, `<div class="text-green-500">Login successful! Redirecting...</div>
		<script>
			setTimeout(() => {
				window.location.href = "/admin";
			}, 1000);
		</script>`)
}

func parseTrackNumber(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}
