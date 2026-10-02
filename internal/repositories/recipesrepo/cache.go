package recipesrepo

import "fmt"

const (
	recipeKeyPrefix  = "recipe:"
	recipesKeyPrefix = "recipes:"
)

func composeRecipeCacheKey(id string) string {
	return fmt.Sprintf("%s:id:%s", recipeKeyPrefix, id)
}

func composeRecipesCacheKey(cursor, userId string, limit int64) string {
	return fmt.Sprintf("%s:cursor:%s:sort:%s:limit:%d", recipesKeyPrefix, cursor, userId, limit)
}
