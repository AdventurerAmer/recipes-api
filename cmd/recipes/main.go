package main

import (
	"os"

	recipesv1 "github.com/AdventurerAmer/recipes-api/cmd/recipes/v1"
)

func main() {
	os.Exit(recipesv1.Run())
}
