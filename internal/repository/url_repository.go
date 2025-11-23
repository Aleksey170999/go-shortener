package repository

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	"github.com/Aleksey170999/go-shortener/internal/config"
	db "github.com/Aleksey170999/go-shortener/internal/config/db"
	"github.com/Aleksey170999/go-shortener/internal/model"
	_ "github.com/jackc/pgx/v5"
	"github.com/lib/pq"
)

// URLRepository defines the interface for URL storage operations.
// Implementations must be safe for concurrent use by multiple goroutines.
type URLRepository interface {
	// Save stores a new URL or returns an existing one if the original URL already exists.
	// Returns the saved URL and any error encountered.
	Save(url *model.URL) (*model.URL, error)

	// GetByShortURL retrieves a URL by its short identifier.
	// Returns ErrNotFound if no URL with the given short identifier exists.
	GetByShortURL(shortURL string) (*model.URL, error)

	// GetByUserID retrieves all URLs created by a specific user.
	// Returns an empty slice if no URLs are found for the user.
	GetByUserID(userID string) ([]model.URL, error)

	// BatchDelete marks multiple URLs as deleted for a specific user.
	// This is a soft delete operation that sets the IsDeleted flag on the URLs.
	// ShortURLs that don't belong to the user or don't exist are silently ignored.
	BatchDelete(shortURLs []string, userID string) error
}

// memoryURLRepository is an in-memory implementation of URLRepository.
// It stores URLs in a map and is safe for concurrent access.
type memoryURLRepository struct {
	data map[string]*model.URL
	mu   sync.RWMutex
}

// DataBaseURLRepository is a PostgreSQL implementation of URLRepository.
// It stores URLs in a PostgreSQL database and handles all SQL operations.
type DataBaseURLRepository struct {
	DB *sql.DB
}

// NewMemoryURLRepository creates a new in-memory URL repository.
// This implementation is primarily used for testing and development.
//
// Returns:
//   - *memoryURLRepository: A new instance of in-memory URL repository
func NewMemoryURLRepository() *memoryURLRepository {
	repo := memoryURLRepository{
		data: make(map[string]*model.URL),
	}
	return &repo
}

// NewDataBaseURLRepository creates a new PostgreSQL URL repository.
// It establishes a connection to the database using the provided configuration.
//
// Parameters:
//   - cfg: Application configuration containing database connection details
//
// Returns:
//   - *DataBaseURLRepository: A new instance of database URL repository
//   - error: If database connection fails
func NewDataBaseURLRepository(cfg *config.Config) *DataBaseURLRepository {
	dbCon, err := sql.Open("postgres", cfg.DatabaseDSN)
	if err != nil {
		fmt.Println(err)
	}
	repo := DataBaseURLRepository{
		DB: dbCon,
	}

	db.ApplyMigrations(dbCon)
	return &repo
}

// Save stores a URL in the in-memory repository.
// If a URL with the same original URL already exists, it returns the existing URL.
//
// Implements URLRepository interface.
func (r *memoryURLRepository) Save(url *model.URL) (*model.URL, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[url.Short] = url
	return url, nil
}

// GetByShortURL retrieves a URL by its short identifier from memory.
// Returns ErrNotFound if no URL with the given ID exists.
//
// Implements URLRepository interface.
func (r *memoryURLRepository) GetByShortURL(id string) (*model.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	url, exists := r.data[id]
	if !exists {
		return nil, fmt.Errorf("url not found: %w", ErrNotFound)
	}
	return url, nil
}

// GetByUserID retrieves all URLs created by a specific user from memory.
// Returns an empty slice if no URLs are found for the user.
//
// Implements URLRepository interface.
func (r *memoryURLRepository) GetByUserID(userID string) ([]model.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var userURLs []model.URL
	for _, url := range r.data {
		if url.UserID == userID {
			userURLs = append(userURLs, *url)
		}
	}

	if len(userURLs) == 0 {
		return nil, ErrNotFound
	}

	return userURLs, nil
}

// BatchDelete marks multiple URLs as deleted for a specific user in memory.
// This is a soft delete operation that sets the IsDeleted flag on the URLs.
// ShortURLs that don't belong to the user or don't exist are silently ignored.
// Implements URLRepository interface with in-memory implementation.
func (r *memoryURLRepository) BatchDelete(shortURLs []string, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, short := range shortURLs {
		if url, exists := r.data[short]; exists {
			if url.UserID == userID && !url.IsDeleted {
				url.IsDeleted = true
				r.data[short] = url
			}
		}
	}

	return nil
}

// Save stores a URL in the database.
// If a URL with the same original URL already exists, it returns the existing URL.
// Implements URLRepository interface with PostgreSQL-specific implementation.
func (r *DataBaseURLRepository) Save(url *model.URL) (*model.URL, error) {
	var isConflict bool
	insertSQL := `WITH inserted AS (
						INSERT INTO urls (id, short_url, original_url, user_id)
						VALUES ($1, $2, $3, $4)
						ON CONFLICT (original_url) DO NOTHING
						RETURNING *
					)
					select id, short_url, false as is_conflict FROM inserted
					UNION
					SELECT id, short_url, true as is_conflict FROM urls 
					WHERE original_url = $3 AND NOT EXISTS (SELECT 1 FROM inserted)`
	err := r.DB.QueryRow(insertSQL, url.ID, url.Short, url.Original, url.UserID).
		Scan(&url.ID, &url.Short, &isConflict)

	if err != nil {
		return nil, err
	}
	if isConflict {
		return url, model.ErrURLAlreadyExists
	}
	return url, nil
}

// GetByShortURL retrieves a URL by its short identifier from the database.
// Returns ErrNotFound if no URL with the given ID exists.
// Implements URLRepository interface with PostgreSQL-specific implementation.
func (r *DataBaseURLRepository) GetByShortURL(id string) (*model.URL, error) {
	var url model.URL
	err := r.DB.QueryRow("SELECT id, short_url, original_url, user_id, is_deleted FROM urls WHERE short_url = $1", id).
		Scan(&url.ID, &url.Short, &url.Original, &url.UserID, &url.IsDeleted)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("url not found: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get url: %w", err)
	}
	return &url, nil
}

// GetByUserID retrieves all URLs created by a specific user from the database.
// Returns an empty slice if no URLs are found for the user.
// Implements URLRepository interface with PostgreSQL-specific implementation.
func (r *DataBaseURLRepository) GetByUserID(userID string) ([]model.URL, error) {
	rows, err := r.DB.Query("SELECT id, short_url, original_url, user_id FROM urls WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user urls: %w", err)
	}
	defer rows.Close()

	var urls []model.URL
	for rows.Next() {
		var url model.URL
		if err := rows.Scan(&url.ID, &url.Short, &url.Original, &url.UserID); err != nil {
			return nil, fmt.Errorf("failed to scan url: %w", err)
		}
		urls = append(urls, url)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating urls: %w", err)
	}

	if len(urls) == 0 {
		return nil, ErrNotFound
	}

	return urls, nil
}

// BatchDelete marks multiple URLs as deleted for a specific user in the database.
// This is a soft delete operation that sets the is_deleted flag on the URLs.
// ShortURLs that don't belong to the user or don't exist are silently ignored.
// Implements URLRepository interface with PostgreSQL-specific implementation.
func (r *DataBaseURLRepository) BatchDelete(shortURLs []string, userID string) error {
	if len(shortURLs) == 0 {
		return nil
	}
	query := `UPDATE urls SET is_deleted = TRUE WHERE short_url = ANY($1) AND user_id = $2`
	_, err := r.DB.Exec(query, pq.Array(shortURLs), userID)
	if err != nil {
		log.Printf("BatchDelete error: %v", err)
		return err
	}
	return nil
}

// ErrNotFound is a singleton instance of NotFoundError that is returned
// when a requested resource is not found in the repository.
// It should be used for all "not found" error returns to ensure consistency.
var ErrNotFound = &NotFoundError{}

// NotFoundError is returned when a requested resource is not found.
// It implements the error interface.
type NotFoundError struct{}

// Error returns the string representation of the NotFoundError.
// This method makes NotFoundError implement the error interface.
//
// Returns:
//   - string: The error message indicating the URL was not found
func (e *NotFoundError) Error() string {
	return "url not found"
}
