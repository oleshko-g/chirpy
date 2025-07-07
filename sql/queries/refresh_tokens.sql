-- name: InsertRefreshToken :exec
INSERT INTO
    refresh_tokens (
        token,
        created_at,
        updated_at,
        user_id,
        expires_at
    )
VALUES ($1, NOW(), NOW(), $2, $3);

-- name: SelectRefreshToken :one
SELECT * FROM refresh_tokens WHERE token = $1;

-- name: UpdateRefreshToken :exec
UPDATE refresh_tokens
SET
    updated_at = now(),
    revoked_at = $2
WHERE
    token = $1;