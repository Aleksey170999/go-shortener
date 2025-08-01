package service

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/Aleksey170999/go-shortener/internal/model"
	"github.com/Aleksey170999/go-shortener/internal/repository"

	"io"
)

type URLService struct {
	repo repository.URLRepository
}

func NewURLService(repo repository.URLRepository) *URLService {
	return &URLService{repo: repo}
}

func (s *URLService) Shorten(original string) (string, error) {
	id, err := generateID(6)
	if err != nil {
		return "", err
	}
	url := &model.URL{
		ID:       id,
		Original: original,
	}
	err = s.repo.Save(url)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *URLService) Resolve(id string) (string, error) {
	url, err := s.repo.FindByID(id)
	if err != nil {
		return "", err
	}
	return url.Original, nil
}

func generateID(n int) (string, error) {
	b := make([]byte, n)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b)[:n], nil
}
