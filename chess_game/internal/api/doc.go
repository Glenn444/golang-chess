// Package api implements the Chess Game HTTP and WebSocket server.
//
//     Schemes: http, https
//     Host: localhost:8080
//     BasePath: /
//     Version: 1.0.0
//     License: MIT
//     Contact: Chess Game API<info@chess-game.local>
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
//     SecurityDefinitions:
//       bearer:
//         type: apiKey
//         name: Authorization
//         in: header
//
// swagger:meta
package api

// ── Users ───────────────────────────────────────────────────────────────────────

// swagger:route POST /users/signup Auth createUser
// Register a new user. Sends an OTP to the provided email address.
// responses:
//   200: description:User created, OTP sent
//   400: description:Invalid request
//   409: description:Username or email already taken
//   500: description:Internal server error

// swagger:route POST /users/signin Auth loginUser
// Sign in with email and password. Returns access and refresh tokens.
// responses:
//   200: description:Login successful, returns tokens
//   400: description:Invalid request
//   401: description:Invalid credentials
//   404: description:User not found
//   500: description:Internal server error

// swagger:route POST /users/confirm-email Auth confirmEmail
// Confirm email address with a 6-digit OTP code.
// responses:
//   200: description:Email verified
//   400: description:Invalid request
//   403: description:Invalid OTP
//   404: description:User or OTP not found
//   409: description:Email already verified
//   500: description:Internal server error

// swagger:route POST /users/send-emailotp Auth sendEmailOTP
// Resend the OTP verification code to email.
// responses:
//   200: description:OTP sent
//   400: description:Invalid request
//   403: description:Cooldown active
//   404: description:User not found
//   409: description:Email already verified
//   500: description:Internal server error

// swagger:route POST /users/refresh-token Auth refreshToken
// Issue a new access token using a refresh token.
// responses:
//   200: description:New access token issued
//   400: description:Invalid request
//   401: description:Invalid or expired token
//   500: description:Internal server error

// swagger:route GET /users/check-username Users checkUsernameExists
// Check whether a username is available.
// responses:
//   200: description:Username availability
//   400: description:Invalid request
//   500: description:Internal server error

// swagger:route GET /users/me Users getMe
// Get the current authenticated user's profile.
// Security: bearer
// responses:
//   200: description:User profile
//   401: description:Unauthorized
//   500: description:Internal server error

// ── Games ───────────────────────────────────────────────────────────────────────

// swagger:route POST /games Games createGame
// Create a new chess game.
// Security: bearer
// responses:
//   201: description:Game created
//   400: description:Invalid request
//   401: description:Unauthorized
//   500: description:Internal server error

// swagger:route GET /games Games listWaitingGames
// List all games waiting for a second player.
// Security: bearer
// responses:
//   200: description:List of waiting games
//   401: description:Unauthorized
//   500: description:Internal server error

// swagger:route GET /games/mine Games listMyGames
// List the current user's games.
// Security: bearer
// responses:
//   200: description:List of user's games
//   401: description:Unauthorized
//   500: description:Internal server error

// swagger:route GET /games/{id} Games getGame
// Get a game by ID.
// Security: bearer
// responses:
//   200: description:Game details
//   400: description:Invalid game ID
//   401: description:Unauthorized
//   404: description:Game not found
//   500: description:Internal server error

// swagger:route POST /games/{id}/join Games joinGame
// Join a waiting game as the second player.
// Security: bearer
// responses:
//   200: description:Joined successfully
//   400: description:Invalid request
//   401: description:Unauthorized
//   403: description:Cannot join own game
//   404: description:Game not found
//   409: description:Game already full
//   500: description:Internal server error

// swagger:route POST /games/{id}/resign Games resignGame
// Resign from an active game.
// Security: bearer
// responses:
//   200: description:Resigned successfully
//   400: description:Invalid request
//   401: description:Unauthorized
//   403: description:Not a player in this game
//   409: description:Game not active
//   500: description:Internal server error

// swagger:route GET /games/{id}/moves Games getGameMoves
// Get the move history for a game.
// Security: bearer
// responses:
//   200: description:List of moves
//   400: description:Invalid game ID
//   401: description:Unauthorized
//   500: description:Internal server error

// ── Chat ────────────────────────────────────────────────────────────────────────

// swagger:route POST /games/{id}/chat Chat sendChatMessage
// Send a chat message in a game.
// Security: bearer
// responses:
//   201: description:Message sent
//   400: description:Invalid request
//   401: description:Unauthorized
//   403: description:Not a player
//   500: description:Internal server error

// swagger:route GET /games/{id}/chat Chat getChatMessages
// Get chat history for a game.
// Security: bearer
// responses:
//   200: description:Chat messages
//   400: description:Invalid game ID
//   401: description:Unauthorized
//   500: description:Internal server error

// ── Voice ───────────────────────────────────────────────────────────────────────

// swagger:route POST /games/{id}/voice Voice startVoiceSession
// Start a WebRTC voice session.
// Security: bearer
// responses:
//   201: description:Voice session created
//   400: description:Invalid request
//   401: description:Unauthorized
//   403: description:Not a player
//   409: description:Active session already exists
//   500: description:Internal server error

// swagger:route GET /games/{id}/voice Voice getActiveVoiceSession
// Get the active voice session for a game.
// Security: bearer
// responses:
//   200: description:Active voice session
//   400: description:Invalid game ID
//   401: description:Unauthorized
//   404: description:No active session
//   500: description:Internal server error

// swagger:route PATCH /games/{id}/voice/{vid}/activate Voice activateVoiceSession
// Accept an incoming voice call.
// Security: bearer
// responses:
//   200: description:Voice session activated
//   400: description:Invalid request
//   401: description:Unauthorized
//   403: description:Not the call recipient
//   404: description:Session not found
//   500: description:Internal server error

// swagger:route DELETE /games/{id}/voice/{vid} Voice endVoiceSession
// End a voice session.
// Security: bearer
// responses:
//   200: description:Voice session ended
//   400: description:Invalid request
//   401: description:Unauthorized
//   403: description:Not a player
//   404: description:Session not found
//   500: description:Internal server error

// ── Health ──────────────────────────────────────────────────────────────────────

// swagger:route GET /healthz Health healthz
// Liveness check — is the process alive?
// responses:
//   200: description:Server is alive

// swagger:route GET /readyz Health readyz
// Readiness check — can the server accept traffic?
// responses:
//   200: description:Server is ready
//   503: description:Server not ready

// ── WebSocket ───────────────────────────────────────────────────────────────────

// swagger:route GET /ws WebSocket handleWebSocket
// Upgrade to WebSocket for real-time gameplay.
//
// Query: game_id (uuid, required) — the game to join.
//
// After upgrade, authenticate by sending an auth message first:
//
//   {"type":"auth","payload":{"token":"<jwt>"}}
//
// Then send game events (make_move, chat, voice_offer, etc).
//
// Server sends keepalive pings: {"type":"ping"}. Reply with {"type":"pong"}.
// responses:
//   101: description:Switching Protocols
//   400: description:Missing or invalid game_id
//   404: description:Game not found
