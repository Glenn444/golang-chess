package api

import (
	"fmt"
	"net/http"

	"github.com/Glenn444/golang-chess/config"
	"github.com/Glenn444/golang-chess/internal/db"
	"github.com/Glenn444/golang-chess/internal/token"
	"github.com/Glenn444/golang-chess/internal/utils/emails"
	"github.com/gin-gonic/gin"
	"github.com/olahol/melody"
)

type Server struct {
	config      config.Config
	emailClient emails.EmailClient
	tokenMaker  token.Maker
	store       db.Store
	router      *gin.Engine
	melody      *melody.Melody
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
		emailClient: *emails.NewEmailClient(config.RESEND_API_KEY),
		melody:      melody.New(),
	}
	server.setupMelody()

	gin.ForceConsoleColor()
	router := gin.Default()

	// ── Public ───────────────────────────────────────────────────────────────
	router.GET("/", server.welcome)

	users := router.Group("/users")
	users.GET("/check-username", server.checkUsernameExists)
	users.POST("/signup", server.createUser)
	users.POST("/confirm-email", server.confirmEmail)
	users.POST("/send-emailotp", server.sendEmailOTP)
	users.POST("/signin", server.loginUser)
	users.POST("/refresh-token", server.refreshToken)

	// ── Protected (Bearer JWT required) ─────────────────────────────────────
	auth := router.Group("/").Use(authMiddleware(server.tokenMaker))

	// profile
	auth.GET("/users/me", server.getMe)

	// games
	auth.POST("/games", server.createGame)
	auth.GET("/games", server.listWaitingGames)
	auth.GET("/games/mine", server.listMyGames)
	auth.GET("/games/:id", server.getGame)
	auth.POST("/games/:id/join", server.joinGame)
	auth.POST("/games/:id/resign", server.resignGame)
	auth.GET("/games/:id/moves", server.getGameMoves)

	// chat
	auth.POST("/games/:id/chat", server.sendChatMessage)
	auth.GET("/games/:id/chat", server.getChatMessages)

	// voice (WebRTC session lifecycle; signalling travels over /ws)
	auth.POST("/games/:id/voice", server.startVoiceSession)
	auth.GET("/games/:id/voice", server.getActiveVoiceSession)
	auth.PATCH("/games/:id/voice/:vid/activate", server.activateVoiceSession)
	auth.DELETE("/games/:id/voice/:vid", server.endVoiceSession)

	// ── WebSocket ────────────────────────────────────────────────────────────
	// Bearer token must be sent as ?token=<access_token> (WS clients can't set headers).
	// Game room is selected via ?game_id=<uuid>.
	router.GET("/ws", server.handleWebSocket)

	server.router = router
	return server, nil
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func (server *Server) welcome(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "Welcome to the Chess Game Server"})
}
