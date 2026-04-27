-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    username        TEXT        NOT NULL UNIQUE,
    email           TEXT        NOT NULL UNIQUE,
    password_hash   TEXT        NOT NULL,

    -- email confirmation (OTP flow lives in email_otps table)
    email_confirmed BOOLEAN     NOT NULL DEFAULT FALSE,
    confirmed_at    TIMESTAMPTZ,                        -- set once on confirmation, never cleared

    -- account status
    is_active       BOOLEAN     NOT NULL DEFAULT TRUE,  -- FALSE = suspended/banned
    last_login_at   TIMESTAMPTZ,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- email_confirmed and confirmed_at must always agree
    CONSTRAINT chk_confirmation_consistent
        CHECK (
            (email_confirmed = FALSE AND confirmed_at IS NULL) OR
            (email_confirmed = TRUE  AND confirmed_at IS NOT NULL)
        )
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
