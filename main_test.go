package main

import (
	"encoding/json"
	"io/ioutil"
	"millions-of-words/models"
	"os"
	"testing"
)

func TestLoadAlbumsDataFromJsonFiles(t *testing.T) {
	// Setup: Create a temporary directory
	tempDir, err := ioutil.TempDir("", "albums")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up

	// Prepare a sample album data
	sampleAlbum := models.BandcampAlbumData{
		ArtistName: "Test Artist",
		AlbumName:  "Test Album",
	}
	sampleData, err := json.Marshal(sampleAlbum)
	if err != nil {
		t.Fatalf("Failed to marshal sample album data: %v", err)
	}

	// Write sample data to a temporary file
	tempFile, err := ioutil.TempFile(tempDir, "*.json")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	if _, err := tempFile.Write(sampleData); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temporary file: %v", err)
	}

	// Point the function to the temporary directory
	oldDataDir := "data"                    // Backup the original data directory
	dataDir = tempDir                       // Use the temporary directory for the test
	defer func() { dataDir = oldDataDir }() // Restore after test

	// Execute the function
	loadAlbumsDataFromJsonFiles()

	// Test: Verify the albums variable contains the loaded data
	if len(albums) != 1 || albums[0].ArtistName != "Test Artist" || albums[0].AlbumName != "Test Album" {
		t.Errorf("Expected albums to contain the loaded sample data")
	}
}
