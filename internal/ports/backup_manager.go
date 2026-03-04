package ports

import (
	"context"
	"time"
)

type BackupResult struct {
	FilePath  string
	CreatedAt time.Time
	Removed   []string
}

type BackupManager interface {
	CreateBackup(ctx context.Context, backupDir string, keep int) (BackupResult, error)
}
