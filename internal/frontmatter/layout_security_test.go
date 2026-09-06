package frontmatter

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRenderKanbanEscapesDynamicContentAndDisablesRawHTML(t *testing.T) {
	input := `# Header <img src=x onerror=alert(1)>

#### Board " onmouseover="alert(2) <script>alert(3)</script>
##### Column </span><svg onload=alert(4)>
- [ ] **formatted** <svg onload=alert(5)> [unsafe](javascript:alert(6))
  - [x] [safe](https://example.com/docs?q=wiki)
`

	got := RenderKanbanBasic(input)
	lower := strings.ToLower(got)
	for _, forbidden := range []string{
		"<script", "<svg", "<img", `onerror="`, `onload="`, `onmouseover="`,
		`href="javascript:`, `style="`, `onclick="`, `onchange="`, `oninput="`,
	} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("kanban output contains unsafe HTML %q:\n%s", forbidden, got)
		}
	}

	for _, expected := range []string{
		`&lt;script&gt;alert(3)&lt;/script&gt;`,
		`&lt;/span&gt;&lt;svg onload=alert(4)&gt;`,
		`<strong>formatted</strong>`,
		`href="https://example.com/docs?q=wiki"`,
		`data-indent-level="1"`,
	} {
		if !strings.Contains(got, expected) {
			t.Errorf("kanban output is missing %q:\n%s", expected, got)
		}
	}

	idPattern := regexp.MustCompile(`data-board-id="([^"]+)"`)
	match := idPattern.FindStringSubmatch(got)
	if len(match) != 2 || !regexp.MustCompile(`^board-[a-z0-9-]+-0$`).MatchString(match[1]) {
		t.Fatalf("kanban board ID was not normalized: %q", got)
	}
}

func TestKanbanTemplateEscapesTranslatedLabels(t *testing.T) {
	data := kanbanTemplateData{
		ID:                "board-safe-0",
		RenameColumnLabel: `rename" onmouseover="alert(1)`,
		AddTaskLabel:      `<svg onload=alert(2)>`,
		DeleteColumnLabel: `</button><script>alert(3)</script>`,
		AddColumnLabel:    `<img src=x onerror=alert(4)>`,
	}

	var output strings.Builder
	if err := kanbanBoardTemplate.Execute(&output, data); err != nil {
		t.Fatalf("execute kanban template: %v", err)
	}
	got := output.String()
	lower := strings.ToLower(got)
	for _, forbidden := range []string{"<script", "<svg", "<img", `onmouseover="`, `onload="`, `onerror="`} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("translated kanban label escaped its attribute context via %q: %s", forbidden, got)
		}
	}
}

func TestNormalizeBoardIDUsesOnlyBoundedSafeCharacters(t *testing.T) {
	got := normalizeBoardID(`  Release / \ " onclick=alert(1) `+strings.Repeat("x", 100), 7)
	if !regexp.MustCompile(`^board-[a-z0-9-]{1,64}-7$`).MatchString(got) {
		t.Fatalf("normalizeBoardID returned an unsafe or unbounded ID: %q", got)
	}
}

func TestRenderLinksEscapesFieldsRejectsDangerousURLsAndHasNoInlineControls(t *testing.T) {
	input := `# Links <img src=x onerror=alert(1)>

## Category </option><svg onload=alert(2)>
- [Title <img src=x onerror=alert(3)>](https://example.com/path?q=%22%20onmouseover%3Dalert%284%29) - Description <script>alert(5)</script> | 2026-08-31
- [Rejected](javascript:alert(6)) - must not render | 2026-08-31
`

	got, err := RenderLinks(input)
	if err != nil {
		t.Fatalf("RenderLinks returned an error: %v", err)
	}
	lower := strings.ToLower(got)
	for _, forbidden := range []string{
		"<script", "<svg", "<img src=x", `onerror="`, `onload="`, `onmouseover="`,
		`href="javascript:`, `style="`, `onclick="`, `onchange="`, `oninput="`, "rejected",
	} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("links output contains unsafe or rejected content %q:\n%s", forbidden, got)
		}
	}

	for _, expected := range []string{
		`Links &lt;img src=x onerror=alert(1)&gt;`,
		`Category &lt;/option&gt;&lt;svg onload=alert(2)&gt;`,
		`Title &lt;img src=x onerror=alert(3)&gt;`,
		`Description &lt;script&gt;alert(5)&lt;/script&gt;`,
		`id="searchClear" title="Clear search" hidden`,
		`id="searchResultsInfo" hidden`,
		`class="floating-add-link-btn"`,
	} {
		if !strings.Contains(got, expected) {
			t.Errorf("links output is missing %q:\n%s", expected, got)
		}
	}
}

func TestLinksTemplateEscapesTranslatedLabelsAndURLContexts(t *testing.T) {
	data := NewLinksData()
	data.Title = `<svg onload=alert(1)>`
	data.Categories[`</option><img src=x onerror=alert(2)>`] = []Link{{
		Title:       `<script>alert(3)</script>`,
		URL:         `https://example.com/" onmouseover="alert(4)`,
		Description: `<img src=x onerror=alert(5)>`,
		Category:    `</div><svg onload=alert(6)>`,
		AddedAt:     time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
	}}
	data.calculateStats()
	labels := linksLabels{
		NoResultsTitle:   `<svg onload=alert(7)>`,
		NoResultsMessage: `</div><script>alert(8)</script>`,
		AddNewLink:       `add" onfocus="alert(9)`,
	}

	got, err := renderLinksTemplate(data, labels)
	if err != nil {
		t.Fatalf("render links template: %v", err)
	}
	lower := strings.ToLower(got)
	for _, forbidden := range []string{"<script", "<svg", "<img src=x", `onerror="`, `onload="`, `onmouseover="`, `onfocus="`} {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("links template escaped its output context via %q: %s", forbidden, got)
		}
	}
}

func TestLayoutRenderingIsRequestLocalUnderConcurrency(t *testing.T) {
	const requestCount = 32
	start := make(chan struct{})
	errors := make(chan error, requestCount*2)
	var wg sync.WaitGroup

	for i := 0; i < requestCount; i++ {
		marker := fmt.Sprintf("REQUEST%08d", i)
		wg.Add(2)

		go func() {
			defer wg.Done()
			<-start
			input := fmt.Sprintf("#### BOARD%s\n##### COLUMN%s\n- [ ] TASK%s\n", marker, marker, marker)
			got := RenderKanbanBasic(input)
			for _, expected := range []string{"BOARD" + marker, "COLUMN" + marker, "TASK" + marker} {
				if !strings.Contains(got, expected) {
					errors <- fmt.Errorf("kanban response for %s is missing %s: %s", marker, expected, got)
					return
				}
			}
		}()

		go func() {
			defer wg.Done()
			<-start
			input := fmt.Sprintf("# TITLE%s\n\n## CATEGORY%s\n- [LINK%s](https://example.com/%s) - DESCRIPTION%s | 2026-08-31\n", marker, marker, marker, marker, marker)
			got, err := RenderLinks(input)
			if err != nil {
				errors <- fmt.Errorf("links response for %s failed: %w", marker, err)
				return
			}
			for _, expected := range []string{"TITLE" + marker, "CATEGORY" + marker, "LINK" + marker, "DESCRIPTION" + marker} {
				if !strings.Contains(got, expected) {
					errors <- fmt.Errorf("links response for %s is missing %s: %s", marker, expected, got)
					return
				}
			}
		}()
	}

	close(start)
	wg.Wait()
	close(errors)
	for err := range errors {
		t.Error(err)
	}
}
