package handler

import (
	"github.com/djomlaa/socnet/internal/service"
	"net/http"
	"strconv"
)

func (h *handler) timeline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()
	last, _ := strconv.Atoi(q.Get("last"))
	before, _ := strconv.Atoi(q.Get("before"))

	pp, err := h.Timeline(ctx, last, before)

	if err == service.ErrUnauthenticated {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err != nil {
		respondError(w, err)
		return
	}

	respond(w, pp, http.StatusOK)
}
