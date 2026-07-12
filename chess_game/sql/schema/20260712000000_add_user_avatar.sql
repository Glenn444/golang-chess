-- +goose Up
-- Profile pictures. A separate table (not a users column) so the many
-- `SELECT * FROM users` queries never carry the image blob. Avatars are
-- re-encoded server-side to a small square JPEG (~10-30KB), so inline
-- bytea beats external storage.
CREATE TABLE user_avatars (
    user_id    UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    image      BYTEA NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE user_avatars;
