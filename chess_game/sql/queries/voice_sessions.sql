-- name: CreateVoiceSession :one
INSERT INTO voice_sessions (game_id, initiator_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetVoiceSessionByID :one
SELECT * FROM voice_sessions
WHERE id = $1;

-- name: GetActiveVoiceSessionByGameID :one
SELECT * FROM voice_sessions
WHERE game_id = $1
  AND state IN ('pending', 'active')
ORDER BY started_at DESC
LIMIT 1;

-- name: ActivateVoiceSession :one
UPDATE voice_sessions
SET state = 'active'
WHERE id = $1
  AND state = 'pending'
RETURNING *;

-- name: EndVoiceSession :one
UPDATE voice_sessions
SET
    state    = 'ended',
    ended_at = NOW()
WHERE id = $1
  AND state != 'ended'
RETURNING *;

-- name: GetVoiceSessionsByGameID :many
SELECT * FROM voice_sessions
WHERE game_id = $1
ORDER BY started_at DESC;
