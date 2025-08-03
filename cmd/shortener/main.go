package main

import (
	"net/http"

	"github.com/Aleksey170999/go-shortener/internal/handler"
	"github.com/Aleksey170999/go-shortener/internal/repository"
	"github.com/Aleksey170999/go-shortener/internal/service"
)

func main() {
	repo := repository.NewMemoryURLRepository()
	urlService := service.NewURLService(repo)
	h := handler.NewHandler(urlService)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" && r.Method == http.MethodPost {
			h.ShortenURLHandler(w, r)
			return
		}
		if r.URL.Path != "/" && r.Method == http.MethodGet {
			h.RedirectHandler(w, r)
			return
		}
		http.Error(w, "Not found", http.StatusNotFound)
	})

	http.ListenAndServe(":8080", mux)
}
