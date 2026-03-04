package db

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSQLiteBackupManager_CreateBackupAndRotate(t *testing.T) {
	root := t.TempDir()
	dbPath := filepath.Join(root, "app.db")
	backupDir := filepath.Join(root, "backups")

	sqliteDB, err := OpenSQLite(dbPath)
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer sqliteDB.Close()

	if _, err := sqliteDB.Exec(`CREATE TABLE IF NOT EXISTS test_items (id INTEGER PRIMARY KEY, body TEXT);`); err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	if _, err := sqliteDB.Exec(`INSERT INTO test_items(body) VALUES ('a'), ('b');`); err != nil {
		t.Fatalf("failed to insert rows: %v", err)
	}

	manager := NewSQLiteBackupManager(sqliteDB)
	for i := 0; i < 3; i++ {
		result, err := manager.CreateBackup(context.Background(), backupDir, 2)
		if err != nil {
			t.Fatalf("CreateBackup failed: %v", err)
		}
		if result.FilePath == "" {
			t.Fatal("backup file path is empty")
		}
		if _, err := os.Stat(result.FilePath); err != nil {
			t.Fatalf("backup file does not exist: %v", err)
		}
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		t.Fatalf("failed to read backup dir: %v", err)
	}
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "app-") && strings.HasSuffix(entry.Name(), ".db") {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("expected 2 backup files after rotation, got %d", count)
	}
}
