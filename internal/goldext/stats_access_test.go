package goldext

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yuin/goldmark"
)

func TestStatsRenderAccessIsolationAndFailClosed(t *testing.T) {
	const workers = 20
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(allow bool) {
			defer wg.Done()
			predicate := DocumentAccessPredicate(func(string) bool { return allow })
			extension := TrustedNodesForDocumentWithAccess("test", predicate).(*trustedNodesExtension)
			if got := extension.canAccess("/finance"); got != allow {
				t.Errorf("render predicate returned %t, want %t", got, allow)
			}
			var output bytes.Buffer
			if err := goldmark.New(goldmark.WithExtensions(extension)).Convert([]byte(":::stats recent=100:::"), &output); err != nil {
				t.Error(err)
			}
			if strings.Contains(output.String(), "doc-path") {
				t.Error("stats exposed unfiltered document metadata")
			}
		}(i%2 == 0)
	}
	wg.Wait()
	var output bytes.Buffer
	if err := goldmark.New(goldmark.WithExtensions(TrustedNodesForDocument("test"))).Convert([]byte(":::stats count=*:::"), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `class="count-number">0</div>`) {
		t.Fatal("missing predicate exposed a document count")
	}
}

func TestStatsFilterBeforeLimitingAndCounting(t *testing.T) {
	root := t.TempDir()
	for _, doc := range []struct {
		path string
		when time.Time
	}{
		{"public/older", time.Unix(1000, 0)},
		{"finance/secret", time.Unix(2000, 0)},
	} {
		dir := filepath.Join(root, filepath.FromSlash(doc.path))
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		file := filepath.Join(dir, "document.md")
		if err := os.WriteFile(file, []byte("# "+doc.path), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(file, doc.when, doc.when); err != nil {
			t.Fatal(err)
		}
	}
	allowed := func(path string) bool { return !strings.HasPrefix(path, "/finance") }
	docs := getRecentDocuments(root, 1, allowed)
	if len(docs) != 1 || docs[0].Path != "public/older" {
		t.Fatalf("filtered recent documents = %+v", docs)
	}
	if got := countDocuments(root, root, allowed); got != 1 {
		t.Fatalf("global count = %d, want 1", got)
	}
	if got := countDocuments(root, filepath.Join(root, "finance"), allowed); got != 0 {
		t.Fatalf("restricted folder count = %d, want 0", got)
	}
	if got := countDocuments(root, root, nil); got != 0 {
		t.Fatalf("missing predicate count = %d, want 0", got)
	}
	if got := countDocuments(root, root, func(string) bool { return true }); got != 2 {
		t.Fatalf("administrator count = %d, want 2", got)
	}
}
