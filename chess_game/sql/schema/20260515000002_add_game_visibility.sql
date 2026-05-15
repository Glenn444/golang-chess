-- +goose Up
-- +goose StatementBegin
ALTER TABLE games ADD COLUMN visibility VARCHAR(10) NOT NULL DEFAULT 'public';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE games DROP COLUMN IF EXISTS visibility;
-- +goose StatementEnd
