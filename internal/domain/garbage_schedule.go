package domain

type GarbageSchedule struct {
	Rules []GarbageRule `json:"rules"`
}

type GarbageRule struct {
	Weekday string `json:"weekday"`
	Type    string `json:"type"`
}
