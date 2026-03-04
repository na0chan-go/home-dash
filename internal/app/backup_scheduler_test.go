package app

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	usebackup "github.com/na0chan-go/home-dash/internal/usecase/backup"
)

func TestBackupScheduler_RunsPeriodically(t *testing.T) {
	var calls int32
	scheduler := newBackupScheduler(20*time.Millisecond, func(context.Context) (usebackup.BackupDTO, error) {
		atomic.AddInt32(&calls, 1)
		return usebackup.BackupDTO{}, nil
	})

	scheduler.Start()
	time.Sleep(70 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := scheduler.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}

	if atomic.LoadInt32(&calls) < 2 {
		t.Fatalf("expected scheduler to run at least twice, got %d", atomic.LoadInt32(&calls))
	}
}

