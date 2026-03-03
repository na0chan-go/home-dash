package garbage

type Weekday string

const (
	Sunday    Weekday = "sunday"
	Monday    Weekday = "monday"
	Tuesday   Weekday = "tuesday"
	Wednesday Weekday = "wednesday"
	Thursday  Weekday = "thursday"
	Friday    Weekday = "friday"
	Saturday  Weekday = "saturday"
)

type Schedule struct {
	Sunday    []string `json:"sunday"`
	Monday    []string `json:"monday"`
	Tuesday   []string `json:"tuesday"`
	Wednesday []string `json:"wednesday"`
	Thursday  []string `json:"thursday"`
	Friday    []string `json:"friday"`
	Saturday  []string `json:"saturday"`
}

func (s Schedule) ItemsByWeekday(weekday Weekday) []string {
	switch weekday {
	case Sunday:
		return copyItems(s.Sunday)
	case Monday:
		return copyItems(s.Monday)
	case Tuesday:
		return copyItems(s.Tuesday)
	case Wednesday:
		return copyItems(s.Wednesday)
	case Thursday:
		return copyItems(s.Thursday)
	case Friday:
		return copyItems(s.Friday)
	case Saturday:
		return copyItems(s.Saturday)
	default:
		return []string{}
	}
}

func copyItems(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	out := make([]string, len(items))
	copy(out, items)
	return out
}
