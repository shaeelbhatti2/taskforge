package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shaeelbhatti2/taskforge/internal/domain"
)

func (s *Server) cancelRun(w http.ResponseWriter, r *http.Request) {
	ns := namespaceID(r)
	id := chi.URLParam(r, "id")
	run, err := s.store.GetJobRun(r.Context(), ns, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	if run.Status != domain.RunRunning && run.Status != domain.RunPending {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "not cancellable"})
		return
	}
	if _, err := domain.TransitionRun(run, domain.RunCancelled); err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	now := time.Now().UTC()
	run.FinishedAt = &now
	if err := s.store.UpdateJobRun(r.Context(), run); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, run)
}
