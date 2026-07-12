package api

import (
	"net/http"
	"strings"
	"sync"
	"time"

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
	// UCI Skill Level for engine games: 0 (weakest) to 20 (full strength).
	StockfishLevel int `json:"stockfish_level" binding:"omitempty,min=0,max=20"`
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
	defer func() {
		mu.Unlock()
		// Drop the entry so the map doesn't grow with every user who ever
		// created a game over the server's lifetime.
		server.createGameMUsMu.Lock()
		delete(server.createGameMUs, user.ID)
		server.createGameMUsMu.Unlock()
	}()

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

	if req.Opponent == "stockfish" && !server.engineAvailable() {
		ctx.JSON(http.StatusServiceUnavailable, errorMessage("stockfish is not available on this server"))
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
		StockfishLevel:       int32(req.StockfishLevel),
		WhiteTimeRemainingMs: timeMs,
		BlackTimeRemainingMs: timeMs,
		LastMoveAt:           time.Now(),
		TimeoutCh:            make(chan struct{}),
	}

	// Engine games never wait for a human opponent — they start active and
	// stay out of the public lobby regardless of the requested visibility.
	visibility := req.Visibility
	state := db.GameStateWaiting
	if req.Opponent == "stockfish" {
		visibility = "private"
		state = db.GameStateActive
	}

	// Single insert: game row, visibility, opponent, board state and clocks.
	var game db.Game
	boardState := board.SerializeGameState(gameState)
	switch req.PlayerColor {
	case "w":
		game, err = server.store.CreateGameAsWhite(ctx, db.CreateGameAsWhiteParams{
			WhitePlayerID:        user.ID,
			Visibility:           visibility,
			BoardState:           boardState,
			WhiteTimeRemainingMs: timeMs,
			Opponent:             req.Opponent,
			StockfishLevel:       int16(req.StockfishLevel),
			State:                state,
		})
	case "b":
		game, err = server.store.CreateGameAsBlack(ctx, db.CreateGameAsBlackParams{
			BlackPlayerID:        user.ID,
			Visibility:           visibility,
			BoardState:           boardState,
			WhiteTimeRemainingMs: timeMs,
			Opponent:             req.Opponent,
			StockfishLevel:       int16(req.StockfishLevel),
			State:                state,
		})
	}

	if handleDBError(ctx, err, WithLogArgs("createGame: failed", "user_id", user.ID)) {
		return
	}

	//save the game in memory
	server.activeGamesMu.Lock()
	server.activeGames[game.ID] = gameState
	server.activeGamesMu.Unlock()

	// If the user plays black against the engine, white (the engine) opens.
	server.maybeTriggerEngine(game.ID, gameState)

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

	// Moves, chat messages and voice sessions reference games with
	// ON DELETE CASCADE, so a single delete cleans everything atomically.
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

	if !server.canViewGame(ctx, game) {
		return
	}

	ctx.JSON(http.StatusOK, server.toGameResponse(ctx, game))
}

// canViewGame allows participants always, and anyone for public games.
// Private games answer 404 to non-participants so their existence isn't leaked.
func (server *Server) canViewGame(ctx *gin.Context, game db.Game) bool {
	if game.Visibility == "public" {
		return true
	}
	user, ok := server.getCurrentUser(ctx)
	if !ok {
		return false
	}
	if uuidEq(game.WhitePlayerID, user.ID) || uuidEq(game.BlackPlayerID, user.ID) {
		return true
	}
	ctx.JSON(http.StatusNotFound, errorMessage(ErrGameNotFound))
	return false
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

	// engine games have no open seat for humans
	if game.Opponent != "person" {
		ctx.JSON(http.StatusForbidden, errorMessage("cannot join an engine game"))
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

	// Send push notification to the player who was waiting. `updated` carries
	// both player IDs after the join — no need to refetch the game.
	server.notifyOpponent(ctx, updated, user.ID, user.Username)

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
		LastMoveAt:           game.LastMoveAt,
	})
	if handleDBError(ctx, err, WithLogArgs("resignGame: UpdateGameState", "game_id", ctx.Param("id"))) {
		return
	}

	// Cancel timeout watcher and evict from memory on resign. The channel is
	// swapped under GameStateMu (never while holding activeGamesMu) so it can't
	// race a concurrent close in wsHandleMove.
	server.activeGamesMu.Lock()
	gs, inMemory := server.activeGames[gameID]
	delete(server.activeGames, gameID)
	server.activeGamesMu.Unlock()
	if inMemory {
		gs.GameStateMu.Lock()
		gs.Status = db.GameStateResign
		stopTimeoutWatcher(gs)
		gs.GameStateMu.Unlock()
	}

	// The resigner's opponent wins.
	winner := "w"
	if uuidEq(game.WhitePlayerID, user.ID) {
		winner = "b"
	}
	server.finishGame(ctx, gameID, winner)

	// Tell the opponent (and any spectators) over the WebSocket — without this
	// they would only discover the resignation on their next page load.
	server.wsBroadcastToWatchers(gameID, mustMarshalEvent(EventMakeMove, MoveResult{
		CurrentPlayer:        string(game.CurrentPlayer),
		EndReason:            "resign",
		EndedByPlayerID:      uidStr(user.ID),
		Winner:               winner,
		WhiteTimeRemainingMs: game.WhiteTimeRemainingMs,
		BlackTimeRemainingMs: game.BlackTimeRemainingMs,
	}))

	ctx.JSON(http.StatusOK, server.toGameResponse(ctx, updated))
}

// LiveGameResponse is one row of the public "watch live" list.
type LiveGameResponse struct {
	ID                   string    `json:"id"`
	WhiteUsername        string    `json:"white_username"`
	WhiteRating          int32     `json:"white_rating"`
	BlackUsername        string    `json:"black_username"`
	BlackRating          int32     `json:"black_rating"`
	CurrentPlayer        string    `json:"current_player"`
	MoveCount            int32     `json:"move_count"`
	WhiteTimeRemainingMs int64     `json:"white_time_remaining_ms"`
	BlackTimeRemainingMs int64     `json:"black_time_remaining_ms"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// @Summary      List live games
// @Description  Returns public person-vs-person games currently being played, newest activity first. No auth required — spectators use this.
// @Tags         Games
// @Accept       json
// @Produce      json
// @Success      200  {array}   api.LiveGameResponse
// @Failure      500  {object}  map[string]string
// @Router       /games/live [get]
func (server *Server) listLiveGames(ctx *gin.Context) {
	rows, err := server.store.ListLiveGames(ctx)
	if handleDBError(ctx, err, WithLogArgs("listLiveGames: failed")) {
		return
	}
	out := make([]LiveGameResponse, len(rows))
	for i, r := range rows {
		out[i] = LiveGameResponse{
			ID:                   uidStr(r.ID),
			WhiteUsername:        r.WhiteUsername,
			WhiteRating:          r.WhiteRating,
			BlackUsername:        r.BlackUsername,
			BlackRating:          r.BlackRating,
			CurrentPlayer:        string(r.CurrentPlayer),
			MoveCount:            r.MoveCount,
			WhiteTimeRemainingMs: r.WhiteTimeRemainingMs,
			BlackTimeRemainingMs: r.BlackTimeRemainingMs,
			CreatedAt:            r.CreatedAt.Time,
			UpdatedAt:            r.UpdatedAt.Time,
		}
	}
	ctx.JSON(http.StatusOK, out)
}

// GameReplayResponse carries everything a client needs to replay a finished
// (or ongoing) game: the ordered UCI move list plus result metadata.
type GameReplayResponse struct {
	GameID          string    `json:"game_id"`
	WhitePlayerName string    `json:"white_player_name"`
	BlackPlayerName string    `json:"black_player_name"`
	Opponent        string    `json:"opponent"`
	StockfishLevel  int16     `json:"stockfish_level"`
	State           string    `json:"state"`
	EndReason       string    `json:"end_reason"`
	EndedByPlayerID string    `json:"ended_by_player_id"`
	Moves           []string  `json:"moves"` // UCI, in play order
	MoveCount       int32     `json:"move_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// @Summary      Replay a game
// @Description  Returns the ordered move list (UCI) and result metadata so the client can step through the game.
// @Tags         Games
// @Accept       json
// @Produce      json
// @Param        id  path  string  true  "Game UUID"
// @Security     Bearer
// @Success      200  {object}  api.GameReplayResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /games/{id}/replay [get]
func (server *Server) getGameReplay(ctx *gin.Context) {
	gameID, ok := parseUUIDParam(ctx, "id")
	if !ok {
		return
	}

	game, err := server.store.GetGameByID(ctx, gameID)
	if handleDBError(ctx, err,
		WithNotFoundMsg(ErrGameNotFound),
		WithLogArgs("getGameReplay: GetGameByID", "game_id", ctx.Param("id"))) {
		return
	}
	if !server.canViewGame(ctx, game) {
		return
	}

	snap := board.DeserializeGameState(game.BoardState)
	moves := snap.StockfishGame
	if moves == nil {
		moves = []string{}
	}

	ctx.JSON(http.StatusOK, GameReplayResponse{
		GameID:          uidStr(game.ID),
		WhitePlayerName: server.lookupUsername(ctx, game.WhitePlayerID),
		BlackPlayerName: server.lookupUsername(ctx, game.BlackPlayerID),
		Opponent:        game.Opponent,
		StockfishLevel:  game.StockfishLevel,
		State:           string(game.State),
		EndReason:       game.EndReason,
		EndedByPlayerID: uidStr(game.EndedByPlayerID),
		Moves:           moves,
		MoveCount:       game.MoveCount,
		CreatedAt:       game.CreatedAt.Time,
		UpdatedAt:       game.UpdatedAt.Time,
	})
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

	game, err := server.store.GetGameByID(ctx, gameID)
	if handleDBError(ctx, err,
		WithNotFoundMsg(ErrGameNotFound),
		WithLogArgs("getGameMoves: GetGameByID", "game_id", ctx.Param("id"))) {
		return
	}
	if !server.canViewGame(ctx, game) {
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
