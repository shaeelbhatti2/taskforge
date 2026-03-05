CREATE TABLE IF NOT EXISTS namespaces (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS retry_policies (
    id TEXT PRIMARY KEY,
    max_attempts INT NOT NULL DEFAULT 3,
    delay_seconds INT NOT NULL DEFAULT 30,
    backoff_base INT NOT NULL DEFAULT 2,
    max_delay_sec INT NOT NULL DEFAULT 3600,
    retry_on_codes TEXT DEFAULT '[]'
);

CREATE TABLE IF NOT EXISTS concurrency_groups (
    id TEXT PRIMARY KEY,
    namespace_id TEXT NOT NULL,
    name TEXT NOT NULL,
    max_running INT NOT NULL DEFAULT 1,
    UNIQUE(namespace_id, name)
);

CREATE TABLE IF NOT EXISTS job_definitions (
    id TEXT PRIMARY KEY,
    namespace_id TEXT NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    command TEXT,
    http_method TEXT,
    http_url TEXT,
    http_headers TEXT DEFAULT '{}',
    http_body TEXT,
    expected_status INT,
    timeout_sec INT NOT NULL DEFAULT 300,
    retry_policy_id TEXT,
    concurrency_group_id TEXT,
    concurrency_limit INT NOT NULL DEFAULT 0,
    rate_limit_per_min INT NOT NULL DEFAULT 0,
    idempotency_key TEXT,
    env TEXT DEFAULT '{}',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(namespace_id, name)
);

CREATE TABLE IF NOT EXISTS schedules (
    id TEXT PRIMARY KEY,
    namespace_id TEXT NOT NULL,
    job_id TEXT NOT NULL,
    cron_expr TEXT,
    interval_sec INT,
    timezone TEXT NOT NULL DEFAULT 'UTC',
    paused INTEGER NOT NULL DEFAULT 0,
    missed_policy TEXT NOT NULL DEFAULT 'run',
    next_run_at DATETIME,
    last_run_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS workflows (
    id TEXT PRIMARY KEY,
    namespace_id TEXT NOT NULL,
    name TEXT NOT NULL,
    nodes TEXT NOT NULL DEFAULT '[]',
    edges TEXT NOT NULL DEFAULT '[]',
    timeout_sec INT NOT NULL DEFAULT 3600,
    fail_fast INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(namespace_id, name)
);

CREATE TABLE IF NOT EXISTS workflow_runs (
    id TEXT PRIMARY KEY,
    workflow_id TEXT NOT NULL,
    namespace_id TEXT NOT NULL,
    status TEXT NOT NULL,
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    finished_at DATETIME
);

CREATE TABLE IF NOT EXISTS job_runs (
    id TEXT PRIMARY KEY,
    namespace_id TEXT NOT NULL,
    job_id TEXT NOT NULL,
    workflow_run_id TEXT,
    parent_run_id TEXT,
    status TEXT NOT NULL,
    attempt INT NOT NULL DEFAULT 1,
    scheduled_at DATETIME NOT NULL,
    started_at DATETIME,
    finished_at DATETIME,
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
    run_id TEXT NOT NULL,
    from_status TEXT NOT NULL,
    to_status TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS workers (
    id TEXT PRIMARY KEY,
    namespace_id TEXT NOT NULL,
    name TEXT NOT NULL,
    tags TEXT NOT NULL DEFAULT '[]',
    capacity INT NOT NULL DEFAULT 1,
    status TEXT NOT NULL DEFAULT 'online',
    last_seen_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(namespace_id, name)
);

CREATE TABLE IF NOT EXISTS dead_letter_entries (
    id TEXT PRIMARY KEY,
    namespace_id TEXT NOT NULL,
    job_run_id TEXT NOT NULL,
    job_id TEXT NOT NULL,
    reason TEXT NOT NULL,
    payload TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS webhooks (
    id TEXT PRIMARY KEY,
    namespace_id TEXT NOT NULL,
    url TEXT NOT NULL,
    secret TEXT NOT NULL,
    events TEXT NOT NULL DEFAULT '[]',
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS api_keys (
    id TEXT PRIMARY KEY,
    namespace_id TEXT NOT NULL,
    name TEXT NOT NULL,
    key_hash TEXT NOT NULL,
    prefix TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(namespace_id, name)
);

CREATE INDEX IF NOT EXISTS idx_job_runs_namespace_status ON job_runs(namespace_id, status);
CREATE INDEX IF NOT EXISTS idx_schedules_next_run ON schedules(next_run_at);
CREATE INDEX IF NOT EXISTS idx_workers_namespace ON workers(namespace_id);
