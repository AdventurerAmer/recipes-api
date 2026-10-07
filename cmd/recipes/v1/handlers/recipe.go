package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"github.com/gin-gonic/gin"
)

type Recipes struct {
	service ports.RecipesService
}

func NewRecipes(service ports.RecipesService) *Recipes {
	return &Recipes{
		service: service,
	}
}

func (h *Recipes) Create(c *gin.Context) {
	userId := c.GetString("user_id")

	var req ports.CreateRecipeRequest
	if err := c.ShouldBind(&req); err != nil {
		c.Error(err)
		return
	}

	if err := json.Unmarshal([]byte(req.RecipeStr), &req.Recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipe JSON format"})
		return
	}

	file, err := req.ImageHeader.Open()
	if err != nil {
		c.Error(err)
		return
	}

	req.UserId = userId
	req.Image = ports.ObjectStorageFile{
		Reader:      file,
		Size:        int(req.ImageHeader.Size),
		ContentType: req.ImageHeader.Header.Get("Content-Type"),
	}
	resp, err := h.service.Create(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *Recipes) List(c *gin.Context) {
	var req ports.ListRecipesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(err)
		return
	}

	resp, err := h.service.List(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Recipes) Search(c *gin.Context) {
	var req ports.SearchRecipesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(err)
		return
	}

	resp, err := h.service.Search(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Recipes) Get(c *gin.Context) {
	var req ports.GetRecipeRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.Error(err)
		return
	}

	resp, err := h.service.Get(c, req)
	if err != nil {
		c.Error(err)
		return
	}
	c.AbortWithStatusJSON(http.StatusOK, resp)
}

func (h *Recipes) Update(c *gin.Context) {
	var req ports.UpdateRecipeRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.Error(err)
		return
	}

	if err := c.ShouldBind(&req); err != nil {
		c.Error(err)
		return
	}

	if req.ImageHeader != nil {
		file, err := req.ImageHeader.Open()
		if err != nil {
			c.Error(err)
			return
		}
		req.Image = &ports.ObjectStorageFile{
			Reader:      file,
			Size:        int(req.ImageHeader.Size),
			ContentType: req.ImageHeader.Header.Get("Content-Type"),
		}
	}

	req.UserId = c.GetString("user_id")
	resp, err := h.service.Update(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Recipes) Delete(c *gin.Context) {
	var req ports.DeleteRecipeRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.Error(err)
		return
	}

	req.UserId = c.GetString("user_id")

	resp, err := h.service.Delete(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
