package store

import (
	"context"
	"database/sql"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type sqliteStore struct {
	pgStore
	mu       sync.Mutex
	leaderID int64
	leader   string
}

func openSQLite(url string) (Store, error) {
	db, err := sql.Open("sqlite3", url)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &sqliteStore{pgStore: pgStore{db: db}}
	return s, nil
}

func (s *sqliteStore) TryAcquireLeader(ctx context.Context, lockID int64, holder string) (bool, error) {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.leader == "" || s.leader == holder {
		s.leaderID = lockID
		s.leader = holder
		return true, nil
	}
	return false, nil
}

func (s *sqliteStore) ReleaseLeader(ctx context.Context, lockID int64, holder string) error {
	_ = ctx
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.leader == holder && s.leaderID == lockID {
		s.leader = ""
		s.leaderID = 0
	}
	return nil
}

func (s *sqliteStore) Migrate(ctx context.Context) error {
	return runMigrations(ctx, s.db)
}

func (s *sqliteStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *sqliteStore) Close() error {
	return s.db.Close()
}
