package http

import (
	"net/http"

	"github.com/heyjorgedev/suss"
	"github.com/heyjorgedev/suss/http/html"
)

func (s *Server) Error(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	if suss.ErrorIsNotFound(err) {
		html.NotFoundPage().Render(r.Context(), w)
		return
	}

	http.Error(w, err.Error(), http.StatusInternalServerError)
}
