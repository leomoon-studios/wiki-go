package goldext

import (
	"bytes"
	"strings"
	"sync"
	"testing"

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
	if strings.Contains(output.String(), "count-number") {
		t.Fatal("missing predicate exposed a document count")
	}
}
