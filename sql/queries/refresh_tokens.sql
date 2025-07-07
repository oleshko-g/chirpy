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