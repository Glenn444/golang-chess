package api

import (
	"fmt"
	"net/http"

	"github.com/Glenn444/golang-chess/config"
	"github.com/Glenn444/golang-chess/internal/db"
	"github.com/Glenn444/golang-chess/internal/token"
	"github.com/Glenn444/golang-chess/internal/utils/emails"
	"github.com/gin-gonic/gin"
)

// Server serves HTTP requests for our chess service
type Server struct {
	config     config.Config
	emailClient emails.EmailClient
	tokenMaker token.Maker
	store      db.Store
	router     *gin.Engine
}

//Initializes the routes
func NewServer(config config.Config, store db.Store) (*Server, error) {
	jwtTokenMaker, err := token.NewJWTMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker %w\n", err)
	}
	server := &Server{
		tokenMaker: jwtTokenMaker,
		store:      store,
		config:     config,
		emailClient: *emails.NewEmailClient(config.RESEND_API_KEY),
	}

	// Force log's color
	gin.ForceConsoleColor()
	router := gin.Default()

	

	//add middleware to refresh token
	router.Use()
	//add routes to router
	router.GET("/", server.welcome)
	
	//users routes
	users := router.Group("/users")

	//auth group routes
	//authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))


	users.GET("/check-username",server.checkUsernameExists)
	users.POST("/confirm-email",server.confirmEmail)
	users.POST("/send-otp",server.sendEmailOTP)
	users.POST("/signup",server.createUser)
	

	//router.POST("/token/refresh-token",server.refreshToken)

	server.router = router
	return server, nil
}

// start runs the HTTP server on a specific address
func (server *Server) Start(address string) error {
	return server.router.Run(address)
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
func (server *Server) welcome(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "Welcome to the Server")

}
