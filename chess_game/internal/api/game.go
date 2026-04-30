package api

import (
	"net/http"

	db "github.com/Glenn444/golang-chess/internal/db"
	"github.com/gin-gonic/gin"
)

func (server *Server) createGame(ctx *gin.Context) {
	user, ok := server.getCurrentUser(ctx)
	if !ok {
		return
	}

	game, err := server.store.CreateGame(ctx, user.ID)
	if handleDBError(ctx, err, WithLogArgs("createGame: failed", "user_id", user.ID)) {
		return
	}

	ctx.JSON(http.StatusCreated, game)
}

func (server *Server) listWaitingGames(ctx *gin.Context) {
	games, err := server.store.ListWaitingGames(ctx)
	if handleDBError(ctx, err, WithLogArgs("listWaitingGames: failed")) {
		return
	}
	ctx.JSON(http.StatusOK, games)
}

func (server *Server) listMyGames(ctx *gin.Context) {
	user, ok := server.getCurrentUser(ctx)
	if !ok {
		return
	}

	games, err := server.store.GetGamesByPlayerID(ctx, user.ID)
	if handleDBError(ctx, err, WithLogArgs("listMyGames: failed", "user_id", user.ID)) {
		return
	}
	ctx.JSON(http.StatusOK, games)
}

func (server *Server) getGame(ctx *gin.Context) {
	gameID, ok := parseUUIDParam(ctx, "id")
	if !ok {
		return
	}

	game, err := server.store.GetGameByID(ctx, gameID)
	if handleDBError(ctx, err,
		WithNotFoundMsg(ErrGameNotFound),
		WithLogArgs("getGame: failed", "game_id", ctx.Param("id"))) {
		return
	}

	ctx.JSON(http.StatusOK, game)
}

func (server *Server) joinGame(ctx *gin.Context) {
	gameID, ok := parseUUIDParam(ctx, "id")
	if !ok {
		return
	}

	user, ok := server.getCurrentUser(ctx)
	if !ok {
		return
	}

	game, err := server.store.GetGameByID(ctx, gameID)
	if handleDBError(ctx, err,
		WithNotFoundMsg(ErrGameNotFound),
		WithLogArgs("joinGame: GetGameByID", "game_id", ctx.Param("id"))) {
		return
	}

	if uuidEq(game.WhitePlayerID, user.ID) {
		ctx.JSON(http.StatusForbidden, errorMessage(ErrCannotJoinOwnGame))
		return
	}

	updated, err := server.store.JoinGame(ctx, db.JoinGameParams{
		ID:            gameID,
		BlackPlayerID: user.ID,
	})
	if handleDBError(ctx, err,
		WithNotFoundMsg(ErrGameNotJoinable),
		WithLogArgs("joinGame: JoinGame", "game_id", ctx.Param("id"), "user_id", user.ID)) {
		return
	}

	ctx.JSON(http.StatusOK, updated)
}

func (server *Server) resignGame(ctx *gin.Context) {
	gameID, ok := parseUUIDParam(ctx, "id")
	if !ok {
		return
	}

	user, ok := server.getCurrentUser(ctx)
	if !ok {
		return
	}

	game, err := server.store.GetGameByID(ctx, gameID)
	if handleDBError(ctx, err,
		WithNotFoundMsg(ErrGameNotFound),
		WithLogArgs("resignGame: GetGameByID", "game_id", ctx.Param("id"))) {
		return
	}

	if !uuidEq(game.WhitePlayerID, user.ID) && !uuidEq(game.BlackPlayerID, user.ID) {
		ctx.JSON(http.StatusForbidden, errorMessage(ErrNotAPlayer))
		return
	}

	if game.State != db.GameStateActive {
		ctx.JSON(http.StatusConflict, errorMessage(ErrGameNotActive))
		return
	}

	updated, err := server.store.UpdateGameState(ctx, db.UpdateGameStateParams{
		ID:      gameID,
		State:   db.GameStateResign,
		InCheck: false,
	})
	if handleDBError(ctx, err, WithLogArgs("resignGame: UpdateGameState", "game_id", ctx.Param("id"))) {
		return
	}

	ctx.JSON(http.StatusOK, updated)
}

func (server *Server) getGameMoves(ctx *gin.Context) {
	gameID, ok := parseUUIDParam(ctx, "id")
	if !ok {
		return
	}

	moves, err := server.store.GetMovesByGameID(ctx, gameID)
	if handleDBError(ctx, err,
		WithNotFoundMsg(ErrGameNotFound),
		WithLogArgs("getGameMoves: failed", "game_id", ctx.Param("id"))) {
		return
	}

	ctx.JSON(http.StatusOK, moves)
}
