package api

import (
	"errors"
	"fmt"
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

const (
	// A confirm-email code must never be usable to reset a password (and vice
	// versa) — every OTP is scoped to one purpose.
	otpPurposeConfirmEmail  = "confirm_email"
	otpPurposePasswordReset = "password_reset"

	// otpLifetime matches the "expires in 15 minutes" wording in the emails.
	otpLifetime = 15 * time.Minute
	// A user may request a fresh code this long after the previous one.
	otpResendCooldown = time.Minute
)

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=30,username"`
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

// @Summary      Register a new user
// @Description  Creates an account and sends a 6-digit OTP to the provided email.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body  CreateUserRequest  true  "Registration payload"
// @Success      200   {object}  CreateUserResponse
// @Failure      400   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /users/signup [post]
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
		WithLogArgs("createUser: failed to check UserNameExists", "username", req.Username),
	) {
		return
	}

	if userExists {
		ctx.JSON(http.StatusConflict, errorMessage(ErrUserAlreadyExists))
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		slog.Error("createUser: failed to hash Password", "err", err, "user_email", req.Email)
		ctx.JSON(http.StatusInternalServerError, errorMessage(ErrInternalServer))
		return
	}

	arg := db.CreateUserParams{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if handleDBError(ctx, err,
		WithUniqueMsg(ErrUserAlreadyExists),
		WithLogArgs("createUser: failed creating user in db ", "user_email", req.Email)) {
		return
	}

	otpCode, err := auth.GenerateOTP(6)
	if err != nil {
		slog.Error("createrUser: failed to generate OTP Code", "err", err, "user_id", user.ID)
		ctx.JSON(http.StatusInternalServerError, errorMessage(ErrGeneratingOTP))
		return
	}
	otpHash, err := auth.SignOtpCode(otpCode, server.config.TokenSymmetricKey)
	if err != nil {
		slog.Error("createUser: failed SignOtpCode", "err", err, "user_id", user.ID)
		ctx.JSON(http.StatusInternalServerError, errorMessage(ErrInternalServer))
		return
	}

	argEmailOtp := db.CreateEmailOTPParams{
		UserID:    user.ID,
		CodeHash:  otpHash,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(otpLifetime), Valid: true},
		Purpose:   otpPurposeConfirmEmail,
	}
	//verify email
	_, err = server.store.CreateEmailOTP(ctx, argEmailOtp)
	if handleDBError(ctx, err,
		WithLogArgs("createUser: failed CreateEmailOTP", "user_id", user.ID),
	) {
		return
	}

	//send the otp to the user email
	err = server.emailClient.SendEmailOTP(user.Email, otpCode)
	if err != nil {
		slog.Error("createUser: failed SendEmailOTP", "err", err, "user_email", user.Email)
		ctx.JSON(http.StatusInternalServerError, errorMessage(ErrSendingOTP))
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

// @Summary      Confirm email
// @Description  Verifies email address with the 6-digit OTP code.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body  ConfirmEmail  true  "OTP payload"
// @Success      200   {object}  ConfirmEmailResponse
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /users/confirm-email [post]
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
		WithNotFoundMsg(ErrUserNotFound),
		WithLogArgs("confirmEmail: failed GetUserByEmail", "user_email", req.Email),
	) {
		return
	}

	//if user is already verified fail the request
	if user.EmailConfirmed {
		ctx.JSON(http.StatusConflict, errorMessage(ErrEmailAlreadyVerified))
		return
	}

	//2. Get valid OTP Code from db
	emailOTP, err := server.store.GetValidOTP(ctx, db.GetValidOTPParams{
		UserID:  user.ID,
		Purpose: otpPurposeConfirmEmail,
	})
	if handleDBError(ctx, err,
		WithNotFoundMsg(ErrOTPNotFound),
		WithLogArgs("confirmEmail: failed GetValidOTP", "user_id", user.ID),
	) {
		return
	}

	//3. Compare the OTP hash to the given OTP
	match, err := auth.ConfirmOTP(req.EMAIL_OTP, emailOTP.CodeHash, server.config.TokenSymmetricKey)
	if err != nil {
		slog.Error("confirmEmail: failed confirmOTP", "err", err, "user_id", user.ID)
		ctx.JSON(http.StatusInternalServerError, errorMessage(ErrInternalServer))
		return
	}
	if !match {
		_, err := server.store.IncrementOTPAttempts(ctx, emailOTP.ID)
		if err != nil {
			slog.Error("confirmEmail: failed IncrementOTPAttempts", "err", err, "user_id", user.ID)
			ctx.JSON(http.StatusInternalServerError, errorMessage(ErrInternalServer))
			return
		}

		//successful increment
		//slog.Error("failed to confirm OTP Code", "err", err, "user_id", user.ID)
		ctx.JSON(http.StatusForbidden, errorMessage(ErrOTPInvalid))
		return
	}
	//4. successful match, mark the OTP as used
	_, err = server.store.MarkOTPUsed(ctx, emailOTP.ID)
	if handleDBError(ctx, err, WithLogArgs("confirmEmail: failed MarkOTPUsed", "OTP_id", emailOTP.ID)) {
		return
	}

	//5. Mark user email as verified
	_, err = server.store.ConfirmEmail(ctx, user.ID)
	if err != nil {
		slog.Error("confirmEmail: failed ConfirmEmail", "err", err, "user_id", user.ID)
		ctx.JSON(http.StatusInternalServerError, errorMessage(ErrInternalServer))
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

type SendEmailOTPResp struct {
	Message string `json:"msg"`
	Email   string `json:"email"`
}

// @Summary      Send OTP
// @Description  Resends the OTP verification code to the user's email.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body  SendEmailOTP  true  "Email payload"
// @Success      200   {object}  SendEmailOTPResp
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /users/send-emailotp [post]
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
		WithNotFoundMsg(ErrUserNotFound),
		WithLogArgs("sendEmailOTP: failed GetUserByEmail", "user_email", req.Email),
	) {
		return
	}

	if user.EmailConfirmed {
		ctx.JSON(http.StatusConflict, errorMessage(ErrEmailAlreadyVerified))
		return
	}

	//get a valid OTP if exists then the user cannot generate new OTP
	emailOTP, err := server.store.GetValidOTP(ctx, db.GetValidOTPParams{
		UserID:  user.ID,
		Purpose: otpPurposeConfirmEmail,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		if handleDBError(ctx, err, WithLogArgs("sendEmailOTP: failed GetValidOTP", "user_id", user.ID)) {
			return
		}
	}

	// A live code exists — enforce a short resend cooldown, then burn it so
	// only one code is ever valid at a time.
	if err == nil {
		if cooldownEnds := emailOTP.CreatedAt.Time.Add(otpResendCooldown); time.Now().Before(cooldownEnds) {
			remaining := time.Until(cooldownEnds).Round(time.Second)
			ctx.JSON(http.StatusTooManyRequests, errorMessage(
				fmt.Sprintf("please wait %s before requesting a new OTP", remaining),
			))
			return
		}
		if err := server.store.InvalidateUserOTPs(ctx, db.InvalidateUserOTPsParams{
			UserID:  user.ID,
			Purpose: otpPurposeConfirmEmail,
		}); err != nil {
			slog.Error("sendEmailOTP: failed InvalidateUserOTPs", "err", err, "user_id", user.ID)
			ctx.JSON(http.StatusInternalServerError, errorMessage(ErrInternalServer))
			return
		}
	}
	//3. generate otp
	otpCode, err := auth.GenerateOTP(6)
	if err != nil {
		slog.Error("sendEmailOTP: failed GenerateOTP", "err", err, "user_id", user.ID)
		ctx.JSON(http.StatusInternalServerError, errorMessage(ErrGeneratingOTP))
		return
	}

	//5. sign,Hash OTP and store in database
	hashedOTP, err := auth.SignOtpCode(otpCode, server.config.TokenSymmetricKey)
	if err != nil {
		slog.Error("sendEmailOTP: failed SignOtpCode",
			"err", err,
			"user_ID", user.ID,
		)
		ctx.JSON(http.StatusInternalServerError, errorMessage(ErrInternalServer))
		return
	}
	_, err = server.store.CreateEmailOTP(ctx, db.CreateEmailOTPParams{
		UserID:   user.ID,
		CodeHash: hashedOTP,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(otpLifetime),
			Valid: true,
		},
		Purpose: otpPurposeConfirmEmail,
	})
	if handleDBError(ctx, err, WithLogArgs("sendEmailOT: failed CreateEmailOTP", "user_id", user.ID)) {
		return
	}
	//4. send email otp to user email
	err = server.emailClient.SendEmailOTP(user.Email, otpCode)
	if err != nil {
		slog.Error("sendEmailOTP: failed SendEmailOTP", "err", err, "user_id", user.ID)
		ctx.JSON(http.StatusInternalServerError, errorMessage(ErrSendingOTP))
		return
	}
	ctx.JSON(http.StatusOK, SendEmailOTPResp{
		Message: "OTP sent successfuly to your email",
		Email:   user.Email,
	})

}

// ── Forgot / Reset Password ─────────────────────────────────────────────────

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// @Summary      Forgot password
// @Description  Sends a 6-digit OTP to the user's email for password reset.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body  ForgotPasswordRequest  true  "Email payload"
// @Success      200   {object}  SendEmailOTPResp
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      429   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /users/forgot-password [post]
func (server *Server) forgotPassword(ctx *gin.Context) {
	var req ForgotPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	req.Email = strings.ToLower(req.Email)

	// Don't reveal whether an email is registered: unknown addresses get the
	// same success response as known ones.
	genericResp := SendEmailOTPResp{
		Message: "if that email is registered, a password reset code has been sent",
		Email:   req.Email,
	}

	user, err := server.store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusOK, genericResp)
			return
		}
		handleDBError(ctx, err, WithLogArgs("forgotPassword: GetUserByEmail", "email", req.Email))
		return
	}

	// Rate-limit: a fresh code can be requested after the cooldown; the old
	// one is burnt so only one live code exists per purpose.
	emailOTP, err := server.store.GetValidOTP(ctx, db.GetValidOTPParams{
		UserID:  user.ID,
		Purpose: otpPurposePasswordReset,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		handleDBError(ctx, err, WithLogArgs("forgotPassword: GetValidOTP", "user_id", user.ID))
		return
	}
	if err == nil {
		if cooldownEnds := emailOTP.CreatedAt.Time.Add(otpResendCooldown); time.Now().Before(cooldownEnds) {
			ctx.JSON(http.StatusTooManyRequests, errorMessage("a reset code was already sent — check your email or wait before requesting another"))
			return
		}
		if err := server.store.InvalidateUserOTPs(ctx, db.InvalidateUserOTPsParams{
			UserID:  user.ID,
			Purpose: otpPurposePasswordReset,
		}); err != nil {
			slog.Error("forgotPassword: InvalidateUserOTPs", "err", err, "user_id", user.ID)
			ctx.JSON(http.StatusInternalServerError, errorMessage(ErrInternalServer))
			return
		}
	}

	otpCode, err := auth.GenerateOTP(6)
	if err != nil {
		slog.Error("forgotPassword: GenerateOTP", "err", err, "user_id", user.ID)
		ctx.JSON(http.StatusInternalServerError, errorMessage(ErrGeneratingOTP))
		return
	}

	otpHash, err := auth.SignOtpCode(otpCode, server.config.TokenSymmetricKey)
	if err != nil {
		slog.Error("forgotPassword: SignOtpCode", "err", err, "user_id", user.ID)
		ctx.JSON(http.StatusInternalServerError, errorMessage(ErrInternalServer))
		return
	}

	_, err = server.store.CreateEmailOTP(ctx, db.CreateEmailOTPParams{
		UserID:   user.ID,
		CodeHash: otpHash,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(otpLifetime),
			Valid: true,
		},
		Purpose: otpPurposePasswordReset,
	})
	if handleDBError(ctx, err, WithLogArgs("forgotPassword: CreateEmailOTP", "user_id", user.ID)) {
		return
	}

	err = server.emailClient.SendPasswordResetOTP(user.Email, otpCode)
	if err != nil {
		slog.Error("forgotPassword: SendPasswordResetOTP", "err", err, "user_id", user.ID)
		ctx.JSON(http.StatusInternalServerError, errorMessage(ErrSendingOTP))
		return
	}

	ctx.JSON(http.StatusOK, genericResp)
}

type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	OTP         string `json:"otp" binding:"required,numeric,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// @Summary      Reset password
// @Description  Resets the user's password using the OTP code sent to their email.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body  ResetPasswordRequest  true  "Reset payload"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /users/reset-password [post]
func (server *Server) resetPassword(ctx *gin.Context) {
	var req ResetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	req.Email = strings.ToLower(req.Email)

	user, err := server.store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Same response as a wrong code — don't reveal whether the email exists.
			ctx.JSON(http.StatusForbidden, errorMessage(ErrOTPInvalid))
			return
		}
		handleDBError(ctx, err, WithLogArgs("resetPassword: GetUserByEmail", "email", req.Email))
		return
	}

	emailOTP, err := server.store.GetValidOTP(ctx, db.GetValidOTPParams{
		UserID:  user.ID,
		Purpose: otpPurposePasswordReset,
	})
	if handleDBError(ctx, err,
		WithNotFoundMsg(ErrOTPNotFound),
		WithLogArgs("resetPassword: GetValidOTP", "user_id", user.ID)) {
		return
	}

	match, err := auth.ConfirmOTP(req.OTP, emailOTP.CodeHash, server.config.TokenSymmetricKey)
	if err != nil {
		slog.Error("resetPassword: ConfirmOTP", "err", err, "user_id", user.ID)
		ctx.JSON(http.StatusInternalServerError, errorMessage(ErrInternalServer))
		return
	}
	if !match {
		_, _ = server.store.IncrementOTPAttempts(ctx, emailOTP.ID)
		ctx.JSON(http.StatusForbidden, errorMessage(ErrOTPInvalid))
		return
	}

	// OTP valid — update the password.
	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		slog.Error("resetPassword: HashPassword", "err", err, "user_id", user.ID)
		ctx.JSON(http.StatusInternalServerError, errorMessage(ErrInternalServer))
		return
	}

	_, err = server.store.UpdateUser(ctx, db.UpdateUserParams{
		ID:           user.ID,
		PasswordHash: pgtype.Text{String: hashedPassword, Valid: true},
	})
	if handleDBError(ctx, err, WithLogArgs("resetPassword: UpdateUser", "user_id", user.ID)) {
		return
	}

	// Mark OTP as used and revoke all existing sessions (security). The
	// password is already changed, so respond 200 either way — but a failure
	// here means a replayable code or live sessions, so log it loudly.
	if _, err := server.store.MarkOTPUsed(ctx, emailOTP.ID); err != nil {
		slog.Error("resetPassword: MarkOTPUsed failed — reset code is still replayable", "err", err, "user_id", user.ID)
	}
	if err := server.store.RevokeAllUserSessions(ctx, user.ID); err != nil {
		slog.Error("resetPassword: RevokeAllUserSessions failed — old sessions remain active", "err", err, "user_id", user.ID)
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "password reset successfully — please sign in again"})
}

type LoginUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginUserResponse struct {
	Username          string    `json:"username"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	LastLoginAt       time.Time `json:"last_login_at"`
	AccessToken       string    `json:"access_token"`
	RefreshToken      string    `json:"refresh_token"`
}

func (r *LoginUserRequest) sanitizeLoginUserReq() {
	r.Email = strings.ToLower(r.Email)
}

type EmailConfirmedResp struct {
	Message string `json:"msg"`
	Email   string `json:"email"`
}

// login user
// @Summary      Sign in
// @Description  Authenticates with email and password. Returns access + refresh tokens.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body  LoginUserRequest  true  "Login payload"
// @Success      200   {object}  LoginUserResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /users/signin [post]
func (server *Server) loginUser(ctx *gin.Context) {
	var req LoginUserRequest
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	//sanitize input
	req.sanitizeLoginUserReq()

	user, err := server.store.GetUserByEmail(ctx, req.Email)
	if handleDBError(ctx, err, WithNotFoundMsg(ErrUserNotFound),
		WithLogArgs("loginUser: failed GetUserByEmail", "user_email", req.Email)) {
		return
	}

	//check if user email is verified
	if !user.EmailConfirmed {
		ctx.JSON(http.StatusForbidden, EmailConfirmedResp{
			Message: "email not confirmed!",
			Email:   user.Email,
		})
		return
	}

	//check user password against saved db password
	err = auth.CheckPassword(user.PasswordHash, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorMessage(ErrInvalidCredentials))
		return
	}

	//create the access token
	access_token, err := server.tokenMaker.CreateToken(user.Username, token.AccessTokenType, server.config.AcessTokenDuration)
	if err != nil {
		slog.Error("loginUser: failed CreateToken - accessToken", "err:", err, "user_email", req.Email)
		ctx.JSON(http.StatusInternalServerError, errorMessage(ErrInternalServer))
		return
	}

	//create the refresh token signed with 1 hour and save to the database
	week := time.Hour * 24 * 7

	refresh_token, err := server.tokenMaker.CreateToken(user.Username, token.RefreshTokenType, week)
	if err != nil {
		slog.Error("loginUser: failed CreateToken - refreshToken", "err:", err, "user_id", user.ID)
		ctx.JSON(http.StatusInternalServerError, errorMessage(ErrInternalServer))
		return
	}

	//create user session + refreshtoken in db

	savedSession, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		UserID:       user.ID,
		RefreshToken: refresh_token,
		UserAgent: pgtype.Text{
			String: ctx.Request.UserAgent(),
			Valid:  true,
		},
		ClientIp: pgtype.Text{
			String: ctx.ClientIP(),
			Valid:  true,
		},
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(week),
			Valid: true,
		},
	})
	if handleDBError(ctx, err, WithLogArgs("loginUser: failed CreateSession", "user_id", user.ID)) {
		return
	}

	last_login := user.LastLoginAt
	err = server.store.UpdateLastLogin(ctx, user.ID)
	if handleDBError(ctx, err, WithLogArgs("loginUser: failed UpdateLastLogin", "user_id", user.ID)) {
		return
	}

	// Set HttpOnly cookies so the browser handles auth automatically.
	maxAgeAccess := int(server.config.AcessTokenDuration.Seconds())
	maxAgeRefresh := int(week.Seconds())
	domain, secure := server.cookieConfig()

	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie("access_token", access_token, maxAgeAccess, "/", domain, secure, true)
	ctx.SetCookie("refresh_token", savedSession.RefreshToken, maxAgeRefresh, "/", domain, secure, true)

	resp := LoginUserResponse{
		Username:          user.Username,
		PasswordChangedAt: user.PasswordUpdatedAt.Time,
		LastLoginAt:       last_login.Time,
		AccessToken:       access_token,
		RefreshToken:      savedSession.RefreshToken,
	}

	ctx.JSON(http.StatusOK, resp)

}

// @Summary      Logout
// @Description  Clears auth cookies and revokes the refresh token session.
// @Tags         Auth
// @Produce      json
// @Security     Bearer
// @Success      200  {object}  map[string]string
// @Router       /users/logout [post]
func (server *Server) logoutUser(ctx *gin.Context) {
	// Try to revoke the session if we can read the refresh token.
	if refreshToken, err := ctx.Cookie("refresh_token"); err == nil && refreshToken != "" {
		if err := server.store.RevokeSession(ctx, refreshToken); err != nil {
			slog.Error("logoutUser: RevokeSession failed", "err", err)
		}
	}

	server.clearAuthCookies(ctx)

	ctx.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// clearAuthCookies expires both auth cookies on the configured domain.
func (server *Server) clearAuthCookies(ctx *gin.Context) {
	domain, secure := server.cookieConfig()
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie("access_token", "", -1, "/", domain, secure, true)
	ctx.SetCookie("refresh_token", "", -1, "/", domain, secure, true)
}

// RefreshTokenRequest accepts the token via cookie or JSON body.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenResponse struct {
	Message     string `json:"message"`
	AccessToken string `json:"access_token"`
}

// @Summary      Refresh token
// @Description  Issues a new access token using a valid refresh token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body  RefreshTokenRequest  false  "Refresh token payload"
// @Success      200   {object}  RefreshTokenResponse
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /users/refresh-token [post]
func (server *Server) refreshToken(ctx *gin.Context) {
	// Read refresh token from cookie first, fall back to JSON body.
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		var req RefreshTokenRequest
		if err := ctx.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
			ctx.JSON(http.StatusUnauthorized, errorMessage(ErrInvalidToken))
			return
		}
		refreshToken = req.RefreshToken
	}

	//verify the refresh token and get the payload
	payload, err := server.tokenMaker.VerifyToken(refreshToken, token.RefreshTokenType)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorMessage(ErrInvalidToken))
		return
	}

	// check session in DB — this is what JWT alone can't do
	session, err := server.store.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		// Clear the stale cookies BEFORE the error body is written — headers
		// can't be added once the response has been flushed.
		server.clearAuthCookies(ctx)
		handleDBError(ctx, err,
			WithNotFoundMsg(ErrSessionNotFound),
			WithLogArgs("refreshToken: GetSessionByRefreshToken"))
		return
	}

	// 3. check not revoked
	if session.IsRevoked {
		ctx.JSON(http.StatusUnauthorized, errorMessage(ErrSessionRevoked))
		return
	}

	// 4. check not expired
	if session.ExpiresAt.Time.Before(time.Now()) {
		ctx.JSON(http.StatusUnauthorized, errorMessage(ErrSessionExpired))
		return
	}

	//refreshtoken is valid issue new access token
	accessToken, err := server.tokenMaker.CreateToken(payload.Subject, token.AccessTokenType, server.config.AcessTokenDuration)
	if err != nil {
		slog.Error("refreshToken: failed CreateToken - accessToken", "user_id", payload.ID)
		ctx.JSON(http.StatusInternalServerError, errorMessage(ErrInvalidToken))
		return
	}

	// Set new access token cookie.
	maxAge := int(server.config.AcessTokenDuration.Seconds())
	domain, secure := server.cookieConfig()
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie("access_token", accessToken, maxAge, "/", domain, secure, true)

	resp := RefreshTokenResponse{
		Message:     "successfully created accessToken",
		AccessToken: accessToken,
	}

	ctx.JSON(http.StatusOK, resp)
}
