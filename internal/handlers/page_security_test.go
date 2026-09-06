package handlers

import (
	"strings"
	"testing"

	"wiki-go/internal/safehtml"
)

func TestDirectoryTemplatesEscapeNavigationTitlesAndPaths(t *testing.T) {
	payload := `"><img src=x onerror="alert(1)">`
	listing, err := safehtml.Execute(directoryListingTemplate, []directoryListingItem{{
		Path:  `/docs/" onmouseover="alert(2)`,
		Title: payload,
	}})
	if err != nil {
		t.Fatal(err)
	}
	heading, err := safehtml.Execute(directoryTitleTemplate, payload)
	if err != nil {
		t.Fatal(err)
	}

	for name, output := range map[string]string{"listing": string(listing), "heading": string(heading)} {
		lower := strings.ToLower(output)
		for _, forbidden := range []string{"<img", `onerror="alert`, `onmouseover="alert`} {
			if strings.Contains(lower, forbidden) {
				t.Fatalf("%s retained active navigation payload %q: %s", name, forbidden, output)
			}
		}
		if !strings.Contains(output, "&lt;img") {
			t.Errorf("%s did not render the title as escaped text: %s", name, output)
		}
	}
}
