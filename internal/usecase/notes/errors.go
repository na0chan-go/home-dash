package notes

import "errors"

var (
	ErrInvalidKind            = errors.New("invalid kind")
	ErrBodyRequired           = errors.New("body is required")
	ErrBodyTooLong            = errors.New("body must be 200 characters or less")
	ErrAuthorTooLong          = errors.New("author must be 20 characters or less")
	ErrInvalidLimit           = errors.New("limit must be between 1 and 100")
	ErrNoteNotFound           = errors.New("note not found")
	ErrKindDoesNotSupportPin  = errors.New("kind does not support pin")
	ErrKindDoesNotSupportDone = errors.New("kind does not support done")
)
