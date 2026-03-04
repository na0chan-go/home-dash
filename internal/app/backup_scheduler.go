package app

import (
	"context"
	"log"
	"sync"
	"time"

	usebackup "github.com/na0chan-go/home-dash/internal/usecase/backup"
)

type backupScheduler struct {
	interval time.Duration
	run      func(context.Context) (usebackup.BackupDTO, error)

	stopCh  chan struct{}
	doneCh  chan struct{}
	started sync.Once
	stopped sync.Once
}

func newBackupScheduler(interval time.Duration, run func(context.Context) (usebackup.BackupDTO, error)) *backupScheduler {
	return &backupScheduler{
		interval: interval,
		run:      run,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

func (s *backupScheduler) Start() {
	s.started.Do(func() {
		go s.loop()
	})
}

func (s *backupScheduler) Shutdown(ctx context.Context) error {
	s.stopped.Do(func() {
		close(s.stopCh)
	})

	select {
	case <-s.doneCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *backupScheduler) loop() {
	defer close(s.doneCh)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			runCtx, cancel := context.WithCancel(context.Background())
			go func() {
				select {
				case <-s.stopCh:
					cancel()
				case <-runCtx.Done():
				}
			}()

			runCtxWithTimeout, timeoutCancel := context.WithTimeout(runCtx, 2*time.Minute)
			result, err := s.run(runCtxWithTimeout)
			timeoutCancel()
			cancel()
			if err != nil {
				log.Printf("level=error component=backup_scheduler err=%v", err)
				continue
			}
			log.Printf("level=info component=backup_scheduler file=%s removed=%d", result.FilePath, result.Removed)
		case <-s.stopCh:
			return
		}
	}
}
