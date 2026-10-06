package recipesrepo

import "fmt"

const (
	recipesVersionKey = "recipes:version"
)

func composeRecipeKey(id string) string {
	return fmt.Sprintf("recipe:id:%s", id)
}

func composeRecipesKey(cursor, userId string, limit int64) string {
	return fmt.Sprintf("recipes:cursor:%s:userId:%s:limit:%d", cursor, userId, limit)
}
