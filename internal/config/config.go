package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTPAddr        string        `yaml:"http_addr"`
	DatabaseURL     string        `yaml:"database_url"`
	LogLevel        string        `yaml:"log_level"`
	SchedulerTick   time.Duration `yaml:"scheduler_tick"`
	WorkerHeartbeat time.Duration `yaml:"worker_heartbeat"`
	JobLogLimit     int           `yaml:"job_log_limit"`
	LeaderLockID    int64         `yaml:"leader_lock_id"`
}

func Load() (*Config, error) {
	cfg := &Config{
		HTTPAddr:        ":8080",
		DatabaseURL:     "file:taskforge.db?cache=shared&mode=rwc",
		LogLevel:        "info",
		SchedulerTick:   time.Second,
		WorkerHeartbeat: 10 * time.Second,
		JobLogLimit:     65536,
		LeaderLockID:    424242,
	}
	if v := os.Getenv("TASKFORGE_HTTP_ADDR"); v != "" {
		cfg.HTTPAddr = v
	}
	if v := os.Getenv("TASKFORGE_DATABASE_URL"); v != "" {
		cfg.DatabaseURL = v
	}
	if v := os.Getenv("TASKFORGE_LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	if path := os.Getenv("TASKFORGE_CONFIG"); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}
