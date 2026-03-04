package backup

import (
	"context"
	"fmt"

	"github.com/na0chan-go/home-dash/internal/ports"
)

const timestampLayout = "2006-01-02T15:04:05Z07:00"

type RunBackupUseCase struct {
	manager   ports.BackupManager
	backupDir string
	keep      int
}

func NewRunBackupUseCase(manager ports.BackupManager, backupDir string, keep int) *RunBackupUseCase {
	return &RunBackupUseCase{
		manager:   manager,
		backupDir: backupDir,
		keep:      keep,
	}
}

func (u *RunBackupUseCase) Execute(ctx context.Context) (BackupDTO, error) {
	result, err := u.manager.CreateBackup(ctx, u.backupDir, u.keep)
	if err != nil {
		return BackupDTO{}, fmt.Errorf("failed to create backup: %w", err)
	}

	return BackupDTO{
		FilePath:  result.FilePath,
		CreatedAt: result.CreatedAt.Format(timestampLayout),
		Removed:   len(result.Removed),
	}, nil
}
