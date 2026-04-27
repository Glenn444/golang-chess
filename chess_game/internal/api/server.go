package api

import (
	"fmt"
	"net/http"

	db "github.com/Glenn444/golang-chess/internal/db"
	"github.com/Glenn444/golang-chess/internal/token"
	"github.com/Glenn444/golang-chess/utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Server serves HTTP requests for our banking service
type Server struct {
	config     util.Config
	tokenMaker token.Maker
	store      db.Store
	router     *gin.Engine
}

//Initializes the routes
func NewServer(config util.Config, store db.Store) (*Server, error) {
	jwtTokenMaker, err := token.NewJWTMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker %w\n", err)
	}
	server := &Server{
		tokenMaker: jwtTokenMaker,
		store:      store,
		config:     config,
	}

	// Force log's color
	gin.ForceConsoleColor()
	router := gin.Default()

	

	//add middleware to refresh token
	router.Use()
	//add routes to router
	router.GET("/", server.welcome)

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))
	authRoutes.POST("/accounts", server.createAccount)
	authRoutes.GET("/accounts/:id", server.getAccountById)
	authRoutes.GET("/accounts", server.listAllAccounts)

	authRoutes.POST("/transfers", server.createTransfer)

	authRoutes.GET("/user", server.getUser)
	authRoutes.GET("/users", server.getAllUsers)


	router.POST("/user", server.createUser)
	
	router.POST("/users/login", server.loginUser)

	router.POST("/token/refresh",server.refreshToken)

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

func (server *Server) welcome(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, "Welcome to the Server")

}
