package db

import (
	"context"
	"database/sql"
	"fmt"
)

type SQLiteHealthChecker struct {
	db *sql.DB
}

func NewSQLiteHealthChecker(db *sql.DB) *SQLiteHealthChecker {
	return &SQLiteHealthChecker{db: db}
}

func (c *SQLiteHealthChecker) Check(ctx context.Context) error {
	if err := c.db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping sqlite: %w", err)
	}
	return nil
}
