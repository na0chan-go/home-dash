package notes

import "time"

type Kind string

const (
	KindNotice   Kind = "notice"
	KindShopping Kind = "shopping"
)

type Note struct {
	ID        int64
	Kind      Kind
	Body      string
	Pinned    bool
	Done      bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
