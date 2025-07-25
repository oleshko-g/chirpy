// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: chirps.sql

package database

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

const createChirp = `-- name: CreateChirp :one
INSERT INTO
    chirps (
        id,
        created_at,
        updated_at,
        body,
        user_id
    )
VALUES (
        gen_random_uuid(),
        now(),
        now(),
        $1,
        $2
    )
RETURNING
    id, created_at, updated_at, body, user_id, deleted_at
`

type CreateChirpParams struct {
	Body   string
	UserID uuid.UUID
}

func (q *Queries) CreateChirp(ctx context.Context, arg CreateChirpParams) (Chirp, error) {
	row := q.db.QueryRowContext(ctx, createChirp, arg.Body, arg.UserID)
	var i Chirp
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Body,
		&i.UserID,
		&i.DeletedAt,
	)
	return i, err
}

const selectChirp = `-- name: SelectChirp :one
SELECT id, created_at, updated_at, body, user_id, deleted_at FROM chirps WHERE id = $1
`

func (q *Queries) SelectChirp(ctx context.Context, id uuid.UUID) (Chirp, error) {
	row := q.db.QueryRowContext(ctx, selectChirp, id)
	var i Chirp
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Body,
		&i.UserID,
		&i.DeletedAt,
	)
	return i, err
}

const selectChirps = `-- name: SelectChirps :many
SELECT id, created_at, updated_at, body, user_id, deleted_at FROM chirps WHERE deleted_at IS NULL ORDER BY created_at
`

func (q *Queries) SelectChirps(ctx context.Context) ([]Chirp, error) {
	rows, err := q.db.QueryContext(ctx, selectChirps)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Chirp
	for rows.Next() {
		var i Chirp
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Body,
			&i.UserID,
			&i.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const selectChirpsByUserID = `-- name: SelectChirpsByUserID :many
SELECT id, created_at, updated_at, body, user_id, deleted_at
FROM chirps
WHERE
    deleted_at IS NULL
    AND user_id = $1
ORDER BY created_at
`

func (q *Queries) SelectChirpsByUserID(ctx context.Context, userID uuid.UUID) ([]Chirp, error) {
	rows, err := q.db.QueryContext(ctx, selectChirpsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Chirp
	for rows.Next() {
		var i Chirp
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Body,
			&i.UserID,
			&i.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateChirp = `-- name: UpdateChirp :exec
UPDATE chirps
SET
    body = COALESCE($3, body),
    deleted_at = COALESCE($4, deleted_at)
WHERE
    id = $1
    AND user_id = $2
`

type UpdateChirpParams struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Body      string
	DeletedAt sql.NullTime
}

func (q *Queries) UpdateChirp(ctx context.Context, arg UpdateChirpParams) error {
	_, err := q.db.ExecContext(ctx, updateChirp,
		arg.ID,
		arg.UserID,
		arg.Body,
		arg.DeletedAt,
	)
	return err
}
