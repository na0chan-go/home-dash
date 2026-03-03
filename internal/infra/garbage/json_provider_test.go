package garbage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJSONScheduleProvider_GetSchedule_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "garbage.json")
	content := `{
  "sunday": [],
  "monday": ["燃えるゴミ"],
  "tuesday": [],
  "wednesday": [],
  "thursday": [],
  "friday": [],
  "saturday": []
}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	provider := NewJSONScheduleProvider(path)
	schedule, err := provider.GetSchedule(context.Background())
	if err != nil {
		t.Fatalf("GetSchedule returned error: %v", err)
	}
	if len(schedule.Monday) != 1 || schedule.Monday[0] != "燃えるゴミ" {
		t.Fatalf("unexpected monday items: %v", schedule.Monday)
	}
}

func TestJSONScheduleProvider_GetSchedule_RejectsLegacyRulesFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "garbage.json")
	content := `{"rules":[{"weekday":"monday","type":"燃えるゴミ"}]}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	provider := NewJSONScheduleProvider(path)
	_, err := provider.GetSchedule(context.Background())
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse garbage schedule file") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestJSONScheduleProvider_GetSchedule_RejectsMissingWeekday(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "garbage.json")
	content := `{
  "monday": ["燃えるゴミ"],
  "tuesday": [],
  "wednesday": [],
  "thursday": [],
  "friday": [],
  "saturday": []
}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}

	provider := NewJSONScheduleProvider(path)
	_, err := provider.GetSchedule(context.Background())
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !strings.Contains(err.Error(), "weekday fields are required") {
		t.Fatalf("unexpected error: %v", err)
	}
}
