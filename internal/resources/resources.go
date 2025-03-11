package resources

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
)

//go:embed static/*
var staticFiles embed.FS

//go:embed templates/*
var templateFiles embed.FS

//go:embed langs/*.json
var languageFiles embed.FS

// GetFileSystem returns an http.FileSystem for the embedded static files
func GetFileSystem() http.FileSystem {
	fsys, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err)
	}
	return http.FS(fsys)
}

// LoadTemplates loads and parses the embedded HTML templates
func LoadTemplates(funcMap template.FuncMap) (*template.Template, error) {
	// Parse base template with function map
	return template.New("base.html").Funcs(funcMap).ParseFS(templateFiles, "templates/*.html")
}

// GetTemplatesFS returns an fs.FS for the embedded template files
func GetTemplatesFS() fs.FS {
	return templateFiles
}

// GetLanguageFS returns an fs.FS for the embedded language files
func GetLanguageFS() fs.FS {
	fsys, err := fs.Sub(languageFiles, "langs")
	if err != nil {
		panic(err)
	}
	return fsys
}
