package recipessrv

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
)

type Config struct {
	RecipesRepo ports.RecipesRepository
	MaxLimit    int
}

type service struct {
	Config
}

func New(cfg Config) ports.RecipesService {
	return &service{
		Config: cfg,
	}
}

func (srv *service) Create(ctx context.Context, user domain.User, req ports.CreateRecipeRequest) (ports.CreateRecipeResponse, error) {
	recipe := domain.Recipe{
		CreatedAt:    time.Now().UTC(),
		UserID:       user.ID,
		Name:         req.Name,
		Tags:         req.Tags,
		Ingredients:  req.Ingredients,
		Instructions: req.Instructions,
	}
	if err := srv.RecipesRepo.Create(ctx, &recipe); err != nil {
		return ports.CreateRecipeResponse{}, fmt.Errorf("'RecipesRepo.Create' failed: %w", err)
	}
	return ports.CreateRecipeResponse{Recipe: recipe}, nil
}

func (srv *service) Get(ctx context.Context, req ports.GetRecipeRequest) (ports.GetRecipeResponse, error) {
	recipe, err := srv.RecipesRepo.Get(ctx, req.ID)
	if err != nil {
		return ports.GetRecipeResponse{}, fmt.Errorf("'RecipesRepo.Get' failed: %w", err)
	}
	return ports.GetRecipeResponse{Recipe: recipe}, nil
}

func (srv *service) List(ctx context.Context, req ports.ListRecipesRequest) (ports.ListRecipesResponse, error) {
	limit := min(req.Limit, srv.MaxLimit)
	recipes, total, err := srv.RecipesRepo.List(ctx, req.LastID, req.Sort, limit)
	if err != nil {
		return ports.ListRecipesResponse{}, fmt.Errorf("'RecipesRepo.List' failed: %w", err)
	}
	return ports.ListRecipesResponse{Recipes: recipes, Total: total}, nil
}

func (srv *service) Update(ctx context.Context, user domain.User, req ports.UpdateRecipeRequest) (ports.UpdateRecipeResponse, error) {
	recipe, err := srv.RecipesRepo.Get(ctx, req.ID)
	if err != nil {
		return ports.UpdateRecipeResponse{}, fmt.Errorf("'RecipesRepo.Get' failed: %w", err)
	}
	if recipe.UserID != user.ID {
		return ports.UpdateRecipeResponse{}, errors.New("permission denied")
	}
	if req.Name != nil {
		recipe.Name = *req.Name
	}
	if req.Tags != nil {
		recipe.Tags = req.Tags
	}
	if req.Instructions != nil {
		recipe.Instructions = req.Instructions
	}
	if req.Ingredients != nil {
		recipe.Ingredients = req.Ingredients
	}
	if err := srv.RecipesRepo.Update(ctx, &recipe); err != nil {
		return ports.UpdateRecipeResponse{}, fmt.Errorf("'RecipesRepo.Update' failed: %w", err)
	}
	return ports.UpdateRecipeResponse{Recipe: recipe}, nil
}

func (srv *service) Delete(ctx context.Context, user domain.User, req ports.DeleteRecipeRequest) (ports.DeleteRecipeResponse, error) {
	if err := srv.RecipesRepo.Delete(ctx, user.ID, req.ID); err != nil {
		return ports.DeleteRecipeResponse{}, fmt.Errorf("'RecipesRepo.Delete' failed: %w", err)
	}
	return ports.DeleteRecipeResponse{
		Message: "Recipe was deleted successfully",
	}, nil
}
