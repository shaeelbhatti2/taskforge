package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/shaeelbhatti2/taskforge/internal/domain"
	"github.com/shaeelbhatti2/taskforge/internal/store"
	"github.com/stretchr/testify/require"
)

func TestSQLiteNamespaceJobFlow(t *testing.T) {
	st, err := store.Open("file::memory:?cache=shared")
	require.NoError(t, err)
	defer st.Close()
	ctx := context.Background()
	require.NoError(t, st.Migrate(ctx))
	now := time.Now().UTC()
	ns := &domain.Namespace{ID: domain.NewID(), Name: "default", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, st.CreateNamespace(ctx, ns))
	job := &domain.JobDefinition{
		ID: domain.NewID(), NamespaceID: ns.ID, Name: "echo-test",
		Type: domain.JobTypeShell, Command: "echo hi", TimeoutSec: 60,
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, st.CreateJob(ctx, job))
	got, err := st.GetJob(ctx, ns.ID, job.ID)
	require.NoError(t, err)
	require.Equal(t, "echo-test", got.Name)
}

func TestNamespaceIsolation(t *testing.T) {
	st, err := store.Open("file::memory:?cache=shared")
	require.NoError(t, err)
	defer st.Close()
	ctx := context.Background()
	require.NoError(t, st.Migrate(ctx))
	now := time.Now().UTC()
	nsA := &domain.Namespace{ID: domain.NewID(), Name: "team-a", CreatedAt: now, UpdatedAt: now}
	nsB := &domain.Namespace{ID: domain.NewID(), Name: "team-b", CreatedAt: now, UpdatedAt: now}
	require.NoError(t, st.CreateNamespace(ctx, nsA))
	require.NoError(t, st.CreateNamespace(ctx, nsB))
	job := &domain.JobDefinition{
		ID: domain.NewID(), NamespaceID: nsA.ID, Name: "job-a",
		Type: domain.JobTypeShell, Command: "true", TimeoutSec: 30,
		CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, st.CreateJob(ctx, job))
	_, err = st.GetJob(ctx, nsB.ID, job.ID)
	require.Error(t, err)
}
