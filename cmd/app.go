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
	"github.com/joho/godotenv"
)

type App struct {
	auth     *login.Authenticator
	registry *sso.ServiceRegistry
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}
	registry, err := sso.LoadServiceRegistry("config.json")
	if err != nil {
		log.Fatalf("Failed to load service registry: %v", err)
	}

	app := &App{
		auth:     login.NewAuthenticator(),
		registry: registry,
	}

	router := setupRouter(app)
	runServer(router)
}

func setupRouter(app *App) *gin.Engine {
	if os.Getenv("DEBUG_MODE") == "" {
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

	store, err := newCookieStore()
	if err != nil {
		log.Fatalf("Failed to create cookie store: %v", err)
	}
	router.Use(sessions.Sessions("session", store))

	router.GET("/", app.handleLogin)
	router.POST("/", app.handleLoginAttempt)
	router.GET("/api/requestToken", app.handleTokenRequest)

	return router
}

func newCookieStore() (cookie.Store, error) {
	buf := make([]byte, 64)
	if _, err := rand.Read(buf); err != nil {
		return nil, fmt.Errorf("failed to generate session key: %w", err)
	}
	store := cookie.NewStore(buf)
	store.Options(sessions.Options{
		Secure:   false,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   3600 * 24 * 7,
	})
	return store, nil
}

func runServer(router *gin.Engine) {
	servAddr := os.Getenv("SERVER")
	if servAddr == "" {
		servAddr = "127.0.0.1:13000"
	}
	log.Printf("Server running on %v", servAddr)
	if err := router.Run(servAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func (app *App) handleLogin(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")
	if user == "admin" {
		app.handleLoggedIn(c, c.Query("service"))
		return
	}
	c.HTML(http.StatusOK, "login.html", gin.H{
		"requiresLogin": true,
	})
}

func (app *App) handleLoginAttempt(c *gin.Context) {
	password := c.PostForm("password")
	if app.auth.Authenticate(password) {
		session := sessions.Default(c)
		session.Set("user", "admin")
		session.Save()
		app.handleLoggedIn(c, c.Query("service"))
		return
	}
	c.HTML(http.StatusOK, "login.html", gin.H{
		"error":         "FORBIDDEN",
		"requiresLogin": true,
	})
}

func (app *App) handleLoggedIn(c *gin.Context, service string) {
	if service == "" {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"msg": "SECURED",
		})
	} else {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/api/requestToken?service=%v", service))
	}
}

func (app *App) handleTokenRequest(c *gin.Context) {
	serviceName := c.Query("service")
	session := sessions.Default(c)
	user := session.Get("user").(string)
	if user != "admin" {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/?service=%v", serviceName))
		return
	}

	service, ok := app.registry.Get(serviceName)
	if !ok {
		c.Status(http.StatusBadRequest)
		return
	}

	token, err := sso.CreateToken(service, user)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	redirectURL := fmt.Sprintf("%s?token=%s", service.CallbackURL, token)
	c.Redirect(http.StatusSeeOther, redirectURL)
}
