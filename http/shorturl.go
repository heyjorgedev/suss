package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/heyjorgedev/suss"
	"github.com/heyjorgedev/suss/http/html"
	qrcode "github.com/skip2/go-qrcode"
)

func (s *Server) handlerHomepage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		html.Homepage().Render(r.Context(), w)
	}
}

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

		http.Redirect(w, r, fmt.Sprintf("/manage/%s?secret=%s", shortUrl.Slug, shortUrl.SecretKey), http.StatusSeeOther)
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

		shortUrl, err := s.ShortURLService.FindDialBySlug(r.Context(), slug)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		html.PreviewPage(html.PreviewPageProps{
			Url:      shortUrl.ShortURL(s.PublicURL(r)),
			ShortURL: shortUrl,
		}).Render(r.Context(), w)
	}
}

func (s *Server) handlerShortUrlVisit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")
		if slug == "" {
			http.Error(w, "slug required", http.StatusBadRequest)
			return
		}

		shortUrl, err := s.ShortURLService.FindDialBySlug(r.Context(), slug)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, shortUrl.LongURL, http.StatusSeeOther)
	}
}

func (s *Server) handlerShortUrlManage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")
		if slug == "" {
			http.Error(w, "slug required", http.StatusBadRequest)
			return
		}

		secret := r.URL.Query().Get("secret")
		if secret == "" {
			http.Error(w, "secret required", http.StatusBadRequest)
			return
		}

		shortUrl, err := s.ShortURLService.FindDialBySlug(r.Context(), slug)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if secret != shortUrl.SecretKey {
			http.Error(w, "invalid secret", http.StatusBadRequest)
			return
		}

		html.ManagePage(html.ManagePageProps{
			Url:      shortUrl.ShortURL(s.PublicURL(r)),
			ShortURL: shortUrl,
		}).Render(r.Context(), w)
	}
}

func (s *Server) handlerShortUrlQrCode() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")
		if slug == "" {
			http.Error(w, "slug required", http.StatusBadRequest)
			return
		}

		shortUrl, err := s.ShortURLService.FindDialBySlug(r.Context(), slug)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		q, err := qrcode.New(shortUrl.ShortURL(s.PublicURL(r)), qrcode.Medium)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Header().Set("Expires", time.Now().Add(time.Hour*24).Format(http.TimeFormat))
		q.Write(256, w)
	}
}
