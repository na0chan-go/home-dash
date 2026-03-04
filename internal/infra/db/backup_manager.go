package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/na0chan-go/home-dash/internal/ports"
)

const backupNameLayout = "20060102-150405"

type SQLiteBackupManager struct {
	db *sql.DB
	mu sync.Mutex
}

func NewSQLiteBackupManager(db *sql.DB) *SQLiteBackupManager {
	return &SQLiteBackupManager{db: db}
}

func (m *SQLiteBackupManager) CreateBackup(ctx context.Context, backupDir string, keep int) (ports.BackupResult, error) {
	if keep < 1 {
		return ports.BackupResult{}, fmt.Errorf("backup keep must be >= 1")
	}
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return ports.BackupResult{}, fmt.Errorf("failed to create backup dir: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	filePath, createdAt, err := nextBackupFilePath(backupDir, time.Now())
	if err != nil {
		return ports.BackupResult{}, err
	}

	if err := m.vacuumInto(ctx, filePath); err != nil {
		return ports.BackupResult{}, err
	}

	removed, err := rotateBackups(backupDir, keep)
	if err != nil {
		return ports.BackupResult{}, err
	}

	return ports.BackupResult{
		FilePath:  filePath,
		CreatedAt: createdAt,
		Removed:   removed,
	}, nil
}

func (m *SQLiteBackupManager) vacuumInto(ctx context.Context, filePath string) error {
	escapedPath := strings.ReplaceAll(filePath, "'", "''")
	stmt := "VACUUM INTO '" + escapedPath + "';"
	if _, err := m.db.ExecContext(ctx, stmt); err != nil {
		return fmt.Errorf("failed to backup sqlite db: %w", err)
	}
	return nil
}

func nextBackupFilePath(backupDir string, now time.Time) (string, time.Time, error) {
	baseName := "app-" + now.Format(backupNameLayout)
	path := filepath.Join(backupDir, baseName+".db")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path, now, nil
	}

	for i := 1; i <= 999; i++ {
		candidate := filepath.Join(backupDir, fmt.Sprintf("%s-%03d.db", baseName, i))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate, now, nil
		}
	}
	return "", time.Time{}, fmt.Errorf("failed to generate unique backup filename")
}

func rotateBackups(backupDir string, keep int) ([]string, error) {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup dir: %w", err)
	}

	files := make([]os.FileInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "app-") || !strings.HasSuffix(name, ".db") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("failed to stat backup file: %w", err)
		}
		files = append(files, info)
	}

	if len(files) <= keep {
		return []string{}, nil
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].ModTime().Equal(files[j].ModTime()) {
			return files[i].Name() < files[j].Name()
		}
		return files[i].ModTime().Before(files[j].ModTime())
	})

	toDelete := files[:len(files)-keep]
	removed := make([]string, 0, len(toDelete))
	for _, file := range toDelete {
		target := filepath.Join(backupDir, file.Name())
		if err := os.Remove(target); err != nil {
			return nil, fmt.Errorf("failed to remove old backup %s: %w", target, err)
		}
		removed = append(removed, target)
	}

	return removed, nil
}
