package api

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	db "github.com/Glenn444/golang-chess/internal/db"
	"github.com/Glenn444/golang-chess/internal/token"
	"github.com/Glenn444/golang-chess/internal/utils/auth"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// createUser (registration IS an auth concern)
// confirmEmail
// loginUser
// refreshToken

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func (r *CreateUserRequest) SanitizeCreateUserReq() {
	r.Username = strings.ToLower(r.Username)
	r.Email = strings.ToLower(r.Email)
}

type CreateUserResponse struct {
	Username          string    `json:"username"`
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

	//sanitize user input
	req.SanitizeCreateUserReq()

	//check if username exists, ignore if does not exist
	userExists, err := server.store.UsernameExists(ctx, req.Username)
	if handleDBError(ctx, err, 
		WithLogArgs("createUser: failed to check UserNameExists", "err", err, "username", req.Username),
		) {
		return
	}

	if userExists {
		ctx.JSON(http.StatusConflict, errorMessage("username exists!!"))
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		slog.Error("failed to hash Password", "err", err, "user_id", req.Email)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateUserParams{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if handleDBError(ctx, err,
		WithUniqueMsg("username or email already taken"),
		WithLogArgs("user_email", req.Email)) {
		return
	}

	otpCode, err := auth.GenerateOTP(6)
	if err != nil {
		slog.Error("failed to generate OTP Code", "err", err, "user_id", user.ID)
		ctx.JSON(http.StatusInternalServerError, errorMessage("error generating otp code"))
		return
	}
	otpHash, err := auth.SignOtpCode(otpCode, server.config.TokenSymmetricKey)
	if err != nil {
		slog.Error("failed HMAC OTP Signing", "err", err, "user_id", user.ID)
		ctx.JSON(http.StatusInternalServerError, errorMessage("error"))
		return
	}

	argEmailOtp := db.CreateEmailOTPParams{
		UserID:    user.ID,
		CodeHash:  otpHash,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(server.config.AcessTokenDuration), Valid: true},
	}
	//verify email
	_, err = server.store.CreateEmailOTP(ctx, argEmailOtp)
	if handleDBError(ctx, err, WithLogArgs("user_id", user.ID)) {
		return
	}

	//send the otp to the user email
	err = server.emailClient.SendEmailOTP(user.Email, otpCode)
	if err != nil {
		slog.Error("failed to send email otp", "err", err, "user_email", user.Email)
		ctx.JSON(http.StatusInternalServerError, errorMessage("error sending OTP to email"))
		return
	}
	resp := CreateUserResponse{
		Username:          user.Username,
		PasswordChangedAt: user.UpdatedAt.Time,
		CreatedAt:         user.CreatedAt.Time,
	}
	ctx.JSON(http.StatusOK, resp)

}

type ConfirmEmail struct {
	Email     string `json:"email" binding:"required,email"`
	EMAIL_OTP string `json:"email_otp" binding:"required,numeric,len=6"`
}

func (r *ConfirmEmail) SanitizeConfirmEmailReq() {
	r.Email = strings.ToLower(r.Email)
}

type ConfirmEmailResponse struct {
	Message string `json:"message"`
	Email   string `json:"email"`
}

func (server *Server) confirmEmail(ctx *gin.Context) {
	var req ConfirmEmail

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	//sanitize user email
	req.SanitizeConfirmEmailReq()

	user, err := server.store.GetUserByEmail(ctx, req.Email)
	if handleDBError(ctx, err,
		WithNotFoundMsg("user does not exist"),
		WithLogArgs("user_email", req.Email),
	) {
		return
	}

	//if user is already verified fail the request
	if user.EmailConfirmed {
		ctx.JSON(http.StatusConflict, errorMessage("email already verified, proceed to login"))
		return
	}

	//2. Get valid OTP Code from db
	emailOTP, err := server.store.GetValidOTP(ctx, user.ID)
	if handleDBError(ctx, err,
		WithNotFoundMsg("valid OTP not found"),
		WithLogArgs("user_id", user.ID),
	) {
		return
	}

	//3. Compare the OTP hash to the given OTP
	match, err := auth.ConfirmOTP(req.EMAIL_OTP, emailOTP.CodeHash, server.config.TokenSymmetricKey)
	if err != nil {
		slog.Error("failed to decode OTPHASH Byte ", "err", err, "user_id", user.ID)
		ctx.JSON(http.StatusInternalServerError, errorMessage("sorry, something wrong happened"))
		return
	}
	if !match {
		_, err := server.store.IncrementOTPAttempts(ctx, emailOTP.ID)
		if err != nil {
			slog.Error("failed to incrementOTP Attempts otp", "err", err, "user_id", user.ID)
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}

		//successful increment
		//slog.Error("failed to confirm OTP Code", "err", err, "user_id", user.ID)
		ctx.JSON(http.StatusForbidden, errorMessage("invalid OTP code"))
		return
	}
	//4. successful match, mark the OTP as used
	_, err = server.store.MarkOTPUsed(ctx, emailOTP.ID)
	if handleDBError(ctx, err, WithLogArgs("OTP_id", emailOTP.ID)) {
		return
	}

	//5. Mark user email as verified
	_, err = server.store.ConfirmEmail(ctx, user.ID)
	if err != nil {
		slog.Error("failed to mark user email as verified", "err", err, "user_id", user.ID)
		ctx.JSON(http.StatusInternalServerError, errorMessage("error"))
		return
	}

	ctx.JSON(http.StatusOK, ConfirmEmailResponse{
		Message: "email verified successfully",
		Email:   user.Email,
	})

}

type SendEmailOTP struct {
	Email string `json:"email" binding:"required,email"`
}

func (r *SendEmailOTP) SanitizeEmailOTP() {
	r.Email = strings.ToLower(r.Email)
}
func (server *Server) sendEmailOTP(ctx *gin.Context) {
	var req SendEmailOTP
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	//sanitize user email
	req.SanitizeEmailOTP()

	//2. Get the user by email
	user, err := server.store.GetUserByEmail(ctx, req.Email)
	if handleDBError(ctx, err,
		WithNotFoundMsg("user does not exist!"),
		WithLogArgs("user_email", req.Email),
	) {
		return
	}

	if user.EmailConfirmed {
		ctx.JSON(http.StatusConflict, errorMessage("email already verified, proceed to login"))
		return
	}

	//get a valid OTP if exists then the user cannot generate new OTP
	emailOTP, err := server.store.GetValidOTP(ctx, user.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		if handleDBError(ctx, err, WithLogArgs("user_id", user.ID)) {
			return
		}
	}

	// valid OTP exists — enforce cooldown
	if err == nil && emailOTP.ExpiresAt.Time.After(time.Now()) {
		ctx.JSON(http.StatusForbidden, errorMessage("wait before requesting another OTP"))
		return
	}
	//3. generate otp
	otpCode, err := auth.GenerateOTP(6)
	if err != nil {
		slog.Error("failed to generate OTP", "err", err, "user_id", user.ID)
		ctx.JSON(http.StatusInternalServerError, errorMessage("error generating OTP Code"))
		return
	}

	//5. sign,Hash OTP and store in database
	hashedOTP, err := auth.SignOtpCode(otpCode, server.config.TokenSymmetricKey)
	if err != nil {
		slog.Error("error signing OTP Code",
			"err", err,
		)
		ctx.JSON(http.StatusInternalServerError, errorMessage("an error occourred"))
		return
	}
	_, err = server.store.CreateEmailOTP(ctx, db.CreateEmailOTPParams{
		UserID:   user.ID,
		CodeHash: hashedOTP,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(server.config.AcessTokenDuration),
			Valid: true,
		},
	})
	if handleDBError(ctx, err, WithLogArgs("user_id", user.ID)) {
		return
	}
	//4. send email otp to user email
	err = server.emailClient.SendEmailOTP(user.Email, otpCode)
	if err != nil {
		slog.Error("failed to send OTP email", "err", err, "user_id", user.ID)
		ctx.JSON(http.StatusInternalServerError, errorMessage("failed to send OTP email"))
		return
	}
	ctx.JSON(http.StatusOK, successMessage("OTP sent successfuly to your email"))

}

type loginUserRequest struct {
	Email string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	Username          string    `json:"username"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	LastLoginAt         time.Time `json:"last_login_at"`
	AccessToken       string    `json:"access_token"`
	RefreshToken      string    `json:"refresh_token"`
}

func (r *loginUserRequest)sanitizeLoginUserReq(){
	r.Email = strings.ToLower(r.Email)
}

// login user
func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	//sanitize input
	req.sanitizeLoginUserReq()

	user, err := server.store.GetUserByEmail(ctx,req.Email)
	if handleDBError(ctx,err,WithNotFoundMsg("user does not exist"),
	WithLogArgs("user_email",req.Email,
	)){
		return
	}

	//check user password against saved db password
	err = auth.CheckPassword(user.PasswordHash,req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorMessage("Invalid email or password"))
		return
	}

	//create the access token
	access_token, err := server.tokenMaker.CreateToken(user.Username, token.AccessToken, server.config.AcessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	//create the refresh token signed with 1 hour and save to the database
	week := time.Hour * 24 * 7

	refresh_token, err := server.tokenMaker.CreateToken(user.Username, token.RefreshToken, week)
	if err != nil {
		slog.Error("failed creating refresh token","user_id",user.ID)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	//create user session + refreshtoken in db
	ctx.Request.UserAgent()
	ctx.ClientIP()

	savedSession, err := server.store.CreateSession(ctx,db.CreateSessionParams{
		UserID: user.ID,
		RefreshToken: refresh_token,
		UserAgent: pgtype.Text{
			String: ctx.Request.UserAgent(),
			Valid: true,
		},
		ClientIp: pgtype.Text{
			String: ctx.ClientIP(),
			Valid: true,
		},
		ExpiresAt: pgtype.Timestamptz{
			Time: time.Now().Add(week),
			Valid: true,
		},
	})
	if handleDBError(ctx,err,WithLogArgs("user_id",user.ID)){
		return
	}

	last_login := user.LastLoginAt
	err = server.store.UpdateLastLogin(ctx,user.ID)
	if handleDBError(ctx,err,WithLogArgs("user_id",user.ID)){
		return
	}

	resp := loginUserResponse{
		Username: user.Username,
		PasswordChangedAt: user.PasswordUpdatedAt.Time,
		LastLoginAt: last_login.Time,
		AccessToken: access_token,
		RefreshToken: savedSession.RefreshToken,
	}

	
	ctx.JSON(http.StatusOK, resp)

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
