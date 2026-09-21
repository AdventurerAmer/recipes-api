package handlers

import (
	"net/http"

	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"github.com/gin-gonic/gin"
)

type UsersHandler struct {
	UsersService ports.UsersService
}

func NewUsersHandler(usersService ports.UsersService) *UsersHandler {
	return &UsersHandler{
		UsersService: usersService,
	}
}

func (h *UsersHandler) Register(c *gin.Context) {
	var req ports.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, _, err := h.UsersService.Register(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
