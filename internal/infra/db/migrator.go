package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func RunMigrations(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT (datetime('now'))
		);
	`); err != nil {
		return fmt.Errorf("failed to ensure schema_migrations: %w", err)
	}

	entries, err := fs.ReadDir(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		applied, err := isApplied(ctx, db, name)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		sqlBytes, err := migrationFS.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", name, err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin tx for %s: %w", name, err)
		}

		if _, err := tx.ExecContext(ctx, string(sqlBytes)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", name, err)
		}

		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(version) VALUES (?)`, name); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to register migration %s: %w", name, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", name, err)
		}
	}

	return nil
}

func isApplied(ctx context.Context, db *sql.DB, version string) (bool, error) {
	var exists int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(1) FROM schema_migrations WHERE version = ?`, version).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check migration %s: %w", version, err)
	}
	return exists > 0, nil
}
