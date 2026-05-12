package ports

import (
	"context"
	"mime/multipart"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
)

type RecipesRepository interface {
	Create(ctx context.Context, recipe *domain.Recipe) error
	Get(ctx context.Context, id string) (domain.Recipe, error)
	List(ctx context.Context, lastID, userID, sortBy string, limit int) ([]domain.Recipe, int, error)
	Update(ctx context.Context, recipe *domain.Recipe) error
	Delete(ctx context.Context, userID, id string) error
}

type RecipesService interface {
	Create(ctx context.Context, user domain.User, req CreateRecipeRequest) (CreateRecipeResponse, error)
	Get(ctx context.Context, req GetRecipeRequest) (GetRecipeResponse, error)
	List(ctx context.Context, req ListRecipesRequest) (ListRecipesResponse, error)
	Update(ctx context.Context, user domain.User, req UpdateRecipeRequest) (UpdateRecipeResponse, error)
	Delete(ctx context.Context, user domain.User, req DeleteRecipeRequest) (DeleteRecipeResponse, error)
}

type CreateRecipeRequest struct {
	Recipe struct {
		Name         string   `json:"name" binding:"required,min=1"`
		Tags         []string `json:"tags"`
		Ingredients  []string `json:"ingredients" binding:"required,min=1"`
		Instructions []string `json:"instructions" binding:"required,min=1"`
	} `form:"recipe" binding:"required"`
	ImageHeader *multipart.FileHeader `form:"image" binding:"required"`
	Image       ObjectStorageFile
}

type CreateRecipeResponse struct {
	Recipe domain.Recipe `json:"recipe"`
}

type GetRecipeRequest struct {
	ID string `json:"id" uri:"id" binding:"required"`
}

type GetRecipeResponse struct {
	Recipe domain.Recipe `json:"recipe"`
}

type ListRecipesRequest struct {
	LastID string `json:"lastID" form:"lastID"`
	UserID string `json:"userID" from:"userID"`
	SortBy string `json:"sortBy" form:"sortBy,default=-createdAt"`
	Limit  int    `json:"limit" form:"limit,default=20"`
}

type ListRecipesResponse struct {
	Recipes []domain.Recipe `json:"recipes"`
	Total   int             `json:"total"`
}

type UpdateRecipeRequest struct {
	ID     string `json:"id" uri:"id" binding:"required"`
	Recipe struct {
		Name         *string  `json:"name" binding:"omitempty,min=1"`
		Tags         []string `json:"tags" binding:"omitempty,min=1"`
		Ingredients  []string `json:"ingredients" binding:"omitempty,min=1"`
		Instructions []string `json:"instructions" binding:"omitempty,min=1"`
	} `form:"recipe"`
	ImageHeader *multipart.FileHeader `form:"image"`
	Image       *ObjectStorageFile
}

type UpdateRecipeResponse struct {
	Recipe domain.Recipe `json:"recipe"`
}

type DeleteRecipeRequest struct {
	ID string `json:"id" uri:"id" binding:"required"`
}

type DeleteRecipeResponse struct {
	Message string `json:"message"`
}
