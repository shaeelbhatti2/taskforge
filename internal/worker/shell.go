package worker

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/shaeelbhatti2/taskforge/internal/domain"
)

type ShellResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

type ShellExecutor struct {
	logLimit int
}

func NewShellExecutor(logLimit int) *ShellExecutor {
	if logLimit <= 0 {
		logLimit = 65536
	}
	return &ShellExecutor{logLimit: logLimit}
}

func (e *ShellExecutor) Run(ctx context.Context, job *domain.JobDefinition) (ShellResult, error) {
	if job.Type != domain.JobTypeShell {
		return ShellResult{}, fmt.Errorf("not a shell job")
	}
	timeout := time.Duration(job.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(runCtx, "sh", "-c", job.Command)
	if len(job.Env) > 0 {
		cmd.Env = append(cmd.Environ(), envPairs(job.Env)...)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			return ShellResult{}, err
		}
	}
	return ShellResult{
		ExitCode: exitCode,
		Stdout:   truncate(stdout.String(), e.logLimit),
		Stderr:   truncate(stderr.String(), e.logLimit),
	}, nil
}

func envPairs(env map[string]string) []string {
	out := make([]string, 0, len(env))
	for k, v := range env {
		out = append(out, k+"="+v)
	}
	return out
}

func truncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit] + "\n...truncated"
}

func ParseCommand(raw string) (string, []string) {
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}
