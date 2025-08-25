package service

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"os"
	"sync"

	"github.com/Aleksey170999/go-shortener/internal/model"
	"github.com/Aleksey170999/go-shortener/internal/repository"
	"github.com/google/uuid"

	"io"
)

var mu sync.Mutex

func AppendURLRecord(filename string, url *model.URL) error {
	mu.Lock()
	defer mu.Unlock()

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(&url); err != nil {
		return err
	}

	return nil
}

type URLService struct {
	repo repository.URLRepository
}

func NewURLService(repo repository.URLRepository) *URLService {
	return &URLService{repo: repo}
}

func (s *URLService) Shorten(original string, storageFilePath string) (string, error) {
	shortURL, err := generateShortURL(6)
	if err != nil {
		return "", err
	}
	url := &model.URL{
		ID:       uuid.New().String(),
		Original: original,
		Short:    shortURL,
	}
	err = s.repo.Save(url)
	AppendURLRecord(storageFilePath, url)
	if err != nil {
		return "", err
	}
	return shortURL, nil
}

func (s *URLService) Resolve(shortURL string) (string, error) {
	url, err := s.repo.FindByShortURL(shortURL)
	if err != nil {
		return "", err
	}
	return url.Original, nil
}

func generateShortURL(n int) (string, error) {
	b := make([]byte, n)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b)[:n], nil
}
