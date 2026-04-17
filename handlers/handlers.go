package handlers

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/AdventurerAmer/recipes-api/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/argon2"
)

type RecipesHandler struct {
	ctx         context.Context
	collection  *mongo.Collection
	redisClient *redis.Client
}

func NewRecipesHandler(ctx context.Context, collection *mongo.Collection, redisClient *redis.Client) *RecipesHandler {
	return &RecipesHandler{
		ctx:         ctx,
		collection:  collection,
		redisClient: redisClient,
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
	h.redisClient.Del(h.ctx, "recipes")
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
	var recipes []models.Recipe

	cacheKey := "recipes"
	cacheData, cacheErr := h.redisClient.Get(h.ctx, cacheKey).Result()
	if cacheErr == nil {
		if err := json.Unmarshal([]byte(cacheData), &recipes); err == nil {
			c.JSON(http.StatusOK, recipes)
			return
		}
	}

	cursor, err := h.collection.Find(h.ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "find recipes failed"})
		return
	}
	defer cursor.Close(context.Background())
	for cursor.Next(h.ctx) {
		var recipe models.Recipe
		if err := cursor.Decode(&recipe); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "find recipes failed"})
			return
		}
		recipes = append(recipes, recipe)
	}
	if cacheErr == redis.Nil {
		h.redisClient.Set(h.ctx, cacheKey, recipes, time.Minute)
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
	h.redisClient.Del(h.ctx, "recipes")
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
		return
	}
	h.redisClient.Del(h.ctx, "recipes")
	c.JSON(http.StatusOK, gin.H{"message": "Recipe has been deleted"})
}

type AuthHandler struct {
	ctx        context.Context
	client     *mongo.Client
	collection *mongo.Collection
}

func NewAuthHandler(ctx context.Context, client *mongo.Client, collection *mongo.Collection) *AuthHandler {
	return &AuthHandler{
		ctx:        ctx,
		client:     client,
		collection: collection,
	}
}

// Recommended parameters (RFC 9106)
const (
	argon2Time    = 1
	argon2Memory  = 64 * 1024 // 64 MB
	argon2Threads = 4
	argon2KeyLen  = 32
	argon2SaltLen = 16
)

func hashPassward(password string) (string, error) {
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argon2Memory, argon2Time, argon2Threads, b64Salt, b64Hash)
	return encoded, nil
}

func verifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	var memory, time uint32
	var threads uint8
	fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	salt, _ := base64.RawStdEncoding.DecodeString(parts[4])
	decodedHash, _ := base64.RawStdEncoding.DecodeString(parts[5])
	comparisonHash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(decodedHash)))
	return subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1, nil
}

func (h *AuthHandler) SignUpHandler(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	session, err := h.client.StartSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var insertUser *models.User
	_, err = session.WithTransaction(h.ctx, func(ctx mongo.SessionContext) (interface{}, error) {
		filter := bson.M{"username": user.Username}
		result := h.collection.FindOne(ctx, filter)
		var findUser models.User
		if err := result.Decode(&findUser); err != nil {
			if !errors.Is(err, mongo.ErrNoDocuments) {
				return nil, err
			}
		}
		hash, err := hashPassward(user.Password)
		if err != nil {
			return nil, err
		}
		insertUser = &models.User{
			ID:       primitive.NewObjectID(),
			Username: user.Username,
			Password: hash,
		}
		_, err = h.collection.InsertOne(ctx, insertUser)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": insertUser.ID})
}

func (h *AuthHandler) SignInHandler(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter := bson.M{"username": user.Username}
	result := h.collection.FindOne(h.ctx, filter)
	var findUser models.User
	if err := result.Decode(&findUser); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ok, err := verifyPassword(user.Password, findUser.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
	}

	token := uuid.New().String()
	session := sessions.Default(c)
	session.Set("username", findUser.Username)
	session.Set("token", token)
	if err := session.Save(); err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
	}

	c.JSON(http.StatusOK, gin.H{"message": "User signed in"})
}

func (handler *AuthHandler) SignOutHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.JSON(http.StatusOK, gin.H{"message": "Signed out..."})
}

func (handler *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		sessionToken := session.Get("token")
		if sessionToken == nil {
			c.JSON(http.StatusForbidden, gin.H{"message": "Not logged"})
			c.Abort()
		}
		c.Next()
	}
}
