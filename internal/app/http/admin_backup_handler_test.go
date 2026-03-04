package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/na0chan-go/home-dash/internal/ports"
	usebackup "github.com/na0chan-go/home-dash/internal/usecase/backup"
)

type stubBackupManager struct {
	err error
}

func (m *stubBackupManager) CreateBackup(context.Context, string, int) (ports.BackupResult, error) {
	if m.err != nil {
		return ports.BackupResult{}, m.err
	}
	return ports.BackupResult{
		FilePath:  "/data/backups/app-20260304-130000.db",
		CreatedAt: time.Date(2026, 3, 4, 13, 0, 0, 0, time.FixedZone("JST", 9*60*60)),
		Removed:   []string{},
	}, nil
}

func TestAdminBackupHandler_Create_Success(t *testing.T) {
	useCase := usebackup.NewRunBackupUseCase(&stubBackupManager{}, "/data/backups", 30)
	handler := NewAdminBackupHandler(useCase)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/backup", nil)
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"filePath":"/data/backups/app-20260304-130000.db"`) {
		t.Fatalf("unexpected response: %s", rec.Body.String())
	}
}

func TestAdminBackupHandler_Create_MethodNotAllowed(t *testing.T) {
	useCase := usebackup.NewRunBackupUseCase(&stubBackupManager{}, "/data/backups", 30)
	handler := NewAdminBackupHandler(useCase)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/backup", nil)
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestAdminBackupHandler_Create_Returns500OnFailure(t *testing.T) {
	useCase := usebackup.NewRunBackupUseCase(&stubBackupManager{err: errors.New("disk full")}, "/data/backups", 30)
	handler := NewAdminBackupHandler(useCase)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/backup", nil)
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"INTERNAL_ERROR"`) {
		t.Fatalf("unexpected response: %s", rec.Body.String())
	}
}
