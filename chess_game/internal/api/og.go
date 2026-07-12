package api

import (
	"bytes"
	"embed"
	"fmt"
	"image/color"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Glenn444/golang-chess/internal/board"
	db "github.com/Glenn444/golang-chess/internal/db"
	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/fogleman/gg"
	"github.com/gin-gonic/gin"
	"github.com/golang/freetype/truetype"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/image/font"
)

// Dynamic og:image cards — 1200×630 PNGs rendered per game on demand.
// A card is fetched roughly once per share per platform, so: render on
// request, keep a small in-memory cache, set Cache-Control, done.

//go:embed fonts/DejaVuSans.ttf fonts/DejaVuSans-Bold.ttf
var ogFontFS embed.FS

const (
	ogW = 1200
	ogH = 630

	ogTTLLive   = time.Minute      // live cards carry the move count
	ogTTLStatic = 10 * time.Minute // finished/invite cards barely change
	ogCacheCap  = 512              // guards against id-scanning memory abuse
)

var (
	ogFontRegular *truetype.Font
	ogFontBold    *truetype.Font
	ogFontOnce    sync.Once
	ogFontErr     error

	ogCacheMu sync.Mutex
	ogCache   = map[string]ogCacheEntry{}
)

type ogCacheEntry struct {
	png []byte
	exp time.Time
}

func ogFonts() error {
	ogFontOnce.Do(func() {
		reg, err := ogFontFS.ReadFile("fonts/DejaVuSans.ttf")
		if err != nil {
			ogFontErr = err
			return
		}
		bold, err := ogFontFS.ReadFile("fonts/DejaVuSans-Bold.ttf")
		if err != nil {
			ogFontErr = err
			return
		}
		if ogFontRegular, ogFontErr = truetype.Parse(reg); ogFontErr != nil {
			return
		}
		ogFontBold, ogFontErr = truetype.Parse(bold)
	})
	return ogFontErr
}

func ogFace(f *truetype.Font, size float64) font.Face {
	return truetype.NewFace(f, &truetype.Options{Size: size})
}

// ── Palette (mirrors the app's CSS variables) ────────────────────────────────

var (
	ogBgTop      = color.RGBA{0x17, 0x1A, 0x22, 0xFF}
	ogBgBottom   = color.RGBA{0x0D, 0x0E, 0x12, 0xFF}
	ogText       = color.RGBA{0xE8, 0xE6, 0xE1, 0xFF}
	ogMuted      = color.RGBA{0xA5, 0xA2, 0x9B, 0xFF}
	ogAmber      = color.RGBA{0xE5, 0xA9, 0x3B, 0xFF}
	ogRed        = color.RGBA{0xD2, 0x6A, 0x6A, 0xFF}
	ogGrey       = color.RGBA{0x3A, 0x3E, 0x4A, 0xFF}
	ogBoardLight = color.RGBA{0xD9, 0xC9, 0xA8, 0xFF}
	ogBoardDark  = color.RGBA{0x2A, 0x2D, 0x36, 0xFF}
	ogPieceLight = color.RGBA{0xF7, 0xF0, 0xE1, 0xFF}
	ogPieceDark  = color.RGBA{0x15, 0x16, 0x1A, 0xFF}
)

type ogCard struct {
	pillLabel string
	pillColor color.RGBA
	pillDark  bool // dark text on the pill (for amber)
	title     string
	context   string
	board     [][]pieces.Square // nil = no board
}

// ── HTTP handlers ────────────────────────────────────────────────────────────

func (server *Server) ogGameCard(ctx *gin.Context) {
	server.serveOGCard(ctx, false)
}

func (server *Server) ogInviteCard(ctx *gin.Context) {
	server.serveOGCard(ctx, true)
}

func (server *Server) serveOGCard(ctx *gin.Context, invite bool) {
	if err := ogFonts(); err != nil {
		slog.Error("og: font load failed", "err", err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	id := strings.TrimSuffix(ctx.Param("id"), ".png")
	card, key, ttl := server.ogCardFor(ctx, id, invite)

	ogCacheMu.Lock()
	if e, ok := ogCache[key]; ok && time.Now().Before(e.exp) {
		ogCacheMu.Unlock()
		writeOGPNG(ctx, e.png, ttl)
		return
	}
	ogCacheMu.Unlock()

	png, err := renderOGCard(card)
	if err != nil {
		slog.Error("og: render failed", "key", key, "err", err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ogCacheMu.Lock()
	if len(ogCache) >= ogCacheCap {
		ogCache = map[string]ogCacheEntry{} // cards are cheap — just reset
	}
	ogCache[key] = ogCacheEntry{png: png, exp: time.Now().Add(ttl)}
	ogCacheMu.Unlock()

	writeOGPNG(ctx, png, ttl)
}

// ogCardFor decides what the card shows, applying the same privacy rules as
// the landing pages: the scraper fetches with no auth, and the same image is
// shown to everyone the link reaches.
func (server *Server) ogCardFor(ctx *gin.Context, id string, invite bool) (ogCard, string, time.Duration) {
	generic := ogCard{
		pillLabel: "PLAY CHESS",
		pillColor: ogAmber, pillDark: true,
		title:   "Chesske",
		context: "Play chess online · voice & chat · chesske.com",
		board:   board.Initialise_board(board.Create_board()),
	}

	parsed, err := uuid.Parse(id)
	if err != nil {
		return generic, "generic", ogTTLStatic
	}
	game, err := server.store.GetGameByID(ctx, pgtype.UUID{Bytes: parsed, Valid: true})
	if err != nil {
		return generic, "generic", ogTTLStatic
	}

	if invite {
		if game.Opponent != "person" || game.State != db.GameStateWaiting {
			return generic, "generic", ogTTLStatic
		}
		creatorID := game.WhitePlayerID
		openColor := "black"
		if !creatorID.Valid {
			creatorID = game.BlackPlayerID
			openColor = "white"
		}
		creator := orName(server.lookupUsername(ctx, creatorID), "A Chesske player")
		snap := board.DeserializeGameState(game.BoardState)
		card := ogCard{
			pillLabel: "GAME INVITE",
			pillColor: ogAmber, pillDark: true,
			title:   creator + " challenges you",
			context: strings.ToLower(shareTimeLabel(game.WhiteTimeRemainingMs, game.BlackTimeRemainingMs)) + " · you play " + openColor + " · chesske.com",
			board:   snap.Board,
		}
		return card, "inv|" + id + "|" + creator, ogTTLStatic
	}

	if game.Visibility != "public" {
		return ogCard{
			pillLabel: "PRIVATE",
			pillColor: ogGrey,
			title:     "A private game",
			context:   "Only its players can watch · play your own on chesske.com",
		}, "locked", ogTTLStatic
	}

	white := orName(server.lookupUsername(ctx, game.WhitePlayerID), "White")
	black := orName(server.lookupUsername(ctx, game.BlackPlayerID), "Black")
	if game.Opponent == "stockfish" {
		sf := fmt.Sprintf("Stockfish lv %d", game.StockfishLevel)
		if !game.WhitePlayerID.Valid {
			white = sf
		}
		if !game.BlackPlayerID.Valid {
			black = sf
		}
	}
	snap := board.DeserializeGameState(game.BoardState)

	card := ogCard{
		title: white + " vs " + black,
		board: snap.Board,
	}
	ttl := ogTTLStatic
	switch game.State {
	case db.GameStateActive:
		card.pillLabel = "LIVE"
		card.pillColor = ogRed
		card.context = fmt.Sprintf("Move %d · %s · watch live on chesske.com", game.MoveCount, strings.ToLower(shareTimeLabel(game.WhiteTimeRemainingMs, game.BlackTimeRemainingMs)))
		ttl = ogTTLLive
	case db.GameStateWaiting:
		card.pillLabel = "STARTING SOON"
		card.pillColor = ogAmber
		card.pillDark = true
		card.context = "Waiting for an opponent · chesske.com"
	default:
		card.pillLabel = "FINISHED"
		card.pillColor = ogGrey
		card.context = gameResultText(game, white, black) + " · " + strconv.Itoa(int(game.MoveCount)) + " moves · chesske.com"
	}
	key := fmt.Sprintf("game|%s|%s|%d|%s|%s", id, game.State, game.MoveCount, white, black)
	return card, key, ttl
}

func writeOGPNG(ctx *gin.Context, png []byte, ttl time.Duration) {
	ctx.Header("Content-Type", "image/png")
	ctx.Header("Content-Length", strconv.Itoa(len(png)))
	ctx.Header("Cache-Control", "public, max-age="+strconv.Itoa(int(ttl.Seconds())))
	ctx.Status(http.StatusOK)
	ctx.Writer.Write(png) //nolint:errcheck
}

// ── Drawing ──────────────────────────────────────────────────────────────────

func renderOGCard(card ogCard) ([]byte, error) {
	dc := gg.NewContext(ogW, ogH)

	// Background: vertical brand gradient + low-alpha decorative blobs.
	grad := gg.NewLinearGradient(0, 0, 0, ogH)
	grad.AddColorStop(0, ogBgTop)
	grad.AddColorStop(1, ogBgBottom)
	dc.SetFillStyle(grad)
	dc.DrawRectangle(0, 0, ogW, ogH)
	dc.Fill()

	dc.SetRGBA(0.90, 0.66, 0.23, 0.05)
	dc.DrawCircle(1080, 60, 300)
	dc.Fill()
	dc.DrawCircle(140, 640, 240)
	dc.Fill()

	// Brand row.
	dc.SetFontFace(ogFace(ogFontRegular, 46))
	dc.SetColor(ogAmber)
	dc.DrawString("♞", 60, 100)
	dc.SetFontFace(ogFace(ogFontBold, 40))
	dc.SetColor(ogText)
	dc.DrawString("Chesske", 118, 98)

	// Content column width depends on whether a board is drawn.
	contentW := float64(ogW - 120)
	if card.board != nil {
		contentW = 620
	}

	// State pill.
	dc.SetFontFace(ogFace(ogFontBold, 24))
	pw, _ := dc.MeasureString(card.pillLabel)
	dc.SetColor(card.pillColor)
	dc.DrawRoundedRectangle(60, 168, pw+48, 52, 26)
	dc.Fill()
	if card.pillDark {
		dc.SetRGB(0.09, 0.07, 0.04)
	} else {
		dc.SetColor(ogText)
	}
	dc.DrawString(card.pillLabel, 84, 168+36)

	// Title: bold, wrapped to at most 2 lines, "…"-truncated.
	dc.SetFontFace(ogFace(ogFontBold, 62))
	dc.SetColor(ogText)
	lines := wrapOGText(dc, card.title, contentW, 2)
	y := 330.0
	for _, line := range lines {
		dc.DrawString(line, 60, y)
		y += 76
	}

	// Context line.
	dc.SetFontFace(ogFace(ogFontRegular, 29))
	dc.SetColor(ogMuted)
	ctxLines := wrapOGText(dc, card.context, contentW, 2)
	for _, line := range ctxLines {
		dc.DrawString(line, 60, y)
		y += 42
	}

	// Footer.
	dc.SetFontFace(ogFace(ogFontRegular, 26))
	dc.SetColor(ogMuted)
	dc.DrawString("chesske.com", 60, ogH-46)

	if card.board != nil {
		drawOGBoard(dc, card.board, 768, 131, 46)
	}

	var buf bytes.Buffer
	if err := dc.EncodePNG(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// drawOGBoard paints the position from White's perspective (engine board row
// 0 is rank 1). Both colors use the filled glyph set, colored like the app.
func drawOGBoard(dc *gg.Context, b [][]pieces.Square, x, y, sq float64) {
	size := sq * 8

	// Soft outer glow / frame.
	dc.SetRGBA(0.90, 0.66, 0.23, 0.30)
	dc.DrawRoundedRectangle(x-6, y-6, size+12, size+12, 14)
	dc.Fill()

	glyphFace := ogFace(ogFontBold, sq*0.78)
	for r := 7; r >= 0; r-- {
		for f := 0; f < 8; f++ {
			cx := x + float64(f)*sq
			cy := y + float64(7-r)*sq
			if (r+f)%2 == 1 {
				dc.SetColor(ogBoardLight)
			} else {
				dc.SetColor(ogBoardDark)
			}
			dc.DrawRectangle(cx, cy, sq, sq)
			dc.Fill()

			if r >= len(b) || f >= len(b[r]) || !b[r][f].Occupied || b[r][f].Piece == nil {
				continue
			}
			glyph := pieceGlyphs[strings.ToUpper(b[r][f].Piece.GetPieceType())]
			if glyph == "" {
				continue
			}
			dc.SetFontFace(glyphFace)
			gx := cx + sq/2
			gy := cy + sq/2
			// Shadow first so light pieces stay readable on light squares.
			dc.SetRGBA(0, 0, 0, 0.35)
			dc.DrawStringAnchored(glyph, gx, gy+2, 0.5, 0.5)
			if b[r][f].Piece.GetColor() == "w" {
				dc.SetColor(ogPieceLight)
			} else {
				dc.SetColor(ogPieceDark)
			}
			dc.DrawStringAnchored(glyph, gx, gy, 0.5, 0.5)
		}
	}
}

// wrapOGText word-wraps s to maxWidth, keeping at most maxLines lines and
// ending the last line with "…" when text is cut. Text must never overflow
// the canvas — measure, wrap, truncate.
func wrapOGText(dc *gg.Context, s string, maxWidth float64, maxLines int) []string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	cur := words[0]
	for _, w := range words[1:] {
		if tw, _ := dc.MeasureString(cur + " " + w); tw <= maxWidth {
			cur += " " + w
			continue
		}
		lines = append(lines, cur)
		cur = w
		if len(lines) == maxLines {
			break
		}
	}
	if len(lines) < maxLines {
		lines = append(lines, cur)
	} else {
		last := lines[maxLines-1]
		for {
			if tw, _ := dc.MeasureString(last + "…"); tw <= maxWidth || !strings.Contains(last, " ") {
				break
			}
			last = last[:strings.LastIndex(last, " ")]
		}
		lines[maxLines-1] = last + "…"
	}
	// Hard-truncate any single overlong word.
	for i, line := range lines {
		for {
			if tw, _ := dc.MeasureString(line); tw <= maxWidth || len(line) < 4 {
				break
			}
			line = line[:len(line)-4] + "…"
		}
		lines[i] = line
	}
	return lines
}
