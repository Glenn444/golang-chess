package api

import (
	"log/slog"
	"time"
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
	Opponent    string `json:"opponent" binding:"required,oneof=person stockfish"`
	TimeControl int    `json:"time_control" binding:"oneof=0 5 10 15 30 45 60"`
	Visibility  string `json:"visibility" binding:"required,oneof=public private"`
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

	// Users can have at most 1 active + 2 pending games.
	allGames, err := server.store.GetGamesByPlayerID(ctx, user.ID)
	if err != nil {
		handleDBError(ctx, err, WithLogArgs("createGame: GetGamesByPlayerID", "user_id", user.ID))
		return
	}
	var pending, active int
	for _, g := range allGames {
		switch g.State {
		case db.GameStateWaiting:
			pending++
		case db.GameStateActive:
			active++
		}
	}
	if active >= 1 {
		ctx.JSON(http.StatusConflict, errorMessage("you already have an active game — finish it before creating a new one"))
		return
	}
	if pending >= 2 {
		ctx.JSON(http.StatusConflict, errorMessage("you already have 2 pending games — delete one or wait for them to fill before creating a new one"))
		return
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
	timeMs := int64(req.TimeControl) * 60 * 1000 // 0 = unlimited
	gameState := &pieces.GameState{
		CurrentPlayer:        "w",
		Board:                board.Initialise_board(board.Create_board()),
		CapturedPieces:       make(map[string][]pieces.PieceInterface),
		UserColor:            req.PlayerColor,
		PlayAgainst:          req.Opponent,
		WhiteTimeRemainingMs: timeMs,
		BlackTimeRemainingMs: timeMs,
		LastMoveAt:           time.Now(),
		TimeoutCh:            make(chan struct{}),
	}

	// persist the initial board state with timer columns
	_, err = server.store.UpdateGameState(ctx, db.UpdateGameStateParams{
		ID:                   game.ID,
		State:                db.GameStateWaiting,
		InCheck:              false,
		CurrentPlayer:        db.PlayerColorW,
		MoveCount:            0,
		BoardState:           board.SerializeGameState(gameState),
		WhiteTimeRemainingMs: timeMs,
		BlackTimeRemainingMs: timeMs,
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

	// Only pending (waiting) games can be deleted. Active games must be
	// completed, and past games are kept for history.
	if game.State != db.GameStateWaiting {
		ctx.JSON(http.StatusConflict, errorMessage("only pending games can be deleted — active games must be completed, and finished games are kept for history"))
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

// @Summary      List public games
// @Description  Returns all public games waiting for a second player. Any authenticated user can join these.
// @Tags         Games
// @Accept       json
// @Produce      json
// @Success      200  {array}   api.GameResponse
// @Failure      500  {object}  map[string]string
// @Router       /games/public [get]
func (server *Server) listPublicGames(ctx *gin.Context) {
	games, err := server.store.ListPublicGames(ctx)
	if handleDBError(ctx, err, WithLogArgs("listPublicGames: failed")) {
		return
	}
	ctx.JSON(http.StatusOK, server.toGameResponses(ctx, games))
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

	// Send push notification to the player who was waiting.
	server.notifyOpponent(ctx, gameID, user.ID, user.Username)

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
		ID:                   gameID,
		State:                db.GameStateResign,
		InCheck:              game.InCheck,
		CurrentPlayer:        game.CurrentPlayer,
		MoveCount:            game.MoveCount,
		BoardState:           game.BoardState,
		WhiteTimeRemainingMs: game.WhiteTimeRemainingMs,
		BlackTimeRemainingMs: game.BlackTimeRemainingMs,
		EndedByPlayerID:      user.ID,
		EndReason:            "resign",
	})
	if handleDBError(ctx, err, WithLogArgs("resignGame: UpdateGameState", "game_id", ctx.Param("id"))) {
		return
	}

	// Cancel timeout watcher and evict from memory on resign.
	server.activeGamesMu.Lock()
	if gs, ok := server.activeGames[gameID]; ok {
		close(gs.TimeoutCh)
		gs.TimeoutCh = make(chan struct{})
		delete(server.activeGames, gameID)
	}
	server.activeGamesMu.Unlock()

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
