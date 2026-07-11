package api

import (
	"log/slog"
	"net/http"

	db "github.com/Glenn444/golang-chess/internal/db"
	"github.com/gin-gonic/gin"
)

type SendChatMessageReq struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

// @Summary      Send chat message
// @Description  Sends a chat message in a game. The message is broadcast to all players via WebSocket.
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Param        id   path  string              true  "Game UUID"
// @Param        body body  SendChatMessageReq  true  "Chat message"
// @Security     Bearer
// @Success      201  {object}  object
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /games/{id}/chat [post]
func (server *Server) sendChatMessage(ctx *gin.Context) {
	gameID, ok := parseUUIDParam(ctx, "id")
	if !ok {
		return
	}

	user, ok := server.getCurrentUser(ctx)
	if !ok {
		return
	}

	var req SendChatMessageReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	game, err := server.store.GetGameByID(ctx, gameID)
	if handleDBError(ctx, err,
		WithNotFoundMsg(ErrGameNotFound),
		WithLogArgs("sendChatMessage: GetGameByID", "game_id", ctx.Param("id"))) {
		return
	}

	if !uuidEq(game.WhitePlayerID, user.ID) && !uuidEq(game.BlackPlayerID, user.ID) {
		ctx.JSON(http.StatusForbidden, errorMessage(ErrNotAPlayer))
		return
	}

	// Cap at 200 messages per game.
	count, err := server.store.CountChatMessagesByGameID(ctx, gameID)
	if err != nil {
		slog.Error("sendChatMessage: CountChatMessagesByGameID", "game_id", ctx.Param("id"), "err", err)
	} else if count >= 200 {
		ctx.JSON(http.StatusConflict, errorMessage("chat limit reached (200 messages per game)"))
		return
	}

	msg, err := server.store.CreateChatMessage(ctx, db.CreateChatMessageParams{
		GameID:   gameID,
		SenderID: user.ID,
		Content:  req.Content,
	})
	if handleDBError(ctx, err, WithLogArgs("sendChatMessage: CreateChatMessage", "game_id", ctx.Param("id"))) {
		return
	}

	ctx.JSON(http.StatusCreated, msg)
}

// @Summary      Get chat messages
// @Description  Returns the chat history for a game.
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Param        id  path  string  true  "Game UUID"
// @Security     Bearer
// @Success      200  {array}   object
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /games/{id}/chat [get]
func (server *Server) getChatMessages(ctx *gin.Context) {
	gameID, ok := parseUUIDParam(ctx, "id")
	if !ok {
		return
	}

	messages, err := server.store.GetChatMessagesByGameID(ctx, gameID)
	if handleDBError(ctx, err,
		WithNotFoundMsg(ErrGameNotFound),
		WithLogArgs("getChatMessages: failed", "game_id", ctx.Param("id"))) {
		return
	}

	ctx.JSON(http.StatusOK, messages)
}
