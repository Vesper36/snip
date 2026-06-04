package handlers

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed static
var staticFS embed.FS

//go:embed templates
var templateFS embed.FS

func StaticHandler() http.Handler {
	sub, _ := fs.Sub(staticFS, "static")
	return http.FileServer(http.FS(sub))
}

// Each page template gets its own template instance with the base layout.
// This prevents {{define "content"}} blocks from clobbering each other.
func parseTemplates(funcMap template.FuncMap) map[string]*template.Template {
	tmpls := make(map[string]*template.Template)

	baseData, _ := fs.ReadFile(templateFS, "templates/base.html")

	entries, _ := fs.ReadDir(templateFS, "templates")
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".html") || e.Name() == "base.html" {
			continue
		}
		pageData, _ := fs.ReadFile(templateFS, "templates/"+e.Name())
		t := template.New("").Funcs(funcMap)
		t.Parse(string(baseData))
		t.Parse(string(pageData))

		name := strings.TrimSuffix(e.Name(), ".html")
		tmpls[name] = t
	}
	return tmpls
}
