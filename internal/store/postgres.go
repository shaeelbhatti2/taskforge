package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/shaeelbhatti2/taskforge/internal/domain"
)

type pgStore struct {
	db *sql.DB
}

func openPostgres(url string) (Store, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(5 * time.Minute)
	return &pgStore{db: db}, nil
}

func (s *pgStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *pgStore) Close() error {
	return s.db.Close()
}

func (s *pgStore) Migrate(ctx context.Context) error {
	return runMigrations(ctx, s.db)
}

func (s *pgStore) CreateNamespace(ctx context.Context, ns *domain.Namespace) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO namespaces (id, name, created_at, updated_at) VALUES ($1,$2,$3,$4)`,
		ns.ID, ns.Name, ns.CreatedAt, ns.UpdatedAt)
	return err
}

func (s *pgStore) GetNamespace(ctx context.Context, id string) (*domain.Namespace, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, name, created_at, updated_at FROM namespaces WHERE id=$1`, id)
	var ns domain.Namespace
	if err := row.Scan(&ns.ID, &ns.Name, &ns.CreatedAt, &ns.UpdatedAt); err != nil {
		return nil, err
	}
	return &ns, nil
}

func (s *pgStore) ListNamespaces(ctx context.Context) ([]domain.Namespace, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, created_at, updated_at FROM namespaces ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Namespace
	for rows.Next() {
		var ns domain.Namespace
		if err := rows.Scan(&ns.ID, &ns.Name, &ns.CreatedAt, &ns.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, ns)
	}
	return out, rows.Err()
}

func (s *pgStore) CreateJob(ctx context.Context, job *domain.JobDefinition) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO job_definitions (
			id, namespace_id, name, type, command, http_method, http_url, http_headers,
			http_body, expected_status, timeout_sec, retry_policy_id, concurrency_group_id,
			concurrency_limit, rate_limit_per_min, idempotency_key, env, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
		job.ID, job.NamespaceID, job.Name, job.Type, job.Command, job.HTTPMethod, job.HTTPURL,
		encodeJSON(job.HTTPHeaders), job.HTTPBody, job.ExpectedStatus, job.TimeoutSec,
		job.RetryPolicyID, job.ConcurrencyGroupID, job.ConcurrencyLimit, job.RateLimitPerMin,
		job.IdempotencyKey, encodeJSON(job.Env), job.CreatedAt, job.UpdatedAt)
	return err
}

func (s *pgStore) GetJob(ctx context.Context, namespaceID, id string) (*domain.JobDefinition, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, namespace_id, name, type, command, http_method, http_url, http_headers,
			http_body, expected_status, timeout_sec, retry_policy_id, concurrency_group_id,
			concurrency_limit, rate_limit_per_min, idempotency_key, env, created_at, updated_at
		FROM job_definitions WHERE namespace_id=$1 AND id=$2`, namespaceID, id)
	return scanJob(row)
}

func (s *pgStore) ListJobs(ctx context.Context, namespaceID string) ([]domain.JobDefinition, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, namespace_id, name, type, command, http_method, http_url, http_headers,
			http_body, expected_status, timeout_sec, retry_policy_id, concurrency_group_id,
			concurrency_limit, rate_limit_per_min, idempotency_key, env, created_at, updated_at
		FROM job_definitions WHERE namespace_id=$1 ORDER BY name`, namespaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.JobDefinition
	for rows.Next() {
		job, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *job)
	}
	return out, rows.Err()
}

func (s *pgStore) UpdateJob(ctx context.Context, job *domain.JobDefinition) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE job_definitions SET name=$3, type=$4, command=$5, http_method=$6, http_url=$7,
			http_headers=$8, http_body=$9, expected_status=$10, timeout_sec=$11,
			retry_policy_id=$12, concurrency_group_id=$13, concurrency_limit=$14,
			rate_limit_per_min=$15, idempotency_key=$16, env=$17, updated_at=$18
		WHERE namespace_id=$1 AND id=$2`,
		job.NamespaceID, job.ID, job.Name, job.Type, job.Command, job.HTTPMethod, job.HTTPURL,
		encodeJSON(job.HTTPHeaders), job.HTTPBody, job.ExpectedStatus, job.TimeoutSec,
		job.RetryPolicyID, job.ConcurrencyGroupID, job.ConcurrencyLimit, job.RateLimitPerMin,
		job.IdempotencyKey, encodeJSON(job.Env), job.UpdatedAt)
	return err
}

func (s *pgStore) DeleteJob(ctx context.Context, namespaceID, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM job_definitions WHERE namespace_id=$1 AND id=$2`, namespaceID, id)
	return err
}

func (s *pgStore) CreateSchedule(ctx context.Context, sch *domain.Schedule) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO schedules (id, namespace_id, job_id, cron_expr, interval_sec, timezone,
			paused, missed_policy, next_run_at, last_run_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		sch.ID, sch.NamespaceID, sch.JobID, sch.CronExpr, sch.IntervalSec, sch.Timezone,
		sch.Paused, sch.MissedPolicy, sch.NextRunAt, sch.LastRunAt, sch.CreatedAt, sch.UpdatedAt)
	return err
}

func (s *pgStore) GetSchedule(ctx context.Context, namespaceID, id string) (*domain.Schedule, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, namespace_id, job_id, cron_expr, interval_sec, timezone, paused, missed_policy,
			next_run_at, last_run_at, created_at, updated_at
		FROM schedules WHERE namespace_id=$1 AND id=$2`, namespaceID, id)
	var sch domain.Schedule
	if err := row.Scan(&sch.ID, &sch.NamespaceID, &sch.JobID, &sch.CronExpr, &sch.IntervalSec,
		&sch.Timezone, &sch.Paused, &sch.MissedPolicy, &sch.NextRunAt, &sch.LastRunAt,
		&sch.CreatedAt, &sch.UpdatedAt); err != nil {
		return nil, err
	}
	return &sch, nil
}

func (s *pgStore) ListSchedules(ctx context.Context, namespaceID string) ([]domain.Schedule, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, namespace_id, job_id, cron_expr, interval_sec, timezone, paused, missed_policy,
			next_run_at, last_run_at, created_at, updated_at
		FROM schedules WHERE namespace_id=$1 ORDER BY created_at DESC`, namespaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Schedule
	for rows.Next() {
		var sch domain.Schedule
		if err := rows.Scan(&sch.ID, &sch.NamespaceID, &sch.JobID, &sch.CronExpr, &sch.IntervalSec,
			&sch.Timezone, &sch.Paused, &sch.MissedPolicy, &sch.NextRunAt, &sch.LastRunAt,
			&sch.CreatedAt, &sch.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, sch)
	}
	return out, rows.Err()
}

func (s *pgStore) UpdateSchedule(ctx context.Context, sch *domain.Schedule) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE schedules SET cron_expr=$3, interval_sec=$4, timezone=$5, paused=$6,
			missed_policy=$7, next_run_at=$8, last_run_at=$9, updated_at=$10
		WHERE namespace_id=$1 AND id=$2`,
		sch.NamespaceID, sch.ID, sch.CronExpr, sch.IntervalSec, sch.Timezone, sch.Paused,
		sch.MissedPolicy, sch.NextRunAt, sch.LastRunAt, sch.UpdatedAt)
	return err
}

func (s *pgStore) ListDueSchedules(ctx context.Context, before time.Time) ([]domain.Schedule, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, namespace_id, job_id, cron_expr, interval_sec, timezone, paused, missed_policy,
			next_run_at, last_run_at, created_at, updated_at
		FROM schedules WHERE paused=FALSE AND next_run_at IS NOT NULL AND next_run_at <= $1
		ORDER BY next_run_at LIMIT 500`, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Schedule
	for rows.Next() {
		var sch domain.Schedule
		if err := rows.Scan(&sch.ID, &sch.NamespaceID, &sch.JobID, &sch.CronExpr, &sch.IntervalSec,
			&sch.Timezone, &sch.Paused, &sch.MissedPolicy, &sch.NextRunAt, &sch.LastRunAt,
			&sch.CreatedAt, &sch.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, sch)
	}
	return out, rows.Err()
}

func (s *pgStore) CreateJobRun(ctx context.Context, run *domain.JobRun) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO job_runs (id, namespace_id, job_id, workflow_run_id, parent_run_id, status,
			attempt, scheduled_at, started_at, finished_at, exit_code, stdout, stderr,
			correlation_id, worker_id, idempotency_key, error_message)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		run.ID, run.NamespaceID, run.JobID, nullStr(run.WorkflowRunID), nullStr(run.ParentRunID),
		run.Status, run.Attempt, run.ScheduledAt, run.StartedAt, run.FinishedAt, run.ExitCode,
		run.Stdout, run.Stderr, run.CorrelationID, nullStr(run.WorkerID), nullStr(run.IdempotencyKey),
		run.ErrorMessage)
	return err
}

func (s *pgStore) GetJobRun(ctx context.Context, namespaceID, id string) (*domain.JobRun, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, namespace_id, job_id, workflow_run_id, parent_run_id, status, attempt,
			scheduled_at, started_at, finished_at, exit_code, stdout, stderr, correlation_id,
			worker_id, idempotency_key, error_message
		FROM job_runs WHERE namespace_id=$1 AND id=$2`, namespaceID, id)
	return scanJobRun(row)
}

func (s *pgStore) UpdateJobRun(ctx context.Context, run *domain.JobRun) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE job_runs SET status=$3, attempt=$4, started_at=$5, finished_at=$6, exit_code=$7,
			stdout=$8, stderr=$9, worker_id=$10, error_message=$11
		WHERE namespace_id=$1 AND id=$2`,
		run.NamespaceID, run.ID, run.Status, run.Attempt, run.StartedAt, run.FinishedAt,
		run.ExitCode, run.Stdout, run.Stderr, nullStr(run.WorkerID), run.ErrorMessage)
	return err
}

func (s *pgStore) ListJobRuns(ctx context.Context, filter RunFilter) ([]domain.JobRun, error) {
	q := `SELECT id, namespace_id, job_id, workflow_run_id, parent_run_id, status, attempt,
		scheduled_at, started_at, finished_at, exit_code, stdout, stderr, correlation_id,
		worker_id, idempotency_key, error_message FROM job_runs WHERE namespace_id=$1`
	args := []any{filter.NamespaceID}
	n := 2
	if filter.JobID != "" {
		q += fmt.Sprintf(" AND job_id=$%d", n)
		args = append(args, filter.JobID)
		n++
	}
	if filter.Status != "" {
		q += fmt.Sprintf(" AND status=$%d", n)
		args = append(args, filter.Status)
		n++
	}
	if filter.Since != nil {
		q += fmt.Sprintf(" AND scheduled_at >= $%d", n)
		args = append(args, *filter.Since)
		n++
	}
	if filter.Until != nil {
		q += fmt.Sprintf(" AND scheduled_at <= $%d", n)
		args = append(args, *filter.Until)
		n++
	}
	q += " ORDER BY scheduled_at DESC"
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	q += fmt.Sprintf(" LIMIT $%d OFFSET $%d", n, n+1)
	args = append(args, limit, filter.Offset)
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.JobRun
	for rows.Next() {
		run, err := scanJobRun(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *run)
	}
	return out, rows.Err()
}

func (s *pgStore) RecordTransition(ctx context.Context, tr *domain.RunTransition) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO run_transitions (id, run_id, from_status, to_status, created_at) VALUES ($1,$2,$3,$4,$5)`,
		tr.ID, tr.RunID, tr.From, tr.To, tr.CreatedAt)
	return err
}

func (s *pgStore) CountRunningInGroup(ctx context.Context, groupID string) (int, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM job_runs jr
		JOIN job_definitions jd ON jr.job_id = jd.id
		WHERE jd.concurrency_group_id=$1 AND jr.status='RUNNING'`, groupID)
	var n int
	return n, row.Scan(&n)
}

func (s *pgStore) CreateWorker(ctx context.Context, w *domain.Worker) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO workers (id, namespace_id, name, tags, capacity, status, last_seen_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		w.ID, w.NamespaceID, w.Name, encodeStringSlice(w.Tags), w.Capacity, w.Status, w.LastSeenAt, w.CreatedAt)
	return err
}

func (s *pgStore) UpdateWorker(ctx context.Context, w *domain.Worker) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE workers SET tags=$3, capacity=$4, status=$5, last_seen_at=$6
		WHERE namespace_id=$1 AND id=$2`,
		w.NamespaceID, w.ID, encodeStringSlice(w.Tags), w.Capacity, w.Status, w.LastSeenAt)
	return err
}

func (s *pgStore) GetWorker(ctx context.Context, namespaceID, id string) (*domain.Worker, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, namespace_id, name, tags, capacity, status, last_seen_at, created_at
		FROM workers WHERE namespace_id=$1 AND id=$2`, namespaceID, id)
	return scanWorker(row)
}

func (s *pgStore) ListWorkers(ctx context.Context, namespaceID string) ([]domain.Worker, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, namespace_id, name, tags, capacity, status, last_seen_at, created_at
		FROM workers WHERE namespace_id=$1 ORDER BY name`, namespaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Worker
	for rows.Next() {
		w, err := scanWorker(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *w)
	}
	return out, rows.Err()
}

func (s *pgStore) CreateWorkflow(ctx context.Context, wf *domain.Workflow) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO workflows (id, namespace_id, name, nodes, edges, timeout_sec, fail_fast, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		wf.ID, wf.NamespaceID, wf.Name, encodeJSON(wf.Nodes), encodeJSON(wf.Edges),
		wf.TimeoutSec, wf.FailFast, wf.CreatedAt, wf.UpdatedAt)
	return err
}

func (s *pgStore) GetWorkflow(ctx context.Context, namespaceID, id string) (*domain.Workflow, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, namespace_id, name, nodes, edges, timeout_sec, fail_fast, created_at, updated_at
		FROM workflows WHERE namespace_id=$1 AND id=$2`, namespaceID, id)
	var wf domain.Workflow
	var nodesRaw, edgesRaw string
	if err := row.Scan(&wf.ID, &wf.NamespaceID, &wf.Name, &nodesRaw, &edgesRaw,
		&wf.TimeoutSec, &wf.FailFast, &wf.CreatedAt, &wf.UpdatedAt); err != nil {
		return nil, err
	}
	decodeJSON(nodesRaw, &wf.Nodes)
	decodeJSON(edgesRaw, &wf.Edges)
	return &wf, nil
}

func (s *pgStore) ListWorkflows(ctx context.Context, namespaceID string) ([]domain.Workflow, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, namespace_id, name, nodes, edges, timeout_sec, fail_fast, created_at, updated_at
		FROM workflows WHERE namespace_id=$1 ORDER BY name`, namespaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Workflow
	for rows.Next() {
		var wf domain.Workflow
		var nodesRaw, edgesRaw string
		if err := rows.Scan(&wf.ID, &wf.NamespaceID, &wf.Name, &nodesRaw, &edgesRaw,
			&wf.TimeoutSec, &wf.FailFast, &wf.CreatedAt, &wf.UpdatedAt); err != nil {
			return nil, err
		}
		decodeJSON(nodesRaw, &wf.Nodes)
		decodeJSON(edgesRaw, &wf.Edges)
		out = append(out, wf)
	}
	return out, rows.Err()
}

func (s *pgStore) CreateWorkflowRun(ctx context.Context, wr *domain.WorkflowRun) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO workflow_runs (id, workflow_id, namespace_id, status, started_at, finished_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		wr.ID, wr.WorkflowID, wr.NamespaceID, wr.Status, wr.StartedAt, wr.FinishedAt)
	return err
}

func (s *pgStore) UpdateWorkflowRun(ctx context.Context, wr *domain.WorkflowRun) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE workflow_runs SET status=$2, finished_at=$3 WHERE id=$1`,
		wr.ID, wr.Status, wr.FinishedAt)
	return err
}

func (s *pgStore) CreateDeadLetter(ctx context.Context, entry *domain.DeadLetterEntry) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO dead_letter_entries (id, namespace_id, job_run_id, job_id, reason, payload, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		entry.ID, entry.NamespaceID, entry.JobRunID, entry.JobID, entry.Reason, entry.Payload, entry.CreatedAt)
	return err
}

func (s *pgStore) ListDeadLetters(ctx context.Context, namespaceID string) ([]domain.DeadLetterEntry, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, namespace_id, job_run_id, job_id, reason, payload, created_at
		FROM dead_letter_entries WHERE namespace_id=$1 ORDER BY created_at DESC`, namespaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.DeadLetterEntry
	for rows.Next() {
		var e domain.DeadLetterEntry
		if err := rows.Scan(&e.ID, &e.NamespaceID, &e.JobRunID, &e.JobID, &e.Reason, &e.Payload, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *pgStore) DeleteDeadLetter(ctx context.Context, namespaceID, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM dead_letter_entries WHERE namespace_id=$1 AND id=$2`, namespaceID, id)
	return err
}

func (s *pgStore) CreateWebhook(ctx context.Context, wh *domain.Webhook) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO webhooks (id, namespace_id, url, secret, events, enabled, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		wh.ID, wh.NamespaceID, wh.URL, wh.Secret, encodeStringSlice(wh.Events), wh.Enabled, wh.CreatedAt)
	return err
}

func (s *pgStore) ListWebhooks(ctx context.Context, namespaceID string) ([]domain.Webhook, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, namespace_id, url, secret, events, enabled, created_at
		FROM webhooks WHERE namespace_id=$1`, namespaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Webhook
	for rows.Next() {
		var wh domain.Webhook
		var eventsRaw string
		if err := rows.Scan(&wh.ID, &wh.NamespaceID, &wh.URL, &wh.Secret, &eventsRaw, &wh.Enabled, &wh.CreatedAt); err != nil {
			return nil, err
		}
		wh.Events = decodeStringSlice(eventsRaw)
		out = append(out, wh)
	}
	return out, rows.Err()
}

func (s *pgStore) CreateAPIKey(ctx context.Context, key *domain.APIKey) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO api_keys (id, namespace_id, name, key_hash, prefix, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		key.ID, key.NamespaceID, key.Name, key.KeyHash, key.Prefix, key.CreatedAt)
	return err
}

func (s *pgStore) GetAPIKeyByPrefix(ctx context.Context, prefix string) (*domain.APIKey, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, namespace_id, name, key_hash, prefix, created_at
		FROM api_keys WHERE prefix=$1`, prefix)
	var key domain.APIKey
	if err := row.Scan(&key.ID, &key.NamespaceID, &key.Name, &key.KeyHash, &key.Prefix, &key.CreatedAt); err != nil {
		return nil, err
	}
	return &key, nil
}

func (s *pgStore) TryAcquireLeader(ctx context.Context, lockID int64, holder string) (bool, error) {
	row := s.db.QueryRowContext(ctx, `SELECT pg_try_advisory_lock($1)`, lockID)
	var ok bool
	if err := row.Scan(&ok); err != nil {
		return false, err
	}
	if ok {
		_, _ = s.db.ExecContext(ctx, `SELECT set_config('taskforge.leader', $1, false)`, holder)
	}
	return ok, nil
}

func (s *pgStore) ReleaseLeader(ctx context.Context, lockID int64, holder string) error {
	_, _ = holder, ctx
	row := s.db.QueryRowContext(ctx, `SELECT pg_advisory_unlock($1)`, lockID)
	var ok bool
	return row.Scan(&ok)
}

func scanJobRun(row scanner) (*domain.JobRun, error) {
	var run domain.JobRun
	var wfRunID, parentID, workerID, idemKey sql.NullString
	if err := row.Scan(&run.ID, &run.NamespaceID, &run.JobID, &wfRunID, &parentID, &run.Status,
		&run.Attempt, &run.ScheduledAt, &run.StartedAt, &run.FinishedAt, &run.ExitCode,
		&run.Stdout, &run.Stderr, &run.CorrelationID, &workerID, &idemKey, &run.ErrorMessage); err != nil {
		return nil, err
	}
	run.WorkflowRunID = wfRunID.String
	run.ParentRunID = parentID.String
	run.WorkerID = workerID.String
	run.IdempotencyKey = idemKey.String
	return &run, nil
}

func scanWorker(row scanner) (*domain.Worker, error) {
	var w domain.Worker
	var tagsRaw string
	if err := row.Scan(&w.ID, &w.NamespaceID, &w.Name, &tagsRaw, &w.Capacity, &w.Status, &w.LastSeenAt, &w.CreatedAt); err != nil {
		return nil, err
	}
	w.Tags = decodeStringSlice(tagsRaw)
	return &w, nil
}

func nullStr(v string) sql.NullString {
	if v == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: v, Valid: true}
}
