package config

import (
	"os"
	"strconv"
	"time"
)

func (c *Config) APIKey() string {
	if v := os.Getenv("TASKFORGE_API_KEY"); v != "" {
		return v
	}
	return "dev-api-key"
}

func (c *Config) DefaultNamespace() string {
	if v := os.Getenv("TASKFORGE_NAMESPACE"); v != "" {
		return v
	}
	return "default"
}

func (c *Config) ParseDuration(key, fallback string) time.Duration {
	raw := fallback
	if v := os.Getenv(key); v != "" {
		raw = v
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		d, _ = time.ParseDuration(fallback)
	}
	return d
}

func (c *Config) IntEnv(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func ApplyEnvOverrides(cfg *Config) {
	if v := os.Getenv("TASKFORGE_SCHEDULER_TICK"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.SchedulerTick = d
		}
	}
	if v := os.Getenv("TASKFORGE_WORKER_HEARTBEAT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.WorkerHeartbeat = d
		}
	}
	if v := os.Getenv("TASKFORGE_JOB_LOG_LIMIT"); v != "" {
		cfg.JobLogLimit = cfg.IntEnv("TASKFORGE_JOB_LOG_LIMIT", cfg.JobLogLimit)
	}
}
