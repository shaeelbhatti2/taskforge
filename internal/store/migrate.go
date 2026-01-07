package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shaeelbhatti2/taskforge/internal/domain"
)

func runMigrations(ctx context.Context, db *sql.DB) error {
	data, err := os.ReadFile(filepath.Join("migrations", "001_initial.sql"))
	if err != nil {
		candidates := []string{
			"migrations/001_initial.sql",
			filepath.Join("..", "migrations", "001_initial.sql"),
		}
		for _, c := range candidates {
			data, err = os.ReadFile(c)
			if err == nil {
				break
			}
		}
		if err != nil {
			return fmt.Errorf("read migration: %w", err)
		}
	}
	sqlText := string(data)
	sqlText = strings.ReplaceAll(sqlText, "JSONB", "TEXT")
	sqlText = strings.ReplaceAll(sqlText, "TIMESTAMPTZ", "DATETIME")
	_, err = db.ExecContext(ctx, sqlText)
	if err != nil {
		return fmt.Errorf("apply migration: %w", err)
	}
	return nil
}

func scanJob(row scanner) (*domain.JobDefinition, error) {
	var job domain.JobDefinition
	var headers, env sql.NullString
	err := row.Scan(
		&job.ID, &job.NamespaceID, &job.Name, &job.Type, &job.Command,
		&job.HTTPMethod, &job.HTTPURL, &headers, &job.HTTPBody, &job.ExpectedStatus,
		&job.TimeoutSec, &job.RetryPolicyID, &job.ConcurrencyGroupID,
		&job.ConcurrencyLimit, &job.RateLimitPerMin, &job.IdempotencyKey, &env,
		&job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	decodeJSON(headers.String, &job.HTTPHeaders)
	decodeJSON(env.String, &job.Env)
	return &job, nil
}

type scanner interface {
	Scan(dest ...any) error
}
