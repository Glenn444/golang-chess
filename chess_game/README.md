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
users ──< game_moves   (via player_id)
users ──< chat_messages (via sender_id)
users ──< voice_sessions (via initiator_id)
```

### Tables

#### `users`
Stores registered player accounts.

| Column | Type | Notes |
|---|---|---|
| `id` | UUID | Primary key |
| `username` | TEXT | Unique display name |
| `email` | TEXT | Unique, used for login |
| `password_hash` | TEXT | bcrypt hash |
| `created_at` | TIMESTAMPTZ | |
| `updated_at` | TIMESTAMPTZ | |

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
