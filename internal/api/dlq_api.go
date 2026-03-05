package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shaeelbhatti2/taskforge/internal/dlq"
)

func (s *Server) mountDLQ(r chi.Router) {
	svc := dlq.New(s.store)
	r.Route("/dlq", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			items, err := svc.List(r.Context(), namespaceID(r))
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, items)
		})
		r.Post("/{id}/replay", func(w http.ResponseWriter, r *http.Request) {
			ns := namespaceID(r)
			items, err := svc.List(r.Context(), ns)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			id := chi.URLParam(r, "id")
			for _, e := range items {
				if e.ID == id {
					run, err := svc.Replay(r.Context(), e)
					if err != nil {
						writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
						return
					}
					writeJSON(w, http.StatusAccepted, run)
					return
				}
			}
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		})
		r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
			if err := svc.Discard(r.Context(), namespaceID(r), chi.URLParam(r, "id")); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	})
}
