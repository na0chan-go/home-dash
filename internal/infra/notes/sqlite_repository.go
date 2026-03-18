package notes

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	domainnotes "github.com/na0chan-go/home-dash/internal/domain/notes"
	"github.com/na0chan-go/home-dash/internal/ports"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) List(ctx context.Context, params ports.ListNotesParams) ([]domainnotes.Note, error) {
	query := `
		SELECT id, kind, body, author, pinned, acknowledged, done, created_at, updated_at
		FROM notes
	`
	args := make([]any, 0, 2)
	if params.Kind != nil {
		query += ` WHERE kind = ?`
		args = append(args, string(*params.Kind))
	}

	query += ` ORDER BY ` + buildOrderBy(params.Order)
	query += ` LIMIT ?`
	args = append(args, params.Limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list notes: %w", err)
	}
	defer rows.Close()

	out := make([]domainnotes.Note, 0)
	for rows.Next() {
		note, err := scanNote(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, note)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate notes: %w", err)
	}
	return out, nil
}

func (r *SQLiteRepository) Create(ctx context.Context, params ports.CreateNoteParams) (domainnotes.Note, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO notes(kind, body, author, pinned, done)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, kind, body, author, pinned, acknowledged, done, created_at, updated_at
	`, string(params.Kind), params.Body, params.Author, params.Pinned, params.Done)

	note, err := scanNote(row)
	if err != nil {
		return domainnotes.Note{}, fmt.Errorf("failed to create note: %w", err)
	}
	return note, nil
}

func (r *SQLiteRepository) Delete(ctx context.Context, id int64) (bool, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM notes WHERE id = ?`, id)
	if err != nil {
		return false, fmt.Errorf("failed to delete note: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get delete result: %w", err)
	}
	return affected > 0, nil
}

func (r *SQLiteRepository) SetPinned(ctx context.Context, id int64, pinned bool) (domainnotes.Note, bool, error) {
	var kind string
	if err := r.db.QueryRowContext(ctx, `SELECT kind FROM notes WHERE id = ?`, id).Scan(&kind); err != nil {
		if err == sql.ErrNoRows {
			return domainnotes.Note{}, false, nil
		}
		return domainnotes.Note{}, false, fmt.Errorf("failed to find note for set pin: %w", err)
	}

	if kind != string(domainnotes.KindNotice) {
		note, err := r.findByID(ctx, id)
		if err != nil {
			return domainnotes.Note{}, false, err
		}
		return note, true, nil
	}

	row := r.db.QueryRowContext(ctx, `
		UPDATE notes
		SET pinned = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%SZ','now')
		WHERE id = ?
		RETURNING id, kind, body, author, pinned, acknowledged, done, created_at, updated_at
	`, pinned, id)

	note, err := scanNote(row)
	if err != nil {
		return domainnotes.Note{}, false, fmt.Errorf("failed to set pin: %w", err)
	}
	return note, true, nil
}

func (r *SQLiteRepository) SetAcknowledged(ctx context.Context, id int64, acknowledged bool) (domainnotes.Note, bool, error) {
	var kind string
	if err := r.db.QueryRowContext(ctx, `SELECT kind FROM notes WHERE id = ?`, id).Scan(&kind); err != nil {
		if err == sql.ErrNoRows {
			return domainnotes.Note{}, false, nil
		}
		return domainnotes.Note{}, false, fmt.Errorf("failed to find note for set acknowledged: %w", err)
	}

	if kind != string(domainnotes.KindNotice) {
		note, err := r.findByID(ctx, id)
		if err != nil {
			return domainnotes.Note{}, false, err
		}
		return note, true, nil
	}

	row := r.db.QueryRowContext(ctx, `
		UPDATE notes
		SET acknowledged = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%SZ','now')
		WHERE id = ?
		RETURNING id, kind, body, author, pinned, acknowledged, done, created_at, updated_at
	`, acknowledged, id)

	note, err := scanNote(row)
	if err != nil {
		return domainnotes.Note{}, false, fmt.Errorf("failed to set acknowledged: %w", err)
	}
	return note, true, nil
}

func (r *SQLiteRepository) SetDone(ctx context.Context, id int64, done bool) (domainnotes.Note, bool, error) {
	var kind string
	if err := r.db.QueryRowContext(ctx, `SELECT kind FROM notes WHERE id = ?`, id).Scan(&kind); err != nil {
		if err == sql.ErrNoRows {
			return domainnotes.Note{}, false, nil
		}
		return domainnotes.Note{}, false, fmt.Errorf("failed to find note for set done: %w", err)
	}

	if kind != string(domainnotes.KindShopping) {
		note, err := r.findByID(ctx, id)
		if err != nil {
			return domainnotes.Note{}, false, err
		}
		return note, true, nil
	}

	row := r.db.QueryRowContext(ctx, `
		UPDATE notes
		SET done = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%SZ','now')
		WHERE id = ?
		RETURNING id, kind, body, author, pinned, acknowledged, done, created_at, updated_at
	`, done, id)

	note, err := scanNote(row)
	if err != nil {
		return domainnotes.Note{}, false, fmt.Errorf("failed to set done: %w", err)
	}
	return note, true, nil
}

func (r *SQLiteRepository) findByID(ctx context.Context, id int64) (domainnotes.Note, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, kind, body, author, pinned, acknowledged, done, created_at, updated_at
		FROM notes
		WHERE id = ?
	`, id)
	note, err := scanNote(row)
	if err != nil {
		return domainnotes.Note{}, fmt.Errorf("failed to find note: %w", err)
	}
	return note, nil
}

func buildOrderBy(order ports.NoteOrder) string {
	switch order {
	case ports.NoteOrderNotice:
		return "pinned DESC, acknowledged ASC, created_at DESC"
	case ports.NoteOrderShopping:
		return "done ASC, created_at DESC"
	default:
		return "created_at DESC"
	}
}

type noteScanner interface {
	Scan(dest ...any) error
}

func scanNote(scanner noteScanner) (domainnotes.Note, error) {
	var rawKind string
	var createdAtRaw string
	var updatedAtRaw string
	var note domainnotes.Note

	if err := scanner.Scan(
		&note.ID,
		&rawKind,
		&note.Body,
		&note.Author,
		&note.Pinned,
		&note.Acknowledged,
		&note.Done,
		&createdAtRaw,
		&updatedAtRaw,
	); err != nil {
		return domainnotes.Note{}, err
	}

	note.Kind = domainnotes.Kind(strings.ToLower(rawKind))
	createdAt, err := time.Parse(time.RFC3339, createdAtRaw)
	if err != nil {
		return domainnotes.Note{}, fmt.Errorf("invalid created_at format: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339, updatedAtRaw)
	if err != nil {
		return domainnotes.Note{}, fmt.Errorf("invalid updated_at format: %w", err)
	}
	note.CreatedAt = createdAt
	note.UpdatedAt = updatedAt
	return note, nil
}
