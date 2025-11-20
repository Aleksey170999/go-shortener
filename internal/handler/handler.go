package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"

	"github.com/Aleksey170999/go-shortener/internal/audit"
	db_pack "github.com/Aleksey170999/go-shortener/internal/config/db"

	"github.com/Aleksey170999/go-shortener/internal/config"
	"github.com/Aleksey170999/go-shortener/internal/middlewares"
	"github.com/Aleksey170999/go-shortener/internal/model"
	"github.com/Aleksey170999/go-shortener/internal/repository"
	"github.com/Aleksey170999/go-shortener/internal/service"
	"github.com/Aleksey170999/go-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

var validate = validator.New()

// Handler provides all the HTTP handlers for the URL shortener service.
// It contains all the necessary dependencies to handle incoming requests.
type Handler struct {
	URLService   *service.URLService
	Cfg          *config.Config
	Storage      *storage.Storage
	AuditManager *audit.AuditManager
}

// NewHandler creates a new instance of Handler with the provided dependencies.
//
// Parameters:
//   - urlService: Service for URL shortening and management operations
//   - cfg: Application configuration
//   - storage: Storage for persisting URLs
//   - auditManager: Manager for handling audit logs
//
// Returns:
//   - *Handler: A new Handler instance with the provided dependencies
func NewHandler(urlService *service.URLService, cfg *config.Config, storage *storage.Storage, auditManager *audit.AuditManager) *Handler {
	return &Handler{
		URLService:   urlService,
		Cfg:          cfg,
		Storage:      storage,
		AuditManager: auditManager,
	}
}

// ShortenURLHandler handles the URL shortening request.
// It reads the URL from the request body, validates it, and returns a shortened version.
//
// Request:
//   - Method: POST
//   - Body: The URL to be shortened as plain text
//
// Responses:
//   - 201 Created: On successful URL shortening, returns the shortened URL
//   - 400 Bad Request: If the request body is empty or invalid
//   - 500 Internal Server Error: If there's an error processing the request
func (h *Handler) ShortenURLHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "can't read body", http.StatusBadRequest)
		return
	}
	original := string(body)
	if original == "" {
		http.Error(w, "empty url", http.StatusBadRequest)
		return
	}
	userID, _ := middlewares.GetUserID(r)

	if userID == "" {
		userID = uuid.New().String()

		http.SetCookie(w, &http.Cookie{
			Name:     "user_id",
			Value:    userID,
			Path:     "/",
			HttpOnly: true,
		})
	}

	url, err := h.URLService.Shorten(original, "", userID)
	if err != nil {
		if errors.Is(err, model.ErrURLAlreadyExists) {
			fullAddress := fmt.Sprintf("%s/%s", h.Cfg.ReturnPrefix, url.Short)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict) // 409
			w.Write([]byte(fullAddress))
			return
		}
		http.Error(w, "failed to shorten url", http.StatusInternalServerError)
		return
	}

	if h.AuditManager != nil {
		go h.AuditManager.LogEvent(r.Context(), "shorten", userID, original)
	}

	h.Storage.LoadToStorage(url)
	fullAddress := fmt.Sprintf("%s/%s", h.Cfg.ReturnPrefix, url.Short)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fullAddress))
}

// RedirectHandler handles the redirection of shortened URLs to their original URLs.
// It extracts the short URL ID from the request path and performs the redirection.
//
// Request:
//   - Method: GET
//   - Path: /{id}
//
// Responses:
//   - 307 Temporary Redirect: Redirects to the original URL
//   - 400 Bad Request: If the short URL ID is missing
//   - 404 Not Found: If the short URL is not found or has been deleted
//   - 500 Internal Server Error: If there's an error processing the request
func (h *Handler) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "id")
	if shortURL == "" {
		http.Error(w, "missing short url id", http.StatusBadRequest)
		return
	}
	url, err := h.URLService.Resolve(shortURL)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if url.IsDeleted {
		http.Error(w, "gone", http.StatusGone)
		return
	}

	userID, _ := middlewares.GetUserID(r)
	if h.AuditManager != nil && userID != "" {
		go h.AuditManager.LogEvent(r.Context(), "follow", userID, url.Original)
	}

	http.Redirect(w, r, url.Original, http.StatusTemporaryRedirect)
}

// ShortenJSONURLHandler handles URL shortening requests in JSON format.
// It accepts a JSON payload with the URL to be shortened and returns a JSON response
// containing the shortened URL.
//
// Request:
//   - Method: POST
//   - Content-Type: application/json
//   - Body: JSON object with an 'url' field containing the URL to be shortened
//
// Example Request:
//
//	{
//	  "url": "https://example.com/long/url/to/be/shortened"
//	}
//
// Responses:
//   - 201 Created: On successful shortening, returns a JSON response with the shortened URL
//   - 400 Bad Request: If the request body is invalid or missing required fields
//   - 500 Internal Server Error: If there's an error processing the request
func (h *Handler) ShortenJSONURLHandler(w http.ResponseWriter, r *http.Request) {
	var req model.ShortenJSONRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.Cfg.Logger.Error("error decoding request body", zap.Error(err))
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "empty url", http.StatusBadRequest)
		return
	}

	userID, _ := middlewares.GetUserID(r)

	if userID == "" {
		userID = uuid.New().String()

		http.SetCookie(w, &http.Cookie{
			Name:     "user_id",
			Value:    userID,
			Path:     "/",
			HttpOnly: true,
		})
	}

	url, err := h.URLService.Shorten(req.URL, "", userID)
	if err != nil {
		if errors.Is(err, model.ErrURLAlreadyExists) {
			response := model.ShortenJSONResponse{
				Result: fmt.Sprintf("%s/%s", h.Cfg.ReturnPrefix, url.Short),
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(response)
			return
		}

		h.Cfg.Logger.Error("error shortening url", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if h.AuditManager != nil {
		go h.AuditManager.LogEvent(r.Context(), "shorten", userID, req.URL)
	}

	h.Storage.LoadToStorage(url)

	response := model.ShortenJSONResponse{
		Result: fmt.Sprintf("%s/%s", h.Cfg.ReturnPrefix, url.Short),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// PingDBHandler handles the /ping endpoint to check database connectivity.
// Returns:
//   - 200 OK if the database is reachable
//   - 500 Internal Server Error if the database is not available
//
// This handler is used for health checks and monitoring.
func (h *Handler) PingDBHandler(w http.ResponseWriter, r *http.Request) {
	err := db_pack.PingDB(h.Cfg.DatabaseDSN)
	if err != nil {
		http.Error(w, "failed to ping DB", http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

// ShortenJSONURLBatchHandler handles batch URL shortening requests.
// Accepts a JSON array of URLs and returns their shortened versions.
//
// Request body should be a JSON array of objects with the following structure:
//
//	[
//	  {"correlation_id": "<unique_id>", "original_url": "<url>"},
//	  ...
//	]
//
// Response is a JSON array of objects with the following structure:
//
//	[
//	  {"correlation_id": "<same_id>", "short_url": "<short_url>"},
//	  ...
//	]
//
// Returns:
//   - 201 Created on successful batch processing
//   - 400 Bad Request for invalid input
//   - 500 Internal Server Error for processing failures
func (h *Handler) ShortenJSONURLBatchHandler(w http.ResponseWriter, r *http.Request) {
	var req []model.RequestURLItem
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		h.Cfg.Logger.Debug("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, item := range req {
		err := validate.Struct(item)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
	}

	var resp []model.ResponseURLItem
	userID, _ := middlewares.GetUserID(r)
	for _, item := range req {
		url, _ := h.URLService.Shorten(item.OriginalURL, item.СorrelationID, userID)
		resp = append(resp, model.ResponseURLItem{
			CorrelationID: item.СorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", h.Cfg.ReturnPrefix, url.Short),
		})
		h.Storage.LoadToStorage(url)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// GetUserURLsHandler retrieves all URLs shortened by the current user.
// The user is identified by the session cookie.
//
// Response is a JSON array of objects with the following structure:
//
//	[
//	  {"short_url": "<short_url>", "original_url": "<original_url>"},
//	  ...
//	]
//
// Returns:
//   - 200 OK with the list of URLs
//   - 204 No Content if no URLs found for the user
//   - 500 Internal Server Error for processing failures
func (h *Handler) GetUserURLsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middlewares.GetUserID(r)

	w.Header().Set("Content-Type", "application/json")
	if userID == "" {
		log.Printf("[GetUserURLsHandler] userID is empty")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		log.Printf("[GetUserURLsHandler] error getting userID: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	urls, err := h.URLService.GetUserURLs(userID)
	if err != nil {
		if err == repository.ErrNotFound {
			log.Printf("[GetUserURLsHandler] no urls found for userID=%s", userID)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		log.Printf("[GetUserURLsHandler] error fetching urls for userID=%s: %v", userID, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		log.Printf("[GetUserURLsHandler] urls list empty for userID=%s", userID)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	log.Printf("[GetUserURLsHandler] found %d urls for userID=%s", len(urls), userID)

	resp := make([]model.UserURLsResponse, 0, len(urls))
	for _, url := range urls {
		resp = append(resp, model.UserURLsResponse{
			ShortURL:    fmt.Sprintf("%s/%s", h.Cfg.ReturnPrefix, url.Short),
			OriginalURL: url.Original,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// BatchDeleteUserURLsHandler handles batch deletion of URLs for the current user.
// The deletion is processed asynchronously.
//
// Request body should be a JSON array of short URL identifiers to delete:
//
//	["id1", "id2", ...]
//
// Returns:
//   - 202 Accepted if the deletion request was accepted for processing
//   - 400 Bad Request for invalid input
//   - 401 Unauthorized if user is not authenticated
//   - 500 Internal Server Error for processing failures
//
// Note: This is an asynchronous operation. The actual deletion happens in a separate goroutine.
func (h *Handler) BatchDeleteUserURLsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middlewares.GetUserID(r)
	if err != nil {
		log.Print(err)
	}
	var shortUrls []string
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&shortUrls); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	go func(shortUrls []string, userID string) {
		err := h.URLService.BatchDelete(shortUrls, userID)
		if err != nil {
			log.Printf("[BatchDeleteUserURLsHandler] async BatchDelete error: %v", err)
		}
	}(shortUrls, userID)
	w.WriteHeader(http.StatusAccepted)
}
