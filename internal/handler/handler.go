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
	"github.com/Aleksey170999/go-shortener/internal/middlewares"
	"github.com/Aleksey170999/go-shortener/internal/model"
	"github.com/Aleksey170999/go-shortener/internal/repository"
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
	userID, err := middlewares.GetUserID(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

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
	userID, err := middlewares.GetUserID(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

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

	url, err := h.URLService.Shorten(req.URL, "", userID)
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

	h.Storage.LoadToStorage(url)
	resp := model.ShortenJSONResponse{
		Result: fmt.Sprintf("%s/%s", h.Cfg.ReturnPrefix, url.Short),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
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
	userID, err := middlewares.GetUserID(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req []model.RequestURLItem
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		h.Cfg.Logger.Debug("cannot decode request JSON body", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	validate := validator.New()
	for _, item := range req {
		err := validate.Struct(item)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
	}

	var resp []model.ResponseURLItem
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

func (h *Handler) GetUserURLsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middlewares.GetUserID(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	urls, err := h.URLService.GetUserURLs(userID)
	if err != nil {
		if err == repository.ErrNotFound {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Error(w, "failed to get user URLs", http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp := make([]model.UserURLsResponse, 0, len(urls))
	for _, url := range urls {
		resp = append(resp, model.UserURLsResponse{
			ShortURL:    fmt.Sprintf("%s/%s", h.Cfg.ReturnPrefix, url.Short),
			OriginalURL: url.Original,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
