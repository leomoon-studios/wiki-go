package resources

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"testing"

	"wiki-go/internal/basepath"
	"wiki-go/internal/types"
)

func TestNavItemsArrowMarkup(t *testing.T) {
	tests := []struct {
		name         string
		alwaysOpen   bool
		parentActive bool
		childActive  bool
		contains     []string
		excludes     []string
	}{
		{
			name:       "collapsible mode uses a toggle button",
			alwaysOpen: false,
			contains: []string{
				`<a href="/parent">`,
				`<button type="button" class="nav-arrow" aria-label="Toggle Parent" aria-expanded="false"></button>`,
			},
			excludes: []string{
				`class="with-nav-arrow"`,
				`<span class="nav-arrow"`,
			},
		},
		{
			name:         "selected directory remains collapsed by default",
			parentActive: true,
			contains: []string{
				`aria-expanded="false"`,
			},
			excludes: []string{
				`nav-item directory active open`,
			},
		},
		{
			name:         "ancestor of selected directory is open by default",
			parentActive: true,
			childActive:  true,
			contains: []string{
				`nav-item directory active open`,
				`aria-expanded="true"`,
			},
		},
		{
			name:       "always-open mode puts a decorative arrow inside the link",
			alwaysOpen: true,
			contains: []string{
				`<a href="/parent" class="with-nav-arrow">`,
				`<span class="nav-arrow" aria-hidden="true"></span>`,
			},
			excludes: []string{
				`<button`,
				`aria-expanded`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := template.New("nav-items.html").Funcs(template.FuncMap{
				"dict": templateDict,
				"url":  basepath.URL,
			}).ParseFS(templateFiles, "templates/nav-items.html")
			if err != nil {
				t.Fatalf("parse nav-items template: %v", err)
			}

			tree := &types.NavTree{
				AlwaysOpen: tt.alwaysOpen,
				Root: &types.NavItem{
					Children: []*types.NavItem{
						{
							Title:    "Parent",
							Path:     "/parent",
							IsDir:    true,
							IsActive: tt.parentActive,
							Children: []*types.NavItem{
								{Title: "Child", Path: "/parent/child", IsDir: true, IsActive: tt.childActive},
							},
						},
					},
				},
			}

			var rendered bytes.Buffer
			if err := tmpl.ExecuteTemplate(&rendered, "nav-items", tree); err != nil {
				t.Fatalf("render nav-items template: %v", err)
			}

			html := rendered.String()
			for _, expected := range tt.contains {
				if !strings.Contains(html, expected) {
					t.Errorf("rendered markup does not contain %q:\n%s", expected, html)
				}
			}
			for _, unexpected := range tt.excludes {
				if strings.Contains(html, unexpected) {
					t.Errorf("rendered markup unexpectedly contains %q:\n%s", unexpected, html)
				}
			}
		})
	}
}

func templateDict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("dict requires an even number of arguments")
	}

	result := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict keys must be strings")
		}
		result[key] = values[i+1]
	}
	return result, nil
}
