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
			s.Error(w, r, suss.Errorf(suss.EINVALID, "slug required"))
			return
		}

		shortUrl, err := s.ShortURLService.FindDialBySlug(r.Context(), slug)
		if err != nil {
			s.Error(w, r, err)
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
			s.Error(w, r, suss.Errorf(suss.EINVALID, "slug required"))
			return
		}

		shortUrl, err := s.ShortURLService.FindDialBySlug(r.Context(), slug)
		if err != nil {
			s.Error(w, r, err)
			return
		}

		http.Redirect(w, r, shortUrl.LongURL, http.StatusSeeOther)
	}
}

func (s *Server) handlerShortUrlManage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")
		if slug == "" {
			s.Error(w, r, suss.Errorf(suss.EINVALID, "slug required"))
			return
		}

		secret := r.URL.Query().Get("secret")
		if secret == "" {
			s.Error(w, r, suss.Errorf(suss.EINVALID, "secret required"))
			return
		}

		shortUrl, err := s.ShortURLService.FindDialBySlug(r.Context(), slug)
		if err != nil {
			s.Error(w, r, err)
			return
		}

		if secret != shortUrl.SecretKey {
			s.Error(w, r, suss.Errorf(suss.EINVALID, "invalid secret"))
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
			w.WriteHeader(http.StatusNotFound)
			return
		}

		shortUrl, err := s.ShortURLService.FindDialBySlug(r.Context(), slug)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		q, err := qrcode.New(shortUrl.ShortURL(s.PublicURL(r)), qrcode.Medium)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Header().Set("Expires", time.Now().Add(time.Hour*24).Format(http.TimeFormat))
		q.Write(512, w)
	}
}
