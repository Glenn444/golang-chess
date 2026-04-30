package api

import (
	"net/http"
	"strings"

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

// type SearchUserParams struct {
// 	Username string `form:"username" binding:"required,alphanum"`
// }

// type GetUserResponse struct {
// 	Username          string    `json:"username"`
// 	FullName          string    `json:"full_name"`
// 	Email             string    `json:"email"`
// 	PasswordChangedAt time.Time `json:"password_changed_at"`
// 	CreatedAt         time.Time `json:"created_at"`
// }

// func (server *Server) getUser(ctx *gin.Context) {
// 	var param SearchUserParams
// 	if err := ctx.ShouldBindQuery(&param); err != nil {
// 		ctx.JSON(http.StatusBadRequest, errorResponse(err))
// 		return
// 	}

// 	user, err := server.store.GetUser(ctx, param.Username)
// 	if err != nil {
// 		if err == sql.ErrNoRows{
// 			ctx.JSON(http.StatusNotFound,errorMessage("user does not exist"))
// 			return
// 		}

// 		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
// 		return
// 	}

// 	resp := GetUserResponse{
// 		Username:          user.Username,
// 		Email:             user.Email,
// 		FullName:          user.FullName,
// 		PasswordChangedAt: user.PasswordChangedAt,
// 		CreatedAt:         user.CreatedAt,
// 	}
// 	ctx.JSON(http.StatusOK, resp)

// }
