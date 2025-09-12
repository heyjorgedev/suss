package http

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/benbjohnson/hashfs"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/heyjorgedev/suss"
	"github.com/heyjorgedev/suss/http/dist"
	"github.com/heyjorgedev/suss/http/html"
)

// time to wait for the server to finish processing requests when shutting down
const ShutdownTimeout = 2 * time.Second

type Server struct {
	ln     net.Listener
	server *http.Server
	router chi.Router

	// address to listen on
	Addr string

	// dependent services to use
	ShortURLService suss.ShortURLService
}

func NewServer() *Server {
	r := chi.NewRouter()
	s := &Server{
		server: &http.Server{},
		router: r,
	}
	s.server.Handler = http.HandlerFunc(s.serveHTTP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.GetHead)
	r.Handle("/assets/*", http.StripPrefix("/assets/", hashfs.FileServer(dist.FS)))
	r.NotFound(templ.Handler(html.NotFoundPage()).ServeHTTP)

	// register routes
	r.Get("/", templ.Handler(html.Homepage()).ServeHTTP)
	r.Post("/shorten", s.handlerShortUrlCreate())
	r.Get("/preview/{slug}", s.handlerShortUrlPreview())
	r.Get("/{slug}+", s.handlerShortUrlPreview())
	r.Get("/{slug}", s.handlerShortUrlVisit())

	return s
}

func (s *Server) Open() (err error) {
	// open a listener on our bind address.
	if s.ln, err = net.Listen("tcp", s.Addr); err != nil {
		return err
	}

	// start the http server
	go s.server.Serve(s.ln)

	return nil
}

func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	return s.server.Shutdown(ctx)
}

func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	// override method for forms passing "_method" value.
	if r.Method == http.MethodPost {
		switch v := r.PostFormValue("_method"); v {
		case http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete:
			r.Method = v
		}
	}

	s.router.ServeHTTP(w, r)
}
