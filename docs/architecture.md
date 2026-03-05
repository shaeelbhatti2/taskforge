# TaskForge Architecture

## Components

```
Operator -> API/CLI -> Store (PostgreSQL)
                |
           Scheduler (leader elected)
                |
           Worker Pool (heartbeats, tags)
                |
           Job Executors (shell, HTTP)
```

## Scheduler flow

1. Leader node acquires advisory lock
2. Tick loop loads due schedules from store
3. Creates pending JobRun rows with dedup constraint
4. Workers pull/execute and report results
5. Retry policy or DLQ on failure

## Workflow execution

Workflows are validated DAGs. The executor topologically sorts nodes,
runs fork branches concurrently, and waits at join nodes.

## Multi-tenancy

Every query is scoped by namespace_id. API keys authenticate requests
and map to a namespace context via headers.
