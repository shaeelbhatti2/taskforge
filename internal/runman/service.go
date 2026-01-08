package runman

import (
	"context"
	"fmt"
	"time"

	"github.com/shaeelbhatti2/taskforge/internal/domain"
	"github.com/shaeelbhatti2/taskforge/internal/store"
)

type Service struct {
	store store.Store
}

func NewService(st store.Store) *Service {
	return &Service{store: st}
}

func (s *Service) StartRun(ctx context.Context, run *domain.JobRun) error {
	if run.Status != domain.RunPending {
		return fmt.Errorf("run must start pending")
	}
	now := time.Now().UTC()
	run.StartedAt = &now
	tr, err := domain.TransitionRun(run, domain.RunRunning)
	if err != nil {
		return err
	}
	if err := s.store.UpdateJobRun(ctx, run); err != nil {
		return err
	}
	return s.store.RecordTransition(ctx, tr)
}

func (s *Service) CompleteRun(ctx context.Context, run *domain.JobRun, exitCode int, stdout, stderr string) error {
	if run.Status != domain.RunRunning {
		return fmt.Errorf("run not running")
	}
	now := time.Now().UTC()
	run.FinishedAt = &now
	run.ExitCode = &exitCode
	run.Stdout = stdout
	run.Stderr = stderr
	target := domain.RunSuccess
	if exitCode != 0 {
		target = domain.RunFailed
	}
	tr, err := domain.TransitionRun(run, target)
	if err != nil {
		return err
	}
	if err := s.store.UpdateJobRun(ctx, run); err != nil {
		return err
	}
	return s.store.RecordTransition(ctx, tr)
}

func (s *Service) CancelRun(ctx context.Context, run *domain.JobRun) error {
	tr, err := domain.TransitionRun(run, domain.RunCancelled)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	run.FinishedAt = &now
	if err := s.store.UpdateJobRun(ctx, run); err != nil {
		return err
	}
	return s.store.RecordTransition(ctx, tr)
}

func (s *Service) MarkDead(ctx context.Context, run *domain.JobRun, reason string) error {
	tr, err := domain.TransitionRun(run, domain.RunDead)
	if err != nil {
		return err
	}
	run.ErrorMessage = reason
	now := time.Now().UTC()
	run.FinishedAt = &now
	if err := s.store.UpdateJobRun(ctx, run); err != nil {
		return err
	}
	return s.store.RecordTransition(ctx, tr)
}

func (s *Service) RetryRun(ctx context.Context, run *domain.JobRun) error {
	tr, err := domain.TransitionRun(run, domain.RunPending)
	if err != nil {
		return err
	}
	run.Attempt++
	run.StartedAt = nil
	run.FinishedAt = nil
	run.ExitCode = nil
	if err := s.store.UpdateJobRun(ctx, run); err != nil {
		return err
	}
	return s.store.RecordTransition(ctx, tr)
}
