package http

import (
	"context"
	"fmt"
	"net/http"
)

func (s *Server) middlewareHost(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := s.PublicURL(r)
		fmt.Println(url)
		ctx := context.WithValue(r.Context(), "url", url)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
