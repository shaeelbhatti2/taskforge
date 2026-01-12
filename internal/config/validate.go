package config

import "time"

func (c *Config) Validate() error {
	if c.HTTPAddr == "" {
		c.HTTPAddr = ":8080"
	}
	if c.DatabaseURL == "" {
		c.DatabaseURL = "file:taskforge.db?cache=shared&mode=rwc"
	}
	if c.SchedulerTick <= 0 {
		c.SchedulerTick = DefaultSchedulerTick
	}
	if c.WorkerHeartbeat <= 0 {
		c.WorkerHeartbeat = DefaultWorkerHeartbeat
	}
	if c.JobLogLimit <= 0 {
		c.JobLogLimit = 65536
	}
	if c.LeaderLockID == 0 {
		c.LeaderLockID = 424242
	}
	ApplyEnvOverrides(c)
	return nil
}

const (
	DefaultSchedulerTick   = time.Second
	DefaultWorkerHeartbeat = 10 * time.Second
)
