package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Glenn444/golang-chess/config"
	db "github.com/Glenn444/golang-chess/internal/db"
	mock_db "github.com/Glenn444/golang-chess/internal/db/mock"
	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/internal/token"
	"github.com/Glenn444/golang-chess/internal/utils/auth"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// ── mock email sender ────────────────────────────────────────────────────────────

type mockEmailSender struct {
	sendEmailOTPFn func(to, otp string) error
}

func (m *mockEmailSender) SendEmailOTP(to, otp string) error {
	return m.sendEmailOTPFn(to, otp)
}

// ── test server (auth) ───────────────────────────────────────────────────────────

func newTestAuthServer(t *testing.T) (*Server, *mock_db.MockStore, *mockEmailSender) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)
	store := mock_db.NewMockStore(ctrl)
	tokenMaker, _ := token.NewJWTMaker("12345678901234567890123456789012")
	emailMock := &mockEmailSender{}
	server := &Server{
		config: config.Config{
			TokenSymmetricKey:  "12345678901234567890123456789012",
			AcessTokenDuration: 15 * time.Minute,
		},
		tokenMaker:  tokenMaker,
		store:       store,
		emailClient: emailMock,
		activeGames: make(map[pgtype.UUID]*pieces.GameState),
	}
	return server, store, emailMock
}

func authCtx(method, path string, body any) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)

	var r *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewBuffer(b)
	} else {
		r = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, r)
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req
	return ctx, rec
}

// ── createUser tests ─────────────────────────────────────────────────────────────

func TestCreateUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, store, emailMock := newTestAuthServer(t)

		reqBody := CreateUserRequest{
			Username: "newuser",
			Email:    "new@example.com",
			Password: "password123",
		}

		hashedPw, _ := auth.HashPassword(reqBody.Password)
		user := db.User{
			ID:             userUUID(),
			Username:       reqBody.Username,
			Email:          reqBody.Email,
			PasswordHash:   hashedPw,
			EmailConfirmed: false,
			IsActive:       true,
			CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
			UpdatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}

		emailMock.sendEmailOTPFn = func(to, otp string) error {
			require.Equal(t, reqBody.Email, to)
			require.Len(t, otp, 6)
			return nil
		}

		store.EXPECT().UsernameExists(gomock.Any(), reqBody.Username).Return(false, nil)
		store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(user, nil)
		store.EXPECT().CreateEmailOTP(gomock.Any(), gomock.Any()).Return(db.EmailOtp{}, nil)

		ctx, rec := authCtx(http.MethodPost, "/users/signup", reqBody)
		server.createUser(ctx)

		require.Equal(t, http.StatusOK, rec.Code)

		var resp CreateUserResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		require.Equal(t, reqBody.Username, resp.Username)
	})

	t.Run("username already exists", func(t *testing.T) {
		server, store, _ := newTestAuthServer(t)

		reqBody := CreateUserRequest{
			Username: "takenuser",
			Email:    "taken@example.com",
			Password: "password123",
		}

		store.EXPECT().UsernameExists(gomock.Any(), reqBody.Username).Return(true, nil)

		ctx, rec := authCtx(http.MethodPost, "/users/signup", reqBody)
		server.createUser(ctx)

		require.Equal(t, http.StatusConflict, rec.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		server, _, _ := newTestAuthServer(t)

		ctx, rec := authCtx(http.MethodPost, "/users/signup", gin.H{
			"username": "ab", // too short? no min length validation, but it should pass basic validation
			"email":    "not-an-email",
			"password": "12345", // min=6, this is only 5
		})
		server.createUser(ctx)

		require.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

// ── confirmEmail tests ───────────────────────────────────────────────────────────

func TestConfirmEmail(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, store, _ := newTestAuthServer(t)

		otpCode, _ := auth.GenerateOTP(6)
		otpHash, _ := auth.SignOtpCode(otpCode, server.config.TokenSymmetricKey)

		user := db.User{
			ID:             userUUID(),
			Username:       "verifyuser",
			Email:          "verify@example.com",
			EmailConfirmed: false,
			IsActive:       true,
		}

		emailOTP := db.EmailOtp{
			ID:       userUUID(),
			UserID:   user.ID,
			CodeHash: otpHash,
			Attempts: 0,
		}

		store.EXPECT().GetUserByEmail(gomock.Any(), user.Email).Return(user, nil)
		store.EXPECT().GetValidOTP(gomock.Any(), user.ID).Return(emailOTP, nil)
		store.EXPECT().MarkOTPUsed(gomock.Any(), emailOTP.ID).Return(db.EmailOtp{}, nil)
		store.EXPECT().ConfirmEmail(gomock.Any(), user.ID).Return(db.User{}, nil)

		ctx, rec := authCtx(http.MethodPost, "/users/confirm-email", ConfirmEmail{
			Email:     user.Email,
			EMAIL_OTP: otpCode,
		})
		server.confirmEmail(ctx)

		require.Equal(t, http.StatusOK, rec.Code)

		var resp ConfirmEmailResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		require.Equal(t, "email verified successfully", resp.Message)
	})

	t.Run("user not found", func(t *testing.T) {
		server, store, _ := newTestAuthServer(t)

		store.EXPECT().GetUserByEmail(gomock.Any(), "unknown@example.com").Return(db.User{}, pgx.ErrNoRows)

		ctx, rec := authCtx(http.MethodPost, "/users/confirm-email", ConfirmEmail{
			Email:     "unknown@example.com",
			EMAIL_OTP: "123456",
		})
		server.confirmEmail(ctx)

		require.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("email already verified", func(t *testing.T) {
		server, store, _ := newTestAuthServer(t)

		user := db.User{
			ID:             userUUID(),
			Email:          "already@example.com",
			EmailConfirmed: true,
		}

		store.EXPECT().GetUserByEmail(gomock.Any(), user.Email).Return(user, nil)

		ctx, rec := authCtx(http.MethodPost, "/users/confirm-email", ConfirmEmail{
			Email:     user.Email,
			EMAIL_OTP: "123456",
		})
		server.confirmEmail(ctx)

		require.Equal(t, http.StatusConflict, rec.Code)
	})

	t.Run("wrong OTP code", func(t *testing.T) {
		server, store, _ := newTestAuthServer(t)

		user := db.User{
			ID:             userUUID(),
			Email:          "verify@example.com",
			EmailConfirmed: false,
		}

		// create an OTP with a different code
		otherCode, _ := auth.GenerateOTP(6)
		otherHash, _ := auth.SignOtpCode(otherCode, server.config.TokenSymmetricKey)

		emailOTP := db.EmailOtp{
			ID:       userUUID(),
			UserID:   user.ID,
			CodeHash: otherHash,
			Attempts: 0,
		}

		store.EXPECT().GetUserByEmail(gomock.Any(), user.Email).Return(user, nil)
		store.EXPECT().GetValidOTP(gomock.Any(), user.ID).Return(emailOTP, nil)
		store.EXPECT().IncrementOTPAttempts(gomock.Any(), emailOTP.ID).Return(db.EmailOtp{}, nil)

		ctx, rec := authCtx(http.MethodPost, "/users/confirm-email", ConfirmEmail{
			Email:     user.Email,
			EMAIL_OTP: "999999", // wrong code
		})
		server.confirmEmail(ctx)

		require.Equal(t, http.StatusForbidden, rec.Code)
	})
}

// ── loginUser tests ──────────────────────────────────────────────────────────────

func TestLoginUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, store, _ := newTestAuthServer(t)

		password := "password123"
		hashedPw, _ := auth.HashPassword(password)

		user := db.User{
			ID:             userUUID(),
			Username:       "loginuser",
			Email:          "login@example.com",
			PasswordHash:   hashedPw,
			EmailConfirmed: true,
			IsActive:       true,
			LastLoginAt:    pgtype.Timestamptz{Time: time.Now().Add(-24 * time.Hour), Valid: true},
		}

		store.EXPECT().GetUserByEmail(gomock.Any(), user.Email).Return(user, nil)
		store.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Return(db.Session{
			ID:           userUUID(),
			UserID:       user.ID,
			RefreshToken: "some-refresh-token",
		}, nil)
		store.EXPECT().UpdateLastLogin(gomock.Any(), user.ID).Return(nil)

		ctx, rec := authCtx(http.MethodPost, "/users/signin", LoginUserRequest{
			Email:    user.Email,
			Password: password,
		})
		server.loginUser(ctx)

		require.Equal(t, http.StatusOK, rec.Code)

		var resp LoginUserResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		require.Equal(t, user.Username, resp.Username)
	
	// Tokens are now set as HttpOnly cookies, not in the body.
	cookies := rec.Result().Cookies()
	var accessCookie, refreshCookie bool
	for _, c := range cookies {
		if c.Name == "access_token" && c.Value != "" {
			accessCookie = true
		}
		if c.Name == "refresh_token" && c.Value != "" {
			refreshCookie = true
		}
	}
	require.True(t, accessCookie, "access_token cookie not set")
	require.True(t, refreshCookie, "refresh_token cookie not set")
	})

	t.Run("wrong password", func(t *testing.T) {
		server, store, _ := newTestAuthServer(t)

		hashedPw, _ := auth.HashPassword("correct_password")

		user := db.User{
			ID:             userUUID(),
			Username:       "loginuser",
			Email:          "login@example.com",
			PasswordHash:   hashedPw,
			EmailConfirmed: true,
			IsActive:       true,
		}

		store.EXPECT().GetUserByEmail(gomock.Any(), user.Email).Return(user, nil)

		ctx, rec := authCtx(http.MethodPost, "/users/signin", LoginUserRequest{
			Email:    user.Email,
			Password: "wrong_password",
		})
		server.loginUser(ctx)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("user not found", func(t *testing.T) {
		server, store, _ := newTestAuthServer(t)

		store.EXPECT().GetUserByEmail(gomock.Any(), "nobody@example.com").Return(db.User{}, pgx.ErrNoRows)

		ctx, rec := authCtx(http.MethodPost, "/users/signin", LoginUserRequest{
			Email:    "nobody@example.com",
			Password: "anything123",
		})
		server.loginUser(ctx)

		require.Equal(t, http.StatusNotFound, rec.Code)
	})
}

// ── refreshToken tests ───────────────────────────────────────────────────────────

func TestRefreshToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, store, _ := newTestAuthServer(t)

		// create a real refresh token
		refreshToken, _ := server.tokenMaker.CreateToken("testuser", token.RefreshTokenType, 7*24*time.Hour)

		session := db.Session{
			ID:           userUUID(),
			UserID:       userUUID(),
			RefreshToken: refreshToken,
			IsRevoked:    false,
			ExpiresAt:    pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour), Valid: true},
		}

		store.EXPECT().GetSessionByRefreshToken(gomock.Any(), refreshToken).Return(session, nil)

		ctx, rec := authCtx(http.MethodPost, "/users/refresh-token", RefreshTokenRequest{
			RefreshToken: refreshToken,
		})
		server.refreshToken(ctx)

		require.Equal(t, http.StatusOK, rec.Code)

		var resp RefreshTokenResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		require.NotEmpty(t, resp.AccessToken)
	})

	t.Run("invalid token", func(t *testing.T) {
		server, _, _ := newTestAuthServer(t)

		ctx, rec := authCtx(http.MethodPost, "/users/refresh-token", RefreshTokenRequest{
			RefreshToken: "not-a-valid-jwt-token",
		})
		server.refreshToken(ctx)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("session revoked", func(t *testing.T) {
		server, store, _ := newTestAuthServer(t)

		refreshToken, _ := server.tokenMaker.CreateToken("testuser", token.RefreshTokenType, 7*24*time.Hour)

		session := db.Session{
			ID:           userUUID(),
			UserID:       userUUID(),
			RefreshToken: refreshToken,
			IsRevoked:    true,
			ExpiresAt:    pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour), Valid: true},
		}

		store.EXPECT().GetSessionByRefreshToken(gomock.Any(), refreshToken).Return(session, nil)

		ctx, rec := authCtx(http.MethodPost, "/users/refresh-token", RefreshTokenRequest{
			RefreshToken: refreshToken,
		})
		server.refreshToken(ctx)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("session expired", func(t *testing.T) {
		server, store, _ := newTestAuthServer(t)

		refreshToken, _ := server.tokenMaker.CreateToken("testuser", token.RefreshTokenType, 7*24*time.Hour)

		session := db.Session{
			ID:           userUUID(),
			UserID:       userUUID(),
			RefreshToken: refreshToken,
			IsRevoked:    false,
			ExpiresAt:    pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour), Valid: true},
		}

		store.EXPECT().GetSessionByRefreshToken(gomock.Any(), refreshToken).Return(session, nil)

		ctx, rec := authCtx(http.MethodPost, "/users/refresh-token", RefreshTokenRequest{
			RefreshToken: refreshToken,
		})
		server.refreshToken(ctx)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

// ── getMe tests ──────────────────────────────────────────────────────────────────

func TestGetMe(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, store, _ := newTestAuthServer(t)

		user := db.User{
			ID:             userUUID(),
			Username:       "meuser",
			Email:          "me@example.com",
			EmailConfirmed: true,
			IsActive:       true,
			LastLoginAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
			CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}

		store.EXPECT().GetUserByUsername(gomock.Any(), user.Username).Return(user, nil)

		ctx, rec := authCtx(http.MethodGet, "/users/me", nil)
		setAuth(ctx, user.Username)

		server.getMe(ctx)

		require.Equal(t, http.StatusOK, rec.Code)

		var resp getMeResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		require.Equal(t, user.Username, resp.Username)
		require.Equal(t, user.Email, resp.Email)
	})
}

// ── checkUsernameExists tests ─────────────────────────────────────────────────────

func TestCheckUsernameExists(t *testing.T) {
	t.Run("username exists", func(t *testing.T) {
		server, store, _ := newTestAuthServer(t)

		store.EXPECT().UsernameExists(gomock.Any(), "taken").Return(true, nil)

		ctx, rec := authCtx(http.MethodGet, "/users/check-username?username=taken", nil)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/users/check-username?username=taken", nil)
		ctx.Request.Header.Set("Content-Type", "application/json")

		server.checkUsernameExists(ctx)

		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("username does not exist", func(t *testing.T) {
		server, store, _ := newTestAuthServer(t)

		store.EXPECT().UsernameExists(gomock.Any(), "free").Return(false, nil)

		ctx, rec := authCtx(http.MethodGet, "/users/check-username?username=free", nil)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/users/check-username?username=free", nil)
		ctx.Request.Header.Set("Content-Type", "application/json")

		server.checkUsernameExists(ctx)

		require.Equal(t, http.StatusOK, rec.Code)
	})
}

