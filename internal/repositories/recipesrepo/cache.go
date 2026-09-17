package recipesrepo

import "fmt"

const (
	recipeKeyPrefix  = "recipe:"
	recipesKeyPrefix = "recipes:"
)

func composeRecipeCacheKey(id string) string {
	return fmt.Sprintf("%s:id:%s", recipeKeyPrefix, id)
}

func composeRecipesCacheKey(userId, lastId, sort string, limit int) string {
	return fmt.Sprintf("%s:userId:%s:lastId:%s:sort:%s:limit:%d", recipesKeyPrefix, lastId, userId, sort, limit)
}
