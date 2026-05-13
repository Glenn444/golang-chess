package api

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Glenn444/golang-chess/config"
	"github.com/Glenn444/golang-chess/internal/db"
	"github.com/Glenn444/golang-chess/internal/pieces"
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
	activeGames   map[pgtype.UUID]*pieces.GameState
	activeGamesMu sync.RWMutex
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
		activeGames: make(map[pgtype.UUID]*pieces.GameState),
	}
	server.setupMelody()

	gin.ForceConsoleColor()
	router := gin.Default()
	router.RedirectTrailingSlash = false  // add this
	// Register custom validators
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
			re := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
			return re.MatchString(fl.Field().String())
		})
	}
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
	users.POST("/signin", server.loginUser)
	users.POST("/refresh-token", server.refreshToken)

	// ── Protected (Bearer JWT + rate-limited) ─────────────────────────────────
	authUsers := router.Group("/users").Use(authMiddleware(server.tokenMaker), authLimiter)
	authUsers.GET("/me", server.getMe)

	authGames := router.Group("/games").Use(authMiddleware(server.tokenMaker), authLimiter)
	authGames.POST("", server.createGame)
	authGames.GET("", server.listWaitingGames)
	authGames.GET("/mine", server.listMyGames)
	authGames.GET("/:id", server.getGame)
	authGames.POST("/:id/join", server.joinGame)
	authGames.POST("/:id/resign", server.resignGame)
	authGames.GET("/:id/moves", server.getGameMoves)

	// chat + voice (scoped under games)
	authGames.POST("/:id/chat", server.sendChatMessage)
	authGames.GET("/:id/chat", server.getChatMessages)
	authGames.POST("/:id/voice", server.startVoiceSession)
	authGames.GET("/:id/voice", server.getActiveVoiceSession)
	authGames.PATCH("/:id/voice/:vid/activate", server.activateVoiceSession)
	authGames.DELETE("/:id/voice/:vid", server.endVoiceSession)

	// --- Turn server ------------------------------------------
	authRoutes.GET("/turn-credentials", server.getTURNCredentials)

	// ── WebSocket ─────────────────────────────────────────────────────────────
	// No token in query string — client sends an "auth" message as the first
	// frame after upgrading. Game room is selected via ?game_id=<uuid>.
	router.GET("/ws", server.handleWebSocket)

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
