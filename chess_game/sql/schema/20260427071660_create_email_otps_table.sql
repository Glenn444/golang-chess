-- +goose Up
-- +goose StatementBegin

-- Stores 6-digit numeric OTPs for email confirmation (and future password-reset flows).
--
-- Security contract (enforced in application layer):
--   1. Generate with crypto/rand, NOT math/rand.
--   2. Store HMAC-SHA256(secret_key, otp_code) — never the plain code.
--   3. Expire after 15–30 minutes (expires_at).
--   4. Allow at most 5 verification attempts (attempts < 5) before the code is
--      considered burnt; issue a fresh one after that.
--   5. Call MarkOTPUsed immediately on success to prevent replay.
--   6. Call InvalidateUserOTPs before issuing a new code so only one live OTP
--      exists per user at a time.
CREATE TABLE email_otps (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash   TEXT        NOT NULL,                  -- HMAC-SHA256(server_secret, "847291")
    expires_at  TIMESTAMPTZ NOT NULL,                  -- typically NOW() + 15 min
    attempts    SMALLINT    NOT NULL DEFAULT 0,         -- incremented on each wrong guess
    used_at     TIMESTAMPTZ,                           -- NULL = not yet consumed
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_attempts_non_negative CHECK (attempts >= 0)
);

CREATE INDEX idx_email_otps_user_id   ON email_otps(user_id);
CREATE INDEX idx_email_otps_expires_at ON email_otps(expires_at); -- for periodic cleanup

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS email_otps;
-- +goose StatementEnd
