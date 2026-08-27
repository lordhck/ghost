package main

import (
	"embed"
	"html/template"
)

//go:embed src/html/*.html
var htmlFS embed.FS

//go:embed src/icons/*.svg
var iconFS embed.FS

var (
	indexTmpl    = template.Must(template.ParseFS(htmlFS, "src/html/indexof.html"))
	notFoundTmpl = template.Must(template.ParseFS(htmlFS, "src/html/notfound.html"))
)

// template.HTML skips escaping, which is safe only because these are our own
// embedded files and never anything from the served directory.
func mustIcon(name string) template.HTML {
	b, err := iconFS.ReadFile("src/icons/" + name)
	if err != nil {
		panic(err)
	}
	return template.HTML(b)
}
