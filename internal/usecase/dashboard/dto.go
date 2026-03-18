package dashboard

type DashboardDTO struct {
	GeneratedAt string            `json:"generatedAt"`
	Notes       DashboardNotesDTO `json:"notes"`
	Garbage     DashboardGarbage  `json:"garbage"`
}

type DashboardNotesDTO struct {
	Notice   []NoteDTO `json:"notice"`
	Shopping []NoteDTO `json:"shopping"`
}

type DashboardGarbage struct {
	Today    DailyGarbageDTO `json:"today"`
	Tomorrow DailyGarbageDTO `json:"tomorrow"`
}

type NoteDTO struct {
	ID           int64  `json:"id"`
	Kind         string `json:"kind"`
	Body         string `json:"body"`
	Author       string `json:"author"`
	Pinned       bool   `json:"pinned"`
	Acknowledged bool   `json:"acknowledged"`
	Done         bool   `json:"done"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type DailyGarbageDTO struct {
	Date    string   `json:"date"`
	Weekday string   `json:"weekday"`
	Items   []string `json:"items"`
	Label   string   `json:"label"`
}
