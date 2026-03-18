package ports

import (
	"context"

	"github.com/na0chan-go/home-dash/internal/domain/notes"
)

type NoteOrder int

const (
	NoteOrderCreatedAtDesc NoteOrder = iota
	NoteOrderNotice
	NoteOrderShopping
)

type ListNotesParams struct {
	Kind  *notes.Kind
	Limit int
	Order NoteOrder
}

type CreateNoteParams struct {
	Kind   notes.Kind
	Body   string
	Author string
	Pinned bool
	Done   bool
}

type NotesRepository interface {
	List(ctx context.Context, params ListNotesParams) ([]notes.Note, error)
	Create(ctx context.Context, params CreateNoteParams) (notes.Note, error)
	Delete(ctx context.Context, id int64) (bool, error)
	SetPinned(ctx context.Context, id int64, pinned bool) (notes.Note, bool, error)
	SetDone(ctx context.Context, id int64, done bool) (notes.Note, bool, error)
}
