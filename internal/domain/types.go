package domain

import (
	"time"

	"github.com/google/uuid"
)

type RunStatus string

const (
	RunPending   RunStatus = "PENDING"
	RunRunning   RunStatus = "RUNNING"
	RunSuccess   RunStatus = "SUCCESS"
	RunFailed    RunStatus = "FAILED"
	RunCancelled RunStatus = "CANCELLED"
	RunDead      RunStatus = "DEAD"
)

type JobType string

const (
	JobTypeShell JobType = "shell"
	JobTypeHTTP  JobType = "http"
	JobTypePlugin JobType = "plugin"
)

type Namespace struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RetryPolicy struct {
	ID           string `json:"id"`
	MaxAttempts  int    `json:"max_attempts"`
	DelaySeconds int    `json:"delay_seconds"`
	BackoffBase  int    `json:"backoff_base"`
	MaxDelaySec  int    `json:"max_delay_sec"`
	RetryOnCodes []int  `json:"retry_on_codes,omitempty"`
}

type ConcurrencyGroup struct {
	ID          string `json:"id"`
	NamespaceID string `json:"namespace_id"`
	Name        string `json:"name"`
	MaxRunning  int    `json:"max_running"`
}

type JobDefinition struct {
	ID                string            `json:"id"`
	NamespaceID       string            `json:"namespace_id"`
	Name              string            `json:"name"`
	Type              JobType           `json:"type"`
	Command           string            `json:"command,omitempty"`
	HTTPMethod        string            `json:"http_method,omitempty"`
	HTTPURL           string            `json:"http_url,omitempty"`
	HTTPHeaders       map[string]string `json:"http_headers,omitempty"`
	HTTPBody          string            `json:"http_body,omitempty"`
	ExpectedStatus    int               `json:"expected_status,omitempty"`
	TimeoutSec        int               `json:"timeout_sec"`
	RetryPolicyID     string            `json:"retry_policy_id,omitempty"`
	ConcurrencyGroupID string           `json:"concurrency_group_id,omitempty"`
	ConcurrencyLimit  int               `json:"concurrency_limit"`
	RateLimitPerMin   int               `json:"rate_limit_per_min"`
	IdempotencyKey    string            `json:"idempotency_key,omitempty"`
	Env               map[string]string `json:"env,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

type Schedule struct {
	ID            string     `json:"id"`
	NamespaceID   string     `json:"namespace_id"`
	JobID         string     `json:"job_id"`
	CronExpr      string     `json:"cron_expr,omitempty"`
	IntervalSec   int        `json:"interval_sec,omitempty"`
	Timezone      string     `json:"timezone"`
	Paused        bool       `json:"paused"`
	MissedPolicy  string     `json:"missed_policy"`
	NextRunAt     *time.Time `json:"next_run_at,omitempty"`
	LastRunAt     *time.Time `json:"last_run_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type WorkflowNodeType string

const (
	NodeJob       WorkflowNodeType = "job"
	NodeFork      WorkflowNodeType = "fork"
	NodeJoin      WorkflowNodeType = "join"
	NodeCondition WorkflowNodeType = "condition"
)

type WorkflowNode struct {
	ID       string           `json:"id"`
	Type     WorkflowNodeType `json:"type"`
	JobID    string           `json:"job_id,omitempty"`
	Label    string           `json:"label"`
	Config   map[string]string `json:"config,omitempty"`
}

type WorkflowEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Workflow struct {
	ID            string         `json:"id"`
	NamespaceID   string         `json:"namespace_id"`
	Name          string         `json:"name"`
	Nodes         []WorkflowNode `json:"nodes"`
	Edges         []WorkflowEdge `json:"edges"`
	TimeoutSec    int            `json:"timeout_sec"`
	FailFast      bool           `json:"fail_fast"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type WorkflowRun struct {
	ID          string     `json:"id"`
	WorkflowID  string     `json:"workflow_id"`
	NamespaceID string     `json:"namespace_id"`
	Status      RunStatus  `json:"status"`
	StartedAt   time.Time  `json:"started_at"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
}

type JobRun struct {
	ID              string     `json:"id"`
	NamespaceID     string     `json:"namespace_id"`
	JobID           string     `json:"job_id"`
	WorkflowRunID   string     `json:"workflow_run_id,omitempty"`
	ParentRunID     string     `json:"parent_run_id,omitempty"`
	Status          RunStatus  `json:"status"`
	Attempt         int        `json:"attempt"`
	ScheduledAt     time.Time  `json:"scheduled_at"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
	ExitCode        *int       `json:"exit_code,omitempty"`
	Stdout          string     `json:"stdout,omitempty"`
	Stderr          string     `json:"stderr,omitempty"`
	CorrelationID   string     `json:"correlation_id"`
	WorkerID        string     `json:"worker_id,omitempty"`
	IdempotencyKey  string     `json:"idempotency_key,omitempty"`
	ErrorMessage    string     `json:"error_message,omitempty"`
}

type WorkerStatus string

const (
	WorkerOnline  WorkerStatus = "online"
	WorkerOffline WorkerStatus = "offline"
	WorkerBusy    WorkerStatus = "busy"
)

type Worker struct {
	ID          string       `json:"id"`
	NamespaceID string       `json:"namespace_id"`
	Name        string       `json:"name"`
	Tags        []string     `json:"tags"`
	Capacity    int          `json:"capacity"`
	Status      WorkerStatus `json:"status"`
	LastSeenAt  time.Time    `json:"last_seen_at"`
	CreatedAt   time.Time    `json:"created_at"`
}

type DeadLetterEntry struct {
	ID          string    `json:"id"`
	NamespaceID string    `json:"namespace_id"`
	JobRunID    string    `json:"job_run_id"`
	JobID       string    `json:"job_id"`
	Reason      string    `json:"reason"`
	Payload     string    `json:"payload"`
	CreatedAt   time.Time `json:"created_at"`
}

type Webhook struct {
	ID          string   `json:"id"`
	NamespaceID string   `json:"namespace_id"`
	URL         string   `json:"url"`
	Secret      string   `json:"secret"`
	Events      []string `json:"events"`
	Enabled     bool     `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
}

type APIKey struct {
	ID          string    `json:"id"`
	NamespaceID string    `json:"namespace_id"`
	Name        string    `json:"name"`
	KeyHash     string    `json:"-"`
	Prefix      string    `json:"prefix"`
	CreatedAt   time.Time `json:"created_at"`
}

type RunTransition struct {
	ID        string    `json:"id"`
	RunID     string    `json:"run_id"`
	From      RunStatus `json:"from"`
	To        RunStatus `json:"to"`
	CreatedAt time.Time `json:"created_at"`
}

func NewID() string {
	return uuid.NewString()
}
