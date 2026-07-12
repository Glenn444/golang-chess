package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Glenn444/golang-chess/config"
	"github.com/Glenn444/golang-chess/internal/db"
	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/internal/stockfish"
	"github.com/Glenn444/golang-chess/internal/token"
	"github.com/Glenn444/golang-chess/internal/utils/emails"
	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/olahol/melody"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
			re := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
			return re.MatchString(fl.Field().String())
		})
	}
}

type Server struct {
	config      config.Config
	emailClient emails.EmailSender
	tokenMaker  token.Maker
	store       db.Store
	router      *gin.Engine
	melody      *melody.Melody
	httpServer  *http.Server
	ready       atomic.Bool

	//chess game
	activeGames    map[pgtype.UUID]*pieces.GameState
	activeGamesMu  sync.RWMutex
	createGameMUs  map[pgtype.UUID]*sync.Mutex // per-user mutex to prevent duplicate games
	createGameMUsMu sync.Mutex                  // protects the createGameMUs map

	// Shared Stockfish engine for online engine games. Lazily started; the
	// mutex serializes UCI dialogue (the engine is stateless per query since
	// the full move history is sent each time).
	engine     *stockfish.Stockfish
	engineInit bool
	engineMu   sync.Mutex
}

// engineAvailable reports whether server-side Stockfish games can be played.
func (server *Server) engineAvailable() bool {
	return os.Getenv("STOCKFISH_ENGINE_PATH") != ""
}

// engineBestMove computes the engine reply for the given UCI history at the
// given skill level. A wedged/dead engine is discarded so the next call
// respawns a fresh one.
func (server *Server) engineBestMove(moves []string, level int32) (string, error) {
	server.engineMu.Lock()
	defer server.engineMu.Unlock()

	if !server.engineInit {
		sf, err := stockfish.NewStockfish()
		if err != nil {
			return "", err
		}
		server.engine = sf
		server.engineInit = true
	}
	if server.engine == nil {
		return "", fmt.Errorf("stockfish engine not configured")
	}

	move, err := func() (string, error) {
		if err := server.engine.SetSkillLevel(int(level)); err != nil {
			return "", err
		}
		return server.engine.GetBestMove(moves)
	}()
	if err != nil {
		server.engine.Close()
		server.engine = nil
		server.engineInit = false
		return "", err
	}
	return move, nil
}

func NewServer(cfg config.Config, store db.Store) (*Server, error) {
	jwtTokenMaker, err := token.NewJWTMaker(cfg.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker %w", err)
	}

	server := &Server{
		tokenMaker:  jwtTokenMaker,
		store:       store,
		config:      cfg,
		emailClient: emails.NewEmailClient(cfg.RESEND_API_KEY),
		melody:      melody.New(),
		activeGames:   make(map[pgtype.UUID]*pieces.GameState),
		createGameMUs: make(map[pgtype.UUID]*sync.Mutex),
	}
	server.setupMelody()

	gin.ForceConsoleColor()
	router := gin.Default()
	router.RedirectTrailingSlash = false

	// ── Global middleware ─────────────────────────────────────────────────────
	router.Use(cors.New(cfg.CORSConfig()))

	// Rate limiting: 100 req/min for public routes, 200/min for auth routes.
	publicLimiter := ratelimit.RateLimiter(
		ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
			Rate:  time.Minute,
			Limit: 100,
		}),
		&ratelimit.Options{
			ErrorHandler: func(c *gin.Context, info ratelimit.Info) {
				c.JSON(http.StatusTooManyRequests, errorMessage("rate limit exceeded"))
				c.Abort()
			},
		},
	)
	authLimiter := ratelimit.RateLimiter(
		ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
			Rate:  time.Minute,
			Limit: 200,
		}),
		&ratelimit.Options{
			ErrorHandler: func(c *gin.Context, info ratelimit.Info) {
				c.JSON(http.StatusTooManyRequests, errorMessage("rate limit exceeded"))
				c.Abort()
			},
		},
	)

	// ── Health / Readiness ────────────────────────────────────────────────────
	router.GET("/healthz", server.healthz)
	router.GET("/readyz", server.readyz)

	// ── Swagger UI ────────────────────────────────────────────────────────────
	router.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/docs/index.html")
	})

	swaggerHandler := ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/docs/doc.json"),
	)

	router.GET("/docs/*any", func(c *gin.Context) {
		if c.Param("any") == "/" {
			c.Redirect(http.StatusFound, "/docs/index.html")
			return
		}

		swaggerHandler(c)
	})

	//router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker), authLimiter)
	// ── Welcome ───────────────────────────────────────────────────────────────
	router.GET("/", server.welcome)

	// ── Public (rate-limited) ─────────────────────────────────────────────────
	users := router.Group("/users")
	users.Use(publicLimiter)

	users.GET("/check-username", server.checkUsernameExists)
	users.POST("/signup", server.createUser)
	users.POST("/confirm-email", server.confirmEmail)
	users.POST("/send-emailotp", server.sendEmailOTP)
	users.POST("/forgot-password", server.forgotPassword)
	users.POST("/reset-password", server.resetPassword)
	users.POST("/signin", server.loginUser)
	users.POST("/refresh-token", server.refreshToken)

	// Logout must work even with an expired access token — it only clears
	// cookies and revokes the refresh-token session it can read.
	users.POST("/logout", server.logoutUser)

	// ── Protected (Bearer JWT + rate-limited) ─────────────────────────────────
	authUsers := router.Group("/users").Use(authMiddleware(server.tokenMaker), authLimiter)
	authUsers.GET("/me", server.getMe)
	authUsers.POST("/me/avatar", server.uploadAvatar)
	authUsers.DELETE("/me/avatar", server.deleteAvatar)

	// Public — avatars show on player cards and spectator pages.
	router.GET("/users/:id/avatar", server.getAvatar)

	// Public — no auth required, just returns visible games.
	router.GET("/games/public", server.listPublicGames)
	router.GET("/games/live", server.listLiveGames)

	authGames := router.Group("/games").Use(authMiddleware(server.tokenMaker), authLimiter)
	authGames.POST("", server.createGame)
	authGames.GET("", server.listWaitingGames)
	authGames.GET("/mine", server.listMyGames)
	authGames.GET("/:id", server.getGame)
	authGames.POST("/:id/join", server.joinGame)
	authGames.POST("/:id/resign", server.resignGame)
	authGames.DELETE("/:id", server.deleteGame)
	authGames.GET("/:id/moves", server.getGameMoves)
	authGames.GET("/:id/replay", server.getGameReplay)

	// chat + voice (scoped under games)
	authGames.POST("/:id/chat", server.sendChatMessage)
	authGames.GET("/:id/chat", server.getChatMessages)
	authGames.POST("/:id/voice", server.startVoiceSession)
	authGames.GET("/:id/voice", server.getActiveVoiceSession)
	authGames.PATCH("/:id/voice/:vid/activate", server.activateVoiceSession)
	authGames.DELETE("/:id/voice/:vid", server.endVoiceSession)

	// --- Push notifications -----------------------------------
	pushRoutes := router.Group("/api/push").Use(authMiddleware(server.tokenMaker), authLimiter)
	pushRoutes.POST("/subscribe", server.subscribePush)
	pushRoutes.GET("/subscription", server.getPushSubscription)

	// --- Turn server ------------------------------------------
	authRoutes.GET("/turn-credentials", server.getTURNCredentials)

	// ── WebSocket ─────────────────────────────────────────────────────────────
	// No token in query string — client sends an "auth" message as the first
	// frame after upgrading. Game room is selected via ?game_id=<uuid>.
	router.GET("/ws", server.handleWebSocket)

	// ── Share pages (server-rendered — link scrapers don't run JS) ───────────
	// Served on the apex domain via nginx path routing: chesske.com/invite/…,
	// chesske.com/game/… (live spectator page) and the og:image cards.
	router.GET("/invite/:id", server.invitePage)
	router.GET("/game/:id", server.spectatePage)
	router.GET("/og/invite/:id", server.ogInviteCard)
	router.GET("/og/game/:id", server.ogGameCard)

	server.router = router
	server.ready.Store(true)
	return server, nil
}

func (server *Server) Start(address string) error {
	server.httpServer = &http.Server{
		Addr:    address,
		Handler: server.router,
	}
	return server.httpServer.ListenAndServe()
}

func (server *Server) Shutdown(ctx context.Context) error {
	server.ready.Store(false)
	return server.httpServer.Shutdown(ctx)
}

// @Summary      Health check
// @Description  Returns server liveness status.
// @Tags         Health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /healthz [get]
func (server *Server) healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "alive"})
}

// @Summary      Readiness check
// @Description  Checks database connectivity and returns readiness status.
// @Tags         Health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      503  {object}  map[string]string
// @Router       /readyz [get]
func (server *Server) readyz(c *gin.Context) {
	if !server.ready.Load() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
		return
	}
	if err := server.store.Ping(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"reason": "database unreachable",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

// @Summary      Welcome
// @Description  Returns a welcome message.
// @Tags         General
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       / [get]
func (server *Server) welcome(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "Welcome to the Chess Game Server"})
}

// cookieConfig returns the domain and secure flag for SetCookie calls.
// In development (localhost) we use empty domain + Secure=false so cookies
// work over plain HTTP. In production we use the real domain + Secure=true.
// A port in PUBLIC_HOST would make the Domain attribute invalid (browsers
// reject the cookie), so it is stripped.
func (server *Server) cookieConfig() (domain string, secure bool) {
	if server.config.Environment == "production" {
		host := server.config.PUBLIC_HOST
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		return host, true
	}
	return "", false
}
