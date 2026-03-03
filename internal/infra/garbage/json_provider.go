package garbage

import (
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
	if err := json.Unmarshal(data, &schedule); err != nil {
		return domaingarbage.Schedule{}, fmt.Errorf("failed to parse garbage schedule file: %w", err)
	}

	return schedule, nil
}
