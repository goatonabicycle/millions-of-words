package loader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"millions-of-words/models"
	"millions-of-words/words"
	"os"
	"strings"

	"github.com/supabase-community/postgrest-go"
	storageClient "github.com/supabase-community/storage-go"
	supa "github.com/supabase-community/supabase-go"
)

type SupabaseConfig struct {
	URL        string `json:"supabase_url"`
	ServiceKey string `json:"service_role_key"`
	AnonKey    string `json:"supabase_key"`
}

var (
	publicClient *supa.Client
	adminClient  *supa.Client
)

func init() {
	config, err := loadConfig("auth.json")
	if err != nil {
		log.Fatalf("Failed to load Supabase config: %v", err)
	}

	publicClient, err = supa.NewClient(config.URL, config.AnonKey, nil)
	if err != nil {
		log.Fatalf("Failed to create public Supabase client: %v", err)
	}

	adminClient, err = supa.NewClient(config.URL, config.ServiceKey, nil)
	if err != nil {
		log.Fatalf("Failed to create admin Supabase client: %v", err)
	}
}

func loadConfig(path string) (SupabaseConfig, error) {
	if url := os.Getenv("SUPABASE_URL"); url != "" {
		return SupabaseConfig{
			URL:        url,
			ServiceKey: os.Getenv("SUPABASE_SERVICE_KEY"),
			AnonKey:    os.Getenv("SUPABASE_ANON_KEY"),
		}, nil
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return SupabaseConfig{}, fmt.Errorf("error reading config file: %w", err)
	}

	var config SupabaseConfig
	if err := json.Unmarshal(file, &config); err != nil {
		return SupabaseConfig{}, fmt.Errorf("error parsing config: %w", err)
	}

	return config, nil
}

func LoadAlbumsData(limit ...int) ([]models.BandcampAlbumData, error) {
	albums, err := fetchAlbums(limit...)
	if err != nil {
		return nil, err
	}

	for i := range albums {
		if err := fetchTracks(&albums[i]); err != nil {
			log.Printf("Error fetching tracks for album %s: %v", albums[i].ID, err)
			continue
		}
		calculateAlbumMetrics(&albums[i])
	}
	return albums, nil
}

func fetchAlbums(limit ...int) ([]models.BandcampAlbumData, error) {
	query := publicClient.From("albums").
		Select("*", "exact", false).
		Order("date_added", &postgrest.OrderOpts{
			Ascending:  false,
			NullsFirst: false,
		})

	if len(limit) > 0 && limit[0] > 0 {
		query = query.Limit(limit[0], "")
	}

	data, _, err := query.Execute()
	if err != nil {
		return nil, fmt.Errorf("error querying albums: %w", err)
	}

	var albums []models.BandcampAlbumData
	if err := json.Unmarshal(data, &albums); err != nil {
		return nil, fmt.Errorf("error scanning album rows: %w", err)
	}

	for i := range albums {
		if albums[i].ImageStoragePath != "" {
			publicURL := adminClient.Storage.GetPublicUrl("album-covers", albums[i].ImageStoragePath)
			albums[i].ImageUrl = publicURL.SignedURL
			log.Printf("Album: %s, Storage Path: %s, URL: %s",
				albums[i].AlbumName,
				albums[i].ImageStoragePath,
				albums[i].ImageUrl)
		}
	}

	return albums, nil
}

func fetchTracks(album *models.BandcampAlbumData) error {
	data, _, err := publicClient.From("tracks").
		Select("name, total_length, formatted_length, lyrics", "exact", false).
		Eq("album_id", album.ID).
		Execute()

	if err != nil {
		return fmt.Errorf("error querying tracks: %w", err)
	}

	var tracks []models.BandcampTrackData
	if err := json.Unmarshal(data, &tracks); err != nil {
		return fmt.Errorf("error scanning track rows: %w", err)
	}

	album.Tracks = tracks
	return nil
}

func ValidateAuthKey(key string) (bool, error) {
	data, _, err := adminClient.From("auth_keys").
		Select("*", "exact", false).
		Eq("key", key).
		Execute()
	if err != nil {
		return false, fmt.Errorf("error checking auth key: %w", err)
	}

	var results []map[string]interface{}
	if err := json.Unmarshal(data, &results); err != nil {
		return false, fmt.Errorf("error scanning auth key result: %w", err)
	}

	return len(results) > 0, nil
}

func UpdateTrackLyrics(req models.UpdateLyricsRequest) error {
	valid, err := ValidateAuthKey(req.AuthKey)
	if err != nil {
		return fmt.Errorf("error validating auth: %w", err)
	}
	if !valid {
		return fmt.Errorf("invalid auth key")
	}

	cleanLyrics := strings.TrimSpace(req.Lyrics)
	if strings.HasPrefix(strings.ToLower(cleanLyrics), "lyrics") {
		cleanLyrics = ""
	}

	log.Printf("Updating lyrics for album: %s, track: %s", req.AlbumID, req.TrackName)

	_, count, err := adminClient.From("tracks").
		Update(map[string]interface{}{"lyrics": cleanLyrics}, "tracks", "id").
		Eq("album_id", req.AlbumID).
		Eq("name", req.TrackName).
		Execute()
	if err != nil {
		return fmt.Errorf("error updating lyrics: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("no matching track found")
	}

	log.Printf("Successfully updated lyrics for album: %s, track: %s", req.AlbumID, req.TrackName)
	return nil
}

func GetAlbumBySlug(slug string) (models.BandcampAlbumData, error) {
	data, _, err := publicClient.From("albums").
		Select("*", "exact", false).
		Eq("slug", slug).
		Single().
		Execute()
	if err != nil {
		return models.BandcampAlbumData{}, fmt.Errorf("error fetching album: %w", err)
	}

	var album models.BandcampAlbumData
	if err := json.Unmarshal(data, &album); err != nil {
		return models.BandcampAlbumData{}, fmt.Errorf("error scanning album data: %w", err)
	}

	if album.ImageStoragePath != "" {
		publicURL := adminClient.Storage.GetPublicUrl("album-covers", album.ImageStoragePath)
		album.ImageUrl = publicURL.SignedURL
	}

	if err := fetchTracks(&album); err != nil {
		return models.BandcampAlbumData{}, fmt.Errorf("error fetching tracks: %w", err)
	}

	calculateAlbumMetrics(&album)
	return album, nil
}

func AlbumUrlExists(url string) (bool, error) {
	data, _, err := publicClient.From("albums").
		Select("id", "exact", false).
		Eq("bandcamp_url", url).
		Execute()
	if err != nil {
		return false, fmt.Errorf("error checking if URL exists: %w", err)
	}

	var results []map[string]interface{}
	if err := json.Unmarshal(data, &results); err != nil {
		return false, fmt.Errorf("error scanning results: %w", err)
	}

	return len(results) > 0, nil
}

func SaveAlbum(album models.BandcampAlbumData) error {
	var storagePath string
	if len(album.ImageData) > 0 {
		filename := fmt.Sprintf("%s.jpg", album.ID)
		imageReader := bytes.NewReader(album.ImageData)

		contentType := "image/jpeg"
		fileOpts := storageClient.FileOptions{
			ContentType: &contentType,
		}

		_, err := adminClient.Storage.UploadFile("album-covers", filename, imageReader, fileOpts)
		if err != nil {
			return fmt.Errorf("error uploading image to storage: %w", err)
		}
		storagePath = filename
	}

	albumData := map[string]interface{}{
		"id":                  album.ID,
		"slug":                album.Slug,
		"artist_name":         album.ArtistName,
		"album_name":          album.AlbumName,
		"image_url":           album.ImageUrl,
		"image_storage_path":  storagePath,
		"bandcamp_url":        album.BandcampUrl,
		"album_color_average": album.AlbumColorAverage,
		"total_length":        album.TotalLength,
		"formatted_length":    album.FormattedLength,
		"date_added":          album.DateAdded,
	}

	_, _, err := adminClient.From("albums").
		Insert(albumData, false, "albums", "id", "").
		Execute()
	if err != nil {
		return fmt.Errorf("error inserting album: %w", err)
	}

	for _, track := range album.Tracks {
		trackData := map[string]interface{}{
			"album_id":         album.ID,
			"name":             track.Name,
			"total_length":     track.TotalLength,
			"formatted_length": track.FormattedLength,
			"lyrics":           track.Lyrics,
		}

		_, _, err = adminClient.From("tracks").
			Insert(trackData, false, "tracks", "id", "").
			Execute()
		if err != nil {
			return fmt.Errorf("error inserting track: %w", err)
		}
	}

	return nil
}

func calculateAlbumMetrics(album *models.BandcampAlbumData) {
	totalWords := 0
	totalVowels := 0
	totalConsonants := 0
	totalChars := 0
	totalLines := 0
	uniqueWords := make(map[string]struct{})
	wordLengths := make(map[int]int)

	for i, track := range album.Tracks {
		wordCounts, vowels, consonants, lengths := words.CalculateAndSortWordFrequencies(track.Lyrics)
		words := len(strings.Fields(track.Lyrics))

		totalWords += words
		totalVowels += vowels
		totalConsonants += consonants
		totalChars += len(track.Lyrics)
		totalLines += len(strings.Split(track.Lyrics, "\n"))

		for l, c := range lengths {
			wordLengths[l] += c
		}
		for _, wc := range wordCounts {
			uniqueWords[wc.Word] = struct{}{}
		}

		track.TotalWords = words
		track.TotalCharacters = len(track.Lyrics)
		track.TotalCharactersNoSpaces = len(strings.ReplaceAll(strings.ReplaceAll(track.Lyrics, " ", ""), "\n", ""))
		track.TotalLines = len(strings.Split(track.Lyrics, "\n"))
		album.Tracks[i] = track
	}

	album.TotalWords = totalWords
	album.TotalCharacters = totalChars
	album.TotalLines = totalLines
	album.TotalVowelCount = totalVowels
	album.TotalConsonantCount = totalConsonants
	album.TotalUniqueWords = len(uniqueWords)
	album.WordLengthDistribution = wordLengths
	album.AverageWordsPerTrack = calculateAverage(totalWords, len(album.Tracks))
}

func calculateAverage(total, count int) int {
	if count > 0 {
		return total / count
	}
	return 0
}
