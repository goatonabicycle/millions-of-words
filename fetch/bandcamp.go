package fetch

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"millions-of-words/models"

	"github.com/PuerkitoBio/goquery"
)

func FetchFromBandcamp(url string) (models.BandcampAlbumData, error) {
	log.Printf("Fetching album data from Bandcamp for URL: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		return models.BandcampAlbumData{}, fmt.Errorf("error fetching Bandcamp page: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return models.BandcampAlbumData{}, fmt.Errorf("error parsing HTML: %w", err)
	}

	artistName := doc.Find("#name-section h3 span a").Text()
	albumName := doc.Find(".trackTitle").First().Text()
	imageUrl := doc.Find("a.popupImage").AttrOr("href", "")

	imageData, err := fetchImageData(imageUrl)
	if err != nil {
		log.Printf("Failed to fetch album image: %v", err)
	}

	tracklist, totalAlbumDuration := processTracklist(doc)

	albumColor, err := calculateAverageColor(imageData)
	if err != nil {
		log.Printf("Failed to calculate average color: %v", err)
		albumColor = "#000000"
	}

	return models.BandcampAlbumData{
		ID:                strings.TrimSpace(artistName) + " - " + strings.TrimSpace(albumName),
		Slug:              generateSlug(strings.TrimSpace(artistName) + " - " + strings.TrimSpace(albumName)),
		ArtistName:        strings.TrimSpace(artistName),
		AlbumName:         strings.TrimSpace(albumName),
		ImageUrl:          imageUrl,
		ImageData:         imageData,
		AlbumColorAverage: albumColor,
		Tracks:            tracklist,
		TotalLength:       int(totalAlbumDuration.Seconds()),
		FormattedLength:   formatDuration(int(totalAlbumDuration.Seconds())),
		BandcampUrl:       url,
		DateAdded:         time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

func processTracklist(doc *goquery.Document) ([]models.BandcampTrackData, time.Duration) {
	var tracklist []models.BandcampTrackData
	var totalAlbumDuration time.Duration

	trackNumber := 1
	doc.Find("tr.track_row_view").Each(func(i int, s *goquery.Selection) {
		trackTitle := s.Find(".title-col .track-title").Text()
		trackDurationStr := strings.TrimSpace(s.Find(".title-col .time").Text())

		if strings.TrimSpace(trackTitle) == "" || trackDurationStr == "" {
			return
		}

		trackDuration, err := parseTrackDuration(trackDurationStr)
		if err != nil {
			return
		}

		totalAlbumDuration += trackDuration

		lyrics := strings.TrimSpace(s.Next().Find("div").Text())
		if strings.HasPrefix(lyrics, "lyrics") || strings.Contains(lyrics, "buy track") {
			lyrics = ""
		}

		track := models.BandcampTrackData{
			Name:            strings.TrimSpace(trackTitle),
			TrackNumber:     trackNumber,
			TotalLength:     int(trackDuration.Seconds()),
			FormattedLength: formatDuration(int(trackDuration.Seconds())),
			Lyrics:          lyrics,
			IgnoredWords:    "",
		}

		tracklist = append(tracklist, track)
		trackNumber++
	})

	return tracklist, totalAlbumDuration
}

func parseTrackDuration(durationStr string) (time.Duration, error) {
	parts := strings.Split(durationStr, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid duration format")
	}

	minutes := 0
	seconds := 0
	fmt.Sscanf(parts[0], "%d", &minutes)
	fmt.Sscanf(parts[1], "%d", &seconds)

	if seconds >= 60 {
		return 0, fmt.Errorf("invalid seconds value")
	}

	return time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second, nil
}

func fetchImageData(imageUrl string) ([]byte, error) {
	if imageUrl == "" {
		return nil, fmt.Errorf("no image URL provided")
	}

	resp, err := http.Get(imageUrl)
	if err != nil {
		return nil, fmt.Errorf("error fetching image: %w", err)
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func calculateAverageColor(imageData []byte) (string, error) {
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return "", fmt.Errorf("error decoding image: %w", err)
	}

	var rSum, gSum, bSum, pixelCount uint64

	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			rSum += uint64(r)
			gSum += uint64(g)
			bSum += uint64(b)
			pixelCount++
		}
	}

	avgR := uint8(rSum / pixelCount / 256)
	avgG := uint8(gSum / pixelCount / 256)
	avgB := uint8(bSum / pixelCount / 256)

	return fmt.Sprintf("#%02x%02x%02x", avgR, avgG, avgB), nil
}

var multipleHyphens = regexp.MustCompile(`-+`)

func generateSlug(id string) string {
	slug := strings.ToLower(id)
	slug = strings.ReplaceAll(slug, "&", "and")
	slug = strings.ReplaceAll(slug, ",", "")
	slug = strings.ReplaceAll(slug, ".", "")
	slug = strings.ReplaceAll(slug, ";", "")
	slug = strings.ReplaceAll(slug, ":", "")
	slug = strings.ReplaceAll(slug, "'", "")
	slug = strings.ReplaceAll(slug, "(", "")
	slug = strings.ReplaceAll(slug, ")", "")
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = multipleHyphens.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

func formatDuration(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	seconds = seconds % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
