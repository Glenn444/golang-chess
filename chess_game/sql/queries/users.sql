-- name: CreateUser :one
INSERT INTO users (username, email, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1;

-- name: UsernameExists :one
SELECT EXISTS (
    SELECT 1 FROM users WHERE username = $1
);

-- name: ConfirmEmail :one
UPDATE users
SET
    email_confirmed = TRUE,
    confirmed_at    = NOW(),
    updated_at      = NOW()
WHERE id = $1
  AND email_confirmed = FALSE
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET
    username      = COALESCE(sqlc.narg(username), username),
    email         = COALESCE(sqlc.narg(email), email),
    password_hash = COALESCE(sqlc.narg(password_hash), password_hash),
    updated_at    = NOW()
WHERE id = $1
RETURNING *;

-- name: SetLastLogin :exec
UPDATE users
SET last_login_at = NOW()
WHERE id = $1;

-- name: DeactivateUser :exec
UPDATE users
SET is_active = FALSE, updated_at = NOW()
WHERE id = $1;

-- name: ActivateUser :exec
UPDATE users
SET is_active = TRUE, updated_at = NOW()
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
