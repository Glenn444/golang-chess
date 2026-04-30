package api

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Glenn444/golang-chess/internal/db"
	"github.com/Glenn444/golang-chess/internal/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// ── DB error handling ─────────────────────────────────────────────────────────

type dbErrOptions struct {
	notFound string
	unique   string
	fk       string
	logArgs  []any
}

type DBErrOption func(*dbErrOptions)

func WithNotFoundMsg(msg string) DBErrOption {
	return func(o *dbErrOptions) { o.notFound = msg }
}

func WithUniqueMsg(msg string) DBErrOption {
	return func(o *dbErrOptions) { o.unique = msg }
}

func WithFKMsg(msg string) DBErrOption {
	return func(o *dbErrOptions) { o.fk = msg }
}

func WithLogArgs(args ...any) DBErrOption {
	return func(o *dbErrOptions) { o.logArgs = args }
}

func handleDBError(ctx *gin.Context, err error, opts ...DBErrOption) bool {
	if err == nil {
		return false
	}

	o := &dbErrOptions{
		notFound: "resource not found",
		unique:   "resource already exists",
		fk:       "referenced resource does not exist",
	}
	for _, opt := range opts {
		opt(o)
	}

	if errors.Is(err, pgx.ErrNoRows) {
		ctx.JSON(http.StatusNotFound, errorMessage(o.notFound))
		return true
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			ctx.JSON(http.StatusConflict, errorMessage(o.unique))
			return true
		case pgerrcode.ForeignKeyViolation:
			ctx.JSON(http.StatusUnprocessableEntity, errorMessage(o.fk))
			return true
		case pgerrcode.NotNullViolation, pgerrcode.CheckViolation:
			ctx.JSON(http.StatusBadRequest, errorMessage("invalid data provided"))
			return true
		default:
			slog.Error("postgres error", append(o.logArgs, "code", pgErr.Code, "err", pgErr.Message)...)
			ctx.JSON(http.StatusInternalServerError, errorMessage(ErrInternalServer))
			return true
		}
	}

	slog.Error("unexpected db error", append(o.logArgs, "err", err)...)
	ctx.JSON(http.StatusInternalServerError, errorMessage(ErrInternalServer))
	return true
}

// ── Response helpers ──────────────────────────────────────────────────────────

func errorResponse(err error) gin.H { return gin.H{"error": err.Error()} }
func errorMessage(msg string) gin.H  { return gin.H{"error": msg} }
func successMessage(msg string) gin.H { return gin.H{"message": msg} }

// ── Auth helpers ──────────────────────────────────────────────────────────────

// getCurrentUser resolves the authenticated user from the JWT payload stored in the context.
// Only call this from routes behind authMiddleware.
func (server *Server) getCurrentUser(ctx *gin.Context) (db.User, bool) {
	payload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	user, err := server.store.GetUserByUsername(ctx, payload.Username)
	if handleDBError(ctx, err, WithNotFoundMsg(ErrUserNotFound)) {
		return db.User{}, false
	}
	return user, true
}

// ── UUID helpers ──────────────────────────────────────────────────────────────

// parseUUIDParam parses a named path parameter as a pgtype.UUID.
// Writes 400 and returns false on invalid input.
func parseUUIDParam(ctx *gin.Context, param string) (pgtype.UUID, bool) {
	parsed, err := uuid.Parse(ctx.Param(param))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorMessage(ErrInvalidInput))
		return pgtype.UUID{}, false
	}
	return pgtype.UUID{Bytes: parsed, Valid: true}, true
}

// uuidEq compares two pgtype.UUID values. Returns false if either is NULL.
func uuidEq(a, b pgtype.UUID) bool {
	return a.Valid && b.Valid && a.Bytes == b.Bytes
}
