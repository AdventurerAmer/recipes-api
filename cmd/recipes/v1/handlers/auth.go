package handlers

import (
	"net/http"

	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type Auth struct {
	service ports.AuthService
}

func NewAuth(service ports.AuthService) *Auth {
	return &Auth{
		service: service,
	}
}

func (h *Auth) Login(c *gin.Context) {
	var req ports.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	resp, err := h.service.Login(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	user := resp.User
	session := sessions.Default(c)
	session.Set("user_id", user.Id)
	if err := session.Save(); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Signed in"})
}

func (handler *Auth) Logout(c *gin.Context) {
	session := sessions.Default(c)

	session.Clear()
	session.Options(sessions.Options{MaxAge: -1})
	if err := session.Save(); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Signed out"})
}

func (handler *Auth) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userId, ok := session.Get("user_id").(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
			return
		}

		c.Set("user_id", userId)
		c.Next()
	}
}
