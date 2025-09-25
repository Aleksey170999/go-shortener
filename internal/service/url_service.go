package service

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/Aleksey170999/go-shortener/internal/model"
	"github.com/Aleksey170999/go-shortener/internal/repository"
	"github.com/google/uuid"

	"io"
)

type URLService struct {
	repo repository.URLRepository
}

func NewURLService(repo repository.URLRepository) *URLService {
	return &URLService{repo: repo}
}

func (s *URLService) Shorten(original, id, userID string) (*model.URL, error) {
	shortURL, err := generateShortURL(6)
	if err != nil {
		return nil, err
	}
	var recID string
	if id == "" {
		recID = uuid.New().String()
	} else {
		recID = id
	}
	url := &model.URL{
		ID:       recID,
		Original: original,
		Short:    shortURL,
		UserID:   userID,
	}
	err = s.repo.Save(url)
	if err != nil {
		return url, err
	}
	return url, nil
}

func (s *URLService) Resolve(shortURL string) (string, error) {
	url, err := s.repo.FindByShortURL(shortURL)
	if err != nil {
		return "", err
	}
	return url.Original, nil
}

func (s *URLService) GetUserURLs(userID string) ([]model.URL, error) {
	return s.repo.GetByUserID(userID)
}

func generateShortURL(n int) (string, error) {
	b := make([]byte, n)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b)[:n], nil
}
