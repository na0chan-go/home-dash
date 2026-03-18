package notes

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/na0chan-go/home-dash/internal/infra/db"
	"github.com/na0chan-go/home-dash/internal/ports"
	usenotes "github.com/na0chan-go/home-dash/internal/usecase/notes"
)

func TestSQLiteRepository_PersistsAuthor(t *testing.T) {
	tempDir := t.TempDir()
	databasePath := filepath.Join(tempDir, "app.db")

	sqliteDB, err := db.OpenSQLite(databasePath)
	if err != nil {
		t.Fatalf("OpenSQLite returned error: %v", err)
	}
	defer sqliteDB.Close()

	if err := db.RunMigrations(context.Background(), sqliteDB); err != nil {
		t.Fatalf("RunMigrations returned error: %v", err)
	}

	repo := NewSQLiteRepository(sqliteDB)
	created, err := repo.Create(context.Background(), ports.CreateNoteParams{
		Kind:   "notice",
		Body:   "連絡",
		Author: "妻",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if created.Author != "妻" {
		t.Fatalf("expected created author 妻, got %q", created.Author)
	}
	if created.Acknowledged {
		t.Fatal("expected created notice to start unacknowledged")
	}

	listed, err := repo.List(context.Background(), ports.ListNotesParams{Limit: 10})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(listed) != 1 {
		t.Fatalf("expected 1 note, got %d", len(listed))
	}
	if listed[0].Author != "妻" {
		t.Fatalf("expected listed author 妻, got %q", listed[0].Author)
	}
}

func TestSQLiteRepository_SetAcknowledgedForNotice(t *testing.T) {
	tempDir := t.TempDir()
	databasePath := filepath.Join(tempDir, "app.db")

	sqliteDB, err := db.OpenSQLite(databasePath)
	if err != nil {
		t.Fatalf("OpenSQLite returned error: %v", err)
	}
	defer sqliteDB.Close()

	if err := db.RunMigrations(context.Background(), sqliteDB); err != nil {
		t.Fatalf("RunMigrations returned error: %v", err)
	}

	repo := NewSQLiteRepository(sqliteDB)
	created, err := repo.Create(context.Background(), ports.CreateNoteParams{
		Kind:   "notice",
		Body:   "連絡",
		Author: "妻",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	updated, found, err := repo.SetAcknowledged(context.Background(), created.ID, true)
	if err != nil {
		t.Fatalf("SetAcknowledged returned error: %v", err)
	}
	if !found {
		t.Fatal("expected note to be found")
	}
	if !updated.Acknowledged {
		t.Fatal("expected acknowledged=true after update")
	}
}

func TestSQLiteRepository_ShoppingAuthorRemainsEmpty(t *testing.T) {
	tempDir := t.TempDir()
	databasePath := filepath.Join(tempDir, "app.db")

	sqliteDB, err := db.OpenSQLite(databasePath)
	if err != nil {
		t.Fatalf("OpenSQLite returned error: %v", err)
	}
	defer sqliteDB.Close()

	if err := db.RunMigrations(context.Background(), sqliteDB); err != nil {
		t.Fatalf("RunMigrations returned error: %v", err)
	}

	repo := NewSQLiteRepository(sqliteDB)
	addUseCase := usenotes.NewAddNoteUseCase(repo)
	created, err := addUseCase.Execute(context.Background(), usenotes.AddNoteInput{
		Kind:   "shopping",
		Body:   "牛乳",
		Author: "夫",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if created.Author != "" {
		t.Fatalf("expected shopping author to be empty, got %q", created.Author)
	}
}
