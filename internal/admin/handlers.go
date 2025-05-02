package admin

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"

	"millions-of-words/fetch"
	loader "millions-of-words/loaders/supabase"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	templates TemplateRenderer
}

type TemplateRenderer interface {
	Render(w io.Writer, name string, data interface{}, c echo.Context) error
}

func NewHandler(templates TemplateRenderer) *Handler {
	return &Handler{
		templates: templates,
	}
}

func (h *Handler) AdminHandler(c echo.Context) error {
	cookie, err := c.Cookie("session")
	if err != nil {
		return h.templates.Render(c.Response().Writer, "admin/pages/login.html", map[string]interface{}{
			"Title": "Admin Login - Millions of Words",
		}, c)
	}

	user, err := loader.ValidateSession(cookie.Value)
	if err != nil || user == nil {
		return h.templates.Render(c.Response().Writer, "admin/pages/login.html", map[string]interface{}{
			"Title": "Admin Login - Millions of Words",
		}, c)
	}

	allAlbums, err := loader.LoadAllAlbumsData()
	if err != nil {
		log.Printf("Error loading albums: %v", err)
		return err
	}

	return h.templates.Render(c.Response().Writer, "admin/pages/index.html", map[string]interface{}{
		"Title":         "Admin - Millions of Words",
		"Authenticated": true,
		"Albums":        allAlbums,
	}, c)
}

func (h *Handler) AdminAuthHandler(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	user, err := loader.SignInWithEmail(email, password)
	if err != nil {
		return c.HTML(http.StatusUnauthorized, `
			<div class="text-red-500 text-center p-2">Invalid email or password</div>
		`)
	}

	c.SetCookie(&http.Cookie{
		Name:     "session",
		Value:    user.AccessToken,
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400, // 24 hours
	})

	c.Response().Header().Set("HX-Redirect", "/admin")
	return c.NoContent(http.StatusOK)
}

func (h *Handler) AdminImportHandler(c echo.Context) error {
	if err := validateAuth(c); err != nil {
		return err
	}

	return h.templates.Render(c.Response().Writer, "admin/components/import-form.html", nil, c)
}

func (h *Handler) FetchMetalArchivesHandler(c echo.Context) error {
	if err := validateAuth(c); err != nil {
		return err
	}

	url := c.FormValue("metalArchivesUrl")
	if url == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Metal Archives URL is required")
	}

	data, err := fetch.FetchFromMetalArchives(url)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to fetch data: %v", err))
	}

	return h.templates.Render(c.Response().Writer, "admin/components/metal-archives-preview.html", map[string]interface{}{
		"Data": data,
	}, c)
}

func (h *Handler) ValidateMetalArchivesUrlHandler(c echo.Context) error {
	if err := validateAuth(c); err != nil {
		return err
	}

	url := c.FormValue("metalArchivesUrl")
	if url == "" {
		return c.HTML(http.StatusOK, "")
	}

	if !strings.HasPrefix(url, "https://www.metal-archives.com/albums/") {
		return c.HTML(http.StatusOK, `<div class="text-red-500">Invalid Metal Archives URL format</div>`)
	}

	return c.HTML(http.StatusOK, "")
}

func (h *Handler) LogoutHandler(c echo.Context) error {
	// Clear the session cookie
	c.SetCookie(&http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
	return c.Redirect(http.StatusSeeOther, "/admin")
}

func (h *Handler) ImportAlbumHandler(c echo.Context) error {
	if err := validateAuth(c); err != nil {
		return err
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

	log.Printf("Import complete")
	return c.HTML(http.StatusOK, strings.Join(results, "\n"))
}

func (h *Handler) AlbumListHandler(c echo.Context) error {
	if err := validateAuth(c); err != nil {
		return err
	}
	albums, err := loader.LoadAllAlbumsData()
	if err != nil {
		return c.HTML(500, "Failed to load albums")
	}
	var sb strings.Builder
	sb.WriteString(`<ul class="list-disc pl-6">`)
	for _, album := range albums {
		sb.WriteString("<li>")
		sb.WriteString(template.HTMLEscapeString(album.ArtistName + " - " + album.AlbumName))
		sb.WriteString("</li>")
	}
	sb.WriteString("</ul>")
	return c.HTML(200, sb.String())
}

func validateAuth(c echo.Context) error {
	cookie, err := c.Cookie("session")
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Not authenticated")
	}

	user, err := loader.ValidateSession(cookie.Value)
	if err != nil || user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid session")
	}

	return nil
}
