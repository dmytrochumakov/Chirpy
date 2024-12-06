// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: refresh_tokens.sql

package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const createRefreshToken = `-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens(
    token,
    created_at,
    updated_at,
    user_id,
    expires_at,
    revoked_at
)
VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING token, created_at, updated_at, user_id, expires_at, revoked_at
`

type CreateRefreshTokenParams struct {
	Token     string
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    uuid.UUID
	ExpiresAt time.Time
	RevokedAt sql.NullTime
}

func (q *Queries) CreateRefreshToken(ctx context.Context, arg CreateRefreshTokenParams) (RefreshToken, error) {
	row := q.db.QueryRowContext(ctx, createRefreshToken,
		arg.Token,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.UserID,
		arg.ExpiresAt,
		arg.RevokedAt,
	)
	var i RefreshToken
	err := row.Scan(
		&i.Token,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UserID,
		&i.ExpiresAt,
		&i.RevokedAt,
	)
	return i, err
}

const revokeRefreshToken = `-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = $1, updated_at = $2
WHERE user_id=$3
`

type RevokeRefreshTokenParams struct {
	RevokedAt sql.NullTime
	UpdatedAt time.Time
	UserID    uuid.UUID
}

func (q *Queries) RevokeRefreshToken(ctx context.Context, arg RevokeRefreshTokenParams) error {
	_, err := q.db.ExecContext(ctx, revokeRefreshToken, arg.RevokedAt, arg.UpdatedAt, arg.UserID)
	return err
}
