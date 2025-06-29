-- name: InsetUser :one
INSERT INTO
    users (
        id,
        created_at,
        updated_at,
        email,
        hashed_password
    )
VALUES (
        gen_random_uuid (),
        now(),
        now(),
        $1,
        $2
    )
RETURNING
        id,
        created_at,
        updated_at,
        email
    ;

-- name: ResetUsers :exec
DELETE FROM users;