package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Glenn444/golang-chess/config"
	"github.com/Glenn444/golang-chess/internal/board"
	db "github.com/Glenn444/golang-chess/internal/db"
	mock_db "github.com/Glenn444/golang-chess/internal/db/mock"
	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/internal/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/olahol/melody"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newTestWSServer(t *testing.T) (*Server, *mock_db.MockStore, string) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	store := mock_db.NewMockStore(ctrl)
	maker, _ := token.NewJWTMaker("12345678901234567890123456789012")

	server := &Server{
		config: config.Config{
			TokenSymmetricKey:  "12345678901234567890123456789012",
			AcessTokenDuration: time.Hour,
		},
		tokenMaker:  maker,
		store:       store,
		melody:      melody.New(),
		activeGames: make(map[pgtype.UUID]*pieces.GameState),
	}
	server.setupMelody()

	return server, store, mustToken(maker, "testuser")
}

func mustToken(maker token.Maker, username string) string {
	t, _ := maker.CreateToken(username, token.AccessTokenType, time.Hour)
	return t
}

// ── handleWebSocket tests (returns JSON before upgrade) ─────────────────────────

func TestHandleWebSocket(t *testing.T) {
	t.Run("missing game_id", func(t *testing.T) {
		server, _, _ := newTestWSServer(t)
		ctx, rec := newGameCtx(http.MethodGet, "/ws", nil)
		server.handleWebSocket(ctx)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("invalid game_id format", func(t *testing.T) {
		server, _, _ := newTestWSServer(t)
		ctx, rec := newGameCtx(http.MethodGet, "/ws?game_id=not-a-uuid", nil)
		server.handleWebSocket(ctx)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("game not found", func(t *testing.T) {
		server, store, _ := newTestWSServer(t)
		gameID := gameUUID()

		store.EXPECT().GetGameByID(gomock.Any(), gameID).Return(db.Game{}, pgx.ErrNoRows)

		ctx, rec := newGameCtx(http.MethodGet,
			"/ws?game_id="+uuid.UUID(gameID.Bytes).String(), nil)
		server.handleWebSocket(ctx)
		require.Equal(t, http.StatusNotFound, rec.Code)
	})
}

// ── WebSocket integration tests ─────────────────────────────────────────────────

func setupWSRouter(t *testing.T, server *Server, store *mock_db.MockStore) (*httptest.Server, string, pgtype.UUID) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	gameID := gameUUID()
	whiteID := userUUID()

	user := db.User{
		ID:       whiteID,
		Username: "whiteplayer",
		Email:    "white@example.com",
	}

	// Mock DB calls that sendGameState makes to fetch opponent info.
	dbGame := db.Game{
		ID:            gameID,
		WhitePlayerID: whiteID,
		BlackPlayerID: userUUID(),
		State:         db.GameStateActive,
	}
	store.EXPECT().GetGameByID(gomock.Any(), gameID).Return(dbGame, nil).AnyTimes()
	store.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).Return(user, nil).AnyTimes()

	gameState := &pieces.GameState{
		CurrentPlayer:  "w",
		Board:          board.Initialise_board(board.Create_board()),
		CapturedPieces: make(map[string][]pieces.PieceInterface),
		PlayAgainst:    "person",
		UserColor:      "w",
	}
	server.activeGamesMu.Lock()
	server.activeGames[gameID] = gameState
	server.activeGamesMu.Unlock()

	server.router = router

	// Pre-authenticated: session already has user, gameID, color, and auth flag.
	router.GET("/ws", func(ctx *gin.Context) {
		if err := server.melody.HandleRequestWithKeys(ctx.Writer, ctx.Request, map[string]any{
			wsKeyUser:          user,
			wsKeyGameID:        gameID,
			wsKeyPlayerColor:   "w",
			wsKeyAuthenticated: true,
		}); err != nil {
			slog.Error("ws: upgrade failed", "err", err)
		}
	})

	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)

	return srv, mustToken(server.tokenMaker, "whiteplayer"), gameID
}

func drainGameState(t *testing.T, conn *websocket.Conn) {
	t.Helper()
	var evt WSEvent
	if err := conn.ReadJSON(&evt); err != nil {
		return
	}
	// Accept both raw game_state and GameStatePayload-wrapped formats.
	if evt.Type == "game_state" {
		return
	}
	t.Fatalf("expected game_state event on connect, got %q", evt.Type)
}


func TestWSMakeMoveIntegration(t *testing.T) {
	t.Run("successful move broadcasts result", func(t *testing.T) {
		server, store, _ := newTestWSServer(t)
		srv, _, gameID := setupWSRouter(t, server, store)

		store.EXPECT().UpdateGameState(gomock.Any(), gomock.Any()).Return(db.Game{}, nil)
		store.EXPECT().CreateMove(gomock.Any(), gomock.Any()).Return(db.GameMove{}, nil)

		wsURL := wsURL(srv.URL) + "/ws?dummy=1"
		conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
		defer conn.Close()

		_ = gameID
		drainGameState(t, conn)

		moveEvent := WSEvent{
			Type:    EventMakeMove,
			Payload: json.RawMessage(`{"move":"e4"}`),
		}
		err = conn.WriteJSON(moveEvent)
		require.NoError(t, err)

		var respEvent WSEvent
		err = conn.ReadJSON(&respEvent)
		require.NoError(t, err)
		require.Equal(t, EventMakeMove, respEvent.Type)

		var moveResult MoveResult
		require.NoError(t, json.Unmarshal(respEvent.Payload, &moveResult))
		require.Equal(t, "e4", moveResult.Move)
		require.Equal(t, "b", moveResult.CurrentPlayer)
		require.False(t, moveResult.InCheck)
		require.False(t, moveResult.IsCheckmate)
		// Regression: unlimited games (both clocks zero) must not end in a
		// phantom "timeout" on the first move.
		require.Equal(t, "", moveResult.EndReason)
	})

	t.Run("move rejected after game is over", func(t *testing.T) {
		server, store, _ := newTestWSServer(t)
		srv, _, gameID := setupWSRouter(t, server, store)

		server.activeGamesMu.Lock()
		server.activeGames[gameID].Status = db.GameStateResign
		server.activeGamesMu.Unlock()

		wsURL := wsURL(srv.URL) + "/ws?dummy=1"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		drainGameState(t, conn)

		moveEvent := WSEvent{
			Type:    EventMakeMove,
			Payload: json.RawMessage(`{"move":"e4"}`),
		}
		require.NoError(t, conn.WriteJSON(moveEvent))

		var respEvent WSEvent
		require.NoError(t, conn.ReadJSON(&respEvent))
		require.Equal(t, EventError, respEvent.Type)

		var errPayload map[string]string
		require.NoError(t, json.Unmarshal(respEvent.Payload, &errPayload))
		require.Equal(t, "game is already over", errPayload["error"])
	})

	t.Run("not your turn returns error", func(t *testing.T) {
		server, store, _ := newTestWSServer(t)
		srv, _, gameID := setupWSRouter(t, server, store)

		store.EXPECT().UpdateGameState(gomock.Any(), gomock.Any()).Return(db.Game{}, nil).AnyTimes()
		store.EXPECT().CreateMove(gomock.Any(), gomock.Any()).Return(db.GameMove{}, nil).AnyTimes()

		server.activeGamesMu.Lock()
		server.activeGames[gameID].GameStateMu.Lock()
		server.activeGames[gameID].CurrentPlayer = "b"
		server.activeGames[gameID].GameStateMu.Unlock()
		server.activeGamesMu.Unlock()

		wsURL := wsURL(srv.URL) + "/ws?dummy=1"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		drainGameState(t, conn)

		moveEvent := WSEvent{
			Type:    EventMakeMove,
			Payload: json.RawMessage(`{"move":"e4"}`),
		}
		err = conn.WriteJSON(moveEvent)
		require.NoError(t, err)

		var respEvent WSEvent
		err = conn.ReadJSON(&respEvent)
		require.NoError(t, err)
		require.Equal(t, EventError, respEvent.Type)

		var errPayload map[string]string
		require.NoError(t, json.Unmarshal(respEvent.Payload, &errPayload))
		require.Equal(t, "not your turn", errPayload["error"])
	})

	t.Run("invalid move returns error", func(t *testing.T) {
		server, store, _ := newTestWSServer(t)
		srv, _, gameID := setupWSRouter(t, server, store)

		store.EXPECT().UpdateGameState(gomock.Any(), gomock.Any()).Return(db.Game{}, nil).AnyTimes()
		store.EXPECT().CreateMove(gomock.Any(), gomock.Any()).Return(db.GameMove{}, nil).AnyTimes()

		wsURL := wsURL(srv.URL) + "/ws?dummy=1"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		_ = gameID
		drainGameState(t, conn)

		moveEvent := WSEvent{
			Type:    EventMakeMove,
			Payload: json.RawMessage(`{"move":"Ke5"}`),
		}
		err = conn.WriteJSON(moveEvent)
		require.NoError(t, err)

		var respEvent WSEvent
		err = conn.ReadJSON(&respEvent)
		require.NoError(t, err)
		require.Equal(t, EventError, respEvent.Type)
	})
}

func TestWSHandleChatIntegration(t *testing.T) {
	t.Run("successful chat message broadcast", func(t *testing.T) {
		server, store, _ := newTestWSServer(t)
		srv, _, gameID := setupWSRouter(t, server, store)

		chatMsg := db.ChatMessage{
			ID:       userUUID(),
			GameID:   gameID,
			SenderID: userUUID(),
			Content:  "hello!",
		}
		store.EXPECT().CreateChatMessage(gomock.Any(), gomock.Any()).Return(chatMsg, nil)

		wsURL := wsURL(srv.URL) + "/ws?dummy=1"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		drainGameState(t, conn)

		chatEvent := WSEvent{
			Type:    EventChat,
			Payload: json.RawMessage(`{"content":"hello!"}`),
		}
		err = conn.WriteJSON(chatEvent)
		require.NoError(t, err)

		var respEvent WSEvent
		err = conn.ReadJSON(&respEvent)
		require.NoError(t, err)
		require.Equal(t, EventChat, respEvent.Type)
	})

	t.Run("empty content returns error", func(t *testing.T) {
		server, store, _ := newTestWSServer(t)
		srv, _, _ := setupWSRouter(t, server, store)

		wsURL := wsURL(srv.URL) + "/ws?dummy=1"
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		drainGameState(t, conn)

		chatEvent := WSEvent{
			Type:    EventChat,
			Payload: json.RawMessage(`{"content":""}`),
		}
		err = conn.WriteJSON(chatEvent)
		require.NoError(t, err)

		var respEvent WSEvent
		err = conn.ReadJSON(&respEvent)
		require.NoError(t, err)
		require.Equal(t, EventError, respEvent.Type)
	})
}

func wsURL(httpURL string) string {
	return "ws" + strings.TrimPrefix(httpURL, "http")
}
