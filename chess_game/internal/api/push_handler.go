package api

import (
	"fmt"
	"log/slog"
	"net/http"

	db "github.com/Glenn444/golang-chess/internal/db"
	"github.com/Glenn444/golang-chess/internal/push"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type pushSubscribeReq struct {
	Subscription pushSubscriptionFields `json:"subscription" binding:"required"`
}

type pushSubscriptionFields struct {
	Endpoint string           `json:"endpoint" binding:"required"`
	Keys     pushSubscribeKeys `json:"keys" binding:"required"`
}

type pushSubscribeKeys struct {
	P256dh string `json:"p256dh" binding:"required"`
	Auth   string `json:"auth" binding:"required"`
}

// @Summary      Subscribe to push notifications
// @Description  Saves a global Web Push subscription for the authenticated user.
// @Tags         Push
// @Accept       json
// @Produce      json
// @Param        body  body  pushSubscribeReq  true  "Push subscription payload"
// @Security     Bearer
// @Success      201  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /push/subscribe [post]
func (server *Server) subscribePush(ctx *gin.Context) {
	var req pushSubscribeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, ok := server.getCurrentUser(ctx)
	if !ok {
		return
	}

	err := server.store.SavePushSubscription(ctx, db.SavePushSubscriptionParams{
		UserID:   user.ID,
		Endpoint: req.Subscription.Endpoint,
		P256dh:   req.Subscription.Keys.P256dh,
		Auth:     req.Subscription.Keys.Auth,
	})
	if handleDBError(ctx, err, WithLogArgs("subscribePush: SavePushSubscription", "user_id", user.ID)) {
		return
	}

	slog.Info("push: subscription saved", "user_id", uidStr(user.ID))
	ctx.JSON(http.StatusCreated, successMessage("push subscription saved"))
}

// @Summary      Get push subscription status
// @Description  Returns 200 if the current user has an active push subscription, 404 otherwise.
// @Tags         Push
// @Produce      json
// @Security     Bearer
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /push/subscription [get]
func (server *Server) getPushSubscription(ctx *gin.Context) {
	user, ok := server.getCurrentUser(ctx)
	if !ok {
		return
	}

	_, err := server.store.GetPushSubscriptionExists(ctx, user.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, errorMessage("no push subscription found"))
		return
	}

	ctx.JSON(http.StatusOK, successMessage("push subscription exists"))
}

// notifyOpponent sends a push notification to the player who was waiting
// (the player in the game who is NOT joiningUserID), but only if that player
// is not currently connected via WebSocket.
func (server *Server) notifyOpponent(ctx *gin.Context, gameID pgtype.UUID, joiningUserID pgtype.UUID, joiningUsername string) {
	game, err := server.store.GetGameByID(ctx, gameID)
	if err != nil {
		slog.Warn("push: notifyOpponent — GetGameByID failed", "game_id", uidStr(gameID), "err", err)
		return
	}

	// Determine the waiting player.
	var opponentID pgtype.UUID
	if uuidEq(game.WhitePlayerID, joiningUserID) {
		opponentID = game.BlackPlayerID
	} else {
		opponentID = game.WhitePlayerID
	}
	if !opponentID.Valid {
		return
	}

	// Skip if the waiting player is already connected via WebSocket.
	if server.wsUserInGame(opponentID, gameID) {
		return
	}

	// Look up the waiting player's global push subscription.
	sub, err := server.store.GetPushSubscriptionByUser(ctx, opponentID)
	if err != nil {
		// No subscription — nothing to do.
		return
	}

	cfg := push.Config{
		VAPIDPublicKey:  server.config.VAPIDPublicKey,
		VAPIDPrivateKey: server.config.VAPIDPrivateKey,
		VAPIDSubject:    server.config.VAPIDSubject,
	}

	payload := push.Payload{
		Title: joiningUsername + " joined your Chesske game!",
		Body:  "Head back to Chesske — your game is ready.",
		URL:   fmt.Sprintf("/game/%s", uidStr(gameID)),
	}

	if err := push.Send(sub.Endpoint, sub.P256dh, sub.Auth, payload, cfg); err != nil {
		if err == push.ErrSubscriptionGone {
			server.store.DeletePushSubscriptionByUser(ctx, opponentID)
		}
		slog.Warn("push: notifyOpponent — send failed", "game_id", uidStr(gameID), "err", err)
		return
	}

	slog.Info("push: notification sent", "game_id", uidStr(gameID), "to_user", uidStr(opponentID))
}
