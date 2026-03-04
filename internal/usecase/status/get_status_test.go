package status

import (
	"context"
	"errors"
	"testing"
	"time"

	domaingarbage "github.com/na0chan-go/home-dash/internal/domain/garbage"
	"github.com/na0chan-go/home-dash/internal/ports"
)

type stubClock struct {
	now time.Time
}

func (c *stubClock) NowUTC() time.Time {
	return c.now
}

type stubDBChecker struct {
	err error
}

func (c *stubDBChecker) Check(context.Context) error {
	return c.err
}

type stubGarbageProvider struct {
	err error
}

func (p *stubGarbageProvider) GetSchedule(context.Context) (domaingarbage.Schedule, error) {
	if p.err != nil {
		return domaingarbage.Schedule{}, p.err
	}
	return domaingarbage.Schedule{
		Sunday:    []string{},
		Monday:    []string{"燃えるゴミ"},
		Tuesday:   []string{},
		Wednesday: []string{},
		Thursday:  []string{},
		Friday:    []string{},
		Saturday:  []string{},
	}, nil
}

type stubBackupReader struct {
	lastBackup *time.Time
	err        error
}

func (r *stubBackupReader) LastBackupAt(context.Context) (*time.Time, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.lastBackup, nil
}

func TestGetStatusUseCase_Execute_Success(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)
	startedAt := now.Add(-10 * time.Minute)
	lastBackup := now.Add(-30 * time.Minute)

	useCase := NewGetStatusUseCase(
		&stubClock{now: now},
		&stubDBChecker{},
		&stubGarbageProvider{},
		&stubBackupReader{lastBackup: &lastBackup},
		"/data/app.db",
		true,
		"abc1234",
		startedAt,
	)

	got, err := useCase.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if got.AppVersion != "abc1234" {
		t.Fatalf("unexpected appVersion: %s", got.AppVersion)
	}
	if got.UptimeSeconds != 600 {
		t.Fatalf("unexpected uptimeSeconds: %d", got.UptimeSeconds)
	}
	if got.DB.Path != "/data/app.db" || !got.DB.OK {
		t.Fatalf("unexpected db status: %+v", got.DB)
	}
	if !got.Config.GarbageScheduleLoaded {
		t.Fatal("garbageScheduleLoaded should be true")
	}
	if !got.Auth.Enabled {
		t.Fatal("auth.enabled should be true")
	}
	if got.LastBackup == nil || *got.LastBackup == "" {
		t.Fatal("lastBackup should be set")
	}
	if got.DBError != "" || got.ConfigError != "" || got.LastBackupError != "" {
		t.Fatalf("unexpected diagnostic values: %+v", got)
	}
	if got.ServerTime != "2026-03-05T09:00:00+09:00" {
		t.Fatalf("unexpected serverTime: %s", got.ServerTime)
	}
}

func TestGetStatusUseCase_Execute_Degraded(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)
	useCase := NewGetStatusUseCase(
		&stubClock{now: now},
		&stubDBChecker{err: errors.New("db unavailable")},
		&stubGarbageProvider{err: errors.New("broken schedule file")},
		&stubBackupReader{err: errors.New("backup dir not readable")},
		"/data/app.db",
		false,
		"",
		now.Add(10*time.Second),
	)

	got, err := useCase.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if got.AppVersion != defaultAppVersion {
		t.Fatalf("expected appVersion=%s, got %s", defaultAppVersion, got.AppVersion)
	}
	if got.UptimeSeconds != 0 {
		t.Fatalf("expected uptimeSeconds=0, got %d", got.UptimeSeconds)
	}
	if got.DB.OK {
		t.Fatal("db.ok should be false")
	}
	if got.Config.GarbageScheduleLoaded {
		t.Fatal("garbageScheduleLoaded should be false")
	}
	if got.Auth.Enabled {
		t.Fatal("auth.enabled should be false")
	}
	if got.LastBackup != nil {
		t.Fatalf("lastBackup should be nil, got %v", *got.LastBackup)
	}
	if got.DBError == "" || got.ConfigError == "" || got.LastBackupError == "" {
		t.Fatalf("expected diagnostic errors, got: %+v", got)
	}
}

var (
	_ ports.DBHealthChecker        = (*stubDBChecker)(nil)
	_ ports.GarbageScheduleProvider = (*stubGarbageProvider)(nil)
	_ ports.BackupStatusReader     = (*stubBackupReader)(nil)
	_ ports.Clock                  = (*stubClock)(nil)
)
