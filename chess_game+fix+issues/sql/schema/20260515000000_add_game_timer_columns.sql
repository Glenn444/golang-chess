-- +goose Up
-- +goose StatementBegin
ALTER TABLE games
    ADD COLUMN white_time_remaining_ms BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN black_time_remaining_ms BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN last_move_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE games
    DROP COLUMN IF EXISTS white_time_remaining_ms,
    DROP COLUMN IF EXISTS black_time_remaining_ms,
    DROP COLUMN IF EXISTS last_move_at;
-- +goose StatementEnd
