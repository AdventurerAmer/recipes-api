package handlers

import (
	"net/http"

	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	service ports.AuthService
}

func NewAuthHandler(service ports.AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req ports.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.Login(c, req)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
	}

	user := resp.User
	token := uuid.New().String()
	session := sessions.Default(c)
	session.Set("id", user.Id)
	session.Set("email", user.Email)
	session.Set("displayName", user.DisplayName)
	session.Set("token", token)
	if err := session.Save(); err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
	}

	c.JSON(http.StatusOK, gin.H{"message": "User signed in"})
}

func (handler *AuthHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Options(sessions.Options{MaxAge: -1})
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Signed out..."})
}

func (handler *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		sessionToken := session.Get("token")
		if sessionToken == nil {
			c.JSON(http.StatusForbidden, gin.H{"message": "Not loggedin"})
			c.Abort()
		}
		c.Next()
	}
}
