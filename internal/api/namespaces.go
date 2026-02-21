package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shaeelbhatti2/taskforge/internal/domain"
)

func (s *Server) mountNamespaces(r chi.Router) {
	r.Route("/namespaces", func(r chi.Router) {
		r.Get("/", s.listNamespaces)
		r.Post("/", s.createNamespace)
		r.Get("/{id}", s.getNamespace)
	})
}

func (s *Server) listNamespaces(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.ListNamespaces(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) createNamespace(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := readJSON(r, &req); err != nil || req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	now := time.Now().UTC()
	ns := &domain.Namespace{ID: domain.NewID(), Name: req.Name, CreatedAt: now, UpdatedAt: now}
	if err := s.store.CreateNamespace(r.Context(), ns); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, ns)
}

func (s *Server) getNamespace(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ns, err := s.store.GetNamespace(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, ns)
}
