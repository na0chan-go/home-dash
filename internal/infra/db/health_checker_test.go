package db

import (
	"context"
	"testing"
)

func TestSQLiteHealthChecker_Check(t *testing.T) {
	t.Parallel()

	sqliteDB, err := OpenSQLite(t.TempDir() + "/app.db")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	defer sqliteDB.Close()

	checker := NewSQLiteHealthChecker(sqliteDB)
	if err := checker.Check(context.Background()); err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
}

func TestSQLiteHealthChecker_Check_ReturnsErrorAfterClose(t *testing.T) {
	t.Parallel()

	sqliteDB, err := OpenSQLite(t.TempDir() + "/app.db")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	if err := sqliteDB.Close(); err != nil {
		t.Fatalf("failed to close sqlite: %v", err)
	}

	checker := NewSQLiteHealthChecker(sqliteDB)
	if err := checker.Check(context.Background()); err == nil {
		t.Fatal("expected error but got nil")
	}
}
