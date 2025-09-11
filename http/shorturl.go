package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/heyjorgedev/suss"
)

func (s *Server) handlerShortUrlCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
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
