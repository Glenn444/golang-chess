-- name: SavePushSubscription :exec
INSERT INTO push_subscriptions (user_id, endpoint, p256dh, auth)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id) DO UPDATE
SET endpoint = EXCLUDED.endpoint,
    p256dh   = EXCLUDED.p256dh,
    auth     = EXCLUDED.auth;

-- name: GetPushSubscriptionByUser :one
SELECT endpoint, p256dh, auth
FROM push_subscriptions
WHERE user_id = $1;

-- name: DeletePushSubscriptionByUser :exec
DELETE FROM push_subscriptions
WHERE user_id = $1;

-- name: GetPushSubscriptionExists :one
SELECT id FROM push_subscriptions
WHERE user_id = $1;
