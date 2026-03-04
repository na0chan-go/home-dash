package app

import (
	"context"
	"errors"
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

func TestBackupScheduler_ShutdownCancelsRunningTask(t *testing.T) {
	started := make(chan struct{})
	scheduler := newBackupScheduler(10*time.Millisecond, func(ctx context.Context) (usebackup.BackupDTO, error) {
		close(started)
		<-ctx.Done()
		return usebackup.BackupDTO{}, ctx.Err()
	})

	scheduler.Start()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("scheduler task did not start")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := scheduler.Shutdown(ctx); err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected shutdown error: %v", err)
	}
}
