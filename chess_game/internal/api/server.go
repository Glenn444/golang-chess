package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/Glenn444/golang-chess/config"
	"github.com/Glenn444/golang-chess/internal/db"
	"github.com/Glenn444/golang-chess/internal/pieces"
	"github.com/Glenn444/golang-chess/internal/token"
	"github.com/Glenn444/golang-chess/internal/utils/emails"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/olahol/melody"
)

type Server struct {
	config      config.Config
	emailClient emails.EmailSender
	tokenMaker  token.Maker
	store       db.Store
	router      *gin.Engine
	melody      *melody.Melody
	httpServer  *http.Server

	//chess game
	activeGames   map[pgtype.UUID]*pieces.GameState
	activeGamesMu sync.RWMutex //protect concurremt ws access
}

func NewServer(config config.Config, store db.Store) (*Server, error) {
	jwtTokenMaker, err := token.NewJWTMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker %w\n", err)
	}

	server := &Server{
		tokenMaker:  jwtTokenMaker,
		store:       store,
		config:      config,
		emailClient: emails.NewEmailClient(config.RESEND_API_KEY),
		melody:      melody.New(),
		activeGames: make(map[pgtype.UUID]*pieces.GameState),
		
	}
	server.setupMelody()

	gin.ForceConsoleColor()
	router := gin.Default()

	router.Use(cors.New(config.CORSConfig()))
	// ── Public ───────────────────────────────────────────────────────────────
	router.GET("/", server.welcome)

	//users routes
	users := router.Group("/users")
	
	users.GET("/check-username", server.checkUsernameExists)
	users.POST("/signup", server.createUser)
	users.POST("/confirm-email", server.confirmEmail)
	users.POST("/send-emailotp", server.sendEmailOTP)
	users.POST("/signin", server.loginUser)
	users.POST("/refresh-token", server.refreshToken)

	// ── Protected (Bearer JWT required) ─────────────────────────────────────
	
	//users
	authUsers := router.Group("/users").Use(authMiddleware(server.tokenMaker))

	authUsers.GET("/me",server.getMe)

	// games
	authGames := router.Group("/games").Use(authMiddleware(server.tokenMaker))

	authGames.POST("/", server.createGame)
	authGames.GET("/", server.listWaitingGames)
	authGames.GET("/mine", server.listMyGames)
	authGames.GET("/:id", server.getGame)
	authGames.POST("/:id/join", server.joinGame)
	authGames.POST("/:id/resign", server.resignGame)
	authGames.GET("/:id/moves", server.getGameMoves)

	// chat (scoped under games)
	authGames.POST("/:id/chat", server.sendChatMessage)
	authGames.GET("/:id/chat", server.getChatMessages)

	// voice (WebRTC session lifecycle; signalling travels over /ws)
	authGames.POST("/:id/voice", server.startVoiceSession)
	authGames.GET("/:id/voice", server.getActiveVoiceSession)
	authGames.PATCH("/:id/voice/:vid/activate", server.activateVoiceSession)
	authGames.DELETE("/:id/voice/:vid", server.endVoiceSession)

	// ── WebSocket ────────────────────────────────────────────────────────────
	// Bearer token must be sent as ?token=<access_token> (WS clients can't set headers).
	// Game room is selected via ?game_id=<uuid>.
	//Example url
	//ws://localhost:8080/ws/game?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyXzEyMyJ9.abc123&game_id=550e8400-e29b-41d4-a716-446655440000
	router.GET("/ws", server.handleWebSocket)

	server.router = router
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
	return server.httpServer.Shutdown(ctx)
}

func (server *Server) welcome(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "Welcome to the Chess Game Server"})
}
