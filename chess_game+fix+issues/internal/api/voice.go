package api

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"strings"

	db "github.com/Glenn444/golang-chess/internal/db"
	"github.com/gin-gonic/gin"
	
	"encoding/json"
	"fmt"
)


type cloudflareICEResponse struct {
	IceServers []ICEServer `json:"iceServers"`
}

type ICEServer struct {
	URLs       []string `json:"urls"`
	Username   string   `json:"username,omitempty"`
	Credential string   `json:"credential,omitempty"`
}

// @Summary      Get TURN credentials
// @Description  Returns short-lived Cloudflare TURN credentials for WebRTC voice calls.
// @Tags         Voice
// @Produce      json
// @Security     Bearer
// @Success      200  {object}  cloudflareICEResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /turn-credentials [get]
func (server *Server) getTURNCredentials(ctx *gin.Context) {
	
	cfURL := fmt.Sprintf(
		"https://rtc.live.cloudflare.com/v1/turn/keys/%s/credentials/generate-ice-servers",
		server.config.CloudflareTURNKeyID,
	)

	body, err := json.Marshal(map[string]int{"ttl": 86400})
	if err != nil {
		slog.Error("getTURNCredentials: failed to marshal request body", "err", err)
		ctx.JSON(http.StatusInternalServerError, errorMessage(ErrInternalServer))
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfURL, bytes.NewReader(body))
	if err != nil {
		slog.Error("getTURNCredentials: failed to build Cloudflare request", "err", err)
		ctx.JSON(http.StatusInternalServerError, errorMessage(ErrInternalServer))
		return
	}
	req.Header.Set("Authorization", "Bearer "+server.config.CloudflareTURNAPIToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("getTURNCredentials: failed to reach Cloudflare TURN API", "err", err)
		ctx.JSON(http.StatusInternalServerError, errorMessage(ErrInternalServer))
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("getTURNCredentials: failed to read Cloudflare response body", "err", err, "status", resp.StatusCode)
		ctx.JSON(http.StatusInternalServerError, errorMessage(ErrInternalServer))
		return
	}

	if resp.StatusCode != http.StatusCreated {
		slog.Error("getTURNCredentials: unexpected Cloudflare response",
			"status", resp.StatusCode,
			"body", string(respBody),
		)
		ctx.JSON(http.StatusInternalServerError, errorMessage(ErrInternalServer))
		return
	}

	var iceResp cloudflareICEResponse
	if err := json.Unmarshal(respBody, &iceResp); err != nil {
		slog.Error("getTURNCredentials: failed to parse Cloudflare response", "err", err, "body", string(respBody))
		ctx.JSON(http.StatusInternalServerError, errorMessage(ErrInternalServer))
		return
	}

	// Filter out port 53 URLs — blocked by browsers
	for i, iceServer := range iceResp.IceServers {
		filtered := make([]string, 0, len(iceServer.URLs))
		for _, u := range iceServer.URLs {
			if !strings.Contains(u, ":53?") && !strings.Contains(u, ":53") {
				filtered = append(filtered, u)
			}
		}
		iceResp.IceServers[i].URLs = filtered
	}

	ctx.JSON(http.StatusOK, iceResp)
}

type TURNCredentials struct {
	Username   string `json:"username"`
	Credential string `json:"credential"`
	TTL        int    `json:"ttl"`
	URLs       []string `json:"urls"`
}



// @Summary      Start voice session
// @Description  Initiates a WebRTC voice call in a game. Signalling travels over WebSocket.
// @Tags         Voice
// @Accept       json
// @Produce      json
// @Param        id  path  string  true  "Game UUID"
// @Security     Bearer
// @Success      201  {object}  object
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /games/{id}/voice [post]
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

// @Summary      Get active voice session
// @Description  Returns the active voice session for a game, if one exists.
// @Tags         Voice
// @Accept       json
// @Produce      json
// @Param        id  path  string  true  "Game UUID"
// @Security     Bearer
// @Success      200  {object}  object
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /games/{id}/voice [get]
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

// @Summary      Accept voice call
// @Description  Accepts an incoming voice call. Only the non-initiating player can accept.
// @Tags         Voice
// @Accept       json
// @Produce      json
// @Param        id   path  string  true  "Game UUID"
// @Param        vid  path  string  true  "Voice session UUID"
// @Security     Bearer
// @Success      200  {object}  object
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /games/{id}/voice/{vid}/activate [patch]
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

// @Summary      End voice call
// @Description  Ends a voice session. Either player in the game can end the call.
// @Tags         Voice
// @Accept       json
// @Produce      json
// @Param        id   path  string  true  "Game UUID"
// @Param        vid  path  string  true  "Voice session UUID"
// @Security     Bearer
// @Success      200  {object}  object
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /games/{id}/voice/{vid} [delete]
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
