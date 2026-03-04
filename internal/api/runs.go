package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shaeelbhatti2/taskforge/internal/domain"
	"github.com/shaeelbhatti2/taskforge/internal/store"
)

func (s *Server) mountRuns(r chi.Router) {
	r.Route("/runs", func(r chi.Router) {
		r.Get("/", s.listRuns)
		r.Get("/{id}", s.getRun)
		r.Get("/{id}/logs", s.streamRunLogs)
		r.Post("/{id}/cancel", s.cancelRun)
	})
}

func (s *Server) listRuns(w http.ResponseWriter, r *http.Request) {
	ns := namespaceID(r)
	filter := store.RunFilter{NamespaceID: ns}
	if v := r.URL.Query().Get("job_id"); v != "" {
		filter.JobID = v
	}
	if v := r.URL.Query().Get("status"); v != "" {
		filter.Status = domain.RunStatus(v)
	}
	if v := r.URL.Query().Get("since"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.Since = &t
		}
	}
	if v := r.URL.Query().Get("until"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.Until = &t
		}
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.Limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			filter.Offset = n
		}
	}
	items, err := s.store.ListJobRuns(r.Context(), filter)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) getRun(w http.ResponseWriter, r *http.Request) {
	ns := namespaceID(r)
	id := chi.URLParam(r, "id")
	run, err := s.store.GetJobRun(r.Context(), ns, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (s *Server) streamRunLogs(w http.ResponseWriter, r *http.Request) {
	ns := namespaceID(r)
	id := chi.URLParam(r, "id")
	run, err := s.store.GetJobRun(r.Context(), ns, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("event: stdout\ndata: " + run.Stdout + "\n\n"))
	_, _ = w.Write([]byte("event: stderr\ndata: " + run.Stderr + "\n\n"))
}
