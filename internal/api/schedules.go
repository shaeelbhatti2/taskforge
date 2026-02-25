package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/shaeelbhatti2/taskforge/internal/domain"
)

func (s *Server) mountSchedules(r chi.Router) {
	r.Route("/schedules", func(r chi.Router) {
		r.Get("/", s.listSchedules)
		r.Post("/", s.createSchedule)
		r.Get("/{id}", s.getSchedule)
		r.Put("/{id}", s.updateSchedule)
		r.Post("/{id}/pause", s.pauseSchedule)
		r.Post("/{id}/resume", s.resumeSchedule)
	})
}

func (s *Server) listSchedules(w http.ResponseWriter, r *http.Request) {
	ns := namespaceID(r)
	items, err := s.store.ListSchedules(r.Context(), ns)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) createSchedule(w http.ResponseWriter, r *http.Request) {
	var sch domain.Schedule
	if err := readJSON(r, &sch); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	now := time.Now().UTC()
	sch.ID = domain.NewID()
	sch.NamespaceID = namespaceID(r)
	sch.CreatedAt = now
	sch.UpdatedAt = now
	if err := s.store.CreateSchedule(r.Context(), &sch); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, sch)
}

func (s *Server) getSchedule(w http.ResponseWriter, r *http.Request) {
	ns := namespaceID(r)
	id := chi.URLParam(r, "id")
	sch, err := s.store.GetSchedule(r.Context(), ns, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	writeJSON(w, http.StatusOK, sch)
}

func (s *Server) updateSchedule(w http.ResponseWriter, r *http.Request) {
	ns := namespaceID(r)
	id := chi.URLParam(r, "id")
	var sch domain.Schedule
	if err := readJSON(r, &sch); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	sch.ID = id
	sch.NamespaceID = ns
	sch.UpdatedAt = time.Now().UTC()
	if err := s.store.UpdateSchedule(r.Context(), &sch); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, sch)
}

func (s *Server) pauseSchedule(w http.ResponseWriter, r *http.Request) {
	s.setSchedulePaused(w, r, true)
}

func (s *Server) resumeSchedule(w http.ResponseWriter, r *http.Request) {
	s.setSchedulePaused(w, r, false)
}

func (s *Server) setSchedulePaused(w http.ResponseWriter, r *http.Request, paused bool) {
	ns := namespaceID(r)
	id := chi.URLParam(r, "id")
	sch, err := s.store.GetSchedule(r.Context(), ns, id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	sch.Paused = paused
	sch.UpdatedAt = time.Now().UTC()
	if err := s.store.UpdateSchedule(r.Context(), sch); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, sch)
}
