package api

import (
	"net/http"

	db "github.com/Glenn444/golang-chess/internal/db"
	"github.com/gin-gonic/gin"
)

func (server *Server) startVoiceSession(ctx *gin.Context) {
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
		WithLogArgs("startVoiceSession: GetGameByID", "game_id", ctx.Param("id"))) {
		return
	}

	if !uuidEq(game.WhitePlayerID, user.ID) && !uuidEq(game.BlackPlayerID, user.ID) {
		ctx.JSON(http.StatusForbidden, errorMessage(ErrNotAPlayer))
		return
	}

	// reject if a live session already exists
	_, err = server.store.GetActiveVoiceSessionByGameID(ctx, gameID)
	if err == nil {
		ctx.JSON(http.StatusConflict, errorMessage(ErrVoiceSessionAlreadyActive))
		return
	}

	session, err := server.store.CreateVoiceSession(ctx, db.CreateVoiceSessionParams{
		GameID:      gameID,
		InitiatorID: user.ID,
	})
	if handleDBError(ctx, err, WithLogArgs("startVoiceSession: CreateVoiceSession", "game_id", ctx.Param("id"))) {
		return
	}

	ctx.JSON(http.StatusCreated, session)
}

func (server *Server) getActiveVoiceSession(ctx *gin.Context) {
	gameID, ok := parseUUIDParam(ctx, "id")
	if !ok {
		return
	}

	session, err := server.store.GetActiveVoiceSessionByGameID(ctx, gameID)
	if handleDBError(ctx, err,
		WithNotFoundMsg(ErrVoiceSessionNotFound),
		WithLogArgs("getActiveVoiceSession: failed", "game_id", ctx.Param("id"))) {
		return
	}

	ctx.JSON(http.StatusOK, session)
}

func (server *Server) activateVoiceSession(ctx *gin.Context) {
	vid, ok := parseUUIDParam(ctx, "vid")
	if !ok {
		return
	}

	user, ok := server.getCurrentUser(ctx)
	if !ok {
		return
	}

	voiceSession, err := server.store.GetVoiceSessionByID(ctx, vid)
	if handleDBError(ctx, err,
		WithNotFoundMsg(ErrVoiceSessionNotFound),
		WithLogArgs("activateVoiceSession: GetVoiceSessionByID", "vid", ctx.Param("vid"))) {
		return
	}

	// only the recipient (non-initiator) accepts the call
	if uuidEq(voiceSession.InitiatorID, user.ID) {
		ctx.JSON(http.StatusForbidden, errorMessage("only the call recipient can accept the call"))
		return
	}

	updated, err := server.store.ActivateVoiceSession(ctx, vid)
	if handleDBError(ctx, err,
		WithNotFoundMsg(ErrVoiceSessionNotFound),
		WithLogArgs("activateVoiceSession: ActivateVoiceSession", "vid", ctx.Param("vid"))) {
		return
	}

	ctx.JSON(http.StatusOK, updated)
}

func (server *Server) endVoiceSession(ctx *gin.Context) {
	gameID, ok := parseUUIDParam(ctx, "id")
	if !ok {
		return
	}

	vid, ok := parseUUIDParam(ctx, "vid")
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
		WithLogArgs("endVoiceSession: GetGameByID", "game_id", ctx.Param("id"))) {
		return
	}

	if !uuidEq(game.WhitePlayerID, user.ID) && !uuidEq(game.BlackPlayerID, user.ID) {
		ctx.JSON(http.StatusForbidden, errorMessage(ErrNotAPlayer))
		return
	}

	ended, err := server.store.EndVoiceSession(ctx, vid)
	if handleDBError(ctx, err,
		WithNotFoundMsg(ErrVoiceSessionNotFound),
		WithLogArgs("endVoiceSession: EndVoiceSession", "vid", ctx.Param("vid"))) {
		return
	}

	ctx.JSON(http.StatusOK, ended)
}
