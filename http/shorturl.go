package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/heyjorgedev/suss"
)

func (s *Server) handlerShortUrlCreate() http.HandlerFunc {
	rateLimiter := httprate.NewRateLimiter(5, time.Minute, httprate.WithKeyByIP())
	handler := rateLimiter.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		url := r.Form.Get("url")
		if url == "" {
			http.Error(w, "url required", http.StatusBadRequest)
			return
		}

		shortUrl := &suss.ShortURL{
			LongURL: url,
		}
		if err := s.ShortURLService.Create(r.Context(), shortUrl); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/preview/%s", shortUrl.Slug), http.StatusSeeOther)
	}))

	return handler.ServeHTTP
}

func (s *Server) handlerShortUrlPreview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")
		if slug == "" {
			http.Error(w, "slug required", http.StatusBadRequest)
			return
		}

		w.Write([]byte(slug + "preview"))
	}
}

func (s *Server) handlerShortUrlVisit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")
		if slug == "" {
			http.Error(w, "slug required", http.StatusBadRequest)
			return
		}

		w.Write([]byte(slug))
	}
}
