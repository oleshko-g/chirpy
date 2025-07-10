-- name: CreateChirp :one
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
    *;
-- name: UpdateChirp :exec
UPDATE chirps
SET
    body = COALESCE($3, body),
    deleted_at = COALESCE($4, deleted_at)
WHERE
    id = $1
    AND user_id = $2;

-- name: SelectChirps :many
SELECT * FROM chirps WHERE deleted_at IS NULL ORDER BY created_at;

-- name: SelectChirpsByUserID :many
SELECT *
FROM chirps
WHERE
    deleted_at IS NULL
    AND user_id = $1
ORDER BY created_at;

-- name: SelectChirp :one
SELECT * FROM chirps WHERE id = $1;