package main

import (
	"net/http"

	"github.com/Aleksey170999/go-shortener/internal/config"
	"github.com/Aleksey170999/go-shortener/internal/handler"
	"github.com/Aleksey170999/go-shortener/internal/middlewares"
	"github.com/Aleksey170999/go-shortener/internal/repository"
	"github.com/Aleksey170999/go-shortener/internal/service"
	"github.com/Aleksey170999/go-shortener/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.NewConfig()

	storage := storage.NewStorage(cfg.StorageFilePath)
	var repo repository.URLRepository
	if cfg.DatabaseDSN != "" {
		repo = repository.NewDataBaseURLRepository(cfg)
	} else {
		repo = repository.NewMemoryURLRepository()
		storage.LoadFromStorage(repo)
	}
	urlService := service.NewURLService(repo)
	logger := cfg.Logger
	h := handler.NewHandler(urlService, cfg, storage)
	r := chi.NewRouter()
	r.Use(middlewares.WithLogging(&logger))
	r.Use(middlewares.GzipMiddleware)
	r.Use(middleware.StripSlashes)

	r.Route("/", func(r chi.Router) {
		r.Get("/ping", h.PingDBHandler)
		r.Post("/api/shorten", h.ShortenJSONURLHandler)
		r.Post("/", h.ShortenURLHandler)
		r.Get("/{id}", h.RedirectHandler)
	})
	logger.Sugar().Infoln(
		"msg", "Server starting",
		"url", cfg.RunAddr,
	)
	http.ListenAndServe(cfg.RunAddr, r)
}
