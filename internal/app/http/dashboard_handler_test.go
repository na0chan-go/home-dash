package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	domaingarbage "github.com/na0chan-go/home-dash/internal/domain/garbage"
	domainnotes "github.com/na0chan-go/home-dash/internal/domain/notes"
	"github.com/na0chan-go/home-dash/internal/ports"
	usedashboard "github.com/na0chan-go/home-dash/internal/usecase/dashboard"
)

type dashboardStubClock struct {
	now time.Time
}

func (c *dashboardStubClock) NowUTC() time.Time {
	return c.now
}

type dashboardStubNotesRepo struct {
	err error
}

func (r *dashboardStubNotesRepo) List(context.Context, ports.ListNotesParams) ([]domainnotes.Note, error) {
	if r.err != nil {
		return nil, r.err
	}
	return []domainnotes.Note{}, nil
}

func (r *dashboardStubNotesRepo) Create(context.Context, ports.CreateNoteParams) (domainnotes.Note, error) {
	return domainnotes.Note{}, nil
}

func (r *dashboardStubNotesRepo) Delete(context.Context, int64) (bool, error) {
	return false, nil
}

func (r *dashboardStubNotesRepo) SetPinned(context.Context, int64, bool) (domainnotes.Note, bool, error) {
	return domainnotes.Note{}, false, nil
}

func (r *dashboardStubNotesRepo) SetAcknowledged(context.Context, int64, bool) (domainnotes.Note, bool, error) {
	return domainnotes.Note{}, false, nil
}

func (r *dashboardStubNotesRepo) SetDone(context.Context, int64, bool) (domainnotes.Note, bool, error) {
	return domainnotes.Note{}, false, nil
}

type dashboardStubGarbageProvider struct {
	err error
}

func (p *dashboardStubGarbageProvider) GetSchedule(context.Context) (domaingarbage.Schedule, error) {
	if p.err != nil {
		return domaingarbage.Schedule{}, p.err
	}
	return domaingarbage.Schedule{
		Sunday:    []string{},
		Monday:    []string{},
		Tuesday:   []string{},
		Wednesday: []string{},
		Thursday:  []string{},
		Friday:    []string{},
		Saturday:  []string{},
	}, nil
}

func TestDashboardHandler_Get_Success(t *testing.T) {
	notesRepo := &dashboardStubNotesRepo{}
	provider := &dashboardStubGarbageProvider{}
	clock := &dashboardStubClock{now: time.Date(2026, 3, 4, 0, 0, 0, 0, time.UTC)}

	usecase := usedashboard.NewGetDashboardUseCase(notesRepo, provider, clock)
	handler := NewDashboardHandler(usecase)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := payload["generatedAt"]; !ok {
		t.Fatalf("generatedAt is missing: %s", rec.Body.String())
	}
}

func TestDashboardHandler_Get_ReturnsGeneric500Error(t *testing.T) {
	notesRepo := &dashboardStubNotesRepo{err: errors.New("sqlite: no such table notes")}
	provider := &dashboardStubGarbageProvider{}
	clock := &dashboardStubClock{now: time.Now().UTC()}

	usecase := usedashboard.NewGetDashboardUseCase(notesRepo, provider, clock)
	handler := NewDashboardHandler(usecase)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	rec := httptest.NewRecorder()

	handler.Get(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"internal server error"`) {
		t.Fatalf("expected generic internal error, got: %s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "sqlite") {
		t.Fatalf("response leaked internal details: %s", rec.Body.String())
	}
}
