package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

func runMigrations(ctx context.Context, db *sql.DB, sqlite bool) error {
	file := "001_initial.sql"
	if sqlite {
		file = "001_sqlite.sql"
	}
	candidates := []string{
		filepath.Join("migrations", file),
		filepath.Join("..", "migrations", file),
		filepath.Join("..", "..", "migrations", file),
	}
	var data []byte
	var err error
	for _, c := range candidates {
		data, err = os.ReadFile(c)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("read migration: %w", err)
	}
	_, err = db.ExecContext(ctx, string(data))
	if err != nil {
		return fmt.Errorf("apply migration: %w", err)
	}
	return nil
}
