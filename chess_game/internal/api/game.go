package api

import (
	"net/http"
	"strings"

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
// @Success      201  {object}  object
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

	ctx.JSON(http.StatusCreated, game)
}

// @Summary      List waiting games
// @Description  Returns all games waiting for a second player.
// @Tags         Games
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {array}   object
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /games [get]
func (server *Server) listWaitingGames(ctx *gin.Context) {
	games, err := server.store.ListWaitingGames(ctx)
	if handleDBError(ctx, err, WithLogArgs("listWaitingGames: failed")) {
		return
	}
	ctx.JSON(http.StatusOK, games)
}

// @Summary      List my games
// @Description  Returns all games the current user is participating in.
// @Tags         Games
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Success      200  {array}   object
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
	ctx.JSON(http.StatusOK, games)
}

// @Summary      Get a game
// @Description  Returns a game by its UUID.
// @Tags         Games
// @Accept       json
// @Produce      json
// @Param        id  path  string  true  "Game UUID"
// @Security     Bearer
// @Success      200  {object}  object
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

	ctx.JSON(http.StatusOK, game)
}

// @Summary      Join a game
// @Description  Joins a waiting game as the second player.
// @Tags         Games
// @Accept       json
// @Produce      json
// @Param        id  path  string  true  "Game UUID"
// @Security     Bearer
// @Success      200  {object}  object
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

	ctx.JSON(http.StatusOK, updated)
}

// @Summary      Resign from a game
// @Description  Resigns from an active game.
// @Tags         Games
// @Accept       json
// @Produce      json
// @Param        id  path  string  true  "Game UUID"
// @Security     Bearer
// @Success      200  {object}  object
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

	ctx.JSON(http.StatusOK, updated)
}

// @Summary      Get game moves
// @Description  Returns the move history for a game.
// @Tags         Games
// @Accept       json
// @Produce      json
// @Param        id  path  string  true  "Game UUID"
// @Security     Bearer
// @Success      200  {array}   object
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

	ctx.JSON(http.StatusOK, moves)
}
