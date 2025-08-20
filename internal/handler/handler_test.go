package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Aleksey170999/go-shortener/internal/config"
	"github.com/Aleksey170999/go-shortener/internal/repository"
	"github.com/Aleksey170999/go-shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

func setupTestHandler() *Handler {
	repo := repository.NewMemoryURLRepository()
	urlService := service.NewURLService(repo)
	cfg := config.Config{
		RunAddr:      "localhost:8080",
		ReturnPrefix: "http://localhost:8080",
	}
	return NewHandler(urlService, &cfg)
}

func TestShortenURLHandler(t *testing.T) {
	h := setupTestHandler()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	w := httptest.NewRecorder()

	h.ShortenURLHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Error("expected non-empty body")
	}
}

func TestRedirectHandler(t *testing.T) {
	h := setupTestHandler()
	shortenReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://example.com"))
	shortenW := httptest.NewRecorder()
	h.ShortenURLHandler(shortenW, shortenReq)
	shortenResp := shortenW.Result()
	defer shortenResp.Body.Close()
	if shortenResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", shortenResp.StatusCode)
	}
	shortURL, _ := io.ReadAll(shortenResp.Body)
	short := strings.TrimPrefix(string(shortURL), "http://localhost:8080/")
	redirectReq := httptest.NewRequest(http.MethodGet, "/"+short, nil)
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", short)
	redirectReq = redirectReq.WithContext(context.WithValue(redirectReq.Context(), chi.RouteCtxKey, chiCtx))
	redirectW := httptest.NewRecorder()
	h.RedirectHandler(redirectW, redirectReq)
	redirectResp := redirectW.Result()
	defer redirectResp.Body.Close()
	if redirectResp.StatusCode != http.StatusTemporaryRedirect {
		t.Errorf("expected status 307, got %d", redirectResp.StatusCode)
	}
	loc := redirectResp.Header.Get("Location")
	if loc != "https://example.com" {
		t.Errorf("expected Location 'https://example.com', got '%s'", loc)
	}
}
