package store

import (
	"context"
	"time"

	"github.com/shaeelbhatti2/taskforge/internal/domain"
)

type Store interface {
	Ping(ctx context.Context) error
	Close() error
	Migrate(ctx context.Context) error

	CreateNamespace(ctx context.Context, ns *domain.Namespace) error
	GetNamespace(ctx context.Context, id string) (*domain.Namespace, error)
	ListNamespaces(ctx context.Context) ([]domain.Namespace, error)

	CreateJob(ctx context.Context, job *domain.JobDefinition) error
	GetJob(ctx context.Context, namespaceID, id string) (*domain.JobDefinition, error)
	ListJobs(ctx context.Context, namespaceID string) ([]domain.JobDefinition, error)
	UpdateJob(ctx context.Context, job *domain.JobDefinition) error
	DeleteJob(ctx context.Context, namespaceID, id string) error

	CreateSchedule(ctx context.Context, s *domain.Schedule) error
	GetSchedule(ctx context.Context, namespaceID, id string) (*domain.Schedule, error)
	ListSchedules(ctx context.Context, namespaceID string) ([]domain.Schedule, error)
	UpdateSchedule(ctx context.Context, s *domain.Schedule) error
	ListDueSchedules(ctx context.Context, before time.Time) ([]domain.Schedule, error)

	CreateJobRun(ctx context.Context, run *domain.JobRun) error
	GetJobRun(ctx context.Context, namespaceID, id string) (*domain.JobRun, error)
	UpdateJobRun(ctx context.Context, run *domain.JobRun) error
	ListJobRuns(ctx context.Context, filter RunFilter) ([]domain.JobRun, error)
	RecordTransition(ctx context.Context, tr *domain.RunTransition) error
	CountRunningInGroup(ctx context.Context, groupID string) (int, error)

	CreateWorker(ctx context.Context, w *domain.Worker) error
	UpdateWorker(ctx context.Context, w *domain.Worker) error
	GetWorker(ctx context.Context, namespaceID, id string) (*domain.Worker, error)
	ListWorkers(ctx context.Context, namespaceID string) ([]domain.Worker, error)

	CreateWorkflow(ctx context.Context, wf *domain.Workflow) error
	GetWorkflow(ctx context.Context, namespaceID, id string) (*domain.Workflow, error)
	ListWorkflows(ctx context.Context, namespaceID string) ([]domain.Workflow, error)

	CreateWorkflowRun(ctx context.Context, wr *domain.WorkflowRun) error
	UpdateWorkflowRun(ctx context.Context, wr *domain.WorkflowRun) error

	CreateDeadLetter(ctx context.Context, entry *domain.DeadLetterEntry) error
	ListDeadLetters(ctx context.Context, namespaceID string) ([]domain.DeadLetterEntry, error)
	DeleteDeadLetter(ctx context.Context, namespaceID, id string) error

	CreateWebhook(ctx context.Context, wh *domain.Webhook) error
	ListWebhooks(ctx context.Context, namespaceID string) ([]domain.Webhook, error)

	CreateAPIKey(ctx context.Context, key *domain.APIKey) error
	GetAPIKeyByPrefix(ctx context.Context, prefix string) (*domain.APIKey, error)

	TryAcquireLeader(ctx context.Context, lockID int64, holder string) (bool, error)
	ReleaseLeader(ctx context.Context, lockID int64, holder string) error
}

type RunFilter struct {
	NamespaceID string
	JobID       string
	Status      domain.RunStatus
	Since       *time.Time
	Until       *time.Time
	Limit       int
	Offset      int
}

func Open(databaseURL string) (Store, error) {
	if isSQLite(databaseURL) {
		return openSQLite(databaseURL)
	}
	return openPostgres(databaseURL)
}

func isSQLite(url string) bool {
	return len(url) >= 5 && url[:5] == "file:"
}
