package garbage

type DailyGarbageDTO struct {
	Date    string   `json:"date"`
	Weekday string   `json:"weekday"`
	Items   []string `json:"items"`
	Label   string   `json:"label"`
}

type SummaryDTO struct {
	Today    DailyGarbageDTO `json:"today"`
	Tomorrow DailyGarbageDTO `json:"tomorrow"`
}
