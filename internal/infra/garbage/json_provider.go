package garbage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	domaingarbage "github.com/na0chan-go/home-dash/internal/domain/garbage"
)

type JSONScheduleProvider struct {
	path string
}

func NewJSONScheduleProvider(path string) *JSONScheduleProvider {
	return &JSONScheduleProvider{path: path}
}

func (p *JSONScheduleProvider) GetSchedule(_ context.Context) (domaingarbage.Schedule, error) {
	data, err := os.ReadFile(p.path)
	if err != nil {
		return domaingarbage.Schedule{}, fmt.Errorf("failed to read garbage schedule file: %w", err)
	}

	var schedule domaingarbage.Schedule
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&schedule); err != nil {
		return domaingarbage.Schedule{}, fmt.Errorf("failed to parse garbage schedule file: %w", err)
	}
	if dec.More() {
		return domaingarbage.Schedule{}, fmt.Errorf("failed to parse garbage schedule file: trailing data")
	}
	if err := validateSchedule(schedule); err != nil {
		return domaingarbage.Schedule{}, err
	}

	return schedule, nil
}

func validateSchedule(schedule domaingarbage.Schedule) error {
	if schedule.Sunday == nil ||
		schedule.Monday == nil ||
		schedule.Tuesday == nil ||
		schedule.Wednesday == nil ||
		schedule.Thursday == nil ||
		schedule.Friday == nil ||
		schedule.Saturday == nil {
		return fmt.Errorf("failed to parse garbage schedule file: weekday fields are required")
	}
	return nil
}
