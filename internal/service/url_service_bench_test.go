package service

import (
	"sync"
	"testing"

	"github.com/Aleksey170999/go-shortener/internal/model"
	"github.com/Aleksey170999/go-shortener/internal/repository"
	"github.com/stretchr/testify/require"
)

// memoryURLRepository is a copy of the private memoryURLRepository from repository package
type memoryURLRepository struct {
	data map[string]*model.URL
	mu   sync.RWMutex
}

func newMemoryURLRepository() *memoryURLRepository {
	return &memoryURLRepository{
		data: make(map[string]*model.URL),
	}
}

func (r *memoryURLRepository) Save(url *model.URL) (*model.URL, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data[url.Short] = url
	return url, nil
}

func (r *memoryURLRepository) GetByShortURL(shortURL string) (*model.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	url, exists := r.data[shortURL]
	if !exists {
		return nil, repository.ErrNotFound
	}
	return url, nil
}

func (r *memoryURLRepository) GetByUserID(userID string) ([]model.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var urls []model.URL
	for _, url := range r.data {
		if url.UserID == userID {
			urls = append(urls, *url)
		}
	}

	return urls, nil
}

func (r *memoryURLRepository) BatchDelete(shortURLs []string, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, shortURL := range shortURLs {
		if url, exists := r.data[shortURL]; exists && url.UserID == userID {
			url.IsDeleted = true
			r.data[shortURL] = url
		}
	}

	return nil
}

func BenchmarkURLService_Shorten(b *testing.B) {
	repo := newMemoryURLRepository()
	service := NewURLService(repo)
	userID := "test-user"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.Shorten("https://example.com", "", userID)
		require.NoError(b, err)
	}
}

func BenchmarkURLService_Resolve(b *testing.B) {
	repo := newMemoryURLRepository()
	service := NewURLService(repo)
	userID := "test-user"

	// Pre-populate with test data
	urls := make([]*model.URL, 1000)
	for i := 0; i < 1000; i++ {
		url, err := service.Shorten("https://example.com", "", userID)
		if err != nil {
			b.Fatal(err)
		}
		urls[i] = url
	}

	shortURLs := make([]string, len(urls))
	for i, url := range urls {
		shortURLs[i] = url.Short
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.Resolve(shortURLs[i%len(shortURLs)])
		require.NoError(b, err)
	}
}

func BenchmarkURLService_GetUserURLs(b *testing.B) {
	repo := newMemoryURLRepository()
	service := NewURLService(repo)
	userID := "test-user"

	// Pre-populate with test data
	for i := 0; i < 1000; i++ {
		_, err := service.Shorten("https://example.com", "", userID)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetUserURLs(userID)
		require.NoError(b, err)
	}
}

func BenchmarkURLService_BatchDelete(b *testing.B) {
	repo := newMemoryURLRepository()
	service := NewURLService(repo)
	userID := "test-user"

	// Pre-populate with test data
	urls := make([]*model.URL, 1000)
	shortURLs := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		url, err := service.Shorten("https://example.com", "", userID)
		if err != nil {
			b.Fatal(err)
		}
		urls[i] = url
		shortURLs[i] = url.Short
	}

	batchSize := 100
	batches := make([][]string, 0, (len(shortURLs)+batchSize-1)/batchSize)

	for batchSize < len(shortURLs) {
		shortURLs, batches = shortURLs[batchSize:], append(batches, shortURLs[0:batchSize:batchSize])
	}
	batches = append(batches, shortURLs)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batch := batches[i%len(batches)]
		err := service.BatchDelete(batch, userID)
		require.NoError(b, err)
	}
}
