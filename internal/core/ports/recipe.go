package ports

import (
	"context"
	"mime/multipart"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
)

type RecipesRepository interface {
	Create(ctx context.Context, recipe *domain.Recipe) error
	Get(ctx context.Context, id string) (*domain.Recipe, error)
	List(ctx context.Context, cursor, userId string, limit int64) (*domain.Page[domain.Recipe], error)
	Search(ctx context.Context, name string, page, pageSize int) ([]domain.Recipe, int, error)
	Update(ctx context.Context, recipe *domain.Recipe) error
	Delete(ctx context.Context, recipe *domain.Recipe) error
}

type RecipesService interface {
	Create(ctx context.Context, req CreateRecipeRequest) (CreateRecipeResponse, error)
	Get(ctx context.Context, req GetRecipeRequest) (GetRecipeResponse, error)
	List(ctx context.Context, req ListRecipesRequest) (ListRecipesResponse, error)
	Search(ctx context.Context, name SearchRecipesRequest) (SearchRecipesResponse, error)
	Update(ctx context.Context, req UpdateRecipeRequest) (UpdateRecipeResponse, error)
	Delete(ctx context.Context, req DeleteRecipeRequest) (DeleteRecipeResponse, error)
}

type CreateRecipeRequest struct {
	UserId    string `json:"userId" validate:"required"`
	RecipeStr string `form:"recipe"`
	Recipe    struct {
		Name         string   `json:"name" validate:"required,min=1,max=32"`
		Tags         []string `json:"tags" validate:"omitempty"`
		Ingredients  []string `json:"ingredients" validate:"required,min=1,max=128"`
		Instructions []string `json:"instructions" validate:"required,min=1,max=128"`
	} `json:"recipe" validate:"required"`
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
	Cursor string `json:"cursor" form:"cursor"`
	UserId string `json:"userId" form:"userId"`
	Limit  int    `json:"limit" form:"limit,default=20"`
}

type ListRecipesResponse struct {
	domain.Page[domain.Recipe] `json:",inline"`
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

type UpdateRecipeRequest struct {
	Id        string `json:"id" uri:"id" validate:"required"`
	UserId    string `json:"userId" validate:"required"`
	RecipeStr string `form:"recipe"`
	Recipe    struct {
		Name         *string  `json:"name" validate:"omitempty,min=1"`
		Tags         []string `json:"tags" validate:"omitempty,min=1"`
		Ingredients  []string `json:"ingredients" validate:"omitempty,min=1"`
		Instructions []string `json:"instructions" validate:"omitempty,min=1"`
	} `json:"recipe"`
	ImageHeader *multipart.FileHeader `form:"image"`
	Image       *ObjectStorageFile
}

type UpdateRecipeResponse struct {
	Recipe *domain.Recipe `json:"recipe"`
}

type DeleteRecipeRequest struct {
	Id     string `json:"id" uri:"id" validate:"required"`
	UserId string `json:"userId" validate:"required"`
}

type DeleteRecipeResponse struct {
}
