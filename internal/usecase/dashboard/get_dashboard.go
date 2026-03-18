package dashboard

import (
	"context"
	"fmt"
	"strings"
	"time"

	domaingarbage "github.com/na0chan-go/home-dash/internal/domain/garbage"
	domainnotes "github.com/na0chan-go/home-dash/internal/domain/notes"
	"github.com/na0chan-go/home-dash/internal/ports"
)

const (
	noticeLimit     = 10
	shoppingLimit   = 20
	tokyoTimezone   = "Asia/Tokyo"
	timestampLayout = "2006-01-02T15:04:05Z07:00"
)

type GetDashboardUseCase struct {
	notesRepo       ports.NotesRepository
	garbageProvider ports.GarbageScheduleProvider
	clock           ports.Clock
}

func NewGetDashboardUseCase(
	notesRepo ports.NotesRepository,
	garbageProvider ports.GarbageScheduleProvider,
	clock ports.Clock,
) *GetDashboardUseCase {
	return &GetDashboardUseCase{
		notesRepo:       notesRepo,
		garbageProvider: garbageProvider,
		clock:           clock,
	}
}

func (u *GetDashboardUseCase) Execute(ctx context.Context) (DashboardDTO, error) {
	noticeKind := domainnotes.KindNotice
	noticeNotes, err := u.notesRepo.List(ctx, ports.ListNotesParams{
		Kind:  &noticeKind,
		Limit: noticeLimit,
		Order: ports.NoteOrderNotice,
	})
	if err != nil {
		return DashboardDTO{}, fmt.Errorf("failed to load notice notes: %w", err)
	}

	shoppingKind := domainnotes.KindShopping
	shoppingNotes, err := u.notesRepo.List(ctx, ports.ListNotesParams{
		Kind:  &shoppingKind,
		Limit: shoppingLimit,
		Order: ports.NoteOrderShopping,
	})
	if err != nil {
		return DashboardDTO{}, fmt.Errorf("failed to load shopping notes: %w", err)
	}

	schedule, err := u.garbageProvider.GetSchedule(ctx)
	if err != nil {
		return DashboardDTO{}, fmt.Errorf("failed to load garbage schedule: %w", err)
	}

	now := nowInTokyo(u.clock)
	return DashboardDTO{
		GeneratedAt: now.Format(timestampLayout),
		Notes: DashboardNotesDTO{
			Notice:   toNoteDTOs(noticeNotes),
			Shopping: toNoteDTOs(shoppingNotes),
		},
		Garbage: DashboardGarbage{
			Today:    buildDailyGarbage(schedule, now, 0),
			Tomorrow: buildDailyGarbage(schedule, now, 1),
		},
	}, nil
}

func toNoteDTOs(notes []domainnotes.Note) []NoteDTO {
	out := make([]NoteDTO, 0, len(notes))
	for _, n := range notes {
		out = append(out, NoteDTO{
			ID:        n.ID,
			Kind:      string(n.Kind),
			Body:      n.Body,
			Author:    n.Author,
			Pinned:    n.Pinned,
			Done:      n.Done,
			CreatedAt: n.CreatedAt.Format(timestampLayout),
			UpdatedAt: n.UpdatedAt.Format(timestampLayout),
		})
	}
	return out
}

func buildDailyGarbage(schedule domaingarbage.Schedule, now time.Time, dayOffset int) DailyGarbageDTO {
	target := now.AddDate(0, 0, dayOffset)
	weekday := toWeekday(target.Weekday())
	items := schedule.ItemsByWeekday(weekday)

	label := "なし"
	if len(items) > 0 {
		label = strings.Join(items, " / ")
	}

	return DailyGarbageDTO{
		Date:    target.Format("2006-01-02"),
		Weekday: string(weekday),
		Items:   items,
		Label:   label,
	}
}

func nowInTokyo(clock ports.Clock) time.Time {
	loc, err := time.LoadLocation(tokyoTimezone)
	if err != nil {
		loc = time.FixedZone("JST", 9*60*60)
	}
	return clock.NowUTC().In(loc)
}

func toWeekday(weekday time.Weekday) domaingarbage.Weekday {
	switch weekday {
	case time.Sunday:
		return domaingarbage.Sunday
	case time.Monday:
		return domaingarbage.Monday
	case time.Tuesday:
		return domaingarbage.Tuesday
	case time.Wednesday:
		return domaingarbage.Wednesday
	case time.Thursday:
		return domaingarbage.Thursday
	case time.Friday:
		return domaingarbage.Friday
	default:
		return domaingarbage.Saturday
	}
}
