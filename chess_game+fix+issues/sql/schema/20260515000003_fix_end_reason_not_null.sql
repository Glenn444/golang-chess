-- +goose Up
-- +goose StatementBegin
UPDATE games SET end_reason = '' WHERE end_reason IS NULL;
ALTER TABLE games ALTER COLUMN end_reason SET NOT NULL;
ALTER TABLE games ALTER COLUMN end_reason SET DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE games ALTER COLUMN end_reason DROP NOT NULL;
ALTER TABLE games ALTER COLUMN end_reason DROP DEFAULT;
-- +goose StatementEnd
