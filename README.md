# TaskForge

Distributed job scheduler and workflow orchestration engine built in Go.

TaskForge gives platform teams a reliable way to run cron jobs, one-off tasks, and multi-step DAG workflows across a pool of workers. Operators define schedules and dependencies declaratively; the engine handles retries, timeouts, concurrency limits, and dead-letter queues.

## Features

- Cron and interval scheduling with timezone support
- DAG workflows with dependency edges and parallel execution
- Worker pool with heartbeats and capacity tags
- Retry policies and dead-letter queue
- REST API, CLI, and web dashboard
- Prometheus metrics and structured logging

## Quick start

```bash
docker compose up -d postgres
go run ./cmd/taskforge
```

Copy `.env.example` to `.env` and adjust settings as needed.

## Project layout

```
cmd/taskforge/          server + scheduler
cmd/taskforge-worker/   worker process
internal/               core packages
migrations/             SQL migrations
web/                    dashboard assets
docker/                 container build
config/                 example configuration
```

## License

MIT
