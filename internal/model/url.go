// Package model defines the core data structures used throughout the application.
// It contains the domain models and DTOs (Data Transfer Objects) for the API.
package model

import "errors"

// URL represents a shortened URL in the system.
// It contains both the original URL and its shortened version,
// along with metadata about the URL.
type URL struct {
	// ID is a unique identifier for the URL (UUID format)
	ID string `json:"uuid" db:"id"`

	// Original is the original long URL that was shortened
	Original string `json:"original_url" db:"original_url"`

	// Short is the shortened URL identifier
	Short string `json:"short_url" db:"short_url"`

	// UserID is the ID of the user who created this URL
	UserID string `json:"user_id,omitempty" db:"user_id"`

	// IsDeleted indicates if the URL has been soft-deleted
	IsDeleted bool `json:"-" db:"is_deleted"`
}

// UserURLsResponse represents the response structure when
// retrieving all URLs for a specific user.
type UserURLsResponse struct {
	// ShortURL is the shortened URL
	ShortURL string `json:"short_url"`

	// OriginalURL is the original URL that was shortened
	OriginalURL string `json:"original_url"`
}

// ShortenJSONRequest represents the request body for creating a new short URL
type ShortenJSONRequest struct {
	// URL is the original URL to be shortened
	URL string `json:"url" validate:"required,url"`
}

// ShortenJSONResponse represents the response after creating a short URL
type ShortenJSONResponse struct {
	// Result contains the shortened URL
	Result string `json:"result"`
}

// RequestURLItem represents a single URL in a batch create request
type RequestURLItem struct {
	// CorrelationID is a client-generated ID to match requests with responses
	СorrelationID string `json:"correlation_id" validate:"required"`

	// OriginalURL is the URL to be shortened
	OriginalURL string `json:"original_url" validate:"required,url"`
}

// ResponseURLItem represents a single URL in a batch create response
type ResponseURLItem struct {
	// CorrelationID matches the ID from the corresponding request
	CorrelationID string `json:"correlation_id"`

	// ShortURL is the generated short URL
	ShortURL string `json:"short_url"`
}

// Common errors
var (
	// ErrURLAlreadyExists is returned when attempting to create a URL that already exists
	ErrURLAlreadyExists = errors.New("url already exists")

	// ErrURLNotFound is returned when a requested URL doesn't exist
	ErrURLNotFound = errors.New("url not found")

	// ErrURLDeleted is returned when attempting to access a deleted URL
	ErrURLDeleted = errors.New("url has been deleted")
)
