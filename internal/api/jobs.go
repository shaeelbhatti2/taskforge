package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shaeelbhatti2/taskforge/internal/domain"
)

func (s *Server) mountJobs(r chi.Router) {
	r.Route("/jobs", func(r chi.Router) {
		r.Get("/", s.listJobs)
		r.Post("/", s.createJob)
		r.Get("/{id}", s.getJob)
		r.Put("/{id}", s.updateJob)
		r.Delete("/{id}", s.deleteJob)
	})
}

func (s *Server) listJobs(w http.ResponseWriter, r *http.Request) {
	ns := namespaceID(r)
	items, err := s.store.ListJobs(r.Context(), ns)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) createJob(w http.ResponseWriter, r *http.Request) {
	var job domain.JobDefinition
	if err := readJSON(r, &job); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	now := time.Now().UTC()
	job.ID = domain.NewID()
	job.NamespaceID = namespaceID(r)
	job.CreatedAt = now
	job.UpdatedAt = now
	if err := s.store.CreateJob(r.Context(), &job); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, job)
}

func (s *Server) getJob(w http.ResponseWriter, r *http.Request) {
	ns := namespaceID(r)
	id := chi.URLParam(r, "id")
	job, err := s.store.GetJob(r.Context(), ns, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (s *Server) updateJob(w http.ResponseWriter, r *http.Request) {
	ns := namespaceID(r)
	id := chi.URLParam(r, "id")
	var job domain.JobDefinition
	if err := readJSON(r, &job); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	job.ID = id
	job.NamespaceID = ns
	job.UpdatedAt = time.Now().UTC()
	if err := s.store.UpdateJob(r.Context(), &job); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (s *Server) deleteJob(w http.ResponseWriter, r *http.Request) {
	ns := namespaceID(r)
	id := chi.URLParam(r, "id")
	if err := s.store.DeleteJob(r.Context(), ns, id); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
