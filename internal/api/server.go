package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/shaeelbhatti2/taskforge/internal/auth"
	"github.com/shaeelbhatti2/taskforge/internal/store"
)

type Server struct {
	store  store.Store
	auth   *auth.Service
	router chi.Router
}

func NewServer(st store.Store, authSvc *auth.Service) *Server {
	s := &Server{store: st, auth: authSvc}
	s.router = chi.NewRouter()
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Recoverer)
	s.router.Get("/health", s.health)
	s.router.Route("/api/v1", func(r chi.Router) {
		r.Use(s.auth.Middleware)
		s.mountRoutes(r)
	})
	return s
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) mountRoutes(r chi.Router) {
	s.mountNamespaces(r)
	s.mountJobs(r)
	s.mountSchedules(r)
	s.mountWorkflows(r)
	s.mountRuns(r)
	s.mountTrigger(r)
	s.mountDashboard(r)
	s.mountDLQ(r)
	s.mountWorkers(r)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func readJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func namespaceID(r *http.Request) string {
	if v := r.Header.Get("X-Namespace-ID"); v != "" {
		return v
	}
	if v := r.URL.Query().Get("namespace_id"); v != "" {
		return v
	}
	return "default"
}
