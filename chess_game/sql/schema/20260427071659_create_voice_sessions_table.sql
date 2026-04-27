-- +goose Up
-- +goose StatementBegin
CREATE TYPE voice_session_state AS ENUM ('pending', 'active', 'ended');

-- Tracks WebRTC voice call lifetime per game. Signaling payloads
-- (offer/answer/ICE candidates) are exchanged over the WebSocket hub
-- and are NOT persisted here — only session state is stored.
CREATE TABLE voice_sessions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id      UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    initiator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    state        voice_session_state NOT NULL DEFAULT 'pending',
    started_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at     TIMESTAMPTZ
);

CREATE INDEX idx_voice_sessions_game_id ON voice_sessions(game_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS voice_sessions;
DROP TYPE  IF EXISTS voice_session_state;
-- +goose StatementEnd
