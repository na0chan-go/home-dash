package ports

import (
	"context"

	"github.com/na0chan-go/home-dash/internal/domain/garbage"
)

type GarbageScheduleProvider interface {
	GetSchedule(ctx context.Context) (garbage.Schedule, error)
}
