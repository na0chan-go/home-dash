package db

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileBackupStatusReader_LastBackupAt(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	oldPath := filepath.Join(dir, "app-20260301-010000.db")
	newPath := filepath.Join(dir, "app-20260304-090000.db")
	ignorePath := filepath.Join(dir, "memo.txt")

	if err := os.WriteFile(oldPath, []byte("old"), 0o644); err != nil {
		t.Fatalf("failed to write old backup file: %v", err)
	}
	if err := os.WriteFile(newPath, []byte("new"), 0o644); err != nil {
		t.Fatalf("failed to write new backup file: %v", err)
	}
	if err := os.WriteFile(ignorePath, []byte("ignore"), 0o644); err != nil {
		t.Fatalf("failed to write ignore file: %v", err)
	}

	oldTime := time.Date(2026, 3, 1, 1, 0, 0, 0, time.UTC)
	newTime := time.Date(2026, 3, 4, 9, 0, 0, 0, time.UTC)
	if err := os.Chtimes(oldPath, oldTime, oldTime); err != nil {
		t.Fatalf("failed to chtimes old backup file: %v", err)
	}
	if err := os.Chtimes(newPath, newTime, newTime); err != nil {
		t.Fatalf("failed to chtimes new backup file: %v", err)
	}

	reader := NewFileBackupStatusReader(dir)
	got, err := reader.LastBackupAt(context.Background())
	if err != nil {
		t.Fatalf("LastBackupAt returned error: %v", err)
	}
	if got == nil {
		t.Fatal("expected latest backup time, got nil")
	}
	if !got.Equal(newTime) {
		t.Fatalf("unexpected latest backup time: got=%s want=%s", got.Format(time.RFC3339), newTime.Format(time.RFC3339))
	}
}

func TestFileBackupStatusReader_LastBackupAt_NoFile(t *testing.T) {
	t.Parallel()

	reader := NewFileBackupStatusReader(t.TempDir())
	got, err := reader.LastBackupAt(context.Background())
	if err != nil {
		t.Fatalf("LastBackupAt returned error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestFileBackupStatusReader_LastBackupAt_NoDirectory(t *testing.T) {
	t.Parallel()

	reader := NewFileBackupStatusReader(filepath.Join(t.TempDir(), "missing"))
	got, err := reader.LastBackupAt(context.Background())
	if err != nil {
		t.Fatalf("LastBackupAt returned error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}
