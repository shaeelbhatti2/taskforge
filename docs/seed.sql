INSERT INTO namespaces (id, name, created_at, updated_at)
VALUES ('default-ns', 'default', NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO job_definitions (id, namespace_id, name, type, command, timeout_sec, created_at, updated_at)
VALUES
  ('fetch-artifacts', 'default-ns', 'fetch-artifacts', 'shell', 'echo fetching', 120, NOW(), NOW()),
  ('cleanup-cache-a', 'default-ns', 'cleanup-cache-a', 'shell', 'echo clean a', 60, NOW(), NOW()),
  ('cleanup-cache-b', 'default-ns', 'cleanup-cache-b', 'shell', 'echo clean b', 60, NOW(), NOW()),
  ('notify-complete', 'default-ns', 'notify-complete', 'http', NULL, 30, NOW(), NOW())
ON CONFLICT DO NOTHING;
