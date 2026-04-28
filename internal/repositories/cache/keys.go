package cache

import "fmt"

const (
	userKeyPrefix    = "user:"
	recipeKeyPrefix  = "recipe:"
	recipesKeyPrefix = "recipes:"
)

func composeUserByIDKey(id string) string {
	return fmt.Sprintf("%sid=%s", userKeyPrefix, id)
}

func composeUserByNameKey(name string) string {
	return fmt.Sprintf("%sname=%s", userKeyPrefix, name)
}

func composeRecipeKey(id string) string {
	return fmt.Sprintf("%sid=%s", recipeKeyPrefix, id)
}

func composeRecipesKey(lastID, userID, sort string, limit int) string {
	return fmt.Sprintf("%slastID=%s,userID=%s,sort=%s,limit=%d", recipesKeyPrefix, lastID, userID, sort, limit)
}
