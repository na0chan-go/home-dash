package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/na0chan-go/home-dash/internal/domain"
)

func LoadGarbageSchedule(path string) (domain.GarbageSchedule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.GarbageSchedule{}, fmt.Errorf("failed to read garbage schedule: %w", err)
	}

	var schedule domain.GarbageSchedule
	if err := json.Unmarshal(data, &schedule); err != nil {
		return domain.GarbageSchedule{}, fmt.Errorf("failed to parse garbage schedule: %w", err)
	}

	return schedule, nil
}
