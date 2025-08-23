package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Aleksey170999/go-shortener/internal/config"
	"github.com/Aleksey170999/go-shortener/internal/model"
	"github.com/Aleksey170999/go-shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Handler struct {
	URLService *service.URLService
	Cfg        *config.Config
}

func NewHandler(urlService *service.URLService, cfg *config.Config) *Handler {
	return &Handler{URLService: urlService, Cfg: cfg}
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
	id, err := h.URLService.Shorten(original)
	if err != nil {
		http.Error(w, "failed to shorten url", http.StatusInternalServerError)
		return
	}
	fullAddress := fmt.Sprintf("%s/%s", h.Cfg.ReturnPrefix, id)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fullAddress))
}

func (h *Handler) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing short url id", http.StatusBadRequest)
		return
	}
	original, err := h.URLService.Resolve(id)
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
	id, err := h.URLService.Shorten(req.URL)
	if err != nil {
		http.Error(w, "failed to shorten url", http.StatusInternalServerError)
		return
	}
	resp := model.ShortenJSONResponse{
		Result: fmt.Sprintf("%s/%s", h.Cfg.ReturnPrefix, id),
	}
	enc := json.NewEncoder(w)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := enc.Encode(resp); err != nil {
		h.Cfg.Logger.Debug("error encoding response", zap.Error(err))
		return
	}
}
