package garbage

import (
	"context"
	"fmt"
	"strings"
	"time"

	domaingarbage "github.com/na0chan-go/home-dash/internal/domain/garbage"
	"github.com/na0chan-go/home-dash/internal/ports"
)

const tokyoTimezone = "Asia/Tokyo"

type GetTodayUseCase struct {
	provider ports.GarbageScheduleProvider
	clock    ports.Clock
}

type GetTomorrowUseCase struct {
	provider ports.GarbageScheduleProvider
	clock    ports.Clock
}

type GetSummaryUseCase struct {
	provider ports.GarbageScheduleProvider
	clock    ports.Clock
}

func NewGetTodayUseCase(provider ports.GarbageScheduleProvider, clock ports.Clock) *GetTodayUseCase {
	return &GetTodayUseCase{provider: provider, clock: clock}
}

func NewGetTomorrowUseCase(provider ports.GarbageScheduleProvider, clock ports.Clock) *GetTomorrowUseCase {
	return &GetTomorrowUseCase{provider: provider, clock: clock}
}

func NewGetSummaryUseCase(provider ports.GarbageScheduleProvider, clock ports.Clock) *GetSummaryUseCase {
	return &GetSummaryUseCase{provider: provider, clock: clock}
}

func (u *GetTodayUseCase) Execute(ctx context.Context) (DailyGarbageDTO, error) {
	schedule, err := u.provider.GetSchedule(ctx)
	if err != nil {
		return DailyGarbageDTO{}, fmt.Errorf("failed to load garbage schedule: %w", err)
	}
	return buildDaily(schedule, nowInTokyo(u.clock), 0), nil
}

func (u *GetTomorrowUseCase) Execute(ctx context.Context) (DailyGarbageDTO, error) {
	schedule, err := u.provider.GetSchedule(ctx)
	if err != nil {
		return DailyGarbageDTO{}, fmt.Errorf("failed to load garbage schedule: %w", err)
	}
	return buildDaily(schedule, nowInTokyo(u.clock), 1), nil
}

func (u *GetSummaryUseCase) Execute(ctx context.Context) (SummaryDTO, error) {
	schedule, err := u.provider.GetSchedule(ctx)
	if err != nil {
		return SummaryDTO{}, fmt.Errorf("failed to load garbage schedule: %w", err)
	}

	now := nowInTokyo(u.clock)
	return SummaryDTO{
		Today:    buildDaily(schedule, now, 0),
		Tomorrow: buildDaily(schedule, now, 1),
	}, nil
}

func buildDaily(schedule domaingarbage.Schedule, now time.Time, dayOffset int) DailyGarbageDTO {
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
