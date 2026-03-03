package notes

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	domainnotes "github.com/na0chan-go/home-dash/internal/domain/notes"
	"github.com/na0chan-go/home-dash/internal/ports"
)

type ListNotesUseCase struct {
	repo ports.NotesRepository
}

type AddNoteUseCase struct {
	repo ports.NotesRepository
}

type DeleteNoteUseCase struct {
	repo ports.NotesRepository
}

type SetPinUseCase struct {
	repo ports.NotesRepository
}

type SetDoneUseCase struct {
	repo ports.NotesRepository
}

type ListNotesInput struct {
	Kind  string
	Limit int
}

type AddNoteInput struct {
	Kind   string
	Body   string
	Pinned bool
	Done   bool
}

type SetPinInput struct {
	ID     int64
	Pinned bool
}

type SetDoneInput struct {
	ID   int64
	Done bool
}

func NewListNotesUseCase(repo ports.NotesRepository) *ListNotesUseCase {
	return &ListNotesUseCase{repo: repo}
}

func NewAddNoteUseCase(repo ports.NotesRepository) *AddNoteUseCase {
	return &AddNoteUseCase{repo: repo}
}

func NewDeleteNoteUseCase(repo ports.NotesRepository) *DeleteNoteUseCase {
	return &DeleteNoteUseCase{repo: repo}
}

func NewSetPinUseCase(repo ports.NotesRepository) *SetPinUseCase {
	return &SetPinUseCase{repo: repo}
}

func NewSetDoneUseCase(repo ports.NotesRepository) *SetDoneUseCase {
	return &SetDoneUseCase{repo: repo}
}

func (u *ListNotesUseCase) Execute(ctx context.Context, input ListNotesInput) ([]NoteDTO, error) {
	limit := input.Limit
	if limit == 0 {
		limit = 50
	}
	if limit < 1 || limit > 100 {
		return nil, ErrInvalidLimit
	}

	var kindPtr *domainnotes.Kind
	order := ports.NoteOrderCreatedAtDesc
	if input.Kind != "" {
		kind, err := parseKind(input.Kind)
		if err != nil {
			return nil, err
		}
		kindPtr = &kind
		if kind == domainnotes.KindNotice {
			order = ports.NoteOrderNotice
		}
		if kind == domainnotes.KindShopping {
			order = ports.NoteOrderShopping
		}
	}

	notes, err := u.repo.List(ctx, ports.ListNotesParams{
		Kind:  kindPtr,
		Limit: limit,
		Order: order,
	})
	if err != nil {
		return nil, err
	}

	out := make([]NoteDTO, 0, len(notes))
	for _, n := range notes {
		out = append(out, toDTO(n))
	}
	return out, nil
}

func (u *AddNoteUseCase) Execute(ctx context.Context, input AddNoteInput) (NoteDTO, error) {
	kind, err := parseKind(input.Kind)
	if err != nil {
		return NoteDTO{}, err
	}

	body, err := validateAndNormalizeBody(input.Body)
	if err != nil {
		return NoteDTO{}, err
	}

	pinned, done := normalizeFlags(kind, input.Pinned, input.Done)
	note, err := u.repo.Create(ctx, ports.CreateNoteParams{
		Kind:   kind,
		Body:   body,
		Pinned: pinned,
		Done:   done,
	})
	if err != nil {
		return NoteDTO{}, err
	}
	return toDTO(note), nil
}

func (u *DeleteNoteUseCase) Execute(ctx context.Context, id int64) error {
	deleted, err := u.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrNoteNotFound
	}
	return nil
}

func (u *SetPinUseCase) Execute(ctx context.Context, input SetPinInput) (NoteDTO, error) {
	note, found, err := u.repo.SetPinned(ctx, input.ID, input.Pinned)
	if err != nil {
		return NoteDTO{}, err
	}
	if !found {
		return NoteDTO{}, ErrNoteNotFound
	}
	if note.Kind != domainnotes.KindNotice {
		return NoteDTO{}, ErrKindDoesNotSupportPin
	}
	return toDTO(note), nil
}

func (u *SetDoneUseCase) Execute(ctx context.Context, input SetDoneInput) (NoteDTO, error) {
	note, found, err := u.repo.SetDone(ctx, input.ID, input.Done)
	if err != nil {
		return NoteDTO{}, err
	}
	if !found {
		return NoteDTO{}, ErrNoteNotFound
	}
	if note.Kind != domainnotes.KindShopping {
		return NoteDTO{}, ErrKindDoesNotSupportDone
	}
	return toDTO(note), nil
}

func parseKind(raw string) (domainnotes.Kind, error) {
	kind := domainnotes.Kind(raw)
	if kind != domainnotes.KindNotice && kind != domainnotes.KindShopping {
		return "", ErrInvalidKind
	}
	return kind, nil
}

func validateAndNormalizeBody(body string) (string, error) {
	normalized := strings.TrimSpace(body)
	if normalized == "" {
		return "", ErrBodyRequired
	}
	if utf8.RuneCountInString(normalized) > 200 {
		return "", ErrBodyTooLong
	}
	return normalized, nil
}

func normalizeFlags(kind domainnotes.Kind, pinned bool, done bool) (bool, bool) {
	switch kind {
	case domainnotes.KindNotice:
		return pinned, false
	case domainnotes.KindShopping:
		return false, done
	default:
		return false, false
	}
}

func toDTO(note domainnotes.Note) NoteDTO {
	return NoteDTO{
		ID:        note.ID,
		Kind:      string(note.Kind),
		Body:      note.Body,
		Pinned:    note.Pinned,
		Done:      note.Done,
		CreatedAt: note.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: note.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func IsUserError(err error) bool {
	return errors.Is(err, ErrInvalidKind) ||
		errors.Is(err, ErrBodyRequired) ||
		errors.Is(err, ErrBodyTooLong) ||
		errors.Is(err, ErrInvalidLimit) ||
		errors.Is(err, ErrKindDoesNotSupportPin) ||
		errors.Is(err, ErrKindDoesNotSupportDone)
}
