package usersrepo

import "fmt"

const userKeyPrefix = "user:"

func composeUserByIdCacheKey(id string) string {
	return fmt.Sprintf("%s:id:%s", userKeyPrefix, id)
}

func composeUserByEmailCacheKey(name string) string {
	return fmt.Sprintf("%s:email:%s", userKeyPrefix, name)
}
