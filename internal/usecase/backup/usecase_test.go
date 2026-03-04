package backup

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/na0chan-go/home-dash/internal/ports"
)

type stubBackupManager struct {
	result ports.BackupResult
	err    error
}

func (m *stubBackupManager) CreateBackup(context.Context, string, int) (ports.BackupResult, error) {
	if m.err != nil {
		return ports.BackupResult{}, m.err
	}
	return m.result, nil
}

func TestRunBackupUseCase_Execute(t *testing.T) {
	manager := &stubBackupManager{
		result: ports.BackupResult{
			FilePath:  "/data/backups/app-20260304-130000.db",
			CreatedAt: time.Date(2026, 3, 4, 13, 0, 0, 0, time.FixedZone("JST", 9*60*60)),
			Removed:   []string{"old1.db", "old2.db"},
		},
	}
	useCase := NewRunBackupUseCase(manager, "/data/backups", 30)

	got, err := useCase.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if got.FilePath != "/data/backups/app-20260304-130000.db" {
		t.Fatalf("unexpected file path: %s", got.FilePath)
	}
	if got.CreatedAt != "2026-03-04T13:00:00+09:00" {
		t.Fatalf("unexpected createdAt: %s", got.CreatedAt)
	}
	if got.Removed != 2 {
		t.Fatalf("unexpected removed count: %d", got.Removed)
	}
}

func TestRunBackupUseCase_Execute_ReturnsError(t *testing.T) {
	useCase := NewRunBackupUseCase(&stubBackupManager{err: errors.New("disk full")}, "/data/backups", 30)

	if _, err := useCase.Execute(context.Background()); err == nil {
		t.Fatal("expected error but got nil")
	}
}
