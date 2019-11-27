package handler

import (
	"github.com/djomlaa/socnet/internal/service"
	"github.com/matryer/way"
	"net/http"
	"strconv"
)

func (h *handler) notifications(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	last, _ := strconv.Atoi(q.Get("last"))
	before, _ := strconv.ParseInt(q.Get("last"), 10, 64)
	nn, err := h.Notifications(r.Context(), last, before)
	if err == service.ErrUnauthenticated {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err != nil {
		respondError(w, err)
		return
	}

	respond(w, nn, http.StatusOK)

}

func (h *handler) markNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	notificationID, _ := strconv.ParseInt(way.Param(ctx, "notification_id"), 10, 64)
	err := h.MarkNotificationAsRead(ctx, notificationID)
	if err == service.ErrUnauthenticated {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err != nil {
		respondError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

func (h *handler) markNotificationsAsRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	err := h.MarkNotificationsAsRead(ctx)
	if err == service.ErrUnauthenticated {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err != nil {
		respondError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
