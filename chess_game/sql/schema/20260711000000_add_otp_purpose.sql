-- +goose Up
-- +goose StatementBegin

-- An OTP issued to confirm an email address must not be usable to reset a
-- password (and vice versa). Existing rows are all email-confirmation codes.
ALTER TABLE email_otps
    ADD COLUMN purpose VARCHAR(20) NOT NULL DEFAULT 'confirm_email';

ALTER TABLE email_otps
    ADD CONSTRAINT chk_otp_purpose CHECK (purpose IN ('confirm_email', 'password_reset'));

CREATE INDEX idx_email_otps_user_purpose ON email_otps(user_id, purpose);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_email_otps_user_purpose;
ALTER TABLE email_otps DROP CONSTRAINT IF EXISTS chk_otp_purpose;
ALTER TABLE email_otps DROP COLUMN IF EXISTS purpose;
-- +goose StatementEnd
