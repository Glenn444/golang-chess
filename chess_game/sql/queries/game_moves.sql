-- name: CreateMove :one
INSERT INTO game_moves (game_id, player_id, player_color, move_notation, move_number)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetMovesByGameID :many
SELECT * FROM game_moves
WHERE game_id = $1
ORDER BY move_number ASC, created_at ASC;

-- name: GetLastMoveByGameID :one
SELECT * FROM game_moves
WHERE game_id = $1
ORDER BY move_number DESC, created_at DESC
LIMIT 1;

-- name: CountMovesByGameID :one
SELECT COUNT(*) FROM game_moves
WHERE game_id = $1;

-- name: DeleteMovesByGameID :exec
DELETE FROM game_moves
WHERE game_id = $1;
