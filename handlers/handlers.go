package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/AdventurerAmer/recipes-api/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RecipesHandler struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewRecipesHandler(ctx context.Context, collection *mongo.Collection) *RecipesHandler {
	return &RecipesHandler{
		ctx:        ctx,
		collection: collection,
	}
}

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
func (h *RecipesHandler) NewRecipeHandler(c *gin.Context) {
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	recipe.ID = primitive.NewObjectID()
	recipe.CreatedAt = time.Now()
	if _, err := h.collection.InsertOne(h.ctx, recipe); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "insert reciped failed"})
		return
	}
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
func (h *RecipesHandler) ListRecipesHandler(c *gin.Context) {
	cursor, err := h.collection.Find(h.ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "find recipes failed"})
		return
	}
	defer cursor.Close(context.Background())
	recipes := make([]models.Recipe, 0)
	for cursor.Next(h.ctx) {
		var recipe models.Recipe
		if err := cursor.Decode(&recipe); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "find recipes failed"})
			return
		}
		recipes = append(recipes, recipe)
	}
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
func (h *RecipesHandler) SearchRecipesHandler(c *gin.Context) {
	tag := c.Query("tag")

	filter := bson.M{
		"tags": bson.D{
			{Key: "$in", Value: tag},
		},
	}
	cursor, err := h.collection.Find(h.ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "find recipes failed"})
		return
	}
	defer cursor.Close(context.Background())
	recipes := make([]models.Recipe, 0)
	for cursor.Next(h.ctx) {
		var recipe models.Recipe
		if err := cursor.Decode(&recipe); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "find recipes failed"})
			return
		}
		recipes = append(recipes, recipe)
	}
	c.JSON(http.StatusOK, recipes)
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
func (h *RecipesHandler) UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	filter := bson.M{"_id": objectID}
	set := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "name", Value: recipe.Name},
			{Key: "instructions", Value: recipe.Instructions},
			{Key: "ingredients", Value: recipe.Ingredients},
			{Key: "tags", Value: recipe.Tags},
		}},
	}
	if _, err := h.collection.UpdateOne(h.ctx, filter, set); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Recipe has been updated"})
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
func (h *RecipesHandler) DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	filter := bson.M{"_id": objectID}
	if _, err := h.collection.DeleteOne(h.ctx, filter); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"message": "Recipe has been deleted"})
}
