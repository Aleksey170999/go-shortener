package repository

import (
	"sync"

	"github.com/Aleksey170999/go-shortener/internal/model"
)

type URLRepository interface {
	Save(url *model.URL) error
	FindByShortURL(shortURL string) (*model.URL, error)
}

type memoryURLRepository struct {
	data map[string]*model.URL
	mu   sync.RWMutex
}

func NewMemoryURLRepository() *memoryURLRepository {
	repo := memoryURLRepository{
		data: make(map[string]*model.URL),
	}
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
