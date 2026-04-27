-- +goose Up
-- +goose StatementBegin
CREATE TYPE player_color AS ENUM ('w', 'b');

CREATE TABLE game_moves (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id         UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    player_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    player_color    player_color NOT NULL,
    move_notation   TEXT NOT NULL,       -- e.g. "e2e3", "e1g1" for castling
    move_number     INT  NOT NULL,       -- full-move number (increments after black moves)
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_game_moves_game_id ON game_moves(game_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS game_moves;
DROP TYPE  IF EXISTS player_color;
-- +goose StatementEnd
