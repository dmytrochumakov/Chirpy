-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: DeleteAllUsers :exec
DELETE FROM users
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserFromRefreshToken :one
SELECT users.* FROM users
INNER JOIN refresh_tokens ON users.id = refresh_tokens.user_id
WHERE refresh_tokens.token = $1
AND refresh_tokens.revoked_at IS NULL
AND refresh_tokens.expires_at > NOW();

-- name: UpdateUserEmailAndPassword :one
UPDATE users
SET email = $1, hashed_password = $2, updated_at = $3
WHERE id=$4
RETURNING *;

-- name: UpdateUserChirpyRedByUserID :one
UPDATE users
SET is_chirpy_red = $1
WHERE id = $2
RETURNING *;