-- +goose NO TRANSACTION
-- +goose Up
-- Games that end on the clock were being stored as 'checkmate' because the
-- enum had no timeout value. ALTER TYPE ... ADD VALUE cannot run inside a
-- transaction on older PostgreSQL versions, hence NO TRANSACTION.
ALTER TYPE game_state ADD VALUE IF NOT EXISTS 'timeout';

-- +goose Down
-- Removing an enum value is not supported by PostgreSQL; no-op.
SELECT 1;
