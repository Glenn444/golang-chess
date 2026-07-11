-- name: CreateGameAsWhite :one
INSERT INTO games (white_player_id, visibility, board_state, white_time_remaining_ms, black_time_remaining_ms, last_move_at)
VALUES ($1, $2, $3, $4, $4, NOW())
RETURNING *;

-- name: CreateGameAsBlack :one
INSERT INTO games (black_player_id, visibility, board_state, white_time_remaining_ms, black_time_remaining_ms, last_move_at)
VALUES ($1, $2, $3, $4, $4, NOW())
RETURNING *;

-- name: GetGameByID :one
SELECT * FROM games
WHERE id = $1;

-- name: GetGamesByPlayerID :many
SELECT * FROM games
WHERE white_player_id = $1
   OR black_player_id = $1
ORDER BY created_at DESC;

-- name: GetActiveGamesByUser :many
SELECT * FROM games 
WHERE (white_player_id = $1 OR black_player_id = $1)
AND state IN ('waiting', 'active');

-- name: ListWaitingGames :many
SELECT * FROM games
WHERE state = 'waiting'
ORDER BY created_at ASC;

-- name: ListPublicGames :many
SELECT * FROM games
WHERE state = 'waiting' AND visibility = 'public'
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
    state                  = $2,
    in_check               = $3,
    current_player         = $4,
    move_count             = $5,
    board_state            = $6,
    white_time_remaining_ms = $7,
    black_time_remaining_ms = $8,
    ended_by_player_id     = $9,
    end_reason             = $10,
    last_move_at           = $11,
    updated_at             = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteGame :exec
DELETE FROM games
WHERE id = $1;
