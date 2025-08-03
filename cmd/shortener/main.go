package main

import (
	"net/http"

	"github.com/Aleksey170999/go-shortener/internal/handler"
	"github.com/Aleksey170999/go-shortener/internal/repository"
	"github.com/Aleksey170999/go-shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

func main() {
	repo := repository.NewMemoryURLRepository()
	urlService := service.NewURLService(repo)
	h := handler.NewHandler(urlService)
	r := chi.NewRouter()
	r.Route("/", func(r chi.Router) {
		r.Post("/", h.ShortenURLHandler)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.RedirectHandler)
		})
	},
	)

	http.ListenAndServe(":8080", r)
}
