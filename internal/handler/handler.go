package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"

	"errors"

	"github.com/Aleksey170999/go-shortener/internal/config"
	db_pack "github.com/Aleksey170999/go-shortener/internal/config/db"
	"github.com/Aleksey170999/go-shortener/internal/model"
	"github.com/Aleksey170999/go-shortener/internal/service"
	"github.com/Aleksey170999/go-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var validate = validator.New()

type Handler struct {
	URLService *service.URLService
	Cfg        *config.Config
	Storage    *storage.Storage
}

func NewHandler(urlService *service.URLService, cfg *config.Config, storage *storage.Storage) *Handler {
	return &Handler{URLService: urlService, Cfg: cfg, Storage: storage}
}

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
	url, err := h.URLService.Shorten(original, "")
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
	h.Storage.LoadToStorage(url)
	fullAddress := fmt.Sprintf("%s/%s", h.Cfg.ReturnPrefix, url.Short)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fullAddress))
}

func (h *Handler) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "id")
	if shortURL == "" {
		http.Error(w, "missing short url id", http.StatusBadRequest)
		return
	}
	original, err := h.URLService.Resolve(shortURL)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, original, http.StatusTemporaryRedirect)
}

func (h *Handler) ShortenJSONURLHandler(w http.ResponseWriter, r *http.Request) {
	var req model.ShortenJSONRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		h.Cfg.Logger.Debug("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if req.URL == "" {
		http.Error(w, "empty url", http.StatusBadRequest)
		return
	}
	url, err := h.URLService.Shorten(req.URL, "")
	if err != nil {
		if errors.Is(err, model.ErrURLAlreadyExists) {
			resp := model.ShortenJSONResponse{
				Result: fmt.Sprintf("%s/%s", h.Cfg.ReturnPrefix, url.Short),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict) // 409
			json.NewEncoder(w).Encode(resp)
			return
		}
		http.Error(w, "failed to shorten url", http.StatusInternalServerError)
		return
	}
	if err := h.Storage.LoadToStorage(url); err != nil {
		http.Error(w, "failed to store url to storage", http.StatusInternalServerError)
		return
	}
	resp := model.ShortenJSONResponse{
		Result: fmt.Sprintf("%s/%s", h.Cfg.ReturnPrefix, url.Short),
	}
	enc := json.NewEncoder(w)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := enc.Encode(resp); err != nil {
		h.Cfg.Logger.Debug("error encoding response", zap.Error(err))
		return
	}
}

func (h *Handler) PingDBHandler(w http.ResponseWriter, r *http.Request) {
	err := db_pack.PingDB(h.Cfg.DatabaseDSN)
	if err != nil {
		http.Error(w, "failed to ping DB", http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) ShortenJSONURLBatchHandler(w http.ResponseWriter, r *http.Request) {
	var urls []model.RequestURLItem
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&urls); err != nil {
		h.Cfg.Logger.Debug("cannot decode request JSON body", zap.Error(err))
		http.Error(w, "ERROR", http.StatusInternalServerError)
		return
	}
	for _, url := range urls {
		if err := validate.Struct(url); err != nil {
			http.Error(w, fmt.Sprintf("Ошибка валидации в элементе %s: %v", url, err), http.StatusBadRequest)
			return
		}
	}

	var responses []model.ResponseURLItem

	for _, url := range urls {
		shortenURL, err := h.URLService.Shorten(url.OriginalURL, url.СorrelationID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Ошибка сокращения: %s", err), http.StatusBadRequest)
		}
		responses = append(responses, model.ResponseURLItem{
			CorrelationID: shortenURL.ID,
			ShortURL:      fmt.Sprintf("%s/%s", h.Cfg.ReturnPrefix, shortenURL.Short),
		})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responses)
}
