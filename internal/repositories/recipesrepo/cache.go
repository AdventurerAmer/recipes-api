package recipesrepo

import "fmt"

const (
	recipesVersionKey = "recipes:version"
)

func composeRecipeVersionKey(id string) string {
	return fmt.Sprintf("recipe:id:%s:version", id)
}

func composeRecipeGetKey(id string) string {
	return fmt.Sprintf("recipe:id:%s", id)
}

func composeRecipePutKey(id string, version int64) string {
	return fmt.Sprintf("recipe:id:%s:version:%d", id, version)
}

func composeRecipesGetKey(cursor, userId string, limit int64) string {
	return fmt.Sprintf("recipes:cursor:%s:userId:%s:limit:%d", cursor, userId, limit)
}

func composeRecipesPutKey(cursor, userId string, limit int64, version int64) string {
	return fmt.Sprintf("recipes:cursor:%s:userId:%s:limit:%d:version:%d", cursor, userId, limit, version)
}
