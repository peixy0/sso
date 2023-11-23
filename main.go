package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var tokenService = NewTokenService()
var ssoService = NewSingleSignOnService()

func getToken(c *gin.Context) {
	service := c.Param("service")
	tokenId := c.Param("id")
	token := tokenService.GetToken(tokenId, service)
	c.JSON(http.StatusOK, token)
}

func requestToken(c *gin.Context) {
	service := c.Query("service")
	token := tokenService.RequestToken(service)
	redir := tokenService.GetServiceCallback(token.Id, service)
	if redir == "" {
		c.Status(http.StatusBadRequest)
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, redir)
}

func checkToken(c *gin.Context) {
	token := ssoService.ValidateToken(c)
	if token == nil {
		c.Status(http.StatusUnauthorized)
		return
	}
	c.Status(http.StatusOK)
}

func main() {
	servAddr := os.Getenv("SERVER")
	if servAddr == "" {
		servAddr = "127.0.0.1:13000"
	}
	var debugEnabled = os.Getenv("DEBUG_MODE")
	if debugEnabled == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	gin.DisableConsoleColor()
	router := gin.New()
	router.MaxMultipartMemory = 8 << 20
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		MaxAge:       12 * time.Hour,
	}))
	router.GET("/api/requestToken", requestToken)
	router.GET("/api/token/:service/:id", getToken)
	router.GET("/api/check", checkToken)
	router.GET("/api/signin", ssoService.HandleTokenResp)
	log.Printf("Server running on %v", servAddr)
	router.Run(servAddr)
}
