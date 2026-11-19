package handlers

import (
	"html/template"

	"wiki-go/internal/basepath"
)

// urlTemplateFuncs exposes the configured base path to HTML templates.
// Template data always holds wiki-relative paths; templates add the base
// path with {{basePath}} for literal URLs or {{url .Path}} for data values.
var urlTemplateFuncs = template.FuncMap{
	"basePath": basepath.Get,
	"url":      basepath.URL,
}

// withURLFuncs returns funcs extended with the base path helpers.
func withURLFuncs(funcs template.FuncMap) template.FuncMap {
	for name, fn := range urlTemplateFuncs {
		funcs[name] = fn
	}
	return funcs
}
