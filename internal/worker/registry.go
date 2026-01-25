package worker

import (
	"context"
	"time"

	"github.com/shaeelbhatti2/taskforge/internal/domain"
	"github.com/shaeelbhatti2/taskforge/internal/store"
)

type Registry struct {
	store store.Store
}

func NewRegistry(st store.Store) *Registry {
	return &Registry{store: st}
}

func (r *Registry) Register(ctx context.Context, w *domain.Worker) error {
	w.LastSeenAt = time.Now().UTC()
	w.Status = domain.WorkerOnline
	return r.store.CreateWorker(ctx, w)
}

func (r *Registry) Heartbeat(ctx context.Context, namespaceID, workerID string) error {
	w, err := r.store.GetWorker(ctx, namespaceID, workerID)
	if err != nil {
		return err
	}
	w.LastSeenAt = time.Now().UTC()
	w.Status = domain.WorkerOnline
	return r.store.UpdateWorker(ctx, w)
}

func (r *Registry) List(ctx context.Context, namespaceID string) ([]domain.Worker, error) {
	return r.store.ListWorkers(ctx, namespaceID)
}

func (r *Registry) MarkStale(ctx context.Context, namespaceID string, cutoff time.Time) error {
	workers, err := r.store.ListWorkers(ctx, namespaceID)
	if err != nil {
		return err
	}
	for _, w := range workers {
		if w.LastSeenAt.Before(cutoff) && w.Status != domain.WorkerOffline {
			w.Status = domain.WorkerOffline
			if err := r.store.UpdateWorker(ctx, &w); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Registry) MatchTags(workers []domain.Worker, required []string) []domain.Worker {
	if len(required) == 0 {
		return workers
	}
	var out []domain.Worker
	for _, w := range workers {
		if hasTags(w.Tags, required) {
			out = append(out, w)
		}
	}
	return out
}

func hasTags(workerTags, required []string) bool {
	set := map[string]struct{}{}
	for _, t := range workerTags {
		set[t] = struct{}{}
	}
	for _, req := range required {
		if _, ok := set[req]; !ok {
			return false
		}
	}
	return true
}
