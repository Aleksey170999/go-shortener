package storage

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/Aleksey170999/go-shortener/internal/model"
	"github.com/Aleksey170999/go-shortener/internal/repository"
)

type Storage struct {
	FilePath string
	mu       sync.Mutex
}

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
		if err := repo.Save(&urls[i]); err != nil {
			return err
		}
	}

	return nil
}

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

func NewStorage(filePath string) *Storage {
	return &Storage{
		FilePath: filePath,
		mu:       sync.Mutex{},
	}
}
