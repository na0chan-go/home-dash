package notes

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
