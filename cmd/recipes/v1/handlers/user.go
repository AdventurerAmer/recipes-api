package handlers

import (
	"net/http"

	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"github.com/gin-gonic/gin"
)

type Users struct {
	service ports.UsersService
}

func NewUsers(service ports.UsersService) *Users {
	return &Users{
		service: service,
	}
}

func (h *Users) Register(c *gin.Context) {
	var req ports.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, _, err := h.service.Register(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
