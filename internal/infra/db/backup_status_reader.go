package db

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

type FileBackupStatusReader struct {
	backupDir string
}

func NewFileBackupStatusReader(backupDir string) *FileBackupStatusReader {
	return &FileBackupStatusReader{backupDir: backupDir}
}

func (r *FileBackupStatusReader) LastBackupAt(_ context.Context) (*time.Time, error) {
	entries, err := os.ReadDir(r.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read backup dir: %w", err)
	}

	var latest time.Time
	found := false
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

		modTime := info.ModTime()
		if !found || modTime.After(latest) {
			latest = modTime
			found = true
		}
	}

	if !found {
		return nil, nil
	}

	normalized := latest.UTC()
	return &normalized, nil
}
