# Chess Game

A multiplayer chess game server written in Go. Two players connect over WebSockets, play a full game of chess, and can communicate via live **text chat** or **voice calls** (WebRTC) during the match. Move validation is handled server-side; the Stockfish engine is available for single-player analysis.

## Features

- Full chess rules: legal-move validation, check detection, castling, en-passant
- Real-time WebSocket hub (melody) for move broadcasting and chat
- WebRTC voice signalling relayed through the WebSocket hub
- PostgreSQL persistence via sqlc-generated type-safe Go code
- Goose migrations for schema versioning

---

## Database Schema

### Entity Relationship

```
users ──< games (as white_player)
users ──< games (as black_player)
games ──< game_moves
games ──< chat_messages
games ──< voice_sessions
users ──< game_moves    (via player_id)
users ──< chat_messages (via sender_id)
users ──< voice_sessions (via initiator_id)
users ──< email_otps    (via user_id)
```

### Tables

#### `users`
Stores registered player accounts. New accounts have `email_confirmed = FALSE` until the OTP flow (see `email_otps`) completes.

| Column | Type | Notes |
|---|---|---|
| `id` | UUID | Primary key |
| `username` | TEXT | Unique display name |
| `email` | TEXT | Unique, used for login |
| `password_hash` | TEXT | bcrypt hash |
| `email_confirmed` | BOOLEAN | `FALSE` until OTP verified |
| `confirmed_at` | TIMESTAMPTZ | Set once on confirmation, never cleared |
| `is_active` | BOOLEAN | `FALSE` = suspended/banned |
| `last_login_at` | TIMESTAMPTZ | Updated on every successful login |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

A `CHECK` constraint enforces that `email_confirmed` and `confirmed_at` are always in sync.

---

#### `email_otps`
6-digit numeric OTPs for email confirmation. OTP codes never touch this table in plain form — the application stores `HMAC-SHA256(server_secret, "847291")` and re-hashes the user-supplied code before comparing.

**OTP flow:**
1. On registration: call `InvalidateUserOTPs` then `CreateEmailOTP` (expire in 15–30 min).
2. Rate-gate generation with `CountRecentOTPsForUser` (≤ 5 codes per hour).
3. On each guess: call `GetValidOTP` to fetch the current code, compare hashes.
4. Wrong guess → `IncrementOTPAttempts`; lock out after 5 attempts.
5. Correct guess → `MarkOTPUsed` (prevents replay) then `ConfirmEmail` on the user row.
6. Periodic cleanup via `DeleteExpiredOTPs`.

| Column | Type | Notes |
|---|---|---|
| `id` | UUID | Primary key |
| `user_id` | UUID | FK → users |
| `code_hash` | TEXT | `HMAC-SHA256(server_secret, otp_digits)` |
| `expires_at` | TIMESTAMPTZ | Typically `NOW() + 15 min` |
| `attempts` | SMALLINT | Incremented on each wrong guess; max 5 |
| `used_at` | TIMESTAMPTZ | `NULL` = not yet consumed |
| `created_at` | TIMESTAMPTZ | |

---

#### `games`
One row per match. Mirrors the `Game` struct in the WebSocket hub.

| Column | Type | Notes |
|---|---|---|
| `id` | UUID | Primary key |
| `white_player_id` | UUID | FK → users; set on game creation |
| `black_player_id` | UUID | FK → users; set when opponent joins (nullable) |
| `state` | `game_state` enum | `waiting` → `active` → terminal |
| `in_check` | BOOLEAN | Whether the side-to-move king is in check |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

**`game_state` enum values:** `waiting`, `active`, `checkmate`, `stalemate`, `resign`, `draw`, `abandoned`

---

#### `game_moves`
Ordered move log for every game. Mirrors the `Move` struct.

| Column | Type | Notes |
|---|---|---|
| `id` | UUID | Primary key |
| `game_id` | UUID | FK → games |
| `player_id` | UUID | FK → users |
| `player_color` | `player_color` enum | `w` or `b` |
| `move_notation` | TEXT | Coordinate notation, e.g. `e2e3`, `e1g1` |
| `move_number` | INT | Full-move counter (increments after black moves) |
| `created_at` | TIMESTAMPTZ | |

---

#### `chat_messages`
Live text chat between the two players during a game.

| Column | Type | Notes |
|---|---|---|
| `id` | UUID | Primary key |
| `game_id` | UUID | FK → games |
| `sender_id` | UUID | FK → users |
| `content` | TEXT | 1–2 000 characters |
| `created_at` | TIMESTAMPTZ | |

---

#### `voice_sessions`
Tracks the lifetime of a WebRTC voice call per game. Signalling payloads (offer/answer/ICE candidates) travel over the WebSocket hub and are **not** stored here.

| Column | Type | Notes |
|---|---|---|
| `id` | UUID | Primary key |
| `game_id` | UUID | FK → games |
| `initiator_id` | UUID | FK → users; the player who started the call |
| `state` | `voice_session_state` enum | `pending` → `active` → `ended` |
| `started_at` | TIMESTAMPTZ | |
| `ended_at` | TIMESTAMPTZ | Nullable; set when call ends |

**`voice_session_state` enum values:** `pending`, `active`, `ended`

---

---

## API Routes

### Public

| Method | Path | Handler | Description |
|---|---|---|---|
| `GET` | `/` | `welcome` | Health check |
| `GET` | `/users/check-username?username=` | `checkUsernameExists` | Availability check before signup |
| `POST` | `/users/signup` | `createUser` | Register; sends 6-digit OTP to email |
| `POST` | `/users/confirm-email` | `confirmEmail` | Verify OTP → activates account |
| `POST` | `/users/send-emailotp` | `sendEmailOTP` | Resend OTP (rate-limited) |
| `POST` | `/users/signin` | `loginUser` | Returns `access_token` + `refresh_token` |
| `POST` | `/users/refresh-token` | `refreshToken` | Issue new access token from refresh token |

### Protected — `Authorization: Bearer <access_token>`

#### Profile
| Method | Path | Handler | Description |
|---|---|---|---|
| `GET` | `/users/me` | `getMe` | Current user's profile (no sensitive fields) |

#### Games
| Method | Path | Handler | Description |
|---|---|---|---|
| `POST` | `/games` | `createGame` | Create a new game; caller becomes white player |
| `GET` | `/games` | `listWaitingGames` | List open games waiting for a second player |
| `GET` | `/games/mine` | `listMyGames` | All games the caller is playing in |
| `GET` | `/games/:id` | `getGame` | Get a single game by ID |
| `POST` | `/games/:id/join` | `joinGame` | Join a waiting game as black player |
| `POST` | `/games/:id/resign` | `resignGame` | Resign from an active game |
| `GET` | `/games/:id/moves` | `getGameMoves` | Full move history for a game |

#### Chat
| Method | Path | Handler | Description |
|---|---|---|---|
| `POST` | `/games/:id/chat` | `sendChatMessage` | Send a text message (also available via WebSocket) |
| `GET` | `/games/:id/chat` | `getChatMessages` | Retrieve full chat history for a game |

#### Voice (WebRTC session lifecycle)
| Method | Path | Handler | Description |
|---|---|---|---|
| `POST` | `/games/:id/voice` | `startVoiceSession` | Initiate a call; only one session per game at a time |
| `GET` | `/games/:id/voice` | `getActiveVoiceSession` | Get the current pending/active session |
| `PATCH` | `/games/:id/voice/:vid/activate` | `activateVoiceSession` | Recipient accepts the call |
| `DELETE` | `/games/:id/voice/:vid` | `endVoiceSession` | Either player hangs up |

### WebSocket — `GET /ws`

Upgrade a connection to WebSocket. Auth and room selection via query params:

```
/ws?token=<access_token>&game_id=<uuid>
```

The server verifies the token and confirms the caller is a player in the given game before upgrading. All subsequent messages use the JSON envelope:

```json
{ "type": "<event>", "payload": { ... } }
```

| Event type | Direction | Description |
|---|---|---|
| `make_move` | client → server → both | Broadcast a move to both players |
| `chat` | client → server → both | Persist + broadcast a chat message |
| `voice_offer` | client → server → other | Relay WebRTC offer to the opponent |
| `voice_answer` | client → server → other | Relay WebRTC answer |
| `voice_ice` | client → server → other | Relay ICE candidate |
| `voice_end` | client → server → other | Signal call termination |
| `error` | server → client | Error response |

> **Note:** `make_move` validation against the board package is wired but not yet fully implemented — see `wsHandleMove` in [internal/api/ws.go](internal/api/ws.go).

---

## Project Structure

```
.
├── cmd/server/         # Server entry point
├── internal/
│   ├── board/          # Board state, move execution, check detection
│   ├── pieces/         # Per-piece legal-move generators
│   ├── ws/             # WebSocket hub, game/player structs, event types
│   ├── stockfish/      # Stockfish engine wrapper
│   ├── utils/          # Notation helpers
│   └── db/             # sqlc-generated database layer
├── sql/
│   ├── schema/         # Goose migration files
│   └── queries/        # sqlc query files
├── sqlc.yaml
└── Makefile
```

## Getting Started

```bash
# Start Postgres
make postgres

# Create the database
make createdb

# Run migrations
make migrateup

# Generate Go DB code
make sqlc

# Run the server
make server
```
