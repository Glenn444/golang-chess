package api

import (
	"net/http"

	db "github.com/Glenn444/golang-chess/internal/db"
	"github.com/gin-gonic/gin"
)

type sendChatMessageRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

func (server *Server) sendChatMessage(ctx *gin.Context) {
	gameID, ok := parseUUIDParam(ctx, "id")
	if !ok {
		return
	}

	user, ok := server.getCurrentUser(ctx)
	if !ok {
		return
	}

	var req sendChatMessageRequest
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
