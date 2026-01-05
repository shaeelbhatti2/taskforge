package domain

import (
	"fmt"
	"time"
)

var validTransitions = map[RunStatus][]RunStatus{
	RunPending:   {RunRunning, RunCancelled},
	RunRunning:   {RunSuccess, RunFailed, RunCancelled},
	RunFailed:    {RunPending, RunDead},
	RunSuccess:   {},
	RunCancelled: {},
	RunDead:      {},
}

func CanTransition(from, to RunStatus) bool {
	next, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, s := range next {
		if s == to {
			return true
		}
	}
	return false
}

func TransitionRun(run *JobRun, to RunStatus) (*RunTransition, error) {
	if run == nil {
		return nil, fmt.Errorf("run is nil")
	}
	if !CanTransition(run.Status, to) {
		return nil, fmt.Errorf("invalid transition from %s to %s", run.Status, to)
	}
	tr := &RunTransition{
		ID:        NewID(),
		RunID:     run.ID,
		From:      run.Status,
		To:        to,
		CreatedAt: nowUTC(),
	}
	run.Status = to
	return tr, nil
}

func nowUTC() time.Time {
	return time.Now().UTC()
}
