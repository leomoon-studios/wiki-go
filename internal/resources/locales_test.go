package resources

import (
	"encoding/json"
	"io/fs"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var localePlaceholderPattern = regexp.MustCompile(`\{\{\w+\}\}|\{\d+\}`)

func TestLocalesMatchEnglishSchema(t *testing.T) {
	english := readLocale(t, "en.json")
	entries, err := fs.ReadDir(languageFiles, "langs")
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") || entry.Name() == "en.json" {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			translations := readLocale(t, entry.Name())
			for key, englishText := range english {
				translation, found := translations[key]
				if !found {
					t.Errorf("missing translation %q", key)
					continue
				}
				if strings.TrimSpace(translation) == "" {
					t.Errorf("translation %q is empty", key)
				}
				if !reflect.DeepEqual(localePlaceholders(englishText), localePlaceholders(translation)) {
					t.Errorf("translation %q has placeholders %q, want %q", key, localePlaceholders(translation), localePlaceholders(englishText))
				}
			}
			for key := range translations {
				if _, found := english[key]; !found {
					t.Errorf("unexpected translation key %q", key)
				}
			}
		})
	}
}

func readLocale(t *testing.T, name string) map[string]string {
	t.Helper()
	contents, err := fs.ReadFile(languageFiles, "langs/"+name)
	if err != nil {
		t.Fatal(err)
	}
	translations := make(map[string]string)
	if err := json.Unmarshal(contents, &translations); err != nil {
		t.Fatal(err)
	}
	return translations
}

func localePlaceholders(text string) []string {
	placeholders := localePlaceholderPattern.FindAllString(text, -1)
	sort.Strings(placeholders)
	return placeholders
}
