-- name: CreateGameAsWhite :one
INSERT INTO games (white_player_id)
VALUES ($1)
RETURNING *;

-- name: CreateGameAsBlack :one
INSERT INTO games (black_player_id)
VALUES ($1)
RETURNING *;

-- name: GetGameByID :one
SELECT * FROM games
WHERE id = $1;

-- name: GetGamesByPlayerID :many
SELECT * FROM games
WHERE white_player_id = $1
   OR black_player_id = $1
ORDER BY created_at DESC;

-- name: ListWaitingGames :many
SELECT * FROM games
WHERE state = 'waiting'
ORDER BY created_at ASC;

-- name: JoinGameAsBlack :one
UPDATE games
SET
    black_player_id = $2,
    state           = 'active',
    updated_at      = NOW()
WHERE id = $1
  AND state = 'waiting'
  AND black_player_id IS NULL
RETURNING *;

-- name: JoinGameAsWhite :one
UPDATE games
SET
    white_player_id = $2,
    state           = 'active',
    updated_at      = NOW()
WHERE id = $1
  AND state = 'waiting'
  AND white_player_id IS NULL
RETURNING *;

-- name: UpdateGameState :one
UPDATE games
SET
    state      = $2,
    in_check   = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteGame :exec
DELETE FROM games
WHERE id = $1;
