-- name: CreateEmailOTP :one
INSERT INTO email_otps (user_id, code_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetValidOTP :one
-- Returns the most-recent, unexpired, unused code that still has attempts left.
-- The app layer must hash the user-supplied digit string before comparing code_hash.
SELECT * FROM email_otps
WHERE user_id  = $1
  AND expires_at > NOW()
  AND used_at  IS NULL
  AND attempts < 5
ORDER BY created_at DESC
LIMIT 1;

-- name: IncrementOTPAttempts :one
-- Call this on every failed verification attempt.
UPDATE email_otps
SET attempts = attempts + 1
WHERE id = $1
RETURNING *;

-- name: MarkOTPUsed :one
-- Call this immediately after a successful match to prevent replay.
UPDATE email_otps
SET used_at = NOW()
WHERE id      = $1
  AND used_at IS NULL
RETURNING *;

-- name: InvalidateUserOTPs :exec
-- Burn all live codes for a user before issuing a new one.
UPDATE email_otps
SET used_at = NOW()
WHERE user_id = $1
  AND used_at IS NULL;

-- name: CountRecentOTPsForUser :one
-- Rate-limit OTP generation: reject if a code was issued in the last 5 minutes.
SELECT COUNT(*) FROM email_otps
WHERE user_id    = $1
  AND created_at > NOW() - INTERVAL '5 minutes';

-- name: DeleteExpiredOTPs :exec
-- Run periodically (e.g. a cron job) to keep the table small.
DELETE FROM email_otps
WHERE expires_at < NOW();
