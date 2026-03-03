package garbage

import (
	"context"
	"errors"
	"testing"
	"time"

	domaingarbage "github.com/na0chan-go/home-dash/internal/domain/garbage"
	"github.com/na0chan-go/home-dash/internal/ports"
)

type fixedClock struct {
	now time.Time
}

func (c *fixedClock) NowUTC() time.Time {
	return c.now
}

type stubProvider struct {
	schedule domaingarbage.Schedule
	err      error
}

func (p *stubProvider) GetSchedule(context.Context) (domaingarbage.Schedule, error) {
	if p.err != nil {
		return domaingarbage.Schedule{}, p.err
	}
	return p.schedule, nil
}

func TestGetTodayUseCase_UsesTokyoDateAndNoneLabel(t *testing.T) {
	provider := &stubProvider{schedule: domaingarbage.Schedule{Wednesday: []string{}}}
	clock := &fixedClock{now: time.Date(2026, 3, 3, 15, 30, 0, 0, time.UTC)}
	usecase := NewGetTodayUseCase(provider, clock)

	got, err := usecase.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if got.Date != "2026-03-04" {
		t.Fatalf("expected date 2026-03-04, got %s", got.Date)
	}
	if got.Weekday != "wednesday" {
		t.Fatalf("expected weekday wednesday, got %s", got.Weekday)
	}
	if len(got.Items) != 0 {
		t.Fatalf("expected empty items, got %v", got.Items)
	}
	if got.Label != "なし" {
		t.Fatalf("expected label なし, got %s", got.Label)
	}
}

func TestGetTomorrowUseCase_JoinsItemsLabel(t *testing.T) {
	provider := &stubProvider{schedule: domaingarbage.Schedule{Thursday: []string{"燃えるゴミ", "資源"}}}
	clock := &fixedClock{now: time.Date(2026, 3, 3, 15, 30, 0, 0, time.UTC)}
	usecase := NewGetTomorrowUseCase(provider, clock)

	got, err := usecase.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if got.Date != "2026-03-05" {
		t.Fatalf("expected date 2026-03-05, got %s", got.Date)
	}
	if got.Weekday != "thursday" {
		t.Fatalf("expected weekday thursday, got %s", got.Weekday)
	}
	if got.Label != "燃えるゴミ / 資源" {
		t.Fatalf("unexpected label: %s", got.Label)
	}
}

func TestGetSummaryUseCase_ReturnsTodayAndTomorrow(t *testing.T) {
	provider := &stubProvider{schedule: domaingarbage.Schedule{
		Wednesday: []string{"燃えるゴミ"},
		Thursday:  []string{"びん"},
	}}
	clock := &fixedClock{now: time.Date(2026, 3, 3, 15, 30, 0, 0, time.UTC)}
	usecase := NewGetSummaryUseCase(provider, clock)

	got, err := usecase.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if got.Today.Weekday != "wednesday" || got.Tomorrow.Weekday != "thursday" {
		t.Fatalf("unexpected summary weekdays: today=%s tomorrow=%s", got.Today.Weekday, got.Tomorrow.Weekday)
	}
}

func TestGetTodayUseCase_PropagatesProviderError(t *testing.T) {
	provider := &stubProvider{err: errors.New("broken file")}
	clock := &fixedClock{now: time.Now().UTC()}
	usecase := NewGetTodayUseCase(provider, clock)

	_, err := usecase.Execute(context.Background())
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}

func TestInterfacesCompile(t *testing.T) {
	var _ ports.Clock = (*fixedClock)(nil)
	var _ ports.GarbageScheduleProvider = (*stubProvider)(nil)
}
