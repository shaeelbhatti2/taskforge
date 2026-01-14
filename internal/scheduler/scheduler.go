package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/shaeelbhatti2/taskforge/internal/config"
	"github.com/shaeelbhatti2/taskforge/internal/domain"
	"github.com/shaeelbhatti2/taskforge/internal/store"
)

type Dispatcher func(ctx context.Context, schedule domain.Schedule, job *domain.JobDefinition) error

type Scheduler struct {
	store      store.Store
	cfg        *config.Config
	cron       *CronService
	dispatch   Dispatcher
	holderID   string
}

func New(st store.Store, cfg *config.Config, dispatch Dispatcher, holderID string) *Scheduler {
	return &Scheduler{
		store:    st,
		cfg:      cfg,
		cron:     NewCronService(),
		dispatch: dispatch,
		holderID: holderID,
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.SchedulerTick)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			s.tick(ctx, now)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context, now time.Time) {
	ok, err := s.store.TryAcquireLeader(ctx, s.cfg.LeaderLockID, s.holderID)
	if err != nil || !ok {
		return
	}
	defer func() { _ = s.store.ReleaseLeader(ctx, s.cfg.LeaderLockID, s.holderID) }()
	due, err := s.store.ListDueSchedules(ctx, now)
	if err != nil {
		slog.Error("list due schedules", "err", err)
		return
	}
	for _, sch := range due {
		job, err := s.store.GetJob(ctx, sch.NamespaceID, sch.JobID)
		if err != nil {
			slog.Error("load job for schedule", "schedule", sch.ID, "err", err)
			continue
		}
		if err := s.dispatch(ctx, sch, job); err != nil {
			slog.Error("dispatch schedule", "schedule", sch.ID, "err", err)
			continue
		}
		next := s.computeNext(sch, now)
		sch.LastRunAt = &now
		sch.NextRunAt = &next
		if err := s.store.UpdateSchedule(ctx, &sch); err != nil {
			slog.Error("update schedule next run", "schedule", sch.ID, "err", err)
		}
	}
}

func (s *Scheduler) computeNext(sch domain.Schedule, from time.Time) time.Time {
	if sch.CronExpr != "" {
		next, err := s.cron.Next(sch.CronExpr, from)
		if err != nil {
			return from.Add(time.Minute)
		}
		return next
	}
	if sch.IntervalSec > 0 {
		return NextInterval(from, sch.IntervalSec, sch.Timezone)
	}
	return from.Add(time.Minute)
}

func (s *Scheduler) SeedNextRun(sch *domain.Schedule) error {
	now := time.Now().UTC()
	next := s.computeNext(*sch, now)
	sch.NextRunAt = &next
	return nil
}
