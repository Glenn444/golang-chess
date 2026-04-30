package api

import (
	"net/http"
	"strings"

	"github.com/Glenn444/golang-chess/internal/token"
	"github.com/gin-gonic/gin"
)

const(
	authorizationPayloadKey = "authorization_payload"
	authorizationTypeBearer = "Bearer"
)

type authHeader struct {
	Authorization string `header:"Authorization" binding:"required"`
}

func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var h authHeader

		if err := ctx.ShouldBindHeader(&h); err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}
		authParts := strings.Split(h.Authorization, " ")

		if len(authParts) != 2 || authParts[0] != "Bearer" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorMessage("invalid or missing authorization header"))
			return
		}
		authorizationBearerToken := authParts[1]
		//verify that the token is valid and it's not a refreshToken
		payload, err := tokenMaker.VerifyToken(authorizationBearerToken, token.AccessTokenType)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		ctx.Set(authorizationPayloadKey,payload)

		ctx.Next()
	}
}
