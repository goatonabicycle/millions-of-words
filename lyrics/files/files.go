package files

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func SaveAlbumDataToFile(albumName string, albumData interface{}) {

	artistDir := filepath.Join("./../data", albumName)
	if err := os.MkdirAll(artistDir, os.ModePerm); err != nil {
		log.Fatalf("Error creating directory: %v", err)
	}

	jsonData, err := json.MarshalIndent(albumData, "", "    ")
	if err != nil {
		log.Fatalf("Error marshaling album data to JSON: %v", err)
	}

	filePath := filepath.Join(artistDir, fmt.Sprintf("%s_album_data.json", albumName))
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		log.Fatalf("Error writing album data to file: %v", err)
	}

	fmt.Printf("Album data for %s successfully written to %s\n", albumName, filePath)
}
