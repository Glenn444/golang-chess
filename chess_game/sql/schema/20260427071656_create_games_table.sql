-- +goose Up
-- +goose StatementBegin
CREATE TYPE game_state AS ENUM (
    'waiting',
    'active',
    'checkmate',
    'stalemate',
    'resign',
    'draw',
    'abandoned'
);

CREATE TABLE games (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    white_player_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    black_player_id  UUID          REFERENCES users(id) ON DELETE SET NULL,
    state            game_state NOT NULL DEFAULT 'waiting',
    in_check         BOOLEAN NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_games_white_player ON games(white_player_id);
CREATE INDEX idx_games_black_player ON games(black_player_id);
CREATE INDEX idx_games_state        ON games(state);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS games;
DROP TYPE  IF EXISTS game_state;
-- +goose StatementEnd
