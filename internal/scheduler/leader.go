package scheduler

import (
	"context"

	"github.com/shaeelbhatti2/taskforge/internal/store"
)

type Leader struct {
	store    store.Store
	lockID   int64
	holderID string
}

func NewLeader(st store.Store, lockID int64, holderID string) *Leader {
	return &Leader{store: st, lockID: lockID, holderID: holderID}
}

func (l *Leader) Acquire(ctx context.Context) (bool, error) {
	return l.store.TryAcquireLeader(ctx, l.lockID, l.holderID)
}

func (l *Leader) Release(ctx context.Context) error {
	return l.store.ReleaseLeader(ctx, l.lockID, l.holderID)
}

func (l *Leader) WithLeadership(ctx context.Context, fn func(context.Context) error) error {
	ok, err := l.Acquire(ctx)
	if err != nil || !ok {
		return err
	}
	defer func() { _ = l.Release(ctx) }()
	return fn(ctx)
}
