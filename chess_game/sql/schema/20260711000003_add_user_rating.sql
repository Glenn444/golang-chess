-- +goose Up
-- +goose StatementBegin
-- Elo rating, updated after every finished person-vs-person game.
ALTER TABLE users ADD COLUMN rating INT NOT NULL DEFAULT 1200;
CREATE INDEX idx_users_rating ON users(rating DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_rating;
ALTER TABLE users DROP COLUMN IF EXISTS rating;
-- +goose StatementEnd
