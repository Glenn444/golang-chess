package api

import (
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/Glenn444/golang-chess/internal/board"
	db "github.com/Glenn444/golang-chess/internal/db"
	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/gin-gonic/gin"
)

type CreateGameReq struct{
	PlayerColor string `json:"player_color" binding:"required,len=1,oneof=w b"`
	Opponent string `json:"opponent" binding:"required,oneof=person stockfish"`
}

func (r *CreateGameReq)sanitizeCreateGameReq(){
	r.PlayerColor = strings.ToLower(r.PlayerColor)
}

//create a chess game
// @Summary      Create a game
// @Description  Creates a new chess game. Choose your color (w/b) and opponent type.
// @Tags         Games
// @Accept       json
// @Produce      json
// @Param        body  body  CreateGameReq  true  "Game creation payload"
// @Security     Bearer
// @Success      201  {object}  api.GameResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /games [post]
func (server *Server) createGame(ctx *gin.Context) {
	var req CreateGameReq

	if err := ctx.ShouldBindJSON(&req); err != nil{
		ctx.JSON(http.StatusBadRequest,errorResponse(err))
		return
	}
	req.sanitizeCreateGameReq()

	user, ok := server.getCurrentUser(ctx)
	if !ok {
		return
	}

	// Serialize game creation per user to prevent race-condition duplicates.
	server.createGameMUsMu.Lock()
	if _, exists := server.createGameMUs[user.ID]; !exists {
		server.createGameMUs[user.ID] = &sync.Mutex{}
	}
	mu := server.createGameMUs[user.ID]
	server.createGameMUsMu.Unlock()

	mu.Lock()
	defer mu.Unlock()

	// Fetch all games for this user and reject if any are active or waiting.
	allGames, err := server.store.GetGamesByPlayerID(ctx, user.ID)
	if err != nil {
		handleDBError(ctx, err, WithLogArgs("createGame: GetGamesByPlayerID", "user_id", user.ID))
		return
	}
	for _, g := range allGames {
		if g.State == db.GameStateWaiting || g.State == db.GameStateActive {
			ctx.JSON(http.StatusConflict, errorMessage("you already have an active or pending game — finish or delete it before creating a new one"))
			return
		}
	}

	var game db.Game
	switch req.PlayerColor {
	case "w":
		game,err = server.store.CreateGameAsWhite(ctx,user.ID)
	case "b":
		game,err = server.store.CreateGameAsBlack(ctx,user.ID)
	}
	
	if handleDBError(ctx, err, WithLogArgs("createGame: failed", "user_id", user.ID)) {
		return
	}

	//initialise a game state in memory
	gameState := &pieces.GameState{
		CurrentPlayer: "w",
		Board: board.Initialise_board(board.Create_board()),
		CapturedPieces: make(map[string][]pieces.PieceInterface),
		UserColor: req.PlayerColor,
		PlayAgainst: req.Opponent,
	}

	// persist the initial board state
	_, err = server.store.UpdateGameState(ctx, db.UpdateGameStateParams{
		ID:            game.ID,
		State:         db.GameStateWaiting,
		InCheck:       false,
		CurrentPlayer: db.PlayerColorW,
		MoveCount:     0,
		BoardState:    board.SerializeGameState(gameState),
	})
	if handleDBError(ctx, err, WithLogArgs("createGame: failed to persist board state", "game_id", game.ID)) {
		return
	}

	//save the game in memory
	server.activeGamesMu.Lock()
	server.activeGames[game.ID] = gameState
	server.activeGamesMu.Unlock()

	ctx.JSON(http.StatusCreated, server.toGameResponse(ctx, game))
}

// @Summary      Delete a game
// @Description  Deletes a game and all associated moves/chat. Only a participant can delete their game.
// @Tags         Games
// @Accept       json
// @Produce      json
// @Param        id  path  string  true  "Game UUID"
// @Security     Bearer
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /games/{id} [delete]
func (server *Server) deleteGame(ctx *gin.Context) {
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
		WithLogArgs("deleteGame: GetGameByID", "game_id", ctx.Param("id"))) {
		return
	}

	// Only a participant can delete the game.
	if !uuidEq(game.WhitePlayerID, user.ID) && !uuidEq(game.BlackPlayerID, user.ID) {
		ctx.JSON(http.StatusForbidden, errorMessage(ErrNotAPlayer))
		return
	}

	// Clean up associated data.
	if err := server.store.DeleteMovesByGameID(ctx, gameID); err != nil {
		slog.Error("deleteGame: DeleteMovesByGameID", "game_id", ctx.Param("id"), "err", err)
	}
	if err := server.store.DeleteChatMessagesByGameID(ctx, gameID); err != nil {
		slog.Error("deleteGame: DeleteChatMessagesByGameID", "game_id", ctx.Param("id"), "err", err)
	}

	if err := server.store.DeleteGame(ctx, gameID); err != nil {
		handleDBError(ctx, err, WithLogArgs("deleteGame: DeleteGame", "game_id", ctx.Param("id")))
		return
	}

	// Remove from in-memory active games if present.
	server.activeGamesMu.Lock()
	delete(server.activeGames, gameID)
	server.activeGamesMu.Unlock()

	ctx.JSON(http.StatusOK, gin.H{"message": "game deleted"})
}

// @Summary      List waiting games
// @Description  Returns all games waiting for a second player.
// @Tags         Games
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {array}   api.GameResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /games [get]
func (server *Server) listWaitingGames(ctx *gin.Context) {
	games, err := server.store.ListWaitingGames(ctx)
	if handleDBError(ctx, err, WithLogArgs("listWaitingGames: failed")) {
		return
	}
	ctx.JSON(http.StatusOK, server.toGameResponses(ctx, games))
}

// @Summary      List my games
// @Description  Returns all games the current user is participating in.
// @Tags         Games
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {array}   api.GameResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /games/mine [get]
func (server *Server) listMyGames(ctx *gin.Context) {
	user, ok := server.getCurrentUser(ctx)
	if !ok {
		return
	}

	games, err := server.store.GetGamesByPlayerID(ctx, user.ID)
	if handleDBError(ctx, err, WithLogArgs("listMyGames: failed", "user_id", user.ID)) {
		return
	}
	ctx.JSON(http.StatusOK, server.toGameResponses(ctx, games))

}

// @Summary      Get a game
// @Description  Returns a game by its UUID.
// @Tags         Games
// @Accept       json
// @Produce      json
// @Param        id  path  string  true  "Game UUID"
// @Security     Bearer
// @Success      200  {object}  api.GameResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /games/{id} [get]
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

	ctx.JSON(http.StatusOK, server.toGameResponse(ctx, game))
}

// @Summary      Join a game
// @Description  Joins a waiting game as the second player.
// @Tags         Games
// @Accept       json
// @Produce      json
// @Param        id  path  string  true  "Game UUID"
// @Security     Bearer
// @Success      200  {object}  api.GameResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /games/{id}/join [post]
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

	ctx.JSON(http.StatusOK, server.toGameResponse(ctx, updated))
}

// @Summary      Resign from a game
// @Description  Resigns from an active game.
// @Tags         Games
// @Accept       json
// @Produce      json
// @Param        id  path  string  true  "Game UUID"
// @Security     Bearer
// @Success      200  {object}  api.GameResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /games/{id}/resign [post]
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

	ctx.JSON(http.StatusOK, server.toGameResponse(ctx, updated))
}

// @Summary      Get game moves
// @Description  Returns the move history for a game.
// @Tags         Games
// @Accept       json
// @Produce      json
// @Param        id  path  string  true  "Game UUID"
// @Security     Bearer
// @Success      200  {array}   api.GameMoveResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /games/{id}/moves [get]
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

	ctx.JSON(http.StatusOK, toGameMoveResponses(moves))
}
