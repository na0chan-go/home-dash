package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domaingarbage "github.com/na0chan-go/home-dash/internal/domain/garbage"
	usestatus "github.com/na0chan-go/home-dash/internal/usecase/status"
)

type statusHandlerStubClock struct {
	now time.Time
}

func (c *statusHandlerStubClock) NowUTC() time.Time {
	return c.now
}

type statusHandlerStubDBChecker struct {
	err error
}

func (c *statusHandlerStubDBChecker) Check(context.Context) error {
	return c.err
}

type statusHandlerStubGarbageProvider struct {
	err error
}

func (p *statusHandlerStubGarbageProvider) GetSchedule(context.Context) (domaingarbage.Schedule, error) {
	if p.err != nil {
		return domaingarbage.Schedule{}, p.err
	}
	return domaingarbage.Schedule{
		Sunday:    []string{},
		Monday:    []string{},
		Tuesday:   []string{},
		Wednesday: []string{},
		Thursday:  []string{},
		Friday:    []string{},
		Saturday:  []string{},
	}, nil
}

func TestStatusHandler_Get_Success(t *testing.T) {
	now := time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)
	backupAt := now.Add(-time.Hour)
	useCase := usestatus.NewGetStatusUseCase(
		&statusHandlerStubClock{now: now},
		&statusHandlerStubDBChecker{},
		&statusHandlerStubGarbageProvider{},
		&statusHandlerStubBackupReader{lastBackup: &backupAt},
		"/data/app.db",
		true,
		"abc1234",
		now.Add(-time.Minute),
	)
	handler := NewStatusHandler(useCase)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var payload struct {
		AppVersion    string `json:"appVersion"`
		UptimeSeconds int64  `json:"uptimeSeconds"`
		ServerTime    string `json:"serverTime"`
		DB            struct {
			Path string `json:"path"`
			OK   bool   `json:"ok"`
		} `json:"db"`
		Config struct {
			GarbageScheduleLoaded bool `json:"garbageScheduleLoaded"`
		} `json:"config"`
		Auth struct {
			Enabled bool `json:"enabled"`
		} `json:"auth"`
		LastBackup *string `json:"lastBackup"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if payload.AppVersion != "abc1234" {
		t.Fatalf("unexpected appVersion: %s", payload.AppVersion)
	}
	if payload.DB.Path != "/data/app.db" || !payload.DB.OK {
		t.Fatalf("unexpected db payload: %+v", payload.DB)
	}
	if !payload.Config.GarbageScheduleLoaded {
		t.Fatal("garbageScheduleLoaded should be true")
	}
	if !payload.Auth.Enabled {
		t.Fatal("auth.enabled should be true")
	}
	if payload.LastBackup == nil || *payload.LastBackup == "" {
		t.Fatal("lastBackup should be set")
	}
}

func TestStatusHandler_Get_MethodNotAllowed(t *testing.T) {
	now := time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)
	useCase := usestatus.NewGetStatusUseCase(
		&statusHandlerStubClock{now: now},
		&statusHandlerStubDBChecker{},
		&statusHandlerStubGarbageProvider{},
		&statusHandlerStubBackupReader{},
		"/data/app.db",
		false,
		"unknown",
		now,
	)
	handler := NewStatusHandler(useCase)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/status", nil)
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestStatusHandler_Get_DegradedStillReturns200(t *testing.T) {
	now := time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)
	useCase := usestatus.NewGetStatusUseCase(
		&statusHandlerStubClock{now: now},
		&statusHandlerStubDBChecker{err: errors.New("db down")},
		&statusHandlerStubGarbageProvider{err: errors.New("config broken")},
		&statusHandlerStubBackupReader{err: errors.New("backup dir unreadable")},
		"/data/app.db",
		true,
		"abc1234",
		now.Add(-time.Minute),
	)
	handler := NewStatusHandler(useCase)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var payload struct {
		DB struct {
			OK bool `json:"ok"`
		} `json:"db"`
		Config struct {
			GarbageScheduleLoaded bool `json:"garbageScheduleLoaded"`
		} `json:"config"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload.DB.OK {
		t.Fatal("db.ok should be false")
	}
	if payload.Config.GarbageScheduleLoaded {
		t.Fatal("garbageScheduleLoaded should be false")
	}
}

type statusHandlerStubBackupReader struct {
	lastBackup *time.Time
	err        error
}

func (r *statusHandlerStubBackupReader) LastBackupAt(context.Context) (*time.Time, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.lastBackup, nil
}
