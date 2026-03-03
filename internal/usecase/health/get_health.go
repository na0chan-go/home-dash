package health

import (
	"context"

	"github.com/na0chan-go/home-dash/internal/ports"
)

type GetHealthOutput struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

type GetHealthUseCase struct {
	clock ports.Clock
}

func NewGetHealthUseCase(clock ports.Clock) *GetHealthUseCase {
	return &GetHealthUseCase{clock: clock}
}

func (u *GetHealthUseCase) Execute(_ context.Context) (GetHealthOutput, error) {
	return GetHealthOutput{
		Status: "ok",
		Time:   u.clock.NowUTC().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
