package notes

type NoteDTO struct {
	ID        int64  `json:"id"`
	Kind      string `json:"kind"`
	Body      string `json:"body"`
	Pinned    bool   `json:"pinned"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
