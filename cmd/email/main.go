package main

import (
	"embed"
	"html/template"
	"os"

	emailv1 "github.com/AdventurerAmer/recipes-api/cmd/email/v1"
)

//go:embed templates/*.html
var templatesFS embed.FS

func main() {
	templates := template.Must(template.ParseFS(templatesFS, "templates/*.html"))
	os.Exit(emailv1.Run(templates))
}
