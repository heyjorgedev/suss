package http

import (
	"context"
	"net/http"
)

func (s *Server) middlewareHost(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "url", s.PublicURL(r))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
