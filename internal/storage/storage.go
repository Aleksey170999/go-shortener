package storage

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/Aleksey170999/go-shortener/internal/model"
	"github.com/Aleksey170999/go-shortener/internal/repository"
)

// Storage provides file-based persistence for URLs.
// It handles reading from and writing to a JSON file in a thread-safe manner.
type Storage struct {
	FilePath string
	mu       sync.Mutex
}

// LoadFromStorage reads URLs from the storage file and loads them into the provided repository.
// If the storage file doesn't exist, it returns without an error.
//
// Parameters:
//   - repo: The URLRepository where the loaded URLs will be stored
//
// Returns:
//   - error: If there's an error reading or parsing the storage file
func (s *Storage) LoadFromStorage(repo repository.URLRepository) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if len(data) == 0 {
		return nil
	}

	var urls []model.URL
	if err := json.Unmarshal(data, &urls); err != nil {
		return err
	}

	for i := range urls {
		_, err := repo.Save(&urls[i])
		if err != nil {
			return err
		}
	}

	return nil
}

// LoadToStorage adds a URL to the storage file.
// If the file doesn't exist, it will be created.
// The URLs are stored as a JSON array with pretty-printed formatting.
//
// Parameters:
//   - url: The URL to be stored
//
// Returns:
//   - error: If there's an error reading, writing, or parsing the storage file
func (s *Storage) LoadToStorage(url *model.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var urls []model.URL

	data, err := os.ReadFile(s.FilePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if len(data) > 0 {
		if err := json.Unmarshal(data, &urls); err != nil {
			return err
		}
	}

	urls = append(urls, *url)

	newData, err := json.MarshalIndent(urls, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.FilePath, newData, 0644)
}

// NewStorage creates a new Storage instance with the specified file path.
// The file will be created if it doesn't exist when LoadToStorage is first called.
//
// Parameters:
//   - filePath: Path to the JSON file where URLs will be stored
//
// Returns:
//   - *Storage: A new Storage instance
func NewStorage(filePath string) *Storage {
	return &Storage{
		FilePath: filePath,
		mu:       sync.Mutex{},
	}
}
