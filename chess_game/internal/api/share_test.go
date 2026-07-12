package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Glenn444/golang-chess/internal/board"
	db "github.com/Glenn444/golang-chess/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func shareCtx(t *testing.T, path, id string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	ctx, rec := newGameCtx(http.MethodGet, path, nil)
	ctx.Params = gin.Params{{Key: "id", Value: id}}
	return ctx, rec
}

func TestInvitePage(t *testing.T) {
	t.Run("waiting game renders challenge with meta tags", func(t *testing.T) {
		server, store := newTestGameServer(t)
		game := testGame()
		game.WhiteTimeRemainingMs = 10 * 60 * 1000
		game.BlackTimeRemainingMs = 10 * 60 * 1000

		store.EXPECT().GetGameByID(gomock.Any(), game.ID).Return(game, nil)

		ctx, rec := shareCtx(t, "/invite/"+uidStr(game.ID), uidStr(game.ID))
		server.invitePage(ctx)

		require.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		require.Contains(t, body, "someuser") // creator name from mock lookup
		require.Contains(t, body, `property="og:title"`)
		require.Contains(t, body, "/og/invite/"+uidStr(game.ID)+".png")
		require.Contains(t, body, "/play/"+uidStr(game.ID)+"?join=true")
		require.Contains(t, body, "10 min game")
		require.Contains(t, body, `name="robots" content="noindex"`)
	})

	t.Run("unknown game renders branded 404", func(t *testing.T) {
		server, _ := newTestGameServer(t)
		ctx, rec := shareCtx(t, "/invite/nonsense", "nonsense")
		server.invitePage(ctx)
		require.Equal(t, http.StatusNotFound, rec.Code)
		require.Contains(t, rec.Body.String(), "This game doesn't exist")
	})

	t.Run("engine game renders 404", func(t *testing.T) {
		server, store := newTestGameServer(t)
		game := testGame()
		game.Opponent = "stockfish"
		store.EXPECT().GetGameByID(gomock.Any(), game.ID).Return(game, nil)
		ctx, rec := shareCtx(t, "/invite/"+uidStr(game.ID), uidStr(game.ID))
		server.invitePage(ctx)
		require.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestSpectatePage(t *testing.T) {
	t.Run("public active game renders live page with board", func(t *testing.T) {
		server, store := newTestGameServer(t)
		game := testGame()
		game.State = db.GameStateActive
		game.BlackPlayerID = game.WhitePlayerID
		game.BoardState = board.BuildInitialBoardState()
		game.MoveCount = 4

		store.EXPECT().GetGameByID(gomock.Any(), game.ID).Return(game, nil)

		ctx, rec := shareCtx(t, "/game/"+uidStr(game.ID), uidStr(game.ID))
		server.spectatePage(ctx)

		require.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		require.Contains(t, body, "LIVE")
		require.Contains(t, body, "cburnett/bR.svg") // rendered pieces (app's SVG set)
		require.Contains(t, body, "spectate=1")
		require.Contains(t, body, `property="og:image"`)
		// 64 server-rendered squares (the live-update script contains the same
		// markup once as a JS string literal).
		require.GreaterOrEqual(t, strings.Count(body, `<div class="sq `), 64)
	})

	t.Run("private game renders locked page without details", func(t *testing.T) {
		server, store := newTestGameServer(t)
		game := testGame()
		game.Visibility = "private"
		store.EXPECT().GetGameByID(gomock.Any(), game.ID).Return(game, nil)

		ctx, rec := shareCtx(t, "/game/"+uidStr(game.ID), uidStr(game.ID))
		server.spectatePage(ctx)

		require.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		require.Contains(t, body, "This game is private")
		require.Contains(t, body, `name="robots" content="noindex"`)
		require.NotContains(t, body, "someuser") // no player names leak
	})

	t.Run("finished game shows result text", func(t *testing.T) {
		server, store := newTestGameServer(t)
		game := testGame()
		game.State = db.GameStateCheckmate
		game.CurrentPlayer = db.PlayerColor("b") // black got mated
		game.BlackPlayerID = game.WhitePlayerID
		game.BoardState = board.BuildInitialBoardState()
		store.EXPECT().GetGameByID(gomock.Any(), game.ID).Return(game, nil)

		ctx, rec := shareCtx(t, "/game/"+uidStr(game.ID), uidStr(game.ID))
		server.spectatePage(ctx)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, rec.Body.String(), "Checkmate — someuser won")
	})
}

func TestOGCards(t *testing.T) {
	t.Run("game card renders a PNG", func(t *testing.T) {
		server, store := newTestGameServer(t)
		game := testGame()
		game.State = db.GameStateActive
		game.BlackPlayerID = game.WhitePlayerID
		game.BoardState = board.BuildInitialBoardState()
		store.EXPECT().GetGameByID(gomock.Any(), game.ID).Return(game, nil)

		ctx, rec := shareCtx(t, "/og/game/"+uidStr(game.ID)+".png", uidStr(game.ID)+".png")
		server.ogGameCard(ctx)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "image/png", rec.Header().Get("Content-Type"))
		require.Greater(t, rec.Body.Len(), 10_000) // a real card, not an error
		require.Equal(t, "\x89PNG", rec.Body.String()[:4])
		require.Contains(t, rec.Header().Get("Cache-Control"), "max-age=60")
	})

	t.Run("invite card for unknown id degrades to generic brand card", func(t *testing.T) {
		server, _ := newTestGameServer(t)
		ctx, rec := shareCtx(t, "/og/invite/junk.png", "junk.png")
		server.ogInviteCard(ctx)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "image/png", rec.Header().Get("Content-Type"))
	})

	t.Run("private game card is locked", func(t *testing.T) {
		server, store := newTestGameServer(t)
		game := testGame()
		game.Visibility = "private"
		store.EXPECT().GetGameByID(gomock.Any(), game.ID).Return(game, nil)

		ctx, rec := shareCtx(t, "/og/game/"+uidStr(game.ID)+".png", uidStr(game.ID)+".png")
		server.ogGameCard(ctx)
		require.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, "image/png", rec.Header().Get("Content-Type"))
	})
}

func TestListLiveGames(t *testing.T) {
	server, store := newTestGameServer(t)
	rows := []db.ListLiveGamesRow{{
		ID:            gameUUID(),
		WhiteUsername: "glenn",
		BlackUsername: "kev",
		MoveCount:     12,
		CurrentPlayer: db.PlayerColor("w"),
	}}
	store.EXPECT().ListLiveGames(gomock.Any()).Return(rows, nil)

	ctx, rec := newGameCtx(http.MethodGet, "/games/live", nil)
	server.listLiveGames(ctx)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), `"white_username":"glenn"`)
	require.Contains(t, rec.Body.String(), `"black_username":"kev"`)
}
