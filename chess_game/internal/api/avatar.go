package api

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	// Decoders for the formats users actually upload.
	_ "image/gif"
	_ "image/png"

	db "github.com/Glenn444/golang-chess/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const (
	avatarMaxUpload = 5 << 20 // 5MB raw upload cap
	avatarSize      = 256     // stored as a 256×256 JPEG
	avatarJPEGQual  = 85

	// Decode-bomb guard: decoders allocate based on header-CLAIMED dimensions,
	// so a tiny file claiming 50000×50000 would try to allocate gigabytes. The
	// byte-size cap does not protect against this — dimensions are checked via
	// DecodeConfig (header only) before any pixel decode.
	avatarMaxDim    = 10_000
	avatarMaxPixels = 24_000_000 // 24MP — covers any phone/DSLR export
)

// @Summary      Upload profile picture
// @Description  Sets the current user's avatar. Accepts JPEG/PNG/GIF/WebP up to 5MB; the image is center-cropped to a square and stored as 256×256 JPEG.
// @Tags         Users
// @Accept       multipart/form-data
// @Produce      json
// @Param        avatar  formData  file  true  "Image file"
// @Security     Bearer
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /users/me/avatar [post]
func (server *Server) uploadAvatar(ctx *gin.Context) {
	user, ok := server.getCurrentUser(ctx)
	if !ok {
		return
	}

	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, avatarMaxUpload)
	file, _, err := ctx.Request.FormFile("avatar")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorMessage("send the image as multipart field 'avatar' (max 5MB)"))
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, avatarMaxUpload+1))
	if err != nil || len(data) > avatarMaxUpload {
		ctx.JSON(http.StatusBadRequest, errorMessage("image too large (max 5MB)"))
		return
	}

	// Header-only parse first: reject absurd claimed dimensions before the
	// real decoder allocates pixel buffers for them.
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorMessage("unsupported image — use JPEG, PNG, GIF or WebP"))
		return
	}
	if cfg.Width <= 0 || cfg.Height <= 0 ||
		cfg.Width > avatarMaxDim || cfg.Height > avatarMaxDim ||
		cfg.Width*cfg.Height > avatarMaxPixels {
		ctx.JSON(http.StatusBadRequest, errorMessage("image dimensions too large (max 24 megapixels)"))
		return
	}

	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorMessage("unsupported image — use JPEG, PNG, GIF or WebP"))
		return
	}

	jpg, err := encodeAvatar(src)
	if err != nil {
		slog.Error("uploadAvatar: encode failed", "user_id", uidStr(user.ID), "err", err)
		ctx.JSON(http.StatusInternalServerError, errorMessage(ErrInternalServer))
		return
	}

	if err := server.store.UpsertUserAvatar(ctx, db.UpsertUserAvatarParams{
		UserID: user.ID,
		Image:  jpg,
	}); err != nil {
		handleDBError(ctx, err, WithLogArgs("uploadAvatar: UpsertUserAvatar", "user_id", uidStr(user.ID)))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "avatar updated"})
}

// @Summary      Remove profile picture
// @Tags         Users
// @Produce      json
// @Security     Bearer
// @Success      200  {object}  map[string]string
// @Router       /users/me/avatar [delete]
func (server *Server) deleteAvatar(ctx *gin.Context) {
	user, ok := server.getCurrentUser(ctx)
	if !ok {
		return
	}
	if err := server.store.DeleteUserAvatar(ctx, user.ID); err != nil {
		handleDBError(ctx, err, WithLogArgs("deleteAvatar", "user_id", uidStr(user.ID)))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "avatar removed"})
}

// @Summary      Get a user's profile picture
// @Description  Returns the avatar JPEG. Public — avatars show on player cards and spectator pages.
// @Tags         Users
// @Produce      image/jpeg
// @Param        id  path  string  true  "User UUID"
// @Success      200  {file}    binary
// @Failure      404  {object}  map[string]string
// @Router       /users/{id}/avatar [get]
func (server *Server) getAvatar(ctx *gin.Context) {
	userID, ok := parseUUIDParam(ctx, "id")
	if !ok {
		return
	}

	row, err := server.store.GetUserAvatar(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Cacheable 404 — every avatar-less user triggers this once per
			// page, so let browsers remember briefly.
			ctx.Header("Cache-Control", "public, max-age=60")
			ctx.JSON(http.StatusNotFound, errorMessage("no avatar"))
			return
		}
		handleDBError(ctx, err, WithLogArgs("getAvatar", "user_id", ctx.Param("id")))
		return
	}

	etag := fmt.Sprintf(`"av-%d"`, row.UpdatedAt.Time.Unix())
	ctx.Header("Cache-Control", "public, max-age=300")
	ctx.Header("ETag", etag)
	if ctx.GetHeader("If-None-Match") == etag {
		ctx.Status(http.StatusNotModified)
		return
	}
	ctx.Header("Content-Type", "image/jpeg")
	ctx.Header("Content-Length", strconv.Itoa(len(row.Image)))
	ctx.Status(http.StatusOK)
	ctx.Writer.Write(row.Image) //nolint:errcheck
}

// encodeAvatar center-crops to a square and scales to avatarSize, returning
// JPEG bytes. CatmullRom keeps small faces sharp.
func encodeAvatar(src image.Image) ([]byte, error) {
	b := src.Bounds()
	side := min(b.Dx(), b.Dy())
	if side == 0 {
		return nil, fmt.Errorf("empty image")
	}
	crop := image.Rect(
		b.Min.X+(b.Dx()-side)/2,
		b.Min.Y+(b.Dy()-side)/2,
		b.Min.X+(b.Dx()-side)/2+side,
		b.Min.Y+(b.Dy()-side)/2+side,
	)

	dst := image.NewRGBA(image.Rect(0, 0, avatarSize, avatarSize))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, crop, draw.Src, nil)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: avatarJPEGQual}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
