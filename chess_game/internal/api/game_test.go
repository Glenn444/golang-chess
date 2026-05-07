package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Glenn444/golang-chess/config"
	db "github.com/Glenn444/golang-chess/internal/db"
	mock_db "github.com/Glenn444/golang-chess/internal/db/mock"
	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/internal/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// ── Test helpers ─────────────────────────────────────────────────────────────────

func newTestGameServer(t *testing.T) (*Server, *mock_db.MockStore) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	store := mock_db.NewMockStore(ctrl)
	tokenMaker, _ := token.NewJWTMaker("12345678901234567890123456789012")
	server := &Server{
		config: config.Config{
			TokenSymmetricKey: "12345678901234567890123456789012",
		},
		tokenMaker:  tokenMaker,
		store:       store,
		activeGames: make(map[pgtype.UUID]*pieces.GameState),
	}
	return server, store
}

func setAuth(ctx *gin.Context, username string) {
	payload := &token.Payload{
		Username: username,
		ID:       uuid.New(),
	}
	ctx.Set(authorizationPayloadKey, payload)
}

func newGameCtx(method, path string, body any) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	var bodyReader *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(b)
	} else {
		bodyReader = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req
	return ctx, recorder
}

func userUUID() pgtype.UUID {
	return pgtype.UUID{Bytes: uuid.New(), Valid: true}
}

func gameUUID() pgtype.UUID {
	return pgtype.UUID{Bytes: uuid.New(), Valid: true}
}

func testUser() db.User {
	return db.User{
		ID:             userUUID(),
		Username:       "testuser",
		Email:          "test@example.com",
		PasswordHash:   "hashed",
		EmailConfirmed: true,
		IsActive:       true,
	}
}

func testGame() db.Game {
	return db.Game{
		ID:            gameUUID(),
		WhitePlayerID: userUUID(),
		BlackPlayerID: pgtype.UUID{Valid: false},
		State:         db.GameStateWaiting,
	}
}

// ── createGame tests ─────────────────────────────────────────────────────────────

func TestCreateGame(t *testing.T) {
	t.Run("success as white", func(t *testing.T) {
		server, store := newTestGameServer(t)

		user := testUser()
		game := testGame()

		store.EXPECT().GetUserByUsername(gomock.Any(), user.Username).Return(user, nil)
		store.EXPECT().CreateGameAsWhite(gomock.Any(), user.ID).Return(game, nil)
		store.EXPECT().UpdateGameState(gomock.Any(), gomock.Any()).Return(db.Game{}, nil)

		ctx, rec := newGameCtx(http.MethodPost, "/games", CreateGameReq{
			PlayerColor: "w",
			Opponent:    "person",
		})
		setAuth(ctx, user.Username)
		ctx.Params = gin.Params{} // under /games group

		server.createGame(ctx)

		require.Equal(t, http.StatusCreated, rec.Code)

		var created db.Game
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
		require.Equal(t, game.ID, created.ID)
		require.Equal(t, db.GameStateWaiting, created.State)

		// verify in-memory game was created
		server.activeGamesMu.RLock()
		memGame, ok := server.activeGames[game.ID]
		server.activeGamesMu.RUnlock()
		require.True(t, ok)
		require.Equal(t, "w", memGame.CurrentPlayer)
	})

	t.Run("success as black", func(t *testing.T) {
		server, store := newTestGameServer(t)

		user := testUser()
		game := testGame()
		game.BlackPlayerID = user.ID

		store.EXPECT().GetUserByUsername(gomock.Any(), user.Username).Return(user, nil)
		store.EXPECT().CreateGameAsBlack(gomock.Any(), user.ID).Return(game, nil)
		store.EXPECT().UpdateGameState(gomock.Any(), gomock.Any()).Return(db.Game{}, nil)

		ctx, rec := newGameCtx(http.MethodPost, "/games", CreateGameReq{
			PlayerColor: "b",
			Opponent:    "person",
		})
		setAuth(ctx, user.Username)

		server.createGame(ctx)

		require.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		server, _ := newTestGameServer(t)

		// ShouldBindJSON fails before getCurrentUser, so no mock calls needed
		ctx, rec := newGameCtx(http.MethodPost, "/games", gin.H{"player_color": 123})
		// still need auth set to avoid MustGet panic — though we don't reach it,
		// set it defensively
		setAuth(ctx, "testuser")

		server.createGame(ctx)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

// ── listWaitingGames tests ───────────────────────────────────────────────────────

func TestListWaitingGames(t *testing.T) {
	t.Run("success with games", func(t *testing.T) {
		server, store := newTestGameServer(t)

		games := []db.Game{testGame(), testGame()}

		store.EXPECT().ListWaitingGames(gomock.Any()).Return(games, nil)

		ctx, rec := newGameCtx(http.MethodGet, "/games", nil)
		setAuth(ctx, "testuser")

		server.listWaitingGames(ctx)

		require.Equal(t, http.StatusOK, rec.Code)

		var returned []db.Game
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &returned))
		require.Len(t, returned, 2)
	})

	t.Run("success empty list", func(t *testing.T) {
		server, store := newTestGameServer(t)

		store.EXPECT().ListWaitingGames(gomock.Any()).Return([]db.Game{}, nil)

		ctx, rec := newGameCtx(http.MethodGet, "/games", nil)
		setAuth(ctx, "testuser")

		server.listWaitingGames(ctx)

		require.Equal(t, http.StatusOK, rec.Code)
	})
}

// ── listMyGames tests ─────────────────────────────────────────────────────────────

func TestListMyGames(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, store := newTestGameServer(t)
		user := testUser()

		games := []db.Game{testGame()}

		store.EXPECT().GetUserByUsername(gomock.Any(), user.Username).Return(user, nil)
		store.EXPECT().GetGamesByPlayerID(gomock.Any(), user.ID).Return(games, nil)

		ctx, rec := newGameCtx(http.MethodGet, "/games/mine", nil)
		setAuth(ctx, user.Username)

		server.listMyGames(ctx)

		require.Equal(t, http.StatusOK, rec.Code)

		var returned []db.Game
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &returned))
		require.Len(t, returned, 1)
	})
}

// ── getGame tests ─────────────────────────────────────────────────────────────────

func TestGetGame(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, store := newTestGameServer(t)
		g := testGame()

		store.EXPECT().GetGameByID(gomock.Any(), g.ID).Return(g, nil)

		ctx, rec := newGameCtx(http.MethodGet, "/games/"+uuid.UUID(g.ID.Bytes).String(), nil)
		setAuth(ctx, "testuser")
		ctx.Params = gin.Params{{Key: "id", Value: uuid.UUID(g.ID.Bytes).String()}}

		server.getGame(ctx)

		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("invalid uuid", func(t *testing.T) {
		server, _ := newTestGameServer(t)

		ctx, rec := newGameCtx(http.MethodGet, "/games/invalid", nil)
		setAuth(ctx, "testuser")
		ctx.Params = gin.Params{{Key: "id", Value: "invalid"}}

		server.getGame(ctx)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("not found", func(t *testing.T) {
		server, store := newTestGameServer(t)
		id := gameUUID()

		store.EXPECT().GetGameByID(gomock.Any(), id).Return(db.Game{}, pgxNoRows())

		ctx, rec := newGameCtx(http.MethodGet, "/games/"+uuid.UUID(id.Bytes).String(), nil)
		setAuth(ctx, "testuser")
		ctx.Params = gin.Params{{Key: "id", Value: uuid.UUID(id.Bytes).String()}}

		server.getGame(ctx)

		require.Equal(t, http.StatusNotFound, rec.Code)
	})
}

// ── joinGame tests ────────────────────────────────────────────────────────────────

func TestJoinGame(t *testing.T) {
	t.Run("success join as black", func(t *testing.T) {
		server, store := newTestGameServer(t)
		user := testUser()
		g := testGame()
		g.WhitePlayerID = userUUID() // someone else is white

		joined := g
		joined.BlackPlayerID = user.ID
		joined.State = db.GameStateActive

		store.EXPECT().GetUserByUsername(gomock.Any(), user.Username).Return(user, nil)
		store.EXPECT().GetGameByID(gomock.Any(), g.ID).Return(g, nil)
		store.EXPECT().JoinGameAsBlack(gomock.Any(), gomock.Any()).Return(joined, nil)

		ctx, rec := newGameCtx(http.MethodPost, "/games/"+uuid.UUID(g.ID.Bytes).String()+"/join", nil)
		setAuth(ctx, user.Username)
		ctx.Params = gin.Params{{Key: "id", Value: uuid.UUID(g.ID.Bytes).String()}}

		server.joinGame(ctx)

		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("cannot join own game", func(t *testing.T) {
		server, store := newTestGameServer(t)
		user := testUser()
		g := testGame()
		g.WhitePlayerID = user.ID // user IS the white player

		store.EXPECT().GetUserByUsername(gomock.Any(), user.Username).Return(user, nil)
		store.EXPECT().GetGameByID(gomock.Any(), g.ID).Return(g, nil)

		ctx, rec := newGameCtx(http.MethodPost, "/games/"+uuid.UUID(g.ID.Bytes).String()+"/join", nil)
		setAuth(ctx, user.Username)
		ctx.Params = gin.Params{{Key: "id", Value: uuid.UUID(g.ID.Bytes).String()}}

		server.joinGame(ctx)

		require.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("game already full", func(t *testing.T) {
		server, store := newTestGameServer(t)
		user := testUser()
		g := testGame()
		g.WhitePlayerID = userUUID()
		g.BlackPlayerID = userUUID() // both slots filled

		store.EXPECT().GetUserByUsername(gomock.Any(), user.Username).Return(user, nil)
		store.EXPECT().GetGameByID(gomock.Any(), g.ID).Return(g, nil)

		ctx, rec := newGameCtx(http.MethodPost, "/games/"+uuid.UUID(g.ID.Bytes).String()+"/join", nil)
		setAuth(ctx, user.Username)
		ctx.Params = gin.Params{{Key: "id", Value: uuid.UUID(g.ID.Bytes).String()}}

		server.joinGame(ctx)

		require.Equal(t, http.StatusConflict, rec.Code)
	})
}

// ── resignGame tests ──────────────────────────────────────────────────────────────

func TestResignGame(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, store := newTestGameServer(t)
		user := testUser()
		g := testGame()
		g.WhitePlayerID = user.ID
		g.State = db.GameStateActive

		resigned := g
		resigned.State = db.GameStateResign

		store.EXPECT().GetUserByUsername(gomock.Any(), user.Username).Return(user, nil)
		store.EXPECT().GetGameByID(gomock.Any(), g.ID).Return(g, nil)
		store.EXPECT().UpdateGameState(gomock.Any(), gomock.Any()).Return(resigned, nil)

		ctx, rec := newGameCtx(http.MethodPost, "/games/"+uuid.UUID(g.ID.Bytes).String()+"/resign", nil)
		setAuth(ctx, user.Username)
		ctx.Params = gin.Params{{Key: "id", Value: uuid.UUID(g.ID.Bytes).String()}}

		server.resignGame(ctx)

		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("not a player", func(t *testing.T) {
		server, store := newTestGameServer(t)
		user := testUser()
		g := testGame()
		g.WhitePlayerID = userUUID() // different user
		g.BlackPlayerID = userUUID()

		store.EXPECT().GetUserByUsername(gomock.Any(), user.Username).Return(user, nil)
		store.EXPECT().GetGameByID(gomock.Any(), g.ID).Return(g, nil)

		ctx, rec := newGameCtx(http.MethodPost, "/games/"+uuid.UUID(g.ID.Bytes).String()+"/resign", nil)
		setAuth(ctx, user.Username)
		ctx.Params = gin.Params{{Key: "id", Value: uuid.UUID(g.ID.Bytes).String()}}

		server.resignGame(ctx)

		require.Equal(t, http.StatusForbidden, rec.Code)
	})
}

// ── getGameMoves tests ────────────────────────────────────────────────────────────

func TestGetGameMoves(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, store := newTestGameServer(t)
		id := gameUUID()

		moves := []db.GameMove{
			{ID: userUUID(), GameID: id, MoveNotation: "e4", MoveNumber: 1},
			{ID: userUUID(), GameID: id, MoveNotation: "e5", MoveNumber: 2},
		}

		store.EXPECT().GetMovesByGameID(gomock.Any(), id).Return(moves, nil)

		ctx, rec := newGameCtx(http.MethodGet, "/games/"+uuid.UUID(id.Bytes).String()+"/moves", nil)
		setAuth(ctx, "testuser")
		ctx.Params = gin.Params{{Key: "id", Value: uuid.UUID(id.Bytes).String()}}

		server.getGameMoves(ctx)

		require.Equal(t, http.StatusOK, rec.Code)
		var returned []db.GameMove
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &returned))
		require.Len(t, returned, 2)
	})
}

func pgxNoRows() error { return pgx.ErrNoRows }
