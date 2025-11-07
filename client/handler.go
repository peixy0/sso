package sso

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service}
}

func (h *Handler) HandleLogin(c *gin.Context) {
	c.Redirect(http.StatusFound, h.service.GetLoginURL())
}

func (h *Handler) HandleCallback(c *gin.Context) {
	tokenString := c.Query("token")
	if tokenString == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	claims, err := h.service.ValidateToken(tokenString)
	if err != nil {
		c.Status(http.StatusUnauthorized)
		return
	}

	session := sessions.Default(c)
	session.Set("user", claims.User)
	session.Set("token", tokenString)
	if err := session.Save(); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Redirect(http.StatusFound, "/")
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		token := session.Get("token")
		if token == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}
