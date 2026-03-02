package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shaeelbhatti2/taskforge/internal/domain"
	"github.com/shaeelbhatti2/taskforge/internal/workflow"
)

func (s *Server) mountWorkflows(r chi.Router) {
	r.Route("/workflows", func(r chi.Router) {
		r.Get("/", s.listWorkflows)
		r.Post("/", s.createWorkflow)
		r.Get("/{id}", s.getWorkflow)
		r.Post("/validate", s.validateWorkflow)
	})
}

func (s *Server) listWorkflows(w http.ResponseWriter, r *http.Request) {
	ns := namespaceID(r)
	items, err := s.store.ListWorkflows(r.Context(), ns)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) createWorkflow(w http.ResponseWriter, r *http.Request) {
	var wf domain.Workflow
	if err := readJSON(r, &wf); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	wf.NamespaceID = namespaceID(r)
	if wf.ID == "" {
		wf.ID = domain.NewID()
	}
	now := time.Now().UTC()
	wf.CreatedAt = now
	wf.UpdatedAt = now
	if err := workflow.Validate(&wf); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := s.store.CreateWorkflow(r.Context(), &wf); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, wf)
}

func (s *Server) getWorkflow(w http.ResponseWriter, r *http.Request) {
	ns := namespaceID(r)
	id := chi.URLParam(r, "id")
	wf, err := s.store.GetWorkflow(r.Context(), ns, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, wf)
}

func (s *Server) validateWorkflow(w http.ResponseWriter, r *http.Request) {
	var wf domain.Workflow
	if err := readJSON(r, &wf); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if err := workflow.Validate(&wf); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "valid"})
}
