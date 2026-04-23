package handlers

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"github.com/AdventurerAmer/recipes-api/models"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/argon2"
)

type RecipesHandler struct {
	RecipesService ports.RecipesService
}

func NewRecipesHandler(recipesService ports.RecipesService) *RecipesHandler {
	return &RecipesHandler{
		RecipesService: recipesService,
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
	var req ports.CreateRecipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	session := sessions.Default(c)
	user := domain.User{
		ID:       session.Get("id").(string),
		Username: session.Get("username").(string),
	}
	resp, err := h.RecipesService.Create(c, user, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
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
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req := ports.ListRecipesRequest{
		LastID: c.Query("lastID"),
		Sort:   c.DefaultQuery("sort", "-createdAt"),
		Limit:  limit,
	}
	resp, err := h.RecipesService.List(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
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
	var req ports.UpdateRecipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ID = c.Param("id")
	session := sessions.Default(c)
	user := domain.User{
		ID:       session.Get("id").(string),
		Username: session.Get("username").(string),
	}
	resp, err := h.RecipesService.Update(c, user, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
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
	session := sessions.Default(c)
	user := domain.User{
		ID:       session.Get("id").(string),
		Username: session.Get("username").(string),
	}
	req := ports.DeleteRecipeRequest{ID: c.Param("id")}
	resp, err := h.RecipesService.Delete(c, user, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
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
	session.Set("id", findUser.ID)
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
