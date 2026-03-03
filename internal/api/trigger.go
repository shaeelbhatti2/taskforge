package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shaeelbhatti2/taskforge/internal/domain"
)

func (s *Server) mountTrigger(r chi.Router) {
	r.Post("/jobs/{id}/trigger", s.triggerJob)
	r.Post("/jobs/{id}/backfill", s.backfillJob)
}

func (s *Server) triggerJob(w http.ResponseWriter, r *http.Request) {
	ns := namespaceID(r)
	jobID := chi.URLParam(r, "id")
	job, err := s.store.GetJob(r.Context(), ns, jobID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "job not found"})
		return
	}
	run := &domain.JobRun{
		ID:            domain.NewID(),
		NamespaceID:   ns,
		JobID:         job.ID,
		Status:        domain.RunPending,
		Attempt:       1,
		ScheduledAt:   time.Now().UTC(),
		CorrelationID: domain.NewID(),
	}
	if err := s.store.CreateJobRun(r.Context(), run); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusAccepted, run)
}

func (s *Server) backfillJob(w http.ResponseWriter, r *http.Request) {
	s.triggerJob(w, r)
}
