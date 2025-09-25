package repository

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/Aleksey170999/go-shortener/internal/config"
	db "github.com/Aleksey170999/go-shortener/internal/config/db"
	"github.com/Aleksey170999/go-shortener/internal/model"
	_ "github.com/jackc/pgx/v5"
)

type URLRepository interface {
	Save(url *model.URL) (*model.URL, error)
	GetByShortURL(shortURL string) (*model.URL, error)
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
	url, ok := r.data[id]
	if !ok {
		return nil, ErrNotFound
	}
	return url, nil
}

func (r *dataBaseURLRepository) Save(url *model.URL) (*model.URL, error) {
	var isConflict bool
	insertSQL := `WITH inserted AS (
						INSERT INTO urls (id, short_url, original_url)
						VALUES ($1, $2, $3)
						ON CONFLICT (original_url) DO NOTHING
						RETURNING *
					)
					select id, short_url, false as is_conflict FROM inserted
					UNION
					SELECT id, short_url, true as is_conflict FROM urls 
					WHERE original_url = $3 AND NOT EXISTS (SELECT 1 FROM inserted)`
	err := r.db.QueryRow(insertSQL, url.ID, url.Short, url.Original).
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

	row := r.db.QueryRow("SELECT id, original_url, short_url FROM urls WHERE short_url = $1;", id)
	err := row.Scan(&url.ID, &url.Original, &url.Short)
	if err != nil {
		return nil, err
	}
	return &url, nil
}

var ErrNotFound = &NotFoundError{}

type NotFoundError struct{}

func (e *NotFoundError) Error() string {
	return "url not found"
}
