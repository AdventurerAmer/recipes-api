// Recipes API
//
// This is a sample recipes API.
//
//	Schemes: http
//	Host: localhost:3000
//	BasePath: /
//	Version: 1.0.0
//	Contact: Ahmed Amer
//
// <ahamerdev@gmail.com>
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
// swagger:meta
package main

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Recipe struct {
	ID           string    `json:"id"`
	CreatedAt    time.Time `json:"createdAt"`
	Name         string    `json:"name"`
	Tags         []string  `json:"tags"`
	Ingredients  []string  `json:"ingredients"`
	Instructions []string  `json:"instructions"`
}

var mu sync.RWMutex
var recipes []Recipe

// swagger:operation POST /recipes recipes createRecipe
// Creates a new recipe
// --
// parameters:
//
//   - name: name
//     in: body
//     description: Name of the recipe
//     required: true
//     type: string
//
//   - tags: tags
//     in: body
//     description: Tags of the recipe
//     required: true
//     type: string
//
//   - ingredients: ingredients
//     in: ingredients
//     description: Ingredients of the recipe
//     required: true
//     type: string
//
//   - instructions: instructions
//     in: instructions
//     description: instructions of the recipe
//     required: true
//     type: string
//
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Successful operation
//	'400':
//	    description: Invalid input
func NewRecipeHandler(c *gin.Context) {
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	recipe.ID = uuid.NewString()
	recipe.CreatedAt = time.Now()
	mu.Lock()
	defer mu.Unlock()
	recipes = append(recipes, recipe)
	c.JSON(http.StatusCreated, recipe)
}

// swagger:operation GET /recipes recipes listRecipes
// Returns list of recipes
// --
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Successful operation
func ListRecipesHandler(c *gin.Context) {
	mu.RLock()
	defer mu.RUnlock()
	c.JSON(http.StatusOK, recipes)
}

// swagger:operation POST /recipes/search recipes searchRecipes
// Creates a new recipe
// --
// parameters:
//
//   - name: tag
//     in: query
//     description: tag of the recipe
//     required: false
//     type: string
//
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Successful operation
func SearchRecipesHandler(c *gin.Context) {
	mu.RLock()
	defer mu.RUnlock()
	tag := c.Query("tag")
	result := make([]Recipe, 0)
	for _, r := range recipes {
		for _, t := range r.Tags {
			if strings.EqualFold(tag, t) {
				result = append(result, r)
			}
		}
	}
	c.JSON(http.StatusOK, result)
}

// swagger:operation PUT /recipes/{id} recipes updateRecipe
// Update an existing recipe
// --
// parameters:
//   - name: id
//     in: path
//     description: ID of the recipe
//     required: true
//     type: string
//
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Successful operation
//	'400':
//	    description: Invalid input
//	'404':
//	    description: Invalid recipe ID
func UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	mu.Lock()
	defer mu.Unlock()
	idx := -1
	for i, r := range recipes {
		if r.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
		return
	}
	recipe.ID = recipes[idx].ID
	recipe.CreatedAt = recipes[idx].CreatedAt
	recipes[idx] = recipe
	c.JSON(http.StatusOK, recipe)
}

// swagger:operation DELETE /recipes/{id} recipes deleteRecipe
// Delete an existing recipe
// --
// parameters:
//   - name: id
//     in: path
//     description: ID of the recipe
//     required: true
//     type: string
//
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Successful operation
//	'400':
//	    description: Invalid input
//	'404':
//	    description: Invalid recipe ID
func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	mu.Lock()
	defer mu.Unlock()
	idx := -1
	for i, r := range recipes {
		if r.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
		return
	}
	recipes = append(recipes[:idx], recipes[idx+1:]...)
	c.JSON(http.StatusOK, gin.H{"message": "Recipe has been deleted"})
}

func main() {
	r := gin.Default()
	r.POST("/recipes", NewRecipeHandler)
	r.GET("/recipes", ListRecipesHandler)
	r.GET("/recipes/search", SearchRecipesHandler)
	r.PUT("/recipes/:id", UpdateRecipeHandler)
	r.DELETE("/recipes/:id", DeleteRecipeHandler)
	r.Run(":3000")
}
