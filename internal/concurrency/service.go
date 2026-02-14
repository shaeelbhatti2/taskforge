package concurrency

import (
	"context"
	"fmt"

	"github.com/shaeelbhatti2/taskforge/internal/store"
)

type Service struct {
	store store.Store
}

func New(st store.Store) *Service {
	return &Service{store: st}
}

func (s *Service) CanStart(ctx context.Context, groupID string, maxRunning int) (bool, error) {
	if maxRunning <= 0 {
		return true, nil
	}
	n, err := s.store.CountRunningInGroup(ctx, groupID)
	if err != nil {
		return false, err
	}
	return n < maxRunning, nil
}

func (s *Service) Acquire(ctx context.Context, groupID string, maxRunning int) error {
	ok, err := s.CanStart(ctx, groupID, maxRunning)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("concurrency group %s at capacity", groupID)
	}
	return nil
}
