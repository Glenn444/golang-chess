package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Glenn444/golang-chess/internal/board"
	db "github.com/Glenn444/golang-chess/internal/db"
	"github.com/Glenn444/golang-chess/internal/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/olahol/melody"
)

const (
	wsKeyUser   = "ws_user"
	wsKeyGameID = "ws_game_id"
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
	if uuidEq(game.WhitePlayerID,user.ID){
		playerColor = "w"
	}else{
		playerColor = "b"
	}
	if err := server.melody.HandleRequestWithKeys(ctx.Writer, ctx.Request, map[string]any{
		wsKeyUser:   user,
		wsKeyGameID: gameID,
		wsKeyPlayerColor: playerColor,
	}); err != nil {
		slog.Error("ws: upgrade failed", "err", err)
	}
}

func (server *Server) wsOnConnect(s *melody.Session) {
	u := wsUser(s)
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

// wsHandleMove broadcasts a move to both players.
// TODO: validate the move via the board package, persist via store.CreateMove,
// then call store.UpdateGameState for check/checkmate/stalemate detection.
func (server *Server) wsHandleMove(s *melody.Session, gameID pgtype.UUID, payload json.RawMessage) {
	server.activeGamesMu.Lock()
	defer server.activeGamesMu.Unlock()

	gamestate,ok := server.activeGames[gameID]
	if !ok{
		wsWriteError(s,"game not found")
		return
	}
	var body struct {
		Move string `json:"move"` //e2e3
	}
	if err := json.Unmarshal(payload, &body); err != nil || body.Move == "" {
		wsWriteError(s, "invalid move payload")
		return
	}
	

	previousPlayer := gamestate.CurrentPlayer
	user := wsUser(s)

	playerColor := wsPlayerColor(s)
	if playerColor != gamestate.CurrentPlayer{
		wsWriteError(s,"not your turn")
		return
	}
	//enforce turn
	if playerColor != gamestate.CurrentPlayer{
		wsWriteError(s,"not your turn")
		return
	}
	
	err := board.Move(gamestate,body.Move)
	if err != nil{
		wsWriteError(s,err.Error())
		return
	}

	//increment move number after successful move
	gamestate.MoveNumber++

	_,err = server.store.CreateMove(s.Request.Context(),db.CreateMoveParams{
		GameID: gameID,
		PlayerID: user.ID,
		PlayerColor: db.PlayerColor(previousPlayer),
		MoveNotation: body.Move,
		MoveNumber: gamestate.MoveNumber,
	})


	out, _ := json.Marshal(WSEvent{Type: EventMakeMove, Payload: payload})
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

func wsPlayerColor(s *melody.Session)string{
	v,_ := s.Get("ws_player_color")
	return v.(string)
}