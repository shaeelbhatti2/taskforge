CREATE TABLE IF NOT EXISTS namespaces (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS retry_policies (
    id TEXT PRIMARY KEY,
    max_attempts INT NOT NULL DEFAULT 3,
    delay_seconds INT NOT NULL DEFAULT 30,
    backoff_base INT NOT NULL DEFAULT 2,
    max_delay_sec INT NOT NULL DEFAULT 3600,
    retry_on_codes JSONB DEFAULT '[]'
);

CREATE TABLE IF NOT EXISTS concurrency_groups (
    id TEXT PRIMARY KEY,
    namespace_id TEXT NOT NULL REFERENCES namespaces(id),
    name TEXT NOT NULL,
    max_running INT NOT NULL DEFAULT 1,
    UNIQUE(namespace_id, name)
);

CREATE TABLE IF NOT EXISTS job_definitions (
    id TEXT PRIMARY KEY,
    namespace_id TEXT NOT NULL REFERENCES namespaces(id),
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    command TEXT,
    http_method TEXT,
    http_url TEXT,
    http_headers JSONB DEFAULT '{}',
    http_body TEXT,
    expected_status INT,
    timeout_sec INT NOT NULL DEFAULT 300,
    retry_policy_id TEXT REFERENCES retry_policies(id),
    concurrency_group_id TEXT REFERENCES concurrency_groups(id),
    concurrency_limit INT NOT NULL DEFAULT 0,
    rate_limit_per_min INT NOT NULL DEFAULT 0,
    idempotency_key TEXT,
    env JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(namespace_id, name)
);

CREATE TABLE IF NOT EXISTS schedules (
    id TEXT PRIMARY KEY,
    namespace_id TEXT NOT NULL REFERENCES namespaces(id),
    job_id TEXT NOT NULL REFERENCES job_definitions(id),
    cron_expr TEXT,
    interval_sec INT,
    timezone TEXT NOT NULL DEFAULT 'UTC',
    paused BOOLEAN NOT NULL DEFAULT FALSE,
    missed_policy TEXT NOT NULL DEFAULT 'run',
    next_run_at TIMESTAMPTZ,
    last_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS workflows (
    id TEXT PRIMARY KEY,
    namespace_id TEXT NOT NULL REFERENCES namespaces(id),
    name TEXT NOT NULL,
    nodes JSONB NOT NULL DEFAULT '[]',
    edges JSONB NOT NULL DEFAULT '[]',
    timeout_sec INT NOT NULL DEFAULT 3600,
    fail_fast BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(namespace_id, name)
);

CREATE TABLE IF NOT EXISTS workflow_runs (
    id TEXT PRIMARY KEY,
    workflow_id TEXT NOT NULL REFERENCES workflows(id),
    namespace_id TEXT NOT NULL REFERENCES namespaces(id),
    status TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS job_runs (
    id TEXT PRIMARY KEY,
    namespace_id TEXT NOT NULL REFERENCES namespaces(id),
    job_id TEXT NOT NULL REFERENCES job_definitions(id),
    workflow_run_id TEXT REFERENCES workflow_runs(id),
    parent_run_id TEXT,
    status TEXT NOT NULL,
    attempt INT NOT NULL DEFAULT 1,
    scheduled_at TIMESTAMPTZ NOT NULL,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    exit_code INT,
    stdout TEXT,
    stderr TEXT,
    correlation_id TEXT NOT NULL,
    worker_id TEXT,
    idempotency_key TEXT,
    error_message TEXT,
    UNIQUE(job_id, scheduled_at)
);

CREATE TABLE IF NOT EXISTS run_transitions (
    id TEXT PRIMARY KEY,
    run_id TEXT NOT NULL REFERENCES job_runs(id),
    from_status TEXT NOT NULL,
    to_status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS workers (
    id TEXT PRIMARY KEY,
    namespace_id TEXT NOT NULL REFERENCES namespaces(id),
    name TEXT NOT NULL,
    tags JSONB NOT NULL DEFAULT '[]',
    capacity INT NOT NULL DEFAULT 1,
    status TEXT NOT NULL DEFAULT 'online',
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(namespace_id, name)
);

CREATE TABLE IF NOT EXISTS dead_letter_entries (
    id TEXT PRIMARY KEY,
    namespace_id TEXT NOT NULL REFERENCES namespaces(id),
    job_run_id TEXT NOT NULL REFERENCES job_runs(id),
    job_id TEXT NOT NULL REFERENCES job_definitions(id),
    reason TEXT NOT NULL,
    payload TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS webhooks (
    id TEXT PRIMARY KEY,
    namespace_id TEXT NOT NULL REFERENCES namespaces(id),
    url TEXT NOT NULL,
    secret TEXT NOT NULL,
    events JSONB NOT NULL DEFAULT '[]',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS api_keys (
    id TEXT PRIMARY KEY,
    namespace_id TEXT NOT NULL REFERENCES namespaces(id),
    name TEXT NOT NULL,
    key_hash TEXT NOT NULL,
    prefix TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(namespace_id, name)
);

CREATE INDEX IF NOT EXISTS idx_job_runs_namespace_status ON job_runs(namespace_id, status);
CREATE INDEX IF NOT EXISTS idx_schedules_next_run ON schedules(next_run_at) WHERE paused = FALSE;
CREATE INDEX IF NOT EXISTS idx_workers_namespace ON workers(namespace_id);
