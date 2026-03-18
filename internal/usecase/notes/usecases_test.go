package notes

import (
	"context"
	"testing"
	"time"

	domainnotes "github.com/na0chan-go/home-dash/internal/domain/notes"
	"github.com/na0chan-go/home-dash/internal/ports"
)

type stubNotesRepository struct {
	created ports.CreateNoteParams
}

func (r *stubNotesRepository) List(context.Context, ports.ListNotesParams) ([]domainnotes.Note, error) {
	return nil, nil
}

func (r *stubNotesRepository) Create(_ context.Context, params ports.CreateNoteParams) (domainnotes.Note, error) {
	r.created = params
	now := time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC)
	return domainnotes.Note{
		ID:           1,
		Kind:         params.Kind,
		Body:         params.Body,
		Author:       params.Author,
		Pinned:       params.Pinned,
		Acknowledged: false,
		Done:         params.Done,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (r *stubNotesRepository) Delete(context.Context, int64) (bool, error) {
	return false, nil
}

func (r *stubNotesRepository) SetPinned(context.Context, int64, bool) (domainnotes.Note, bool, error) {
	return domainnotes.Note{}, false, nil
}

func (r *stubNotesRepository) SetAcknowledged(_ context.Context, _ int64, acknowledged bool) (domainnotes.Note, bool, error) {
	return domainnotes.Note{
		ID:           1,
		Kind:         domainnotes.KindNotice,
		Body:         "連絡",
		Author:       "妻",
		Pinned:       false,
		Acknowledged: acknowledged,
		Done:         false,
		CreatedAt:    time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC),
	}, true, nil
}

func (r *stubNotesRepository) SetDone(context.Context, int64, bool) (domainnotes.Note, bool, error) {
	return domainnotes.Note{}, false, nil
}

func TestAddNoteUseCase_NoticeTrimsAuthor(t *testing.T) {
	repo := &stubNotesRepository{}
	usecase := NewAddNoteUseCase(repo)

	got, err := usecase.Execute(context.Background(), AddNoteInput{
		Kind:   "notice",
		Body:   "  明日買い物に行こう  ",
		Author: "  妻  ",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if repo.created.Author != "妻" {
		t.Fatalf("expected author to be trimmed, got %q", repo.created.Author)
	}
	if got.Author != "妻" {
		t.Fatalf("expected dto author to be preserved, got %q", got.Author)
	}
}

func TestAddNoteUseCase_ShoppingIgnoresAuthor(t *testing.T) {
	repo := &stubNotesRepository{}
	usecase := NewAddNoteUseCase(repo)

	got, err := usecase.Execute(context.Background(), AddNoteInput{
		Kind:   "shopping",
		Body:   "牛乳",
		Author: "夫",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if repo.created.Author != "" {
		t.Fatalf("expected shopping author to be normalized to empty, got %q", repo.created.Author)
	}
	if got.Author != "" {
		t.Fatalf("expected dto author to be empty, got %q", got.Author)
	}
}

func TestAddNoteUseCase_ReturnsErrorWhenAuthorTooLong(t *testing.T) {
	repo := &stubNotesRepository{}
	usecase := NewAddNoteUseCase(repo)

	_, err := usecase.Execute(context.Background(), AddNoteInput{
		Kind:   "notice",
		Body:   "連絡",
		Author: "abcdefghijklmnopqrstu",
	})
	if err != ErrAuthorTooLong {
		t.Fatalf("expected ErrAuthorTooLong, got %v", err)
	}
}

func TestSetAcknowledgedUseCase_NoticeCanBeAcknowledged(t *testing.T) {
	repo := &stubNotesRepository{}
	usecase := NewSetAcknowledgedUseCase(repo)

	got, err := usecase.Execute(context.Background(), SetAcknowledgedInput{
		ID:           1,
		Acknowledged: true,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !got.Acknowledged {
		t.Fatal("expected acknowledged=true")
	}
}
