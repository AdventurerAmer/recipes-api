package ports

import (
	"context"
	"mime/multipart"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
)

type RecipesRepository interface {
	Create(ctx context.Context, recipe *domain.Recipe) error
	Get(ctx context.Context, id string) (*domain.Recipe, error)
	List(ctx context.Context, userId, lastId, sort string, limit int) ([]domain.Recipe, int, error)
	Search(ctx context.Context, name string, page, pageSize int) ([]domain.Recipe, int, error)
	Update(ctx context.Context, recipe *domain.Recipe) error
	Delete(ctx context.Context, userId, id string) error
}

type RecipesService interface {
	Create(ctx context.Context, user domain.User, req CreateRecipeRequest) (CreateRecipeResponse, error)
	Get(ctx context.Context, req GetRecipeRequest) (GetRecipeResponse, error)
	List(ctx context.Context, req ListRecipesRequest) (ListRecipesResponse, error)
	Search(ctx context.Context, name SearchRecipesRequest) (SearchRecipesResponse, error)
	Update(ctx context.Context, user domain.User, req UpdateRecipeRequest) (UpdateRecipeResponse, error)
	Delete(ctx context.Context, user domain.User, req DeleteRecipeRequest) (DeleteRecipeResponse, error)
}

type CreateRecipeRequest struct {
	Recipe struct {
		Name         string   `json:"name" validate:"required,min=1"`
		Tags         []string `json:"tags"`
		Ingredients  []string `json:"ingredients" validate:"required,min=1"`
		Instructions []string `json:"instructions" validate:"required,min=1"`
	} `form:"recipe" validate:"required"`
	ImageHeader *multipart.FileHeader `form:"image" validate:"required"`
	Image       ObjectStorageFile
}

type CreateRecipeResponse struct {
	Recipe domain.Recipe `json:"recipe"`
}

type GetRecipeRequest struct {
	ID string `json:"id" uri:"id" binding:"required"`
}

type GetRecipeResponse struct {
	Recipe *domain.Recipe `json:"recipe"`
}

type ListRecipesRequest struct {
	LastId string `json:"lastID" form:"lastID"`
	UserId string `json:"userID" from:"userID"`
	Sort   string `json:"sortBy" form:"sortBy,default=-createdAt"`
	Limit  int    `json:"limit" form:"limit,default=20"`
}

type SearchRecipesRequest struct {
	Name     string `json:"name" form:"name"`
	Page     int    `json:"page" form:"page" binding:"omitempty,min=1"`
	PageSize int    `json:"pageSize" form:"pageSize" binding:"omitempty,min=1"`
}

type SearchRecipesResponse struct {
	Recipes []domain.Recipe `json:"recipes"`
	Total   int             `json:"total"`
}

type ListRecipesResponse struct {
	Recipes []domain.Recipe `json:"recipes"`
	Total   int             `json:"total"`
}

type UpdateRecipeRequest struct {
	Id     string `json:"id" uri:"id" validate:"required"`
	Recipe struct {
		Name         *string  `json:"name" validate:"omitempty,min=1"`
		Tags         []string `json:"tags" validate:"omitempty,min=1"`
		Ingredients  []string `json:"ingredients" validate:"omitempty,min=1"`
		Instructions []string `json:"instructions" validate:"omitempty,min=1"`
	} `form:"recipe"`
	ImageHeader *multipart.FileHeader `form:"image"`
	Image       *ObjectStorageFile
}

type UpdateRecipeResponse struct {
	Recipe *domain.Recipe `json:"recipe"`
}

type DeleteRecipeRequest struct {
	Id string `json:"id" uri:"id" binding:"required"`
}

type DeleteRecipeResponse struct {
}
