package ports

import (
	"context"
	"time"
)

type BackupStatusReader interface {
	LastBackupAt(ctx context.Context) (*time.Time, error)
}
