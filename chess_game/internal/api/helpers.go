package api

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

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

	// apply options
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
			ctx.JSON(http.StatusInternalServerError, errorMessage("an error occurred"))
			return true
		}
	}

	slog.Error("unexpected db error", append(o.logArgs, "err", err)...)
	ctx.JSON(http.StatusInternalServerError, errorMessage("an error occurred"))
	return true
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

func errorMessage(message string) gin.H {
	return gin.H{"error": message}
}
func successMessage(msg string)gin.H{
	return gin.H{"message":msg}
}