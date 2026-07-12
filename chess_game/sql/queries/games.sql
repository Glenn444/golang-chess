-- name: CreateGameAsWhite :one
INSERT INTO games (white_player_id, visibility, board_state, white_time_remaining_ms, black_time_remaining_ms, opponent, stockfish_level, state, last_move_at)
VALUES ($1, $2, $3, $4, $4, $5, $6, $7, NOW())
RETURNING *;

-- name: CreateGameAsBlack :one
INSERT INTO games (black_player_id, visibility, board_state, white_time_remaining_ms, black_time_remaining_ms, opponent, stockfish_level, state, last_move_at)
VALUES ($1, $2, $3, $4, $4, $5, $6, $7, NOW())
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
WHERE state = 'waiting' AND opponent = 'person'
ORDER BY created_at ASC;

-- name: ListPublicGames :many
SELECT * FROM games
WHERE state = 'waiting' AND visibility = 'public' AND opponent = 'person'
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

-- name: ListLiveGames :many
SELECT g.id, g.move_count, g.current_player,
       g.white_time_remaining_ms, g.black_time_remaining_ms,
       g.created_at, g.updated_at,
       wu.username AS white_username, wu.rating AS white_rating,
       bu.username AS black_username, bu.rating AS black_rating
FROM games g
JOIN users wu ON wu.id = g.white_player_id
JOIN users bu ON bu.id = g.black_player_id
WHERE g.state = 'active' AND g.visibility = 'public' AND g.opponent = 'person'
ORDER BY g.updated_at DESC
LIMIT 50;

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
