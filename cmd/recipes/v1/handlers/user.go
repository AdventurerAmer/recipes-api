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
		c.Error(err)
		return
	}

	resp, _, err := h.service.Register(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Users) Get(c *gin.Context) {
	userId := c.GetString("user_id")

	req := ports.GetUserRequest{
		Id: userId,
	}

	resp, err := h.service.Get(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Users) SendVerification(c *gin.Context) {
	var req ports.SendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	resp, err := h.service.SendVerification(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Users) Verify(c *gin.Context) {
	var req ports.VerifyUserRequest
	if err := c.ShouldBind(&req); err != nil {
		c.Error(err)
		return
	}

	resp, err := h.service.Verify(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Users) ForgotPassword(c *gin.Context) {
	var req ports.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	resp, err := h.service.ForgotPassword(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Users) ResetPassword(c *gin.Context) {
	var req ports.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	resp, err := h.service.ResetPassword(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Users) ChangePassword(c *gin.Context) {
	var req ports.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	req.UserId = c.GetString("user_id")

	resp, err := h.service.ChangePassword(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
