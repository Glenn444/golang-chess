package api

import (
	"os"
	"testing"
	"time"

	"github.com/Glenn444/golang-chess/config"
	"github.com/Glenn444/golang-chess/internal/db"
	"github.com/Glenn444/golang-chess/internal/utils/random"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T,store db.Store) *Server{
	config := config.Config{
		TokenSymmetricKey: random.RandomString(32),
		AcessTokenDuration: time.Minute,
	}

	server,err := NewServer(config,store)
	require.NoError(t,err)

	return server
}

func TestMain(m *testing.M)  {

	gin.SetMode(gin.TestMode)
	
	os.Exit(m.Run())
}