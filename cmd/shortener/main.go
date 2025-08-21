package main

import (
	"net/http"

	"github.com/Aleksey170999/go-shortener/internal/config"
	"github.com/Aleksey170999/go-shortener/internal/handler"
	"github.com/Aleksey170999/go-shortener/internal/repository"
	"github.com/Aleksey170999/go-shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.NewConfig()
	repo := repository.NewMemoryURLRepository()
	urlService := service.NewURLService(repo)
	logger := cfg.Logger

	h := handler.NewHandler(urlService, cfg)
	r := chi.NewRouter()
	r.Use(handler.WithLogging(&logger))
	r.Route("/", func(r chi.Router) {
		r.Post("/", h.ShortenURLHandler)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.RedirectHandler)
		})
	},
	)
	logger.Sugar().Infoln(
		"msg", "Server starting",
		"url", cfg.RunAddr,
	)
	http.ListenAndServe(cfg.RunAddr, r)
}
