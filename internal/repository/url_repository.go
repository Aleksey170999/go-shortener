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

type URLRepository interface {
	Save(url *model.URL) (*model.URL, error)
	GetByShortURL(shortURL string) (*model.URL, error)
	GetByUserID(userID string) ([]model.URL, error)
	BatchDelete(shortURLs []string, userID string) error
}

type memoryURLRepository struct {
	data map[string]*model.URL
	mu   sync.RWMutex
}

type dataBaseURLRepository struct {
	db *sql.DB
}

func NewMemoryURLRepository() *memoryURLRepository {
	repo := memoryURLRepository{
		data: make(map[string]*model.URL),
	}
	return &repo
}

func NewDataBaseURLRepository(cfg *config.Config) *dataBaseURLRepository {
	dbCon, err := sql.Open("postgres", cfg.DatabaseDSN)
	if err != nil {
		fmt.Println(err)
	}
	repo := dataBaseURLRepository{
		db: dbCon,
	}

	db.ApplyMigrations(dbCon)
	return &repo
}

func (r *memoryURLRepository) Save(url *model.URL) (*model.URL, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[url.Short] = url
	return url, nil
}

func (r *memoryURLRepository) GetByShortURL(id string) (*model.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	url, exists := r.data[id]
	if !exists {
		return nil, fmt.Errorf("url not found: %w", ErrNotFound)
	}
	return url, nil
}
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
func (r *dataBaseURLRepository) Save(url *model.URL) (*model.URL, error) {
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
	err := r.db.QueryRow(insertSQL, url.ID, url.Short, url.Original, url.UserID).
		Scan(&url.ID, &url.Short, &isConflict)

	if err != nil {
		return nil, err
	}
	if isConflict {
		return url, model.ErrURLAlreadyExists
	}
	return url, nil
}

func (r *dataBaseURLRepository) GetByShortURL(id string) (*model.URL, error) {
	var url model.URL
	err := r.db.QueryRow("SELECT id, short_url, original_url, user_id, is_deleted FROM urls WHERE short_url = $1", id).
		Scan(&url.ID, &url.Short, &url.Original, &url.UserID, &url.IsDeleted)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("url not found: %w", ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get url: %w", err)
	}
	return &url, nil
}

func (r *dataBaseURLRepository) GetByUserID(userID string) ([]model.URL, error) {
	rows, err := r.db.Query("SELECT id, short_url, original_url, user_id FROM urls WHERE user_id = $1", userID)
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

func (r *dataBaseURLRepository) BatchDelete(shortURLs []string, userID string) error {
	if len(shortURLs) == 0 {
		return nil
	}
	query := `UPDATE urls SET is_deleted = TRUE WHERE short_url = ANY($1) AND user_id = $2`
	_, err := r.db.Exec(query, pq.Array(shortURLs), userID)
	if err != nil {
		log.Printf("BatchDelete error: %v", err)
		return err
	}
	return nil
}

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

var ErrNotFound = &NotFoundError{}

type NotFoundError struct{}

func (e *NotFoundError) Error() string {
	return "url not found"
}
