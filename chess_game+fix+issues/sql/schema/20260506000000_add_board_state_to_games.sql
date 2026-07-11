-- +goose Up
-- +goose StatementBegin
ALTER TABLE games ADD COLUMN board_state TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE games DROP COLUMN IF EXISTS board_state;
-- +goose StatementEnd
