-- +goose Up
-- +goose StatementBegin
CREATE TYPE player_color AS ENUM ('w', 'b');

CREATE TYPE game_state AS ENUM (
    'waiting',
    'active',
    'checkmate',
    'stalemate',
    'resign',
    'draw',
    'abandoned'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TYPE IF EXISTS player_color;
DROP TYPE IF EXISTS game_state;
-- +goose StatementEnd