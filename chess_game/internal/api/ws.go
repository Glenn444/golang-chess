package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

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
	wsKeyUser        = "ws_user"
	wsKeyGameID      = "ws_game_id"
	wsKeyPlayerColor = "ws_player_color"
)

// WSEvent is the JSON envelope for every WebSocket message.
type WSEvent struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

const (
	EventMakeMove    = "make_move"
	EventChat        = "chat"
	EventVoiceOffer  = "voice_offer"
	EventVoiceAnswer = "voice_answer"
	EventVoiceICE    = "voice_ice"
	EventVoiceEnd    = "voice_end"
	EventError       = "error"
)

// setupMelody wires the melody lifecycle hooks. Called once in NewServer.
func (server *Server) setupMelody() {
	m := server.melody
	m.HandleConnect(server.wsOnConnect)
	m.HandleDisconnect(server.wsOnDisconnect)
	m.HandleMessage(server.wsOnMessage)
}

// handleWebSocket upgrades the connection to WebSocket.
// Query params:
//
//	?token=<access_token>   — Bearer JWT (WS clients cannot set custom headers)
//	?game_id=<uuid>         — game room to join
func (server *Server) handleWebSocket(ctx *gin.Context) {
	rawToken := ctx.Query("token")
	if rawToken == "" {
		ctx.JSON(http.StatusUnauthorized, errorMessage(ErrUnauthorized))
		return
	}

	payload, err := server.tokenMaker.VerifyToken(rawToken, token.AccessTokenType)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorMessage(ErrInvalidToken))
		return
	}

	user, err := server.store.GetUserByUsername(ctx, payload.Username)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorMessage(ErrUserNotFound))
		return
	}

	parsed, err := uuid.Parse(ctx.Query("game_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorMessage("game_id query param required"))
		return
	}
	gameID := pgtype.UUID{Bytes: parsed, Valid: true}

	// verify the caller is actually a player in this game
	game, err := server.store.GetGameByID(ctx, gameID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, errorMessage(ErrGameNotFound))
		return
	}
	if !uuidEq(game.WhitePlayerID, user.ID) && !uuidEq(game.BlackPlayerID, user.ID) {
		ctx.JSON(http.StatusForbidden, errorMessage(ErrNotAPlayer))
		return
	}
	var playerColor string
	if uuidEq(game.WhitePlayerID, user.ID) {
		playerColor = "w"
	} else {
		playerColor = "b"
	}
	if err := server.melody.HandleRequestWithKeys(ctx.Writer, ctx.Request, map[string]any{
		wsKeyUser:        user,
		wsKeyGameID:      gameID,
		wsKeyPlayerColor: playerColor,
	}); err != nil {
		slog.Error("ws: upgrade failed", "err", err)
	}
}

func (server *Server) wsOnConnect(s *melody.Session) {
	u := wsUser(s)
	gameID := wsGameID(s)

	server.activeGamesMu.Lock()
    defer server.activeGamesMu.Unlock()

	if _,exists := server.activeGames[gameID];exists{
	// send current state to the joining session so their UI is in sync
    gameState := server.activeGames[gameID]
    out, _ := json.Marshal(WSEvent{
        Type:    "game_state",
        Payload: wsMarshal(gameState),
    })
    s.Write(out)
    return
	}

	// first player to connect — load from DB
    game, err := server.store.GetGameByID(s.Request.Context(), gameID)
    if err != nil {
        slog.Error("ws: failed to load game on connect", "game_id", gameID, "err", err)
        s.CloseWithMsg(melody.FormatCloseMessage(1011, "failed to load game"))
        return
    }

    server.activeGames[gameID] = restoreGameState(game)
    slog.Info("ws: game loaded into memory", "game_id", gameID)

	slog.Info("ws: connected", "username", u.Username)
}

func (server *Server) wsOnDisconnect(s *melody.Session) {
	u := wsUser(s)
	slog.Info("ws: disconnected", "username", u.Username)
}

func (server *Server) wsOnMessage(s *melody.Session, raw []byte) {
	var event WSEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		wsWriteError(s, "invalid message format")
		return
	}

	gameID := wsGameID(s)

	switch event.Type {
	case EventChat:
		server.wsHandleChat(s, gameID, event.Payload)
	case EventMakeMove:
		server.wsHandleMove(s, gameID, event.Payload)
	case EventVoiceOffer, EventVoiceAnswer, EventVoiceICE, EventVoiceEnd:
		// relay WebRTC signalling directly to the other player — no DB storage
		server.wsRelayToOthers(s, gameID, raw)
	default:
		wsWriteError(s, "unknown event type: "+event.Type)
	}
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

	//lock access to the activeGames
	server.activeGamesMu.RLock()
	gamestate, ok := server.activeGames[gameID]
	server.activeGamesMu.RUnlock()

	if !ok {
		wsWriteError(s, "game not found")
		return
	}
	var body struct {
		Move string `json:"move"` //e2e3
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
	//enforce turn
	if playerColor != gamestate.CurrentPlayer {
		wsWriteError(s, "not your turn")
		return
	}

	err := board.Move(gamestate, body.Move)
	if err != nil {
		wsWriteError(s, err.Error())
		return
	}

	check := board.IsKinginCheck(gamestate)
	isCheckmate := board.IsCheckmate(gamestate)
	isStalemate := board.IsStalemate(gamestate)

	gameStatus := db.GameStateActive
	switch {
	case isCheckmate:
		gameStatus = db.GameStateCheckmate
	case isStalemate:
		gameStatus = db.GameStateStalemate
	}

	//increment move number after successful move
	gamestate.MoveNumber++

	_, err = server.store.UpdateGameState(s.Request.Context(), db.UpdateGameStateParams{
		ID:      gameID,
		State:   gameStatus,
		InCheck: check,
		CurrentPlayer: db.PlayerColor(gamestate.Status),
		MoveCount: gamestate.MoveNumber,
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
		CurrentPlayer: gamestate.CurrentPlayer, //already flipped by board.Move()
		InCheck:       check,
		IsCheckmate:   isCheckmate,
		IsStalemate:   isStalemate,
	}

	out, _ := json.Marshal(WSEvent{Type: EventMakeMove, Payload: wsMarshal(result)})
	server.wsBroadcastToGame(gameID, out)
}

// wsRelayToOthers forwards a message to every other session in the same game room.
// Used for WebRTC signalling (offer/answer/ICE) which must not echo back to sender.
func (server *Server) wsRelayToOthers(s *melody.Session, gameID pgtype.UUID, msg []byte) {
	server.melody.BroadcastFilter(msg, func(other *melody.Session) bool {
		return other != s && uuidEq(wsGameID(other), gameID)
	})
}

// wsBroadcastToGame sends a message to ALL sessions in the game room.
func (server *Server) wsBroadcastToGame(gameID pgtype.UUID, msg []byte) {
	server.melody.BroadcastFilter(msg, func(s *melody.Session) bool {
		return uuidEq(wsGameID(s), gameID)
	})
}

// ── Session key helpers ───────────────────────────────────────────────────────

func wsUser(s *melody.Session) db.User {
	v, _ := s.Get(wsKeyUser)
	return v.(db.User)
}

func wsGameID(s *melody.Session) pgtype.UUID {
	v, _ := s.Get(wsKeyGameID)
	return v.(pgtype.UUID)
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
	v, _ := s.Get("ws_player_color")
	return v.(string)
}


func restoreGameState(game db.Game) *pieces.GameState {
    
    // restore board from game.BoardState if you serialize it
    return &pieces.GameState{
		CurrentPlayer: string(game.CurrentPlayer),
		MoveNumber: game.MoveCount,
		Status: game.State,
		InCheck: game.InCheck,
	}
}