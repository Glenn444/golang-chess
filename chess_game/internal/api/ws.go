package api

import (
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

		server.activeGamesMu.Lock()
		if _, exists := server.activeGames[gameID]; exists {
			server.activeGamesMu.Unlock()
			server.sendGameState(s, gameID, u)
			slog.Info("ws: reconnected", "username", u.Username)
			return
		}
		server.activeGamesMu.Unlock()

		game, err := server.store.GetGameByID(s.Request.Context(), gameID)
		if err != nil {
			slog.Error("ws: failed to load game on connect", "game_id", gameID, "err", err)
			s.CloseWithMsg(melody.FormatCloseMessage(1011, "failed to load game"))
			return
		}
		server.activeGamesMu.Lock()
		server.activeGames[gameID] = restoreGameState(game)
		server.activeGamesMu.Unlock()

		// Send game_state with opponent info.
		server.sendGameState(s, gameID, u)
		slog.Info("ws: game loaded into memory", "game_id", gameID)
		slog.Info("ws: connected", "username", u.Username)
		return
	}

	slog.Info("ws: connected — waiting for auth")
	s.Write([]byte(`{"type":"auth_required"}`))
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
		if other != s && uuidEq(wsGameID(other), gameID) {
			remaining++
		}
		return false
	})

	if remaining == 0 {
		go func(gid pgtype.UUID) {
			time.Sleep(5 * time.Minute)
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
	server.activeGamesMu.Lock()
	if _, exists := server.activeGames[gameID]; exists {
		server.activeGamesMu.Unlock()
		server.sendGameState(s, gameID, user)
		slog.Info("ws: reconnected", "username", user.Username)
		return
	}
	server.activeGamesMu.Unlock()

	// First player — restore from DB.
	game, err := server.store.GetGameByID(s.Request.Context(), gameID)
	if err != nil {
		wsWriteError(s, "failed to load game")
		return
	}

	server.activeGamesMu.Lock()
	server.activeGames[gameID] = restoreGameState(game)
	server.activeGamesMu.Unlock()

	server.sendGameState(s, gameID, user)
	slog.Info("ws: game loaded into memory", "game_id", gameID)
	slog.Info("ws: connected", "username", user.Username)
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
	IsStalemate   bool   `json:"is_stalemate"`
}

func (server *Server) wsHandleMove(s *melody.Session, gameID pgtype.UUID, payload json.RawMessage) {
	server.activeGamesMu.RLock()
	gamestate, ok := server.activeGames[gameID]
	server.activeGamesMu.RUnlock()

	if !ok {
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
	previousPlayer := gamestate.CurrentPlayer
	user := wsUser(s)

	playerColor := wsPlayerColor(s)
	if playerColor != gamestate.CurrentPlayer {
		wsWriteError(s, "not your turn")
		return
	}

	err := board.Move(gamestate, body.Move)
	if err != nil {
		wsWriteError(s, err.Error())
		return
	}

	// Deduct elapsed time from the player who just moved.
	now := time.Now()
	elapsed := now.Sub(gamestate.LastMoveAt).Milliseconds()
	if previousPlayer == "w" {
		gamestate.WhiteTimeRemainingMs -= elapsed
		if gamestate.WhiteTimeRemainingMs <= 0 {
			gamestate.WhiteTimeRemainingMs = 0
		}
	} else {
		gamestate.BlackTimeRemainingMs -= elapsed
		if gamestate.BlackTimeRemainingMs <= 0 {
			gamestate.BlackTimeRemainingMs = 0
		}
	}
	gamestate.LastMoveAt = now

	check := board.IsKinginCheck(gamestate)
	isCheckmate := board.IsCheckmate(gamestate)
	isStalemate := board.IsStalemate(gamestate)

	// Check for timeout.
	timedOut := gamestate.WhiteTimeRemainingMs <= 0 || gamestate.BlackTimeRemainingMs <= 0

	gameStatus := db.GameStateActive
	switch {
	case isCheckmate || timedOut:
		gameStatus = db.GameStateCheckmate
	case isStalemate:
		gameStatus = db.GameStateStalemate
	}

	gamestate.MoveNumber++

	_, err = server.store.UpdateGameState(s.Request.Context(), db.UpdateGameStateParams{
		ID:                   gameID,
		State:                gameStatus,
		InCheck:              check,
		CurrentPlayer:        db.PlayerColor(gamestate.CurrentPlayer),
		MoveCount:            gamestate.MoveNumber,
		BoardState:           board.SerializeGameState(gamestate),
		WhiteTimeRemainingMs: gamestate.WhiteTimeRemainingMs,
		BlackTimeRemainingMs: gamestate.BlackTimeRemainingMs,
	})
	if err != nil {
		slog.Error("ws: wsHandleMove, failed UpdateGameState", "err", err)
	}

	if isCheckmate || isStalemate {
		server.activeGamesMu.Lock()
		delete(server.activeGames, gameID)
		server.activeGamesMu.Unlock()
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
		Move:          body.Move,
		CurrentPlayer: gamestate.CurrentPlayer,
		InCheck:       check,
		IsCheckmate:   isCheckmate,
		IsStalemate:   isStalemate,
	}

	out, _ := json.Marshal(WSEvent{Type: EventMakeMove, Payload: wsMarshal(result)})
	server.wsBroadcastToGame(gameID, out)
}

// wsRelayToOthers forwards a message to every other session in the same game room.
func (server *Server) wsRelayToOthers(s *melody.Session, gameID pgtype.UUID, msg []byte) {
	server.melody.BroadcastFilter(msg, func(other *melody.Session) bool {
		return other != s && uuidEq(wsGameID(other), gameID)
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

// wsBroadcastToGame sends a message to ALL sessions in the game room.
func (server *Server) wsBroadcastToGame(gameID pgtype.UUID, msg []byte) {
	server.melody.BroadcastFilter(msg, func(s *melody.Session) bool {
		return uuidEq(wsGameID(s), gameID)
	})
}

// sendGameState builds the game_state payload with opponent info and sends it
// to the given session.
func (server *Server) sendGameState(s *melody.Session, gameID pgtype.UUID, user db.User) {
	server.activeGamesMu.RLock()
	gs, ok := server.activeGames[gameID]
	server.activeGamesMu.RUnlock()
	if !ok {
		return
	}

	// Determine opponent username.
	game, err := server.store.GetGameByID(s.Request.Context(), gameID)
	if err != nil {
		// Fallback: send without opponent info.
		out, _ := json.Marshal(WSEvent{Type: "game_state", Payload: wsMarshal(gs)})
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
	out, _ := json.Marshal(WSEvent{Type: "game_state", Payload: wsMarshal(payload)})
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

func restoreGameState(game db.Game) *pieces.GameState {
	boardState, stockfishHistory := board.DeserializeGameState(game.BoardState)
	return &pieces.GameState{
		CurrentPlayer:        string(game.CurrentPlayer),
		Board:                boardState,
		MoveNumber:           game.MoveCount,
		Status:               game.State,
		InCheck:              game.InCheck,
		WhiteTimeRemainingMs: game.WhiteTimeRemainingMs,
		BlackTimeRemainingMs: game.BlackTimeRemainingMs,
		LastMoveAt:           game.LastMoveAt.Time,
		StockfishGame:        stockfishHistory,
		CapturedPieces:       make(map[string][]pieces.PieceInterface),
		TimeoutCh:            make(chan struct{}),
	}
}
