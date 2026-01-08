package runman

import (
	"testing"

	"github.com/shaeelbhatti2/taskforge/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestCanTransitionPendingToRunning(t *testing.T) {
	require.True(t, domain.CanTransition(domain.RunPending, domain.RunRunning))
}

func TestCanTransitionRunningToSuccess(t *testing.T) {
	require.True(t, domain.CanTransition(domain.RunRunning, domain.RunSuccess))
}

func TestCanTransitionSuccessToRunningInvalid(t *testing.T) {
	require.False(t, domain.CanTransition(domain.RunSuccess, domain.RunRunning))
}

func TestTransitionRunUpdatesStatus(t *testing.T) {
	run := &domain.JobRun{ID: domain.NewID(), Status: domain.RunPending}
	tr, err := domain.TransitionRun(run, domain.RunRunning)
	require.NoError(t, err)
	require.Equal(t, domain.RunRunning, run.Status)
	require.Equal(t, domain.RunPending, tr.From)
}
