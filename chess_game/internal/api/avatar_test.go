package api

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func avatarUploadCtx(t *testing.T, imageBytes []byte) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	fw, err := w.CreateFormFile("avatar", "avatar.png")
	require.NoError(t, err)
	_, err = fw.Write(imageBytes)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(http.MethodPost, "/users/me/avatar", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	ctx.Request = req
	return ctx, rec
}

// decodeBombPNG builds a few-hundred-byte file whose header claims absurd
// dimensions. image.DecodeConfig parses it fine; a naive image.Decode would
// try to allocate gigabytes for it.
func decodeBombPNG(width, height uint32) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A})

	ihdr := make([]byte, 13)
	binary.BigEndian.PutUint32(ihdr[0:4], width)
	binary.BigEndian.PutUint32(ihdr[4:8], height)
	ihdr[8] = 8  // bit depth
	ihdr[9] = 6  // RGBA
	// compression/filter/interlace = 0

	var chunk bytes.Buffer
	binary.Write(&chunk, binary.BigEndian, uint32(len(ihdr))) //nolint:errcheck
	chunk.WriteString("IHDR")
	chunk.Write(ihdr)
	crc := crc32.NewIEEE()
	crc.Write([]byte("IHDR")) //nolint:errcheck
	crc.Write(ihdr)           //nolint:errcheck
	binary.Write(&chunk, binary.BigEndian, crc.Sum32()) //nolint:errcheck
	buf.Write(chunk.Bytes())
	return buf.Bytes()
}

func smallPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 40, 40))
	for i := 0; i < 40; i++ {
		img.Set(i, i, color.RGBA{R: 229, G: 169, B: 59, A: 255})
	}
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}

func TestUploadAvatar(t *testing.T) {
	t.Run("decode bomb is rejected by header check before pixel decode", func(t *testing.T) {
		server, store := newTestGameServer(t)
		user := testUser()
		store.EXPECT().GetUserByUsername(gomock.Any(), user.Username).Return(user, nil)

		ctx, rec := avatarUploadCtx(t, decodeBombPNG(60_000, 60_000))
		setAuth(ctx, user.Username)

		server.uploadAvatar(ctx)

		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, rec.Body.String(), "dimensions too large")
	})

	t.Run("oversized-but-plausible dimensions rejected", func(t *testing.T) {
		server, store := newTestGameServer(t)
		user := testUser()
		store.EXPECT().GetUserByUsername(gomock.Any(), user.Username).Return(user, nil)
		ctx, rec := avatarUploadCtx(t, decodeBombPNG(9_000, 9_000)) // 81MP > 24MP cap
		setAuth(ctx, user.Username)

		server.uploadAvatar(ctx)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, rec.Body.String(), "dimensions too large")
	})

	t.Run("valid image is cropped, re-encoded and stored", func(t *testing.T) {
		server, store := newTestGameServer(t)
		user := testUser()
		store.EXPECT().GetUserByUsername(gomock.Any(), user.Username).Return(user, nil)
		store.EXPECT().UpsertUserAvatar(gomock.Any(), gomock.Any()).Return(nil)
		ctx, rec := avatarUploadCtx(t, smallPNG(t))
		setAuth(ctx, user.Username)

		server.uploadAvatar(ctx)
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("garbage bytes rejected as unsupported", func(t *testing.T) {
		server, store := newTestGameServer(t)
		user := testUser()
		store.EXPECT().GetUserByUsername(gomock.Any(), user.Username).Return(user, nil)
		ctx, rec := avatarUploadCtx(t, []byte("this is not an image"))
		setAuth(ctx, user.Username)

		server.uploadAvatar(ctx)
		require.Equal(t, http.StatusBadRequest, rec.Code)
		require.Contains(t, rec.Body.String(), "unsupported image")
	})
}
