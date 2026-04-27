package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"net/http"
	"time"

	db "github.com/Glenn444/golang-chess/internal/db"
	"github.com/Glenn444/golang-chess/internal/token"
	"github.com/Glenn444/golang-chess/internal/utils/auth"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
)


type CheckUsernameExistsParams struct {
	Username string `form:"username" binding:"required,max=20"`
}

func (server *Server) checkUsernameExists(ctx *gin.Context){
	//check if username exists
	var req CheckUsernameExistsParams

	if err := ctx.ShouldBindQuery(&req); err != nil{
		ctx.JSON(http.StatusBadRequest,errorResponse(err))
		return
	}

	exists,err := server.store.UsernameExists(ctx,req.Username)
	if err != nil{
		ctx.JSON(http.StatusInternalServerError,errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK,gin.H{"exists":exists})
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	FullName string `json:"full_name"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type CreateUserResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func (server *Server) createUser(ctx *gin.Context) {
	
	var req CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	//check if username exists
	userExists,err := server.store.UsernameExists(ctx,req.Username)
	if err != nil{
		ctx.JSON(http.StatusInternalServerError,errorMessage("something wrong happend"))
		return
	}
	if userExists{
		ctx.JSON(http.StatusConflict,errorMessage("username exists!!"))
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		Email:          req.Email,
		PasswordHash: hashedPassword,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		if pqError, ok := err.(*pq.Error); ok {
			switch pqError.Code.Name() {
			case "unique_violation":
				ctx.JSON(http.StatusConflict, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	otpCode, err := auth.GenerateOTP(6)
	if err != nil{
		ctx.JSON(http.StatusInternalServerError,errorMessage("error generating otp code"))
		return
	}
	otpHash,err := auth.SignOtpCode(otpCode,server.config.TokenSymmetricKey)
	if err != nil{
		ctx.JSON(http.StatusInternalServerError,errorMessage("error signing the otpcode"))
		return
	}

	argEmailOtp := db.CreateEmailOTPParams{
		UserID: user.ID,
		CodeHash: otpHash,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(server.config.AcessTokenDuration),Valid: true},
	}
	//verify email
	_,err = server.store.CreateEmailOTP(ctx,argEmailOtp)
	if err != nil{
		ctx.JSON(http.StatusInternalServerError,errorMessage("failed creating OTP in the database"))
		return
	}

	//send the otp to the user email
	err = server.emailClient.SendEmailOTP(user.Email,otpCode)
	if err != nil{
		ctx.JSON(http.StatusInternalServerError,errorMessage("error sending OTP to email"))
		return
	}
	resp := CreateUserResponse{
		Username:          user.Username,
		PasswordChangedAt: user.UpdatedAt.Time,
		CreatedAt:         user.CreatedAt.Time,
	}
	ctx.JSON(http.StatusOK, resp)

}


type ConfirmEmail struct{
	Email string `json:"email" binding:"required"`
	EMAIL_OTP string `json:"email_otp" binding:"required"`
}

func (server *Server) confirmEmail(ctx *gin.Context){
	var req ConfirmEmail

	if err := ctx.ShouldBindJSON(&req); err != nil{
		ctx.JSON(http.StatusBadRequest,errorResponse(err))
		return
	}


	user,err := server.store.GetUserByEmail(ctx,req.Email)
	if err != nil{
		if err == sql.ErrNoRows{
			ctx.JSON(http.StatusNotFound,errorMessage("user does not exist"))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	//2. Get OTP Code from db
	emailOTP,err := server.store.GetValidOTP(ctx,user.ID)
	if err != nil{
		if err == sql.ErrNoRows{
			ctx.JSON(http.StatusNotFound,errorMessage("otp not found, create a new one"))
			return
		}
	}
}

type SearchUserParams struct {
	Username string `form:"username" binding:"required,alphanum"`
}

type GetUserResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func (server *Server) getUser(ctx *gin.Context) {
	var param SearchUserParams
	if err := ctx.ShouldBindQuery(&param); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, param.Username)
	if err != nil {
		if err == sql.ErrNoRows{
			ctx.JSON(http.StatusNotFound,errorMessage("user does not exist"))
			return
		}
		
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := GetUserResponse{
		Username:          user.Username,
		Email:             user.Email,
		FullName:          user.FullName,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
	ctx.JSON(http.StatusOK, resp)

}

type loginUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginUserResponse struct {
	Username          string    `json:"username"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
	AccessToken       string    `json:"access_token"`
	RefreshToken      string    `json:"refresh_token"`
}

// login user
func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		// possible errors,
		//1. user not found
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorMessage("user does not exist, sign up"))
			return
		}
		//2. something happened to the database
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}

	//check user password against saved db password
	err = util.CheckPassword(user.HashedPassword, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorMessage("Invalid username or password"))
		return
	}

	//create the access token
	access_token, err := server.tokenMaker.CreateToken(req.Username, token.AccessToken, server.config.AcessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	//create the refresh token signed with 1 hour and save to the database
	week := time.Hour * 24 * 7

	refresh_token, err := server.tokenMaker.CreateToken(req.Username, token.RefreshToken, week)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = server.store.UpdateRefreshToken(ctx, database.UpdateRefreshTokenParams{
		Username:     user.Username,
		RefreshToken: refresh_token,
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := loginUserResponse{
		Username:          req.Username,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
		AccessToken:       access_token,
		RefreshToken:      refresh_token,
	}

	ctx.JSON(http.StatusOK, resp)

}

type allUsersResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

// get all users in the app
func (server *Server) getAllUsers(ctx *gin.Context) {

	users, err := server.store.GetAllUsers(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	var allUsers []allUsersResponse
	for _, user := range users {
		gotUser := allUsersResponse{
			Username:          user.Username,
			FullName:          user.FullName,
			Email:             user.Email,
			PasswordChangedAt: user.PasswordChangedAt,
			CreatedAt:         user.CreatedAt,
		}
		allUsers = append(allUsers, gotUser)
	}

	ctx.JSON(http.StatusOK, allUsers)
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type refreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

func (server *Server) refreshToken(ctx *gin.Context) {
	var req refreshTokenRequest

	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	//verify the refresh token and get the payload
	payload, err := server.tokenMaker.VerifyToken(req.RefreshToken, token.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	//refreshtoken is valid issue new access token
	accessToken, err := server.tokenMaker.CreateToken(payload.Subject, token.AccessToken, server.config.AcessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := refreshTokenResponse{
		AccessToken: accessToken,
	}

	ctx.JSON(http.StatusOK, resp)
}
