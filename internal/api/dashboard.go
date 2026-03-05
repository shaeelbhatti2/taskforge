package api

import (
	"net/http"

	"github.com/shaeelbhatti2/taskforge/internal/domain"
	"github.com/shaeelbhatti2/taskforge/internal/store"
)

func (s *Server) mountWorkers(r chi.Router) {
	r.Get("/workers", func(w http.ResponseWriter, r *http.Request) {
		items, err := s.store.ListWorkers(r.Context(), namespaceID(r))
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, items)
	})
}

func (s *Server) mountDashboard(r chi.Router) {
	r.Get("/dashboard", s.dashboardStats)
}

func (s *Server) dashboardStats(w http.ResponseWriter, r *http.Request) {
	ns := namespaceID(r)
	runs, err := s.store.ListJobRuns(r.Context(), store.RunFilter{NamespaceID: ns, Limit: 1000})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	stats := map[string]int{"running": 0, "queued": 0, "failed": 0, "success": 0}
	for _, run := range runs {
		switch run.Status {
		case domain.RunRunning:
			stats["running"]++
		case domain.RunPending:
			stats["queued"]++
		case domain.RunFailed, domain.RunDead:
			stats["failed"]++
		case domain.RunSuccess:
			stats["success"]++
		}
	}
	writeJSON(w, http.StatusOK, stats)
}
