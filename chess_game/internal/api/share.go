package api

import (
	"embed"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strings"

	"github.com/Glenn444/golang-chess/internal/board"
	db "github.com/Glenn444/golang-chess/internal/db"
	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Server-rendered share pages. Link scrapers (WhatsApp/X/Telegram) never run
// JavaScript, so /invite/:id and /game/:id must carry their og/twitter meta
// tags in the markup. Both are also real landing pages for humans: the invite
// page deep-links into the app's join flow, and the game page is a live
// spectator view (read-only WebSocket).

//go:embed templates/*.html
var shareTemplatesFS embed.FS

var shareTemplates = template.Must(template.ParseFS(shareTemplatesFS, "templates/*.html"))

type sharePageMeta struct {
	Title       string
	Description string
	PageURL     string
	ImageURL    string
	NoIndex     bool
}

type boardCellView struct {
	Glyph string
	Light bool
	White bool
}

type inviteView struct {
	Meta        sharePageMeta
	State       string // waiting | active | finished
	CreatorName string
	TimeLabel   string
	ColorLabel  string
	CTAURL      string
	WatchURL    string
	HomeURL     string
}

type spectateView struct {
	Meta       sharePageMeta
	GameID     string
	Locked     bool
	Live       bool
	Waiting    bool
	Finished   bool
	StatusPill string
	ResultText string
	WhiteName  string
	BlackName  string
	MoveCount  int32
	TimeLabel  string
	WhiteClock string
	BlackClock string
	Board      [][]boardCellView
	HomeURL    string
}

// publicBaseURL is the absolute origin share pages and og:image URLs are
// built from. Never derived from the request Host header — proxies lie.
func (server *Server) publicBaseURL() string {
	if server.config.Environment == "production" {
		host := strings.TrimPrefix(server.config.PUBLIC_HOST, ".")
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		if host != "" {
			return "https://" + host
		}
		return "https://chesske.com"
	}
	return "http://localhost:8080"
}

func renderShareHTML(ctx *gin.Context, status int, name string, data any) {
	ctx.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	ctx.Status(status)
	if err := shareTemplates.ExecuteTemplate(ctx.Writer, name, data); err != nil {
		// Headers are already out — nothing to do but log.
		ctx.Error(err) //nolint:errcheck
	}
}

func (server *Server) renderShare404(ctx *gin.Context) {
	renderShareHTML(ctx, http.StatusNotFound, "share404.html", gin.H{"HomeURL": server.publicBaseURL()})
}

// shareGame loads the game for a share page, rendering the branded 404 on
// any failure. Unknown and malformed IDs look identical from outside.
func (server *Server) shareGame(ctx *gin.Context) (db.Game, bool) {
	parsed, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		server.renderShare404(ctx)
		return db.Game{}, false
	}
	game, err := server.store.GetGameByID(ctx, pgtype.UUID{Bytes: parsed, Valid: true})
	if err != nil {
		server.renderShare404(ctx)
		return db.Game{}, false
	}
	return game, true
}

// invitePage renders the "you're challenged" landing page. The game UUID is
// the bearer token: whoever holds the link may accept, exactly like joinGame.
func (server *Server) invitePage(ctx *gin.Context) {
	game, ok := server.shareGame(ctx)
	if !ok {
		return
	}
	// Engine games have no open seat — an invite link to one is meaningless.
	if game.Opponent != "person" {
		server.renderShare404(ctx)
		return
	}

	base := server.publicBaseURL()
	id := uidStr(game.ID)

	creatorID := game.WhitePlayerID
	openColor := "black"
	if !creatorID.Valid {
		creatorID = game.BlackPlayerID
		openColor = "white"
	}
	creator := server.lookupUsername(ctx, creatorID)
	if creator == "" {
		creator = "A Chesske player"
	}

	state := "finished"
	switch game.State {
	case db.GameStateWaiting:
		state = "waiting"
	case db.GameStateActive:
		state = "active"
	}

	timeLabel := shareTimeLabel(game.WhiteTimeRemainingMs, game.BlackTimeRemainingMs)
	v := inviteView{
		Meta: sharePageMeta{
			Title:       fmt.Sprintf("You're invited: play %s on Chesske", creator),
			Description: fmt.Sprintf("%s challenges you to a chess game — %s, you play %s. Tap to accept.", creator, strings.ToLower(timeLabel), openColor),
			PageURL:     base + "/invite/" + id,
			ImageURL:    base + "/og/invite/" + id + ".png",
			NoIndex:     true, // bearer link — never index
		},
		State:       state,
		CreatorName: creator,
		TimeLabel:   timeLabel,
		ColorLabel:  openColor,
		CTAURL:      base + "/play/" + id + "?join=true",
		HomeURL:     base,
	}
	if state == "active" && game.Visibility == "public" {
		v.WatchURL = base + "/game/" + id
	}
	renderShareHTML(ctx, http.StatusOK, "invite.html", v)
}

// spectatePage renders the live spectator page for public games, a final
// board for finished ones, and a locked page (no details) for private games.
func (server *Server) spectatePage(ctx *gin.Context) {
	game, ok := server.shareGame(ctx)
	if !ok {
		return
	}

	base := server.publicBaseURL()
	id := uidStr(game.ID)

	if game.Visibility != "public" {
		renderShareHTML(ctx, http.StatusOK, "spectate.html", spectateView{
			Meta: sharePageMeta{
				Title:       "A private game on Chesske",
				Description: "This game is private — only its players can watch. Play your own game on Chesske.",
				PageURL:     base + "/game/" + id,
				ImageURL:    base + "/og/game/" + id + ".png",
				NoIndex:     true,
			},
			Locked:     true,
			StatusPill: "PRIVATE",
			HomeURL:    base,
		})
		return
	}

	whiteName := server.lookupUsername(ctx, game.WhitePlayerID)
	blackName := server.lookupUsername(ctx, game.BlackPlayerID)
	if game.Opponent == "stockfish" {
		if whiteName == "" {
			whiteName = fmt.Sprintf("Stockfish (level %d)", game.StockfishLevel)
		}
		if blackName == "" {
			blackName = fmt.Sprintf("Stockfish (level %d)", game.StockfishLevel)
		}
	}

	snap := board.DeserializeGameState(game.BoardState)
	live := game.State == db.GameStateActive
	waiting := game.State == db.GameStateWaiting
	finished := !live && !waiting

	pill := "FINISHED"
	if live {
		pill = "LIVE"
	} else if waiting {
		pill = "WAITING"
	}

	timeLabel := shareTimeLabel(game.WhiteTimeRemainingMs, game.BlackTimeRemainingMs)
	title, desc := spectateMeta(game, whiteName, blackName, timeLabel)

	v := spectateView{
		Meta: sharePageMeta{
			Title:       title,
			Description: desc,
			PageURL:     base + "/game/" + id,
			ImageURL:    base + "/og/game/" + id + ".png",
		},
		GameID:     id,
		Live:       live,
		Waiting:    waiting,
		Finished:   finished,
		StatusPill: pill,
		WhiteName:  whiteName,
		BlackName:  blackName,
		MoveCount:  game.MoveCount,
		TimeLabel:  timeLabel,
		WhiteClock: shareClock(game.WhiteTimeRemainingMs, game.BlackTimeRemainingMs, game.WhiteTimeRemainingMs),
		BlackClock: shareClock(game.WhiteTimeRemainingMs, game.BlackTimeRemainingMs, game.BlackTimeRemainingMs),
		Board:      boardCells(snap.Board),
		HomeURL:    base,
	}
	if finished {
		v.ResultText = gameResultText(game, whiteName, blackName)
	}
	renderShareHTML(ctx, http.StatusOK, "spectate.html", v)
}

func spectateMeta(game db.Game, whiteName, blackName, timeLabel string) (title, desc string) {
	w, b := orName(whiteName, "White"), orName(blackName, "Black")
	switch game.State {
	case db.GameStateActive:
		title = fmt.Sprintf("%s vs %s — LIVE on Chesske", w, b)
		desc = fmt.Sprintf("Move %d · %s — watch %s take on %s live, move by move.", game.MoveCount, timeLabel, w, b)
	case db.GameStateWaiting:
		title = "A game is about to begin on Chesske"
		desc = fmt.Sprintf("%s is waiting for an opponent — %s. Watch it live once it starts.", orName(whiteName+blackName, "A player"), strings.ToLower(timeLabel))
	default:
		title = fmt.Sprintf("%s vs %s on Chesske", w, b)
		desc = fmt.Sprintf("%s after %d moves · %s. See the final position on Chesske.", gameResultText(game, w, b), game.MoveCount, timeLabel)
	}
	return title, desc
}

// gameResultText names the winner for a finished game. Winner conventions
// mirror finishGame: checkmate → the mated side is CurrentPlayer; timeout →
// the flagged side's clock is 0; resign → EndedByPlayerID resigned.
func gameResultText(game db.Game, whiteName, blackName string) string {
	w, b := orName(whiteName, "White"), orName(blackName, "Black")
	winnerName := func(c string) string {
		if c == "w" {
			return w
		}
		return b
	}
	switch game.State {
	case db.GameStateCheckmate:
		winner := "w"
		if game.CurrentPlayer == "w" {
			winner = "b"
		}
		return fmt.Sprintf("Checkmate — %s won", winnerName(winner))
	case db.GameStateTimeout:
		winner := "w"
		if game.WhiteTimeRemainingMs <= 0 {
			winner = "b"
		}
		return fmt.Sprintf("%s won on time", winnerName(winner))
	case db.GameStateResign:
		winner := "w"
		if uuidEq(game.EndedByPlayerID, game.WhitePlayerID) {
			winner = "b"
		}
		return fmt.Sprintf("%s won by resignation", winnerName(winner))
	case db.GameStateStalemate:
		return "Draw by stalemate"
	case db.GameStateDraw:
		return "Drawn game"
	case db.GameStateAbandoned:
		return "Game abandoned"
	}
	return "Game over"
}

func orName(name, fallback string) string {
	if strings.TrimSpace(name) == "" {
		return fallback
	}
	return name
}

// shareTimeLabel describes the time control. Initial time control isn't
// stored, so past the first move we can only say timed vs unlimited — except
// for waiting games, where remaining == initial.
func shareTimeLabel(whiteMs, blackMs int64) string {
	if whiteMs == 0 && blackMs == 0 {
		return "Unlimited time"
	}
	if whiteMs == blackMs && whiteMs%60000 == 0 {
		return fmt.Sprintf("%d min game", whiteMs/60000)
	}
	return "Timed game"
}

func shareClock(whiteMs, blackMs, ms int64) string {
	if whiteMs == 0 && blackMs == 0 {
		return "∞"
	}
	if ms < 0 {
		ms = 0
	}
	s := ms / 1000
	return fmt.Sprintf("%d:%02d", s/60, s%60)
}

var pieceGlyphs = map[string]string{
	"K": "♚", "Q": "♛", "R": "♜", "B": "♝", "N": "♞", "P": "♟",
}

// boardCells converts the engine board (row 0 = rank 1) into display rows
// from White's perspective (rank 8 first). Both colors use the filled glyph
// set; CSS colors them.
func boardCells(b [][]pieces.Square) [][]boardCellView {
	rows := make([][]boardCellView, 0, 8)
	for r := 7; r >= 0; r-- {
		row := make([]boardCellView, 0, 8)
		for f := 0; f < 8; f++ {
			cell := boardCellView{Light: (r+f)%2 == 1}
			if r < len(b) && f < len(b[r]) && b[r][f].Occupied && b[r][f].Piece != nil {
				cell.Glyph = pieceGlyphs[strings.ToUpper(b[r][f].Piece.GetPieceType())]
				cell.White = b[r][f].Piece.GetColor() == "w"
			}
			row = append(row, cell)
		}
		rows = append(rows, row)
	}
	return rows
}
