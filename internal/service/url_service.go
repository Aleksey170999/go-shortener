package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
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

// URLService provides high-level operations for URL shortening and management.
// It handles business logic and coordinates with the repository layer for data persistence.
// URLService is safe for concurrent use by multiple goroutines.
type URLService struct {
	repo        repository.URLRepository // Underlying repository for data access
	deleteReqCh chan deleteRequest       // Channel for asynchronous delete operations
}

// NewURLService creates a new instance of URLService with the provided repository.
// It initializes the background worker for processing batch delete operations.
// The repository parameter must not be nil.
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

// Ping DataBase
func (s *URLService) PingDB() error {
	// Check if the repository is a database repository
	dbRepo, ok := s.repo.(*repository.DataBaseURLRepository)
	if !ok {
		return fmt.Errorf("database repository not available")
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Ping the database
	return dbRepo.DB.PingContext(ctx)
}

// Shorten creates a new shortened URL for the given original URL.
// If the original URL already exists in the repository, the existing short URL is returned.
// Parameters:
//   - original: The original URL to be shortened
//   - id: Optional custom ID for the short URL. If empty, a random string will be generated.
//   - userID: ID of the user creating the short URL
//
// Returns:
//   - *model.URL: The created or existing URL object
//   - error: Non-nil if an error occurs during the operation
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

// Resolve retrieves the original URL for a given short URL.
// Returns model.ErrNotFound if no URL with the given short code exists.
//
// Parameters:
//   - shortURL: The short URL code to resolve
//
// Returns:
//   - *model.URL: The URL object containing the original URL
//   - error: Non-nil if the URL is not found or an error occurs
func (s *URLService) Resolve(shortURL string) (*model.URL, error) {
	url, err := s.repo.GetByShortURL(shortURL)
	if err != nil {
		return nil, err
	}
	return url, nil
}

// GetUserURLs retrieves all URLs created by a specific user.
// Returns an empty slice if the user has no URLs.
//
// Parameters:
//   - userID: The ID of the user
//
// Returns:
//   - []model.URL: A slice of URLs created by the user
//   - error: Non-nil if an error occurs during the operation
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

// BatchDelete schedules URLs for deletion in a background worker.
// This is an asynchronous operation that marks URLs as deleted without blocking.
// Only URLs belonging to the specified user will be deleted.
//
// Parameters:
//   - shortURLs: A slice of short URL codes to delete
//   - userID: The ID of the user performing the deletion
//
// Returns:
//   - error: Always returns nil as the operation is asynchronous
//     (errors are logged but not returned to the caller)
func (s *URLService) BatchDelete(shortURLs []string, userID string) error {
	s.deleteReqCh <- deleteRequest{ShortURLs: shortURLs, UserID: userID}
	return nil
}
