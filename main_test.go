package main

import (
	"io"
	"millions-of-words/models"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type MockRenderer struct{}

func (m *MockRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	_, err := w.Write([]byte(name))
	return err
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_VAR", "test_value")
	assert.Equal(t, "test_value", getEnv("TEST_VAR", "default_value"))
	assert.Equal(t, "default_value", getEnv("NON_EXISTING_VAR", "default_value"))
}

func TestIndexHandler(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	albums = []models.BandcampAlbumData{
		{ID: "1", ArtistName: "Artist1", AlbumName: "Album1"},
		{ID: "2", ArtistName: "Artist2", AlbumName: "Album2"},
	}

	e.Renderer = &MockRenderer{}

	if assert.NoError(t, indexHandler(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "index.html", rec.Body.String())
	}
}

func TestSearchAlbumsHandler(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/search-albums?search=Artist1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	albums = []models.BandcampAlbumData{
		{ID: "1", ArtistName: "Artist1", AlbumName: "Album1"},
		{ID: "2", ArtistName: "Artist2", AlbumName: "Album2"},
	}

	e.Renderer = &MockRenderer{}

	if assert.NoError(t, searchAlbumsHandler(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "album-grid.html", rec.Body.String())
	}
}

func TestAlbumDetailsHandler(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/album/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")

	albums = []models.BandcampAlbumData{
		{ID: "1", ArtistName: "Artist1", AlbumName: "Album1"},
		{ID: "2", ArtistName: "Artist2", AlbumName: "Album2"},
	}

	e.Renderer = &MockRenderer{}

	if assert.NoError(t, albumDetailsHandler(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "album-details.html", rec.Body.String())
	}
}

func TestAllWordsHandler(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/all-words", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	albums = []models.BandcampAlbumData{
		{
			ID: "1",
			Tracks: []models.BandcampTrackData{
				{Lyrics: "test word test"},
			},
		},
	}

	e.Renderer = &MockRenderer{}

	if assert.NoError(t, allWordsHandler(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "all-words.html", rec.Body.String())
	}
}

func TestFilterAlbumsByQuery(t *testing.T) {
	albums = []models.BandcampAlbumData{
		{ID: "1", ArtistName: "Artist1", AlbumName: "Album1"},
		{ID: "2", ArtistName: "Artist2", AlbumName: "Album2"},
		{ID: "3", ArtistName: "Artist3", AlbumName: "TestAlbum"},
	}

	testCases := []struct {
		query    string
		expected int
	}{
		{"Artist1", 1},
		{"Album", 3},
		{"Test", 1},
		{"NonExistent", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.query, func(t *testing.T) {
			filtered := filterAlbumsByQuery(tc.query)
			assert.Equal(t, tc.expected, len(filtered))
		})
	}
}
