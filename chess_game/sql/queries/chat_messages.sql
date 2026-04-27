-- name: CreateChatMessage :one
INSERT INTO chat_messages (game_id, sender_id, content)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetChatMessagesByGameID :many
SELECT * FROM chat_messages
WHERE game_id = $1
ORDER BY created_at ASC;

-- name: GetChatMessagesByGameIDPaginated :many
SELECT * FROM chat_messages
WHERE game_id = $1
  AND created_at < sqlc.arg(before)::timestamptz
ORDER BY created_at DESC
LIMIT sqlc.arg(page_size)::int;

-- name: DeleteChatMessagesByGameID :exec
DELETE FROM chat_messages
WHERE game_id = $1;
