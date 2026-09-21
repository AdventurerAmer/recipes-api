package handlers

import (
	"net/http"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type Recipes struct {
	Service ports.RecipesService
}

func NewRecipesHandler(service ports.RecipesService) *Recipes {
	return &Recipes{
		Service: service,
	}
}

func (h *Recipes) Create(c *gin.Context) {
	var req ports.CreateRecipeRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	session := sessions.Default(c)
	user := domain.User{
		Id:          session.Get("id").(string),
		Email:       session.Get("email").(string),
		DisplayName: session.Get("displayName").(string),
	}
	file, err := req.ImageHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	req.Image = ports.ObjectStorageFile{
		Reader:      file,
		Size:        int(req.ImageHeader.Size),
		ContentType: req.ImageHeader.Header.Get("Content-Type"),
	}
	resp, err := h.Service.Create(c, &user, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Recipes) List(c *gin.Context) {
	var req ports.ListRecipesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.Service.List(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Recipes) Search(c *gin.Context) {
	var req ports.SearchRecipesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.Service.Search(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Recipes) Get(c *gin.Context) {
	var req ports.GetRecipeRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.Service.Get(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Recipes) Update(c *gin.Context) {
	var req ports.UpdateRecipeRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	session := sessions.Default(c)
	user := domain.User{
		Id:          session.Get("id").(string),
		Email:       session.Get("email").(string),
		DisplayName: session.Get("displayName").(string),
	}
	if req.ImageHeader != nil {
		file, err := req.ImageHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		req.Image = &ports.ObjectStorageFile{
			Reader:      file,
			Size:        int(req.ImageHeader.Size),
			ContentType: req.ImageHeader.Header.Get("Content-Type"),
		}
	}
	resp, err := h.Service.Update(c, &user, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Recipes) Delete(c *gin.Context) {
	session := sessions.Default(c)
	user := domain.User{
		Id:          session.Get("id").(string),
		Email:       session.Get("email").(string),
		DisplayName: session.Get("displayName").(string),
	}
	var req ports.DeleteRecipeRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.Service.Delete(c, &user, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
