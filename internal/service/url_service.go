package service

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"time"

	"github.com/Aleksey170999/go-shortener/internal/model"
	"github.com/Aleksey170999/go-shortener/internal/repository"
	"github.com/google/uuid"

	"io"
)

type deleteRequest struct {
	ShortURLs []string
	UserID    string
}

type URLService struct {
	repo        repository.URLRepository
	deleteReqCh chan deleteRequest
}

func NewURLService(repo repository.URLRepository) *URLService {
	s := &URLService{
		repo:        repo,
		deleteReqCh: make(chan deleteRequest, 100),
	}
	go s.deleteWorker()
	return s
}

func (s *URLService) deleteWorker() {
	batch := make([]deleteRequest, 0)
	batchSize := 50
	batchTimeout := 100
	for {
		select {
		case req := <-s.deleteReqCh:
			batch = append(batch, req)
			if len(batch) >= batchSize {
				s.flushBatch(batch)
				batch = batch[:0]
			}
		default:
			if len(batch) > 0 {
				s.flushBatch(batch)
				batch = batch[:0]
			}
			time.Sleep(time.Millisecond * time.Duration(batchTimeout))
		}
	}
}

func (s *URLService) flushBatch(batch []deleteRequest) {
	userURLs := make(map[string][]string)
	for _, req := range batch {
		userURLs[req.UserID] = append(userURLs[req.UserID], req.ShortURLs...)
	}
	for userID, urls := range userURLs {
		if err := s.repo.BatchDelete(urls, userID); err != nil {
			log.Printf("[flushBatch] batch delete error: %v", err)
		}
	}
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
	url, err = s.repo.Save(url)
	if err != nil {
		return url, err
	}
	return url, nil
}

func (s *URLService) Resolve(shortURL string) (*model.URL, error) {
	url, err := s.repo.GetByShortURL(shortURL)
	if err != nil {
		return nil, err
	}
	return url, nil
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

func (s *URLService) BatchDelete(shortURLs []string, userID string) error {
	s.deleteReqCh <- deleteRequest{ShortURLs: shortURLs, UserID: userID}
	return nil
}
