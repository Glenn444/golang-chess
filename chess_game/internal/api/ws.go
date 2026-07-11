package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/Glenn444/golang-chess/internal/board"
	db "github.com/Glenn444/golang-chess/internal/db"
	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/internal/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/olahol/melody"
)

const (
	wsKeyUser          = "ws_user"
	wsKeyGameID        = "ws_game_id"
	wsKeyPlayerColor   = "ws_player_color"
	wsKeyAuthenticated = "ws_authenticated"

	// WebSocket keepalive
	pingInterval = 30 * time.Second
	authTimeout  = 10 * time.Second

	// How long an empty room keeps its game state in memory.
	gameEvictionDelay = 30 * time.Minute
)

// WSEvent is the JSON envelope for every WebSocket message.
type WSEvent struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

const (
	EventMakeMove            = "make_move"
	EventChat                = "chat"
	EventVoiceOffer          = "voice_offer"
	EventVoiceAnswer         = "voice_answer"
	EventVoiceICE            = "voice_ice"
	EventVoiceEnd            = "voice_end"
	EventError               = "error"
	EventAuth                = "auth"
	EventAuthRequired        = "auth_required"
	EventPlayerDisconnected  = "player_disconnected"
	EventPlayerReconnected   = "player_reconnected"
	EventPing                = "ping"
	EventPong                = "pong"
	EventVoiceStats          = "voice_stats"
)

// setupMelody wires lifecycle hooks and keepalive. Called once in NewServer.
func (server *Server) setupMelody() {
    m := server.melody
    m.Upgrader.ReadBufferSize = 4096
    m.Upgrader.WriteBufferSize = 4096

    // This is what actually controls max WS message size
    m.Config.MaxMessageSize = 64 * 1024 // 64KB — SDP offers can be ~8-16KB

    m.HandleConnect(server.wsOnConnect)
    m.HandleDisconnect(server.wsOnDisconnect)
    m.HandleMessage(server.wsOnMessage)

    go func() {
        ticker := time.NewTicker(pingInterval)
        defer ticker.Stop()
        for range ticker.C {
            pingMsg, _ := json.Marshal(WSEvent{Type: EventPing})
            server.melody.Broadcast(pingMsg)
        }
    }()
}

// handleWebSocket upgrades to WebSocket. No auth token in query string —
// the client authenticates by sending an "auth" message as the first frame.
// Only ?game_id=<uuid> is required.
func (server *Server) handleWebSocket(ctx *gin.Context) {
	parsed, err := uuid.Parse(ctx.Query("game_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorMessage("game_id query param required"))
		return
	}
	gameID := pgtype.UUID{Bytes: parsed, Valid: true}

	// Verify the game exists before upgrading.
	if _, err := server.store.GetGameByID(ctx, gameID); err != nil {
		ctx.JSON(http.StatusNotFound, errorMessage(ErrGameNotFound))
		return
	}

	if err := server.melody.HandleRequestWithKeys(ctx.Writer, ctx.Request, map[string]any{
		wsKeyGameID:        gameID,
		wsKeyAuthenticated: false,
	}); err != nil {
		slog.Error("ws: upgrade failed", "err", err)
	}
}

func (server *Server) wsOnConnect(s *melody.Session) {
	if wsIsAuthenticated(s) {
		// Pre-authenticated session — load game state.
		u := wsUser(s)
		gameID := wsGameID(s)

		if _, err := server.getOrRestoreGame(s.Request.Context(), gameID); err != nil {
			slog.Error("ws: failed to load game on connect", "game_id", gameID, "err", err)
			s.CloseWithMsg(melody.FormatCloseMessage(1011, "failed to load game"))
			return
		}

		// Send game_state with opponent info.
		server.sendGameState(s, gameID, u)
		slog.Info("ws: connected", "username", u.Username)
		return
	}

	slog.Info("ws: connected — waiting for auth")
	s.Write([]byte(`{"type":"auth_required"}`))

	// Drop sockets that never authenticate within the deadline.
	go func() {
		time.Sleep(authTimeout)
		if !s.IsClosed() && !wsIsAuthenticated(s) {
			s.CloseWithMsg(melody.FormatCloseMessage(4401, "authentication timeout"))
		}
	}()
}

func (server *Server) wsOnDisconnect(s *melody.Session) {
	user, ok := wsUserSafe(s)
	if !ok {
		slog.Info("ws: unauthenticated session disconnected")
		return
	}
	gameID := wsGameID(s)

	out, _ := json.Marshal(WSEvent{Type: EventPlayerDisconnected, Payload: wsMarshal(gin.H{
		"username": user.Username,
		"color":    wsPlayerColor(s),
	})})
	server.wsBroadcastToGame(gameID, out)

	remaining := 0
	server.melody.BroadcastFilter([]byte{}, func(other *melody.Session) bool {
		if other != s && wsIsAuthenticated(other) && uuidEq(wsGameID(other), gameID) {
			remaining++
		}
		return false
	})

	if remaining == 0 {
		go func(gid pgtype.UUID) {
			time.Sleep(gameEvictionDelay)
			// Only evict if the room is still empty — players may have
			// reconnected and be mid-game.
			if server.countConnectedPlayers(gid) > 0 {
				return
			}
			server.activeGamesMu.Lock()
			delete(server.activeGames, gid)
			server.activeGamesMu.Unlock()
		}(gameID)
	}

	slog.Info("ws: disconnected", "username", user.Username, "remaining", remaining)
}

func (server *Server) wsOnMessage(s *melody.Session, raw []byte) {
	var event WSEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		wsWriteError(s, "invalid message format")
		return
	}

	// Pong responses just reset the keepalive — no further processing.
	if event.Type == EventPong {
		return
	}

	// If not yet authenticated, require an auth message first.
	if !wsIsAuthenticated(s) {
		if event.Type != EventAuth {
			wsWriteError(s, "authentication required — send auth message first")
			s.Close()
			return
		}
		server.wsHandleAuth(s, event.Payload)
		return
	}

	gameID := wsGameID(s)

	switch event.Type {
	case EventChat:
		server.wsHandleChat(s, gameID, event.Payload)
	case EventMakeMove:
		server.wsHandleMove(s, gameID, event.Payload)
	case EventVoiceOffer, EventVoiceAnswer, EventVoiceICE, EventVoiceEnd:
		server.wsRelayToOthers(s, gameID, raw)
	case EventVoiceStats:
		server.wsHandleVoiceStats(s, gameID, event.Payload)
	default:
		wsWriteError(s, "unknown event type: "+event.Type)
	}
}

// wsHandleAuth verifies the JWT token, sets session keys, and sends the
// current game state to the newly authenticated client.
func (server *Server) wsHandleAuth(s *melody.Session, payload json.RawMessage) {
	var body struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(payload, &body); err != nil || body.Token == "" {
		wsWriteError(s, "auth payload must include token")
		return
	}

	payloadToken, err := server.tokenMaker.VerifyToken(body.Token, token.AccessTokenType)
	if err != nil {
		wsWriteError(s, ErrInvalidToken)
		return
	}

	user, err := server.store.GetUserByUsername(s.Request.Context(), payloadToken.Username)
	if err != nil {
		wsWriteError(s, ErrUserNotFound)
		return
	}

	gameID := wsGameID(s)

	dbGame, err := server.store.GetGameByID(s.Request.Context(), gameID)
	if err != nil {
		wsWriteError(s, ErrGameNotFound)
		return
	}
	if !uuidEq(dbGame.WhitePlayerID, user.ID) && !uuidEq(dbGame.BlackPlayerID, user.ID) {
		wsWriteError(s, ErrNotAPlayer)
		return
	}

	var playerColor string
	if uuidEq(dbGame.WhitePlayerID, user.ID) {
		playerColor = "w"
	} else {
		playerColor = "b"
	}

	s.Set(wsKeyUser, user)
	s.Set(wsKeyPlayerColor, playerColor)
	s.Set(wsKeyAuthenticated, true)

	slog.Info("ws: authenticated", "username", user.Username, "game_id", gameID)

	// Load or restore game state.
	if _, err := server.getOrRestoreGame(s.Request.Context(), gameID); err != nil {
		wsWriteError(s, "failed to load game")
		return
	}

	server.sendGameState(s, gameID, user)
	out, _ := json.Marshal(WSEvent{Type: EventPlayerReconnected, Payload: wsMarshal(gin.H{
		"username": user.Username,
		"color":    playerColor,
	})})
	server.wsRelayToOthers(s, gameID, out)
	server.maybeStartWatcher(gameID, dbGame.State)
}

// getOrRestoreGame returns the in-memory game state, restoring it from the DB
// exactly once when it was evicted. Concurrent callers get the same instance.
func (server *Server) getOrRestoreGame(ctx context.Context, gameID pgtype.UUID) (*pieces.GameState, error) {
	server.activeGamesMu.RLock()
	gs, ok := server.activeGames[gameID]
	server.activeGamesMu.RUnlock()
	if ok {
		return gs, nil
	}

	game, err := server.store.GetGameByID(ctx, gameID)
	if err != nil {
		return nil, err
	}

	server.activeGamesMu.Lock()
	defer server.activeGamesMu.Unlock()
	// Another goroutine may have restored it while we were reading the DB.
	if gs, ok := server.activeGames[gameID]; ok {
		return gs, nil
	}
	gs = restoreGameState(game)
	server.activeGames[gameID] = gs
	slog.Info("ws: game loaded into memory", "game_id", uidStr(gameID))
	return gs, nil
}

// wsHandleChat persists the message then broadcasts it to the game room.
func (server *Server) wsHandleChat(s *melody.Session, gameID pgtype.UUID, payload json.RawMessage) {
	var body struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(payload, &body); err != nil || body.Content == "" {
		wsWriteError(s, "chat payload must include content")
		return
	}

	u := wsUser(s)
	msg, err := server.store.CreateChatMessage(s.Request.Context(), db.CreateChatMessageParams{
		GameID:   gameID,
		SenderID: u.ID,
		Content:  body.Content,
	})
	if err != nil {
		slog.Error("ws: CreateChatMessage failed", "err", err)
		wsWriteError(s, ErrInternalServer)
		return
	}

	out, _ := json.Marshal(WSEvent{Type: EventChat, Payload: wsMarshal(msg)})
	server.wsBroadcastToGame(gameID, out)
}

// wsHandleMove validates and broadcasts a move, persisting the result.
type MoveResult struct {
	Move          string `json:"move"`
	CurrentPlayer string `json:"current_player"`
	InCheck       bool   `json:"in_check"`
	IsCheckmate   bool   `json:"is_checkmate"`
	IsStalemate           bool   `json:"is_stalemate"`
	EndReason             string `json:"end_reason"`
	EndedByPlayerID       string `json:"ended_by_player_id"`
	WhiteTimeRemainingMs  int64  `json:"white_time_remaining_ms"`
	BlackTimeRemainingMs  int64  `json:"black_time_remaining_ms"`
}

func (server *Server) wsHandleMove(s *melody.Session, gameID pgtype.UUID, payload json.RawMessage) {
	gamestate, err := server.getOrRestoreGame(s.Request.Context(), gameID)
	if err != nil {
		wsWriteError(s, "game not found")
		return
	}
	var body struct {
		Move string `json:"move"`
	}
	if err := json.Unmarshal(payload, &body); err != nil || body.Move == "" {
		wsWriteError(s, "invalid move payload")
		return
	}

	gamestate.GameStateMu.Lock()
	defer gamestate.GameStateMu.Unlock()

	if wsGameIsOver(gamestate.Status) {
		wsWriteError(s, "game is already over")
		return
	}

	previousPlayer := gamestate.CurrentPlayer
	user := wsUser(s)

	playerColor := wsPlayerColor(s)
	if playerColor != gamestate.CurrentPlayer {
		wsWriteError(s, "not your turn")
		return
	}

	// Unlimited games carry zero on both clocks from creation. Evaluate before
	// deducting: in a timed game at most one clock can reach zero.
	unlimited := gamestate.WhiteTimeRemainingMs == 0 && gamestate.BlackTimeRemainingMs == 0

	if err := board.Move(gamestate, body.Move); err != nil {
		wsWriteError(s, err.Error())
		return
	}

	// Deduct elapsed time from the player who just moved (skip if unlimited).
	now := time.Now()
	if !unlimited {
		elapsed := now.Sub(gamestate.LastMoveAt).Milliseconds()
		if previousPlayer == "w" {
			gamestate.WhiteTimeRemainingMs = max(0, gamestate.WhiteTimeRemainingMs-elapsed)
		} else {
			gamestate.BlackTimeRemainingMs = max(0, gamestate.BlackTimeRemainingMs-elapsed)
		}
	}
	gamestate.LastMoveAt = now

	check := board.IsKinginCheck(gamestate)
	isCheckmate := board.IsCheckmate(gamestate)
	isStalemate := board.IsStalemate(gamestate)
	timedOut := !unlimited && (gamestate.WhiteTimeRemainingMs <= 0 || gamestate.BlackTimeRemainingMs <= 0)

	gameStatus := db.GameStateActive
	var endReason string
	switch {
	case isCheckmate:
		gameStatus = db.GameStateCheckmate
		endReason = "checkmate"
	case timedOut:
		gameStatus = db.GameStateTimeout
		endReason = "timeout"
	case isStalemate:
		gameStatus = db.GameStateStalemate
		endReason = "stalemate"
	}
	gameOver := endReason != ""

	gamestate.MoveNumber++
	gamestate.Status = gameStatus
	gamestate.InCheck = check

	// Only stamp who ended the game when it actually ended.
	var endedBy pgtype.UUID
	if gameOver {
		endedBy = user.ID
	}

	_, err = server.store.UpdateGameState(s.Request.Context(), db.UpdateGameStateParams{
		ID:                   gameID,
		State:                gameStatus,
		InCheck:              check,
		CurrentPlayer:        db.PlayerColor(gamestate.CurrentPlayer),
		MoveCount:            gamestate.MoveNumber,
		BoardState:           board.SerializeGameState(gamestate),
		WhiteTimeRemainingMs: gamestate.WhiteTimeRemainingMs,
		BlackTimeRemainingMs: gamestate.BlackTimeRemainingMs,
		EndedByPlayerID:      endedBy,
		EndReason:            endReason,
		LastMoveAt:           pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		slog.Error("ws: wsHandleMove, failed UpdateGameState", "err", err)
	}

	// Cancel the current timeout watcher; start a new one for the next player if the game continues.
	stopTimeoutWatcher(gamestate)
	if gameOver {
		server.activeGamesMu.Lock()
		delete(server.activeGames, gameID)
		server.activeGamesMu.Unlock()
	} else if !unlimited {
		server.startTimeoutWatcher(gameID, gamestate.TimeoutCh)
	}

	_, err = server.store.CreateMove(s.Request.Context(), db.CreateMoveParams{
		GameID:       gameID,
		PlayerID:     user.ID,
		PlayerColor:  db.PlayerColor(previousPlayer),
		MoveNotation: body.Move,
		MoveNumber:   gamestate.MoveNumber,
	})
	if err != nil {
		slog.Error("ws: CreateMove failed", "err", err)
	}

	result := MoveResult{
		Move:                 body.Move,
		CurrentPlayer:        gamestate.CurrentPlayer,
		InCheck:              check,
		IsCheckmate:          isCheckmate,
		IsStalemate:          isStalemate,
		EndReason:            endReason,
		EndedByPlayerID:      uidStr(user.ID),
		WhiteTimeRemainingMs: gamestate.WhiteTimeRemainingMs,
		BlackTimeRemainingMs: gamestate.BlackTimeRemainingMs,
	}

	out, _ := json.Marshal(WSEvent{Type: EventMakeMove, Payload: wsMarshal(result)})
	server.wsBroadcastToGame(gameID, out)
}

// wsGameIsOver reports whether the game has reached a terminal state. The
// zero value (in-memory games created before their first persist) and
// waiting/active games are playable.
func wsGameIsOver(status db.GameState) bool {
	switch status {
	case db.GameStateCheckmate, db.GameStateStalemate, db.GameStateResign,
		db.GameStateDraw, db.GameStateAbandoned, db.GameStateTimeout:
		return true
	}
	return false
}

// stopTimeoutWatcher cancels the running watcher (if any) and arms a fresh
// channel. Must be called with GameStateMu held so the close/replace is
// atomic — closing the same channel twice panics.
func stopTimeoutWatcher(gs *pieces.GameState) {
	if gs.TimeoutCh != nil {
		close(gs.TimeoutCh)
	}
	gs.TimeoutCh = make(chan struct{})
}

// wsRelayToOthers forwards a message to every other authenticated session in
// the same game room.
func (server *Server) wsRelayToOthers(s *melody.Session, gameID pgtype.UUID, msg []byte) {
	server.melody.BroadcastFilter(msg, func(other *melody.Session) bool {
		return other != s && wsIsAuthenticated(other) && uuidEq(wsGameID(other), gameID)
	})
}

// wsHandleVoiceStats logs the ICE candidate pair that the client selected.
// This tells us whether the connection is direct P2P (free) or relayed via TURN (cost).
// The client sends this after the WebRTC connection is established.
type voiceStatsPayload struct {
	LocalType       string `json:"localType"`
	RemoteType      string `json:"remoteType"`
	RelayProtocol   string `json:"relayProtocol"`
	SelectedPair    string `json:"selectedPair"`
	LocalCandidate  string `json:"localCandidate"`
	RemoteCandidate string `json:"remoteCandidate"`
}

func (server *Server) wsHandleVoiceStats(s *melody.Session, gameID pgtype.UUID, raw json.RawMessage) {
	var p voiceStatsPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		slog.Warn("ws: voice_stats parse failed", "err", err)
		return
	}
	user := wsUser(s)
	slog.Info("ws: voice_stats",
		"game_id", gameID,
		"username", user.Username,
		"local_type", p.LocalType,
		"remote_type", p.RemoteType,
		"relay_protocol", p.RelayProtocol,
		"selected_pair", p.SelectedPair,
		"local_candidate", p.LocalCandidate,
		"remote_candidate", p.RemoteCandidate,
	)
}

// wsUserInGame returns true if the given user has an active WebSocket session
// in the specified game room.
func (server *Server) wsUserInGame(userID pgtype.UUID, gameID pgtype.UUID) bool {
	found := false
	server.melody.BroadcastFilter([]byte{}, func(s *melody.Session) bool {
		if uuidEq(wsGameID(s), gameID) {
			u, ok := wsUserSafe(s)
			if ok && uuidEq(u.ID, userID) {
				found = true
			}
		}
		return false
	})
	return found
}

// wsBroadcastToGame sends a message to all authenticated sessions in the game
// room. Unauthenticated sockets must not see game traffic.
func (server *Server) wsBroadcastToGame(gameID pgtype.UUID, msg []byte) {
	server.melody.BroadcastFilter(msg, func(s *melody.Session) bool {
		return wsIsAuthenticated(s) && uuidEq(wsGameID(s), gameID)
	})
}

// sendGameState builds the game_state payload with opponent info and sends it
// to the given session.
func (server *Server) sendGameState(s *melody.Session, gameID pgtype.UUID, user db.User) {
	server.activeGamesMu.RLock()
	gs, ok := server.activeGames[gameID]
	server.activeGamesMu.RUnlock()
	if !ok || gs == nil {
		slog.Error("ws: sendGameState — game not found in memory", "game_id", gameID)
		wsWriteError(s, "game state not available")
		return
	}

	// Determine opponent username.
	game, err := server.store.GetGameByID(s.Request.Context(), gameID)
	if err != nil {
		slog.Warn("ws: sendGameState — DB lookup failed, sending without opponent info",
			"game_id", gameID, "err", err)
		gs.GameStateMu.RLock()
		payload := wsMarshal(gs)
		gs.GameStateMu.RUnlock()
		out, _ := json.Marshal(WSEvent{Type: "game_state", Payload: payload})
		s.Write(out)
		return
	}

	opponentID := game.WhitePlayerID
	if uuidEq(opponentID, user.ID) {
		opponentID = game.BlackPlayerID
	}
	opponentName := server.lookupUsername(s.Request.Context(), opponentID)

	payload := GameStatePayload{
		Game:             gs,
		OpponentUsername: opponentName,
	}
	// A concurrent move mutates the board — snapshot under the read lock.
	gs.GameStateMu.RLock()
	raw := wsMarshal(payload)
	gs.GameStateMu.RUnlock()
	out, _ := json.Marshal(WSEvent{Type: "game_state", Payload: raw})
	s.Write(out)
}

// ── Session key helpers ───────────────────────────────────────────────────────

func wsUser(s *melody.Session) db.User {
	v, _ := s.Get(wsKeyUser)
	return v.(db.User)
}

func wsUserSafe(s *melody.Session) (db.User, bool) {
	v, exists := s.Get(wsKeyUser)
	if !exists {
		return db.User{}, false
	}
	return v.(db.User), true
}

func wsGameID(s *melody.Session) pgtype.UUID {
	v, _ := s.Get(wsKeyGameID)
	return v.(pgtype.UUID)
}

func wsIsAuthenticated(s *melody.Session) bool {
	v, exists := s.Get(wsKeyAuthenticated)
	if !exists {
		return false
	}
	auth, _ := v.(bool)
	return auth
}

func wsWriteError(s *melody.Session, msg string) {
	out, _ := json.Marshal(WSEvent{Type: EventError, Payload: wsMarshal(gin.H{"error": msg})})
	s.Write(out)
}

func wsMarshal(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func wsPlayerColor(s *melody.Session) string {
	v, _ := s.Get(wsKeyPlayerColor)
	return v.(string)
}

// countConnectedPlayers returns the number of authenticated WebSocket sessions in the game room.
func (server *Server) countConnectedPlayers(gameID pgtype.UUID) int {
	count := 0
	server.melody.BroadcastFilter([]byte{}, func(s *melody.Session) bool {
		if uuidEq(wsGameID(s), gameID) && wsIsAuthenticated(s) {
			count++
		}
		return false
	})
	return count
}

// maybeStartWatcher starts (or restarts) a timeout watcher when both players are
// connected and the game is a timed, active game. Must NOT be called while
// holding activeGamesMu.
func (server *Server) maybeStartWatcher(gameID pgtype.UUID, gameState db.GameState) {
	if gameState != db.GameStateActive {
		return
	}
	if server.countConnectedPlayers(gameID) < 2 {
		return
	}
	server.activeGamesMu.RLock()
	gs, ok := server.activeGames[gameID]
	server.activeGamesMu.RUnlock()
	if !ok {
		return
	}

	gs.GameStateMu.Lock()
	defer gs.GameStateMu.Unlock()
	if gs.WhiteTimeRemainingMs == 0 && gs.BlackTimeRemainingMs == 0 {
		return // unlimited time game
	}
	stopTimeoutWatcher(gs)
	server.startTimeoutWatcher(gameID, gs.TimeoutCh)
}

// startTimeoutWatcher launches a goroutine that fires handleTimeout when the
// active player's clock reaches zero. It exits early if timeoutCh is closed
// (move was made or game ended).
func (server *Server) startTimeoutWatcher(gameID pgtype.UUID, timeoutCh <-chan struct{}) {
	go func() {
		server.activeGamesMu.RLock()
		gs, ok := server.activeGames[gameID]
		server.activeGamesMu.RUnlock()
		if !ok {
			return
		}

		gs.GameStateMu.RLock()
		currentPlayer := gs.CurrentPlayer
		var remaining int64
		if currentPlayer == "w" {
			remaining = gs.WhiteTimeRemainingMs
		} else {
			remaining = gs.BlackTimeRemainingMs
		}
		timeLeft := remaining - time.Since(gs.LastMoveAt).Milliseconds()
		gs.GameStateMu.RUnlock()

		if timeLeft <= 0 {
			server.handleTimeout(gameID, currentPlayer)
			return
		}

		select {
		case <-timeoutCh:
			return
		case <-time.After(time.Duration(timeLeft) * time.Millisecond):
			server.handleTimeout(gameID, currentPlayer)
		}
	}()
}

// handleTimeout fires when a player's clock expires with no move. Idempotent:
// does nothing if the game is no longer in activeGames.
func (server *Server) handleTimeout(gameID pgtype.UUID, timedOutColor string) {
	server.activeGamesMu.Lock()
	gs, ok := server.activeGames[gameID]
	if !ok {
		server.activeGamesMu.Unlock()
		return
	}
	delete(server.activeGames, gameID)
	server.activeGamesMu.Unlock()

	gs.GameStateMu.Lock()
	if timedOutColor == "w" {
		gs.WhiteTimeRemainingMs = 0
	} else {
		gs.BlackTimeRemainingMs = 0
	}
	gs.Status = db.GameStateTimeout
	params := db.UpdateGameStateParams{
		ID:                   gameID,
		State:                db.GameStateTimeout,
		InCheck:              gs.InCheck,
		CurrentPlayer:        db.PlayerColor(gs.CurrentPlayer),
		MoveCount:            gs.MoveNumber,
		BoardState:           board.SerializeGameState(gs),
		WhiteTimeRemainingMs: gs.WhiteTimeRemainingMs,
		BlackTimeRemainingMs: gs.BlackTimeRemainingMs,
		EndReason:            "timeout",
		LastMoveAt:           pgtype.Timestamptz{Time: gs.LastMoveAt, Valid: true},
	}
	result := MoveResult{
		CurrentPlayer:        gs.CurrentPlayer,
		EndReason:            "timeout",
		WhiteTimeRemainingMs: gs.WhiteTimeRemainingMs,
		BlackTimeRemainingMs: gs.BlackTimeRemainingMs,
	}
	gs.GameStateMu.Unlock()

	ctx := context.Background()
	if _, err := server.store.UpdateGameState(ctx, params); err != nil {
		slog.Error("handleTimeout: UpdateGameState failed", "game_id", uidStr(gameID), "err", err)
	}

	out, _ := json.Marshal(WSEvent{Type: EventMakeMove, Payload: wsMarshal(result)})
	server.wsBroadcastToGame(gameID, out)
	slog.Info("handleTimeout: game ended by clock", "game_id", uidStr(gameID), "timed_out_color", timedOutColor)
}

func restoreGameState(game db.Game) *pieces.GameState {
	snap := board.DeserializeGameState(game.BoardState)
	return &pieces.GameState{
		CurrentPlayer:        string(game.CurrentPlayer),
		Board:                snap.Board,
		Castle:               snap.Castle,
		EnPassantTarget:      snap.EnPassantTarget,
		MoveNumber:           game.MoveCount,
		Status:               game.State,
		InCheck:              game.InCheck,
		WhiteTimeRemainingMs: game.WhiteTimeRemainingMs,
		BlackTimeRemainingMs: game.BlackTimeRemainingMs,
		// Clocks pause while the game is out of memory: resume from the stored
		// remaining time instead of charging the mover for the downtime (which
		// used to forfeit any game restored after a long gap).
		LastMoveAt:     time.Now(),
		StockfishGame:  snap.StockfishGame,
		CapturedPieces: make(map[string][]pieces.PieceInterface),
		TimeoutCh:      make(chan struct{}),
	}
}
