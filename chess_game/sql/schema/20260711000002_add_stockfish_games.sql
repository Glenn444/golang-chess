-- +goose Up
-- +goose StatementBegin

-- Persist the opponent type so engine games can be recognised after a restart
-- and kept out of the joinable lobby. stockfish_level maps to the UCI
-- "Skill Level" option (0-20).
ALTER TABLE games
    ADD COLUMN opponent VARCHAR(20) NOT NULL DEFAULT 'person',
    ADD COLUMN stockfish_level SMALLINT NOT NULL DEFAULT 0;

ALTER TABLE games
    ADD CONSTRAINT chk_games_opponent CHECK (opponent IN ('person', 'stockfish')),
    ADD CONSTRAINT chk_games_stockfish_level CHECK (stockfish_level BETWEEN 0 AND 20);

-- Engine moves have no user behind them.
ALTER TABLE game_moves ALTER COLUMN player_id DROP NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE game_moves ALTER COLUMN player_id SET NOT NULL;
ALTER TABLE games
    DROP CONSTRAINT IF EXISTS chk_games_stockfish_level,
    DROP CONSTRAINT IF EXISTS chk_games_opponent;
ALTER TABLE games
    DROP COLUMN IF EXISTS stockfish_level,
    DROP COLUMN IF EXISTS opponent;
-- +goose StatementEnd
