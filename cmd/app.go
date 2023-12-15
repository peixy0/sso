package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"os"
	"sso/login"
	"sso/sso"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

var (
	loginService = login.NewLoginService()
	tokenService = sso.NewTokenService()
	ssoService   = sso.NewSingleSignOnService()
)

func handleLoggedIn(c *gin.Context, service string) {
	if service == "" {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"msg": "SECURED",
		})
	} else {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/api/requestToken?service=%v", service))
	}
}

func loginPage(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")
	if user == "admin" {
		handleLoggedIn(c, c.Query("service"))
		return
	}
	c.HTML(http.StatusOK, "login.html", gin.H{
		"requiresLogin": true,
	})
}

func tryLogin(c *gin.Context) {
	password := c.PostForm("password")
	if loginService.IsValidCredential(password) {
		session := sessions.Default(c)
		session.Set("user", "admin")
		session.Save()
		handleLoggedIn(c, c.Query("service"))
		return
	}
	c.HTML(http.StatusOK, "login.html", gin.H{
		"error":         "FORBIDDEN",
		"requiresLogin": true,
	})
}

func getToken(c *gin.Context) {
	service := c.Param("service")
	tokenId := c.Param("id")
	token := tokenService.GetToken(tokenId, service)
	c.JSON(http.StatusOK, token)
}

func requestToken(c *gin.Context) {
	service := c.Query("service")
	session := sessions.Default(c)
	user := session.Get("user")
	if user != "admin" {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/?service=%v", service))
		return
	}
	token := tokenService.RequestToken(service)
	redir := tokenService.GetServiceCallback(token.Id, service)
	if redir == "" {
		c.Status(http.StatusBadRequest)
		return
	}
	c.Redirect(http.StatusSeeOther, redir)
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
	debugEnabled := os.Getenv("DEBUG_MODE")
	if debugEnabled == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	gin.DisableConsoleColor()

	router := gin.New()
	router.LoadHTMLGlob("templates/*")
	router.MaxMultipartMemory = 8 << 20

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		MaxAge:       12 * time.Hour,
	}))

	buf := make([]byte, 64)
	n, err := rand.Read(buf)
	if err != nil || n != cap(buf) {
		panic("error initializing session key")
	}
	store := cookie.NewStore(buf)
	router.Use(sessions.Sessions("session", store))
	router.GET("/", loginPage)
	router.POST("/", tryLogin)
	router.GET("/api/requestToken", requestToken)
	router.GET("/api/token/:service/:id", getToken)
	router.GET("/api/check", checkToken)
	router.GET("/api/signin", ssoService.HandleTokenResp)
	log.Printf("Server running on %v", servAddr)
	router.Run(servAddr)
}
