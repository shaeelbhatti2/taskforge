package dlq

import (
	"context"
	"encoding/json"
	"time"

	"github.com/shaeelbhatti2/taskforge/internal/domain"
	"github.com/shaeelbhatti2/taskforge/internal/store"
)

type Service struct {
	store store.Store
}

func New(st store.Store) *Service {
	return &Service{store: st}
}

func (s *Service) Enqueue(ctx context.Context, run *domain.JobRun, reason string) error {
	payload, _ := json.Marshal(map[string]any{
		"run_id":     run.ID,
		"attempt":    run.Attempt,
		"exit_code":  run.ExitCode,
		"stderr":     run.Stderr,
		"error":      run.ErrorMessage,
	})
	entry := &domain.DeadLetterEntry{
		ID:          domain.NewID(),
		NamespaceID: run.NamespaceID,
		JobRunID:    run.ID,
		JobID:       run.JobID,
		Reason:      reason,
		Payload:     string(payload),
		CreatedAt:   time.Now().UTC(),
	}
	return s.store.CreateDeadLetter(ctx, entry)
}

func (s *Service) List(ctx context.Context, namespaceID string) ([]domain.DeadLetterEntry, error) {
	return s.store.ListDeadLetters(ctx, namespaceID)
}

func (s *Service) Discard(ctx context.Context, namespaceID, id string) error {
	return s.store.DeleteDeadLetter(ctx, namespaceID, id)
}

func (s *Service) Replay(ctx context.Context, entry domain.DeadLetterEntry) (*domain.JobRun, error) {
	run := &domain.JobRun{
		ID:            domain.NewID(),
		NamespaceID:   entry.NamespaceID,
		JobID:         entry.JobID,
		Status:        domain.RunPending,
		Attempt:       1,
		ScheduledAt:   time.Now().UTC(),
		CorrelationID: domain.NewID(),
	}
	if err := s.store.CreateJobRun(ctx, run); err != nil {
		return nil, err
	}
	if err := s.store.DeleteDeadLetter(ctx, entry.NamespaceID, entry.ID); err != nil {
		return nil, err
	}
	return run, nil
}
