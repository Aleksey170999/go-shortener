package handler

import (
	"io"
	"net/http"

	"github.com/Aleksey170999/go-shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	URLService *service.URLService
}

func NewHandler(urlService *service.URLService) *Handler {
	return &Handler{URLService: urlService}
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
	fullAddress := "http://localhost:8080/" + id
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
