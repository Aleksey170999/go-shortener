package repository

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/Aleksey170999/go-shortener/internal/model"
)

func loadFromStorage(repo *memoryURLRepository, storagePath string) {
	data, err := os.ReadFile(storagePath)
	if err != nil {
		return
	}

	var urls []model.URL
	err = json.Unmarshal(data, &urls)
	if err != nil {
		return
	}
	for _, url := range urls {
		repo.Save(&url)
	}
}

type URLRepository interface {
	Save(url *model.URL) error
	FindByShortURL(shortURL string) (*model.URL, error)
}

type memoryURLRepository struct {
	data map[string]*model.URL
	mu   sync.RWMutex
}

func NewMemoryURLRepository(storagePath string) *memoryURLRepository {
	repo := memoryURLRepository{
		data: make(map[string]*model.URL),
	}
	loadFromStorage(&repo, storagePath)
	return &repo
}

func (r *memoryURLRepository) Save(url *model.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[url.Short] = url
	return nil
}

func (r *memoryURLRepository) FindByShortURL(id string) (*model.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	url, ok := r.data[id]
	if !ok {
		return nil, ErrNotFound
	}
	return url, nil
}

var ErrNotFound = &NotFoundError{}

type NotFoundError struct{}

func (e *NotFoundError) Error() string {
	return "url not found"
}
