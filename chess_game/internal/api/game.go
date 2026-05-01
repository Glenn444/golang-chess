package api

import (
	"net/http"
	"strings"

	db "github.com/Glenn444/golang-chess/internal/db"
	"github.com/gin-gonic/gin"
)

type createGameReq struct{
	PlayerColor string `json:"player_color" binding:"required,len=1,oneof=w b"`
}

func (r *createGameReq)sanitizeCreateGameReq(){
	r.PlayerColor = strings.ToLower(r.PlayerColor)
}
func (server *Server) createGame(ctx *gin.Context) {
	var req createGameReq

	if err := ctx.ShouldBindJSON(&req); err != nil{
		ctx.JSON(http.StatusBadRequest,errorResponse(err))
		return
	}
	req.sanitizeCreateGameReq()

	user, ok := server.getCurrentUser(ctx)
	if !ok {
		return
	}

	var game db.Game
	var err error
	switch req.PlayerColor {
	case "w":
		game,err = server.store.CreateGameAsWhite(ctx,user.ID)
	case "b":
		game,err = server.store.CreateGameAsBlack(ctx,user.ID)
	}
	
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

	// prevent joining own game (check both slots)
	if uuidEq(game.WhitePlayerID, user.ID) || uuidEq(game.BlackPlayerID,user.ID){
		ctx.JSON(http.StatusForbidden, errorMessage(ErrCannotJoinOwnGame))
		return
	}

	// determine which slot is open
	var updated db.Game

	switch {
	case !game.WhitePlayerID.Valid:
		updated,err = server.store.JoinGameAsWhite(ctx,db.JoinGameAsWhiteParams{
			ID: gameID,
			WhitePlayerID: user.ID,
		})
	case !game.BlackPlayerID.Valid:
		updated,err = server.store.JoinGameAsBlack(ctx,db.JoinGameAsBlackParams{
			ID: gameID,
			BlackPlayerID: user.ID,
		})
	default:
		ctx.JSON(http.StatusConflict,errorMessage(ErrGameAlreadyFull))
		return
	}
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
