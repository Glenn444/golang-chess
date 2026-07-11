-- +goose Up
-- +goose StatementBegin
ALTER TABLE games
    ADD COLUMN ended_by_player_id UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN end_reason VARCHAR(50);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE games
    DROP COLUMN IF EXISTS ended_by_player_id,
    DROP COLUMN IF EXISTS end_reason;
-- +goose StatementEnd
