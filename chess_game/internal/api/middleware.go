package api

import (
	"net/http"
	"strings"

	"github.com/Glenn444/golang-chess/internal/token"
	"github.com/gin-gonic/gin"
)

const (
	authorizationPayloadKey = "authorization_payload"
	authorizationTypeBearer = "Bearer"
)

func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Try cookie first (browser clients), then Bearer header (CLI/mobile/desktop).
		accessToken, err := ctx.Cookie("access_token")
		if err != nil || accessToken == "" {
			authHeader := ctx.GetHeader("Authorization")
			if authHeader == "" {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorMessage("missing authorization"))
				return
			}
			fields := strings.Fields(authHeader)
			if len(fields) != 2 || fields[0] != authorizationTypeBearer {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorMessage("invalid authorization header"))
				return
			}
			accessToken = fields[1]
		}

		payload, err := tokenMaker.VerifyToken(accessToken, token.AccessTokenType)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		ctx.Set(authorizationPayloadKey, payload)
		ctx.Next()
	}
}
