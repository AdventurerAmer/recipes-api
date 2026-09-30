package tokensrepo

import (
	"fmt"

	"github.com/AdventurerAmer/recipes-api/internal/core/domain"
)

const tokenKeyPrefix = "token"

func composeCacheKey(tokenType domain.TokenType, hash string) string {
	return fmt.Sprintf("%s:%s:hash:%s", tokenKeyPrefix, tokenType, hash)
}
