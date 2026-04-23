package ports

import (
	"context"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
)

type RecipesRepository interface {
	Create(ctx context.Context, recipe *domain.Recipe) error
	Get(ctx context.Context, id string) (domain.Recipe, error)
	List(ctx context.Context, lastID, sort string, limit int) ([]domain.Recipe, int, error)
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
	Name         string   `json:"name"`
	Tags         []string `json:"tags"`
	Ingredients  []string `json:"ingredients"`
	Instructions []string `json:"instructions"`
}

type CreateRecipeResponse struct {
	Recipe domain.Recipe `json:"recipe"`
}

type GetRecipeRequest struct {
	ID string `json:"id"`
}

type GetRecipeResponse struct {
	Recipe domain.Recipe `json:"recipe"`
}

type ListRecipesRequest struct {
	LastID string `json:"lastID"`
	Sort   string `json:"sort"`
	Limit  int    `json:"limit"`
}

type ListRecipesResponse struct {
	Recipes []domain.Recipe `json:"recipes"`
	Total   int             `json:"total"`
}

type UpdateRecipeRequest struct {
	ID           string   `json:"id"`
	Name         *string  `json:"name"`
	Tags         []string `json:"tags"`
	Ingredients  []string `json:"ingredients"`
	Instructions []string `json:"instructions"`
}

type UpdateRecipeResponse struct {
	Recipe domain.Recipe `json:"recipe"`
}

type DeleteRecipeRequest struct {
	ID string `json:"id"`
}

type DeleteRecipeResponse struct {
	Message string `json:"message"`
}
