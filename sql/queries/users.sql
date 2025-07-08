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

-- name: UpdateUser :exec
UPDATE users
SET
    updated_at = now(),
    email = $2,
    hashed_password = $3
WHERE
    id = $1;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: SelectUserByEmail :one
SELECT * FROM users WHERE email = $1;