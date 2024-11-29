package fetch

import (
	"fmt"
	"log"
	"millions-of-words/models"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func createMetalArchivesRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("DNT", "1")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", "https://www.metal-archives.com/")

	return req, nil
}

func convertMetalArchivesDate(rawDate string) string {
	re := regexp.MustCompile(`(\d+)(st|nd|rd|th)`)
	cleanDate := re.ReplaceAllString(rawDate, "$1")

	log.Printf("Clean date: %s", cleanDate)

	parts := strings.Split(cleanDate, ", ")
	if len(parts) != 2 { // Should be ["December 21", "2023"]
		return ""
	}

	monthDay := strings.Split(parts[0], " ")
	if len(monthDay) != 2 { // Should be ["December", "21"]
		return ""
	}

	month := monthDay[0]
	day := monthDay[1]
	year := parts[1]

	months := map[string]string{
		"January": "01", "February": "02", "March": "03", "April": "04",
		"May": "05", "June": "06", "July": "07", "August": "08",
		"September": "09", "October": "10", "November": "11", "December": "12",
	}

	monthNum := months[month]

	if len(day) == 1 {
		day = "0" + day
	}

	return fmt.Sprintf("%s-%s-%s", year, monthNum, day)
}

func FetchFromMetalArchives(url string) (models.BandcampAlbumData, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	bandURL := strings.Replace(url, "/albums/", "/bands/", 1) + "/"
	log.Printf("Fetching band data from: %s", bandURL)

	bandReq, err := createMetalArchivesRequest(bandURL)
	if err != nil {
		return models.BandcampAlbumData{}, err
	}

	resp, err := client.Do(bandReq)
	if err != nil {
		return models.BandcampAlbumData{}, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return models.BandcampAlbumData{}, err
	}

	country := strings.TrimSpace(doc.Find("#band_stats dt:contains('Country of origin:') + dd").Text())
	log.Printf("Found Country: '%s'", country)

	genre := strings.TrimSpace(doc.Find("#band_stats dt:contains('Genre:') + dd").Text())
	log.Printf("Found Genre: '%s'", genre)

	albumReq, err := createMetalArchivesRequest(url)
	if err != nil {
		return models.BandcampAlbumData{}, err
	}

	resp, err = client.Do(albumReq)
	if err != nil {
		return models.BandcampAlbumData{}, err
	}
	defer resp.Body.Close()

	doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return models.BandcampAlbumData{}, err
	}

	releaseDateRaw := strings.TrimSpace(doc.Find("dl.float_left dt:contains('Release date') + dd").Text())
	log.Printf("Found raw Release Date: '%s'", releaseDateRaw)
	releaseDate := convertMetalArchivesDate(releaseDateRaw)
	log.Printf("Converted Release Date: '%s'", releaseDate)

	label := strings.TrimSpace(doc.Find("dl.float_right dt:contains('Label') + dd").Text())
	log.Printf("Found Label: '%s'", label)

	return models.BandcampAlbumData{
		ReleaseDate: releaseDate,
		Genre:       genre,
		Country:     country,
		Label:       label,
	}, nil
}
