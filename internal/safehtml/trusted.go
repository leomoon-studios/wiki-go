// This file contains Wiki-Go's complete trust boundary for converting generated
// markup into html/template's trusted HTML type.
//
// Callers must use FromRenderer only for output from the configured safe
// Goldmark renderer. Execute and ExecuteTemplate are for html/template output,
// which has already received contextual escaping. User-controlled strings must
// never be passed to FromRenderer directly.
package safehtml

import (
	"bytes"
	"html/template"
)

// NonEmptyPlaceholder is a visually empty value used where templates need to
// distinguish an existing empty document from a missing document.
const NonEmptyPlaceholder = template.HTML(" ")

// FromRenderer marks output from Wiki-Go's configured safe renderers as trusted.
func FromRenderer(output []byte) template.HTML {
	return template.HTML(output)
}

// Execute renders an html/template with contextual escaping and returns its
// output as a value that can be embedded in another html/template.
func Execute(tmpl *template.Template, data any) (template.HTML, error) {
	var output bytes.Buffer
	if err := tmpl.Execute(&output, data); err != nil {
		return "", err
	}
	return template.HTML(output.String()), nil
}

// ExecuteTemplate is Execute for a named html/template definition.
func ExecuteTemplate(tmpl *template.Template, name string, data any) (template.HTML, error) {
	var output bytes.Buffer
	if err := tmpl.ExecuteTemplate(&output, name, data); err != nil {
		return "", err
	}
	return template.HTML(output.String()), nil
}
