package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"modernc.org/sqlite"
)

const sqliteDriverName = "sqlite-home-dash"

var registerSQLiteDriverOnce sync.Once

func OpenSQLite(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	registerSQLiteDriverOnce.Do(func() {
		d := &sqlite.Driver{}
		d.RegisterConnectionHook(func(conn sqlite.ExecQuerierContext, _ string) error {
			_, err := conn.ExecContext(context.Background(), `PRAGMA foreign_keys = ON;`, []driver.NamedValue{})
			return err
		})
		sql.Register(sqliteDriverName, d)
	})

	db, err := sql.Open(sqliteDriverName, path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping sqlite: %w", err)
	}

	return db, nil
}
