package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// checkUsernameExists
// getUser
// getAllUsers

type CheckUsernameExistsParams struct {
	Username string `form:"username" binding:"required,max=20"`
}

func (r *CheckUsernameExistsParams) SanitizeParams() {
	r.Username = strings.ToLower(r.Username)
}

func (server *Server) checkUsernameExists(ctx *gin.Context) {
	//check if username exists
	var req CheckUsernameExistsParams

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	//sanitize input
	req.SanitizeParams()

	exists, err := server.store.UsernameExists(ctx, req.Username)
	if handleDBError(ctx, err, WithLogArgs("checkUsernameExists: failed UsernameExists", req.Username)) {
		return
	}
ctx.JSON(http.StatusOK, gin.H{"username": req.Username, "exists": exists})
}


type getMeResponse struct {
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	EmailConfirmed bool      `json:"email_confirmed"`
	IsActive       bool      `json:"is_active"`
	LastLoginAt    time.Time `json:"last_login_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

func (server *Server) getMe(ctx *gin.Context) {
	user, ok := server.getCurrentUser(ctx)
	if !ok {
		return
	}
	ctx.JSON(http.StatusOK, getMeResponse{
		Username:       user.Username,
		Email:          user.Email,
		EmailConfirmed: user.EmailConfirmed,
		IsActive:       user.IsActive,
		LastLoginAt:    user.LastLoginAt.Time,
		CreatedAt:      user.CreatedAt.Time,
	})
}
