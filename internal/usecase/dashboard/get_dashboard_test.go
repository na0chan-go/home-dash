package dashboard

import (
	"context"
	"errors"
	"testing"
	"time"

	domaingarbage "github.com/na0chan-go/home-dash/internal/domain/garbage"
	domainnotes "github.com/na0chan-go/home-dash/internal/domain/notes"
	"github.com/na0chan-go/home-dash/internal/ports"
)

type fixedClock struct {
	now time.Time
}

func (c *fixedClock) NowUTC() time.Time {
	return c.now
}

type stubNotesRepo struct {
	listCalls []ports.ListNotesParams
	notice    []domainnotes.Note
	shopping  []domainnotes.Note
	err       error
}

func (r *stubNotesRepo) List(_ context.Context, params ports.ListNotesParams) ([]domainnotes.Note, error) {
	r.listCalls = append(r.listCalls, params)
	if r.err != nil {
		return nil, r.err
	}
	if params.Kind != nil && *params.Kind == domainnotes.KindNotice {
		return r.notice, nil
	}
	if params.Kind != nil && *params.Kind == domainnotes.KindShopping {
		return r.shopping, nil
	}
	return nil, nil
}

func (r *stubNotesRepo) Create(context.Context, ports.CreateNoteParams) (domainnotes.Note, error) {
	return domainnotes.Note{}, nil
}

func (r *stubNotesRepo) Delete(context.Context, int64) (bool, error) {
	return false, nil
}

func (r *stubNotesRepo) SetPinned(context.Context, int64, bool) (domainnotes.Note, bool, error) {
	return domainnotes.Note{}, false, nil
}

func (r *stubNotesRepo) SetDone(context.Context, int64, bool) (domainnotes.Note, bool, error) {
	return domainnotes.Note{}, false, nil
}

type stubGarbageProvider struct {
	schedule domaingarbage.Schedule
	err      error
}

func (p *stubGarbageProvider) GetSchedule(context.Context) (domaingarbage.Schedule, error) {
	if p.err != nil {
		return domaingarbage.Schedule{}, p.err
	}
	return p.schedule, nil
}

func TestGetDashboardUseCase_CollectsNotesAndGarbage(t *testing.T) {
	notesRepo := &stubNotesRepo{
		notice: []domainnotes.Note{{
			ID:        1,
			Kind:      domainnotes.KindNotice,
			Body:      "連絡",
			Pinned:    true,
			Done:      false,
			CreatedAt: time.Date(2026, 3, 4, 9, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 3, 4, 9, 0, 0, 0, time.UTC),
		}},
		shopping: []domainnotes.Note{{
			ID:        2,
			Kind:      domainnotes.KindShopping,
			Body:      "牛乳",
			Pinned:    false,
			Done:      false,
			CreatedAt: time.Date(2026, 3, 4, 9, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 3, 4, 9, 0, 0, 0, time.UTC),
		}},
	}
	provider := &stubGarbageProvider{schedule: domaingarbage.Schedule{
		Wednesday: []string{},
		Thursday:  []string{"燃えるゴミ"},
		Sunday:    []string{},
		Monday:    []string{},
		Tuesday:   []string{},
		Friday:    []string{},
		Saturday:  []string{},
	}}
	clock := &fixedClock{now: time.Date(2026, 3, 3, 15, 30, 0, 0, time.UTC)}

	usecase := NewGetDashboardUseCase(notesRepo, provider, clock)
	got, err := usecase.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if got.GeneratedAt != "2026-03-04T00:30:00+09:00" {
		t.Fatalf("unexpected generatedAt: %s", got.GeneratedAt)
	}
	if len(got.Notes.Notice) != 1 || len(got.Notes.Shopping) != 1 {
		t.Fatalf("unexpected notes count: notice=%d shopping=%d", len(got.Notes.Notice), len(got.Notes.Shopping))
	}
	if got.Garbage.Today.Label != "なし" {
		t.Fatalf("unexpected today label: %s", got.Garbage.Today.Label)
	}
	if got.Garbage.Tomorrow.Label != "燃えるゴミ" {
		t.Fatalf("unexpected tomorrow label: %s", got.Garbage.Tomorrow.Label)
	}

	if len(notesRepo.listCalls) != 2 {
		t.Fatalf("expected 2 list calls, got %d", len(notesRepo.listCalls))
	}
	if notesRepo.listCalls[0].Kind == nil || *notesRepo.listCalls[0].Kind != domainnotes.KindNotice {
		t.Fatalf("expected first kind notice, got %+v", notesRepo.listCalls[0].Kind)
	}
	if notesRepo.listCalls[0].Order != ports.NoteOrderNotice || notesRepo.listCalls[0].Limit != 10 {
		t.Fatalf("unexpected first list params: %+v", notesRepo.listCalls[0])
	}
	if notesRepo.listCalls[1].Kind == nil || *notesRepo.listCalls[1].Kind != domainnotes.KindShopping {
		t.Fatalf("expected second kind shopping, got %+v", notesRepo.listCalls[1].Kind)
	}
	if notesRepo.listCalls[1].Order != ports.NoteOrderShopping || notesRepo.listCalls[1].Limit != 20 {
		t.Fatalf("unexpected second list params: %+v", notesRepo.listCalls[1])
	}
}

func TestGetDashboardUseCase_ReturnsErrorOnNotesFailure(t *testing.T) {
	notesRepo := &stubNotesRepo{err: errors.New("db error")}
	provider := &stubGarbageProvider{}
	clock := &fixedClock{now: time.Now().UTC()}

	usecase := NewGetDashboardUseCase(notesRepo, provider, clock)
	if _, err := usecase.Execute(context.Background()); err == nil {
		t.Fatal("expected error but got nil")
	}
}

func TestGetDashboardUseCase_ReturnsErrorOnGarbageFailure(t *testing.T) {
	notesRepo := &stubNotesRepo{}
	provider := &stubGarbageProvider{err: errors.New("file error")}
	clock := &fixedClock{now: time.Now().UTC()}

	usecase := NewGetDashboardUseCase(notesRepo, provider, clock)
	if _, err := usecase.Execute(context.Background()); err == nil {
		t.Fatal("expected error but got nil")
	}
}
