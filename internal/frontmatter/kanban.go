package frontmatter

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"html/template"
	"regexp"
	"strings"

	"wiki-go/internal/i18n"
	"wiki-go/internal/safehtml"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldhtml "github.com/yuin/goldmark/renderer/html"
)

// KanbanBoard represents a complete kanban board with multiple columns.
type KanbanBoard struct {
	Title   string
	Columns []KanbanColumn
}

// KanbanColumn represents a column in the kanban board.
type KanbanColumn struct {
	Title string
	Tasks []KanbanTask
}

// KanbanTask represents a task in the kanban board.
type KanbanTask struct {
	Text        string
	Checked     bool
	HTMLText    string
	IndentLevel int
}

// KanbanSection is kept for backward compatibility.
type KanbanSection struct {
	Title string
	Tasks []KanbanTask
}

// PreprocessorFunc defines a function that transforms Markdown before rendering.
type PreprocessorFunc func(markdown string, docPath string) string

// PostProcessorFunc defines a function that processes HTML after Goldmark rendering.
type PostProcessorFunc func(html string) string

type storedKanbanBoard struct {
	Placeholder string
	Index       int
	Board       KanbanBoard
}

type kanbanTemplateData struct {
	ID                string
	Title             string
	Columns           []kanbanTemplateColumn
	RenameColumnLabel string
	AddTaskLabel      string
	DeleteColumnLabel string
	AddColumnLabel    string
}

type kanbanTemplateColumn struct {
	Title string
	Tasks []kanbanTemplateTask
}

type kanbanTemplateTask struct {
	HTML        template.HTML
	Checked     bool
	IndentLevel int
}

const kanbanBoardTemplateSource = `<div class="kanban-container" data-board-id="{{.ID}}">
{{if .Title}}<h4 class="kanban-board-title">{{.Title}}</h4>{{end}}
<div class="kanban-board">
{{range .Columns}}<div class="kanban-column">
<div class="kanban-column-header">
<span class="column-title">{{.Title}}</span>
<span class="kanban-status"></span>
<button type="button" class="rename-column-btn editor-admin-only" title="{{$.RenameColumnLabel}}"><i class="fa fa-pencil"></i></button>
<button type="button" class="add-task-btn editor-admin-only" title="{{$.AddTaskLabel}}"><i class="fa fa-plus"></i></button>
<button type="button" class="delete-column-btn editor-admin-only" title="{{$.DeleteColumnLabel}}"><i class="fa fa-trash"></i></button>
</div>
<div class="kanban-column-content"><ul class="task-list">
{{range .Tasks}}<li class="task-list-item-container"{{if gt .IndentLevel 0}} data-indent-level="{{.IndentLevel}}"{{end}}>
<span class="task-list-item">
<input type="checkbox" class="task-checkbox"{{if .Checked}} checked{{end}} disabled>
<span class="task-text">{{.HTML}}</span>
<span class="save-state"></span>
</span>
</li>{{end}}
</ul></div>
</div>{{end}}
</div>
<div class="add-column-container">
<button type="button" class="add-column-btn editor-admin-only" title="{{.AddColumnLabel}}"><i class="fa fa-plus"></i> {{.AddColumnLabel}}</button>
</div>
</div>`

var kanbanBoardTemplate = template.Must(template.New("kanban-board").Parse(kanbanBoardTemplateSource))

// RenderKanban converts Markdown content to kanban HTML.
func RenderKanban(content string) string {
	return RenderKanbanBasic(content)
}

// RenderKanbanWithProcessors converts Markdown content to kanban HTML with the
// supplied Wiki-Go preprocessors and trusted Goldmark extensions.
func RenderKanbanWithProcessors(content string, preprocessors []PreprocessorFunc, postProcessors []PostProcessorFunc, extraExtensions ...goldmark.Extender) string {
	processedContent, boards := kanbanAwarePreprocess(content)

	for _, preprocessor := range preprocessors {
		if preprocessor != nil {
			processedContent = preprocessor(processedContent, "")
		}
	}

	renderedHTML := renderWithGoldmark(processedContent, extraExtensions...)
	for _, postProcessor := range postProcessors {
		if postProcessor != nil {
			renderedHTML = postProcessor(renderedHTML)
		}
	}

	return restoreKanbanBoards(renderedHTML, boards, preprocessors, extraExtensions...)
}

// RenderKanbanBasic provides the safe built-in kanban renderer without custom
// Wiki-Go processors.
func RenderKanbanBasic(content string) string {
	return RenderKanbanWithProcessors(content, nil, nil)
}

// kanbanAwarePreprocess extracts boards into request-local storage and leaves
// plain-text placeholders that survive safe Goldmark rendering.
func kanbanAwarePreprocess(content string) (string, []storedKanbanBoard) {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))
	boards := make([]storedKanbanBoard, 0)
	contentDigest := sha256.Sum256([]byte(content))
	placeholderPrefix := fmt.Sprintf("WIKIGOKANBAN%x", contentDigest[:8])

	h4Regex := regexp.MustCompile(`^#{4}\s+(.+)$`)
	h5Regex := regexp.MustCompile(`^#{5}\s+(.+)$`)
	taskRegex := regexp.MustCompile(`^\s*[-*+]\s+\[([ xX])\]\s+(.+)$`)

	inKanbanBoard := false
	inKanbanColumn := false
	currentBoard := KanbanBoard{}
	currentColumn := KanbanColumn{}
	nonKanbanLines := make([]string, 0)

	flushNonKanban := func() {
		result = append(result, nonKanbanLines...)
		nonKanbanLines = nonKanbanLines[:0]
	}
	flushBoard := func() {
		if inKanbanColumn {
			currentBoard.Columns = append(currentBoard.Columns, currentColumn)
			inKanbanColumn = false
		}
		if !inKanbanBoard {
			return
		}
		index := len(boards)
		placeholder := fmt.Sprintf("%s%08d", placeholderPrefix, index)
		boards = append(boards, storedKanbanBoard{
			Placeholder: placeholder,
			Index:       index,
			Board:       currentBoard,
		})
		result = append(result, placeholder)
		inKanbanBoard = false
	}

	for _, line := range lines {
		if h4Match := h4Regex.FindStringSubmatch(line); h4Match != nil {
			flushNonKanban()
			flushBoard()
			inKanbanBoard = true
			currentBoard = KanbanBoard{Title: h4Match[1]}
			continue
		}

		if inKanbanBoard {
			if h5Match := h5Regex.FindStringSubmatch(line); h5Match != nil {
				if inKanbanColumn {
					currentBoard.Columns = append(currentBoard.Columns, currentColumn)
				}
				inKanbanColumn = true
				currentColumn = KanbanColumn{Title: h5Match[1]}
				continue
			}
		}

		if inKanbanBoard && inKanbanColumn {
			trimmedLine := strings.TrimSpace(line)
			indent := line[:len(line)-len(trimmedLine)]
			indentLevel := len(indent) / 2
			if indentLevel == 0 && len(indent) > 0 {
				indentLevel = 1
			}

			if taskMatch := taskRegex.FindStringSubmatch(trimmedLine); taskMatch != nil {
				currentColumn.Tasks = append(currentColumn.Tasks, KanbanTask{
					Text:        taskMatch[2],
					Checked:     taskMatch[1] == "x" || taskMatch[1] == "X",
					IndentLevel: indentLevel,
				})
				continue
			}
			if trimmedLine == "" {
				continue
			}

			flushBoard()
			nonKanbanLines = append(nonKanbanLines, line)
			continue
		}

		if inKanbanBoard {
			if strings.TrimSpace(line) == "" {
				continue
			}
			flushBoard()
		}
		nonKanbanLines = append(nonKanbanLines, line)
	}

	flushBoard()
	flushNonKanban()
	return strings.Join(result, "\n"), boards
}

// renderWithGoldmark renders document Markdown while deliberately leaving raw
// HTML disabled. Trusted Wiki-Go node renderers are supplied as extensions.
func renderWithGoldmark(content string, extraExtensions ...goldmark.Extender) string {
	extensions := []goldmark.Extender{
		extension.Table,
		extension.Strikethrough,
		extension.Linkify,
		extension.Footnote,
		extension.DefinitionList,
		extension.GFM,
	}
	extensions = append(extensions, extraExtensions...)

	markdown := goldmark.New(
		goldmark.WithExtensions(extensions...),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(),
		),
		goldmark.WithRendererOptions(goldhtml.WithHardWraps()),
	)

	var buf bytes.Buffer
	if err := markdown.Convert([]byte(content), &buf); err != nil {
		return "<p>Error rendering Markdown: " + template.HTMLEscapeString(err.Error()) + "</p>"
	}
	return buf.String()
}

func restoreKanbanBoards(htmlContent string, boards []storedKanbanBoard, preprocessors []PreprocessorFunc, extraExtensions ...goldmark.Extender) string {
	for _, stored := range boards {
		boardHTML := renderKanbanBoard(stored.Board, stored.Index, preprocessors, extraExtensions...)
		paragraphPlaceholder := "<p>" + stored.Placeholder + "</p>"
		htmlContent = strings.Replace(htmlContent, paragraphPlaceholder, boardHTML, 1)
	}
	return htmlContent
}

func renderKanbanBoard(board KanbanBoard, boardIndex int, preprocessors []PreprocessorFunc, extraExtensions ...goldmark.Extender) string {
	columns := make([]kanbanTemplateColumn, 0, len(board.Columns))
	for _, column := range board.Columns {
		tasks := make([]kanbanTemplateTask, 0, len(column.Tasks))
		for _, task := range column.Tasks {
			tasks = append(tasks, kanbanTemplateTask{
				HTML:        safehtml.FromRenderer([]byte(applyProcessorsToTaskText(task.Text, preprocessors, extraExtensions...))),
				Checked:     task.Checked,
				IndentLevel: task.IndentLevel,
			})
		}
		columns = append(columns, kanbanTemplateColumn{Title: column.Title, Tasks: tasks})
	}

	data := kanbanTemplateData{
		ID:                normalizeBoardID(board.Title, boardIndex),
		Title:             board.Title,
		Columns:           columns,
		RenameColumnLabel: i18n.Translate("kanban.rename_column"),
		AddTaskLabel:      i18n.Translate("kanban.add_task"),
		DeleteColumnLabel: i18n.Translate("kanban.delete_column_title"),
		AddColumnLabel:    i18n.Translate("kanban.add_column_title"),
	}

	var buf bytes.Buffer
	if err := kanbanBoardTemplate.Execute(&buf, data); err != nil {
		return "<p>Error rendering kanban board: " + template.HTMLEscapeString(err.Error()) + "</p>"
	}
	return buf.String()
}

func normalizeBoardID(title string, boardIndex int) string {
	const maxSlugLength = 64

	var slug strings.Builder
	lastWasDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(title)) {
		if slug.Len() >= maxSlugLength {
			break
		}
		isASCIIAlphaNumeric := r >= 'a' && r <= 'z' || r >= '0' && r <= '9'
		if isASCIIAlphaNumeric {
			slug.WriteRune(r)
			lastWasDash = false
			continue
		}
		if slug.Len() > 0 && !lastWasDash {
			slug.WriteByte('-')
			lastWasDash = true
		}
	}

	normalized := strings.Trim(slug.String(), "-")
	if normalized == "" {
		return fmt.Sprintf("board-%d", boardIndex)
	}
	return fmt.Sprintf("board-%s-%d", normalized, boardIndex)
}

// applyProcessorsToTaskText renders task Markdown with raw HTML disabled.
func applyProcessorsToTaskText(taskText string, preprocessors []PreprocessorFunc, extraExtensions ...goldmark.Extender) string {
	processed := taskText
	for _, preprocessor := range preprocessors {
		if preprocessor != nil {
			processed = preprocessor(processed, "")
		}
	}

	extensions := []goldmark.Extender{
		extension.Strikethrough,
		extension.Linkify,
		extension.GFM,
	}
	extensions = append(extensions, extraExtensions...)
	markdown := goldmark.New(goldmark.WithExtensions(extensions...))

	var buf bytes.Buffer
	if err := markdown.Convert([]byte(processed), &buf); err != nil {
		return template.HTMLEscapeString(taskText)
	}

	result := buf.String()
	result = strings.TrimPrefix(result, "<p>")
	result = strings.TrimSuffix(result, "</p>\n")
	result = strings.TrimSuffix(result, "</p>")
	return result
}
