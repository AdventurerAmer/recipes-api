package main

import (
	"log/slog"
	"os"

	v1 "github.com/AdventurerAmer/recipes-api/cmd/recipes/v1"
)

func main() {
	err := v1.Run()
	if err != nil {
		slog.Error("'v1.Run' failed", "error", err)
		os.Exit(1)
	}
}
