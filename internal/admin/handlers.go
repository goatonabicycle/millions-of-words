package admin

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
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

	// Only fetch artist and album names for all albums (fast)
	allAlbums, err := loader.FetchAlbumNamesOnly()
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
	// Only fetch artist and album names for all albums
	albums, err := loader.FetchAlbumNamesOnly()
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

func (h *Handler) ImportStartHandler(c echo.Context) error {
	if err := validateAuth(c); err != nil {
		return err
	}
	urls := strings.Split(c.FormValue("bandcampUrls"), "\n")
	var albumRows strings.Builder
	var urlList []string
	for i, url := range urls {
		url = strings.TrimSpace(url)
		if url == "" {
			continue
		}
		urlList = append(urlList, url)
		albumRows.WriteString(`<div id="album-status-` + fmt.Sprint(i) + `" class="flex items-center gap-2 py-2 border-b border-gray-700">`)
		albumRows.WriteString(`<span class="w-8 h-8 flex items-center justify-center bg-gray-700 rounded-full text-gray-400">`)
		albumRows.WriteString(fmt.Sprint(i + 1))
		albumRows.WriteString(`</span>`)
		albumRows.WriteString(`<span class="flex-1 text-gray-200">` + template.HTMLEscapeString(url) + `</span>`)
		albumRows.WriteString(`<span class="text-yellow-400">Pending</span>`)
		albumRows.WriteString(`</div>`)
	}
	progressBar := `<div class="w-full bg-gray-700 rounded-full h-4 mb-4">
	  <div id="import-progress-bar" class="bg-blue-600 h-4 rounded-full" style="width:0%"></div>
	</div>`
	html := `<div id="import-progress-container">` + progressBar + `<div id="import-album-status-list">` + albumRows.String() + `</div></div>`
	// Add a script to trigger the first import if there are any URLs
	if len(urlList) > 0 {
		html += `<div hx-post="/admin/import/process"
					 hx-target="#album-status-0"
					 hx-swap="outerHTML"
					 hx-vals='{"albumIndex":0,"total":` + fmt.Sprint(len(urlList)) + `,"bandcampUrls":"` + template.JSEscapeString(strings.Join(urlList, "\\n")) + `"}'
					 hx-trigger="load"></div>`
		// Add a script to reload the import tab after a delay (when all are done)
		html += `<div id="import-finish-trigger"></div>`
		html += `<script>
		  function checkImportDone() {
			var lastRow = document.getElementById('album-status-` + fmt.Sprint(len(urlList)-1) + `');
			if (lastRow && lastRow.querySelector('.text-green-400, .text-red-400, .text-yellow-400')) {
			  setTimeout(function() {
				htmx.ajax('GET', '/admin/content/import', {target: '#admin-content', swap: 'innerHTML'});
				document.getElementById('import-progress-global').innerHTML = '';
				setTimeout(function() {
				  var msg = document.getElementById('import-status-message');
				  if (msg) msg.innerHTML = '<div class="bg-green-700 text-white p-2 rounded mb-2">Import complete!</div>';
				}, 500);
			  }, 1000);
			} else {
			  setTimeout(checkImportDone, 500);
			}
		  }
		  checkImportDone();
		</script>`
	}
	return c.HTML(200, html)
}

func (h *Handler) ImportProcessHandler(c echo.Context) error {
	if err := validateAuth(c); err != nil {
		return err
	}
	albumIndex, _ := strconv.Atoi(c.FormValue("albumIndex"))
	total, _ := strconv.Atoi(c.FormValue("total"))
	urls := strings.Split(c.FormValue("bandcampUrls"), "\n")
	if albumIndex >= len(urls) {
		return nil
	}
	url := strings.TrimSpace(urls[albumIndex])
	status := ""
	color := ""
	icon := ""
	if url == "" {
		status = "Skipped"
		color = "text-gray-400"
		icon = "-"
	} else if !strings.Contains(url, "bandcamp.com") {
		status = "Invalid URL"
		color = "text-red-400"
		icon = "✗"
	} else {
		exists, err := loader.AlbumUrlExists(url)
		if err != nil {
			status = "Error"
			color = "text-red-400"
			icon = "✗"
		} else if exists {
			status = "Already Imported"
			color = "text-yellow-400"
			icon = "!"
		} else {
			albumData, err := fetch.FetchFromBandcamp(url)
			if err != nil {
				status = "Fetch Error"
				color = "text-red-400"
				icon = "✗"
			} else if err := loader.SaveAlbum(albumData); err != nil {
				status = "Save Error"
				color = "text-red-400"
				icon = "✗"
			} else {
				status = "Imported"
				color = "text-green-400"
				icon = "✓"
			}
		}
	}
	row := `<div id="album-status-` + fmt.Sprint(albumIndex) + `" class="flex items-center gap-2 py-2 border-b border-gray-700" hx-swap-oob="true">`
	row += `<span class="w-8 h-8 flex items-center justify-center bg-gray-700 rounded-full ` + color + `">` + icon + `</span>`
	row += `<span class="flex-1 text-gray-200">` + template.HTMLEscapeString(url) + `</span>`
	row += `<span class="` + color + `">` + status + `</span>`
	row += `</div>`
	progress := int(float64(albumIndex+1) / float64(total) * 100)
	progressBar := `<div id="import-progress-bar" class="bg-blue-600 h-4 rounded-full" style="width:` + fmt.Sprint(progress) + `%;" hx-swap-oob="true"></div>`
	triggerNext := ""
	if albumIndex+1 < total {
		triggerNext = `<div hx-post=\"/admin/import/process\" hx-target=\"#album-status-` + fmt.Sprint(albumIndex+1) + `\" hx-swap=\"outerHTML\" hx-vals='{"albumIndex":` + fmt.Sprint(albumIndex+1) + `,"total":` + fmt.Sprint(total) + `,"bandcampUrls":"` + template.JSEscapeString(strings.Join(urls, "\\n")) + `"}' hx-trigger=\"load\"></div>`
	}
	return c.HTML(200, row+progressBar+triggerNext)
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
