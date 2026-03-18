package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	domainnotes "github.com/na0chan-go/home-dash/internal/domain/notes"
	"github.com/na0chan-go/home-dash/internal/ports"
	usenotes "github.com/na0chan-go/home-dash/internal/usecase/notes"
)

type stubNotesRepo struct {
	setPinnedCalled bool
	setDoneCalled   bool
	created         ports.CreateNoteParams
}

func (s *stubNotesRepo) List(context.Context, ports.ListNotesParams) ([]domainnotes.Note, error) {
	return nil, nil
}

func (s *stubNotesRepo) Create(_ context.Context, params ports.CreateNoteParams) (domainnotes.Note, error) {
	s.created = params
	now := time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC)
	return domainnotes.Note{
		ID:        1,
		Kind:      params.Kind,
		Body:      params.Body,
		Author:    params.Author,
		Pinned:    params.Pinned,
		Done:      params.Done,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (s *stubNotesRepo) Delete(context.Context, int64) (bool, error) {
	return false, nil
}

func (s *stubNotesRepo) SetPinned(context.Context, int64, bool) (domainnotes.Note, bool, error) {
	s.setPinnedCalled = true
	return domainnotes.Note{}, false, nil
}

func (s *stubNotesRepo) SetDone(context.Context, int64, bool) (domainnotes.Note, bool, error) {
	s.setDoneCalled = true
	return domainnotes.Note{}, false, nil
}

func TestNotesHandler_SetPin_RequiresPinnedField(t *testing.T) {
	h, repo := newNotesHandlerForTest()

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/notes/1/pin", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()

	h.HandleByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"pinned is required"`) {
		t.Fatalf("unexpected response body: %s", rec.Body.String())
	}
	if repo.setPinnedCalled {
		t.Fatal("set pinned usecase should not be called when pinned field is missing")
	}
}

func TestNotesHandler_SetPin_InvalidKeyDoesNotUnpin(t *testing.T) {
	h, repo := newNotesHandlerForTest()

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/notes/1/pin", strings.NewReader(`{"pin": true}`))
	rec := httptest.NewRecorder()

	h.HandleByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if repo.setPinnedCalled {
		t.Fatal("set pinned usecase should not be called for invalid payload")
	}
}

func TestNotesHandler_SetDone_RequiresDoneField(t *testing.T) {
	h, repo := newNotesHandlerForTest()

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/notes/1/done", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()

	h.HandleByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"done is required"`) {
		t.Fatalf("unexpected response body: %s", rec.Body.String())
	}
	if repo.setDoneCalled {
		t.Fatal("set done usecase should not be called when done field is missing")
	}
}

func TestNotesHandler_Create_PassesAuthorToUseCase(t *testing.T) {
	h, repo := newNotesHandlerForTest()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/notes", strings.NewReader(`{"kind":"notice","body":"連絡","author":"妻"}`))
	rec := httptest.NewRecorder()

	h.HandleNotes(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}
	if repo.created.Author != "妻" {
		t.Fatalf("expected author 妻, got %q", repo.created.Author)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response json: %v", err)
	}
	if body["author"] != "妻" {
		t.Fatalf("expected response author 妻, got %#v", body["author"])
	}
}

func newNotesHandlerForTest() (*NotesHandler, *stubNotesRepo) {
	repo := &stubNotesRepo{}

	listUC := usenotes.NewListNotesUseCase(repo)
	addUC := usenotes.NewAddNoteUseCase(repo)
	deleteUC := usenotes.NewDeleteNoteUseCase(repo)
	setPinUC := usenotes.NewSetPinUseCase(repo)
	setDoneUC := usenotes.NewSetDoneUseCase(repo)

	return NewNotesHandler(listUC, addUC, deleteUC, setPinUC, setDoneUC), repo
}
