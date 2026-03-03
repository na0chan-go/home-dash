package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	domainnotes "github.com/na0chan-go/home-dash/internal/domain/notes"
	"github.com/na0chan-go/home-dash/internal/ports"
	usenotes "github.com/na0chan-go/home-dash/internal/usecase/notes"
)

type stubNotesRepo struct {
	setPinnedCalled bool
	setDoneCalled   bool
}

func (s *stubNotesRepo) List(context.Context, ports.ListNotesParams) ([]domainnotes.Note, error) {
	return nil, nil
}

func (s *stubNotesRepo) Create(context.Context, ports.CreateNoteParams) (domainnotes.Note, error) {
	return domainnotes.Note{}, nil
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

func newNotesHandlerForTest() (*NotesHandler, *stubNotesRepo) {
	repo := &stubNotesRepo{}

	listUC := usenotes.NewListNotesUseCase(repo)
	addUC := usenotes.NewAddNoteUseCase(repo)
	deleteUC := usenotes.NewDeleteNoteUseCase(repo)
	setPinUC := usenotes.NewSetPinUseCase(repo)
	setDoneUC := usenotes.NewSetDoneUseCase(repo)

	return NewNotesHandler(listUC, addUC, deleteUC, setPinUC, setDoneUC), repo
}
