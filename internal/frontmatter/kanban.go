package frontmatter

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"wiki-go/internal/i18n"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// KanbanBoard represents a complete kanban board with multiple columns
type KanbanBoard struct {
	Title   string
	Columns []KanbanColumn
}

// KanbanColumn represents a column in the kanban board
type KanbanColumn struct {
	Title string
	Tasks []KanbanTask
}

// KanbanTask represents a task in the kanban board
type KanbanTask struct {
	Text        string
	Checked     bool
	HTMLText    string // Rendered HTML text of the task
	IndentLevel int    // Indentation level for nested tasks
}

// Store kanban boards until after goldext processing
var (
	kanbanBoards     = make(map[string]KanbanBoard)
	kanbanBoardCount = 0
	kanbanMutex      sync.Mutex
)

// KanbanSection represents a kanban board section (H5 column + tasks) - kept for backward compatibility
type KanbanSection struct {
	Title string
	Tasks []KanbanTask
}

// PreprocessorFunc defines a function that transforms markdown before rendering
type PreprocessorFunc func(markdown string, docPath string) string

// PostProcessorFunc defines a function that processes HTML after goldmark rendering
type PostProcessorFunc func(html string) string

// RenderKanban converts markdown content to a kanban board HTML with full goldext support
func RenderKanban(content string) string {
	// For now, use the basic implementation to avoid circular dependency
	// This will be enhanced when called from the main markdown processing
	return RenderKanbanBasic(content)
}

// RenderKanbanWithProcessors converts markdown content to a kanban board HTML with full goldext support
// This function accepts processors and Goldmark extensions to avoid circular dependencies.
func RenderKanbanWithProcessors(content string, preprocessors []PreprocessorFunc, postProcessors []PostProcessorFunc, extraExtensions ...goldmark.Extender) string {
	// Apply kanban-aware preprocessing to protect kanban structure while allowing goldext processing
	processedContent := kanbanAwarePreprocess(content)

	// Apply all provided preprocessors to the content
	for _, preprocessor := range preprocessors {
		if preprocessor != nil {
			processedContent = preprocessor(processedContent, "")
		}
	}

	// Render the processed content with goldmark
	renderedHTML := renderWithGoldmark(processedContent, extraExtensions...)

	// Apply post-processors
	for _, postProcessor := range postProcessors {
		if postProcessor != nil {
			renderedHTML = postProcessor(renderedHTML)
		}
	}

	// Restore kanban boards and build final kanban HTML
	return restoreKanbanBoards(renderedHTML, preprocessors, extraExtensions...)
}

// RenderKanbanBasic provides basic kanban rendering without full goldext support (fallback)
func RenderKanbanBasic(content string) string {
	// Extract header content and parse boards
	headerContent, boards := parseKanbanContentBasic(content)

	// Generate HTML for all kanban boards
	var html strings.Builder

	// Add header content if it exists
	if headerContent != "" {
		html.WriteString(headerContent)
	}

	// Render each kanban board
	for boardIndex, board := range boards {
		// Generate a unique board ID
		boardId := fmt.Sprintf("board-%d", boardIndex)
		if board.Title != "" {
			boardId = fmt.Sprintf("board-%s-%d", strings.ToLower(strings.ReplaceAll(board.Title, " ", "-")), boardIndex)
		}

		// Add kanban-container class with board ID
		html.WriteString(fmt.Sprintf(`<div class="kanban-container" data-board-id="%s">`, boardId))

		// Add board title if it exists
		if board.Title != "" {
			html.WriteString(fmt.Sprintf(`<h4 class="kanban-board-title">%s</h4>`, board.Title))
		}

		html.WriteString(`<div class="kanban-board">`)

		for _, column := range board.Columns {
			html.WriteString(fmt.Sprintf(`<div class="kanban-column">
				<div class="kanban-column-header">
					<span class="column-title">%s</span>
					<span class="kanban-status"></span>
					<button class="rename-column-btn editor-admin-only" title="%s"><i class="fa fa-pencil"></i></button>
					<button class="add-task-btn editor-admin-only" title="%s"><i class="fa fa-plus"></i></button>
					<button class="delete-column-btn editor-admin-only" title="%s"><i class="fa fa-trash"></i></button>
				</div>
				<div class="kanban-column-content">`, column.Title, i18n.Translate("kanban.rename_column"), i18n.Translate("kanban.add_task"), i18n.Translate("kanban.delete_column")))

			// Add a task list for each column
			html.WriteString(`<ul class="task-list">`)

			for _, task := range column.Tasks {
				checkedAttr := ""
				if task.Checked {
					checkedAttr = "checked"
				}

				// Add indentation level attribute if needed
				indentAttr := ""
				if task.IndentLevel > 0 {
					indentAttr = ` data-indent-level="` + strconv.Itoa(task.IndentLevel) + `"`
				}

				// Use the same structure as the regular task list items
				// This ensures compatibility with tasklist-live.js
				html.WriteString(fmt.Sprintf(`<li class="task-list-item-container" style="list-style-type: none;"%s>
					<span class="task-list-item">
						<input type="checkbox" class="task-checkbox" %s disabled>
						<span class="task-text">%s</span>
						<span class="save-state"></span>
					</span>
				</li>`, indentAttr, checkedAttr, task.HTMLText))
			}

			html.WriteString(`</ul></div></div>`)
		}

		html.WriteString(`</div>`) // Close kanban-board

		// Add the "Add Column" button for each board
		html.WriteString(fmt.Sprintf(`<div class="add-column-container">
			<button class="add-column-btn editor-admin-only" title="%s">
				<i class="fa fa-plus"></i> %s
			</button>
		</div>`, i18n.Translate("kanban.add_column_title"), i18n.Translate("kanban.add_column_title")))

		html.WriteString(`</div>`) // Close kanban-container
	}

	return html.String()
}

// kanbanAwarePreprocess protects kanban structure while allowing goldext processing of other content
func kanbanAwarePreprocess(content string) string {
	kanbanMutex.Lock()
	defer kanbanMutex.Unlock()

	// Reset storage
	kanbanBoards = make(map[string]KanbanBoard)
	kanbanBoardCount = 0

	lines := strings.Split(content, "\n")
	var result []string

	// Regular expressions - H4 for board titles, H5 for column titles
	h4Regex := regexp.MustCompile(`^#{4}\s+(.+)$`)
	h5Regex := regexp.MustCompile(`^#{5}\s+(.+)$`)
	taskRegex := regexp.MustCompile(`^\s*[-*+]\s+\[([ xX])\]\s+(.+)$`)

	// State tracking
	inKanbanBoard := false
	inKanbanColumn := false
	currentBoard := KanbanBoard{}
	currentColumn := KanbanColumn{}
	nonKanbanLines := []string{}

	for _, line := range lines {
		// Check for H4 heading (kanban board title)
		if h4Match := h4Regex.FindStringSubmatch(line); h4Match != nil {
			// Save any accumulated non-kanban content
			if len(nonKanbanLines) > 0 {
				result = append(result, nonKanbanLines...)
				nonKanbanLines = []string{}
			}

			// End previous kanban column and board if any
			if inKanbanColumn {
				currentBoard.Columns = append(currentBoard.Columns, currentColumn)
				inKanbanColumn = false
			}
			if inKanbanBoard {
				saveKanbanBoard(currentBoard, &result)
			}

			// Start new kanban board
			inKanbanBoard = true
			currentBoard = KanbanBoard{
				Title:   h4Match[1],
				Columns: []KanbanColumn{},
			}
			continue
		}

		// Check for H5 heading (kanban column) - only when in a kanban board
		if inKanbanBoard {
			if h5Match := h5Regex.FindStringSubmatch(line); h5Match != nil {
				// End previous kanban column if any
				if inKanbanColumn {
					currentBoard.Columns = append(currentBoard.Columns, currentColumn)
				}

				// Start new kanban column
				inKanbanColumn = true
				currentColumn = KanbanColumn{
					Title: h5Match[1],
					Tasks: []KanbanTask{},
				}
				continue
			}
		}

		if inKanbanBoard && inKanbanColumn {
			// Check if this is a task line
			trimmedLine := strings.TrimSpace(line)
			indent := line[:len(line)-len(trimmedLine)]

			// Calculate indentation level
			indentLevel := len(indent) / 2
			if indentLevel == 0 && len(indent) > 0 {
				indentLevel = 1 // Handle tab indentation
			}

			if taskMatch := taskRegex.FindStringSubmatch(trimmedLine); taskMatch != nil {
				// This is a task line - add to current column
				isChecked := taskMatch[1] == "x" || taskMatch[1] == "X"
				taskText := taskMatch[2]

				currentColumn.Tasks = append(currentColumn.Tasks, KanbanTask{
					Text:        taskText,
					Checked:     isChecked,
					HTMLText:    taskText, // Use raw text for basic fallback
					IndentLevel: indentLevel,
				})
				continue
			} else if trimmedLine == "" {
				// Empty line in kanban column - continue
				continue
			} else {
				// Non-task line in kanban column - end the column and board, treat as regular content
				if inKanbanColumn {
					currentBoard.Columns = append(currentBoard.Columns, currentColumn)
					inKanbanColumn = false
				}
				if inKanbanBoard {
					saveKanbanBoard(currentBoard, &result)
					inKanbanBoard = false
				}
				nonKanbanLines = append(nonKanbanLines, line)
			}
		} else if inKanbanBoard && !inKanbanColumn {
			// In kanban board but not in column - check if this is a non-heading line
			trimmedLine := strings.TrimSpace(line)
			if trimmedLine != "" && !h5Regex.MatchString(line) {
				// Non-H5 line in kanban board - end the board and treat as regular content
				if inKanbanBoard {
					saveKanbanBoard(currentBoard, &result)
					inKanbanBoard = false
				}
				nonKanbanLines = append(nonKanbanLines, line)
			} else if trimmedLine == "" {
				// Empty line in kanban board - continue
				continue
			}
		} else {
			// Regular content line
			nonKanbanLines = append(nonKanbanLines, line)
		}
	}

	// Handle remaining content
	if inKanbanColumn {
		currentBoard.Columns = append(currentBoard.Columns, currentColumn)
	}
	if inKanbanBoard {
		saveKanbanBoard(currentBoard, &result)
	}
	if len(nonKanbanLines) > 0 {
		result = append(result, nonKanbanLines...)
	}

	return strings.Join(result, "\n")
}

// saveKanbanBoard saves a kanban board and adds a placeholder to the result
func saveKanbanBoard(board KanbanBoard, result *[]string) {
	kanbanBoardCount++
	id := fmt.Sprintf("KANBAN_BOARD_%d", kanbanBoardCount)
	kanbanBoards[id] = board

	// Add placeholder that won't be processed by goldext
	placeholder := fmt.Sprintf("<!-- %s -->", id)
	*result = append(*result, placeholder)
}

// renderWithGoldmark renders the processed content using goldmark
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

	// Configure Goldmark with all needed extensions (same as regular markdown processing)
	markdown := goldmark.New(
		goldmark.WithExtensions(extensions...),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
			html.WithHardWraps(),
		),
	)

	var buf bytes.Buffer
	if err := markdown.Convert([]byte(content), &buf); err != nil {
		return "<p>Error rendering markdown: " + err.Error() + "</p>"
	}

	return buf.String()
}

// restoreKanbanBoards replaces placeholders with kanban HTML and builds the final result
func restoreKanbanBoards(htmlContent string, preprocessors []PreprocessorFunc, extraExtensions ...goldmark.Extender) string {
	kanbanMutex.Lock()
	defer kanbanMutex.Unlock()

	// Process the HTML content to find placeholders and build kanban structure
	lines := strings.Split(htmlContent, "\n")
	var finalHTML strings.Builder
	boardIndex := 0

	for _, line := range lines {
		// Check if this line contains a kanban board placeholder
		if strings.Contains(line, "<!-- KANBAN_BOARD_") {
			// Extract the placeholder ID
			re := regexp.MustCompile(`<!-- (KANBAN_BOARD_\d+) -->`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				id := matches[1]
				if board, exists := kanbanBoards[id]; exists {
					// Generate a unique board ID
					boardId := fmt.Sprintf("board-%d", boardIndex)
					if board.Title != "" {
						boardId = fmt.Sprintf("board-%s-%d", strings.ToLower(strings.ReplaceAll(board.Title, " ", "-")), boardIndex)
					}
					boardIndex++

					// Process task text with full goldext support for all columns
					var processedColumns []KanbanColumn
					for _, column := range board.Columns {
						var processedTasks []KanbanTask
						for _, task := range column.Tasks {
							processedHTML := applyProcessorsToTaskText(task.Text, preprocessors, extraExtensions...)
							processedTasks = append(processedTasks, KanbanTask{
								Text:        task.Text,
								Checked:     task.Checked,
								HTMLText:    processedHTML,
								IndentLevel: task.IndentLevel,
							})
						}
						processedColumns = append(processedColumns, KanbanColumn{
							Title: column.Title,
							Tasks: processedTasks,
						})
					}

					// Build kanban board HTML with board ID
					finalHTML.WriteString(fmt.Sprintf(`<div class="kanban-container" data-board-id="%s">`, boardId))

					// Add board title if it exists
					if board.Title != "" {
						finalHTML.WriteString(fmt.Sprintf(`<h4 class="kanban-board-title">%s</h4>`, board.Title))
					}

					finalHTML.WriteString(`<div class="kanban-board">`)

					for _, column := range processedColumns {
						finalHTML.WriteString(fmt.Sprintf(`<div class="kanban-column">
							<div class="kanban-column-header">
								<span class="column-title">%s</span>
								<span class="kanban-status"></span>
								<button class="rename-column-btn editor-admin-only" title="%s"><i class="fa fa-pencil"></i></button>
								<button class="add-task-btn editor-admin-only" title="%s"><i class="fa fa-plus"></i></button>
								<button class="delete-column-btn editor-admin-only" title="%s"><i class="fa fa-trash"></i></button>
							</div>
							<div class="kanban-column-content">`, column.Title, i18n.Translate("kanban.rename_column"), i18n.Translate("kanban.add_task"), i18n.Translate("kanban.delete_column_title")))

						finalHTML.WriteString(`<ul class="task-list">`)

						for _, task := range column.Tasks {
							checkedAttr := ""
							if task.Checked {
								checkedAttr = "checked"
							}

							indentAttr := ""
							if task.IndentLevel > 0 {
								indentAttr = ` data-indent-level="` + strconv.Itoa(task.IndentLevel) + `"`
							}

							finalHTML.WriteString(fmt.Sprintf(`<li class="task-list-item-container" style="list-style-type: none;"%s>
								<span class="task-list-item">
									<input type="checkbox" class="task-checkbox" %s disabled>
									<span class="task-text">%s</span>
									<span class="save-state"></span>
								</span>
							</li>`, indentAttr, checkedAttr, task.HTMLText))
						}

						finalHTML.WriteString(`</ul></div></div>`)
					}

					finalHTML.WriteString(`</div>`) // Close kanban-board

					// Add the "Add Column" button for each board
					finalHTML.WriteString(fmt.Sprintf(`<div class="add-column-container">
						<button class="add-column-btn editor-admin-only" title="%s">
							<i class="fa fa-plus"></i> %s
						</button>
					</div>`, i18n.Translate("kanban.add_column_title"), i18n.Translate("kanban.add_column_title")))

					finalHTML.WriteString(`</div>`) // Close kanban-container
				}
			}
		} else {
			// Regular content line
			finalHTML.WriteString(line)
			finalHTML.WriteString("\n")
		}
	}

	return finalHTML.String()
}

// applyProcessorsToTaskText applies preprocessors to individual task text
func applyProcessorsToTaskText(taskText string, preprocessors []PreprocessorFunc, extraExtensions ...goldmark.Extender) string {
	// Apply preprocessors to task text
	processed := taskText
	for _, preprocessor := range preprocessors {
		if preprocessor != nil {
			processed = preprocessor(processed, "")
		}
	}

	// Render with goldmark for inline processing
	extensions := []goldmark.Extender{
		extension.Strikethrough,
		extension.Linkify,
		extension.GFM,
	}
	extensions = append(extensions, extraExtensions...)
	markdown := goldmark.New(
		goldmark.WithExtensions(extensions...),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	if err := markdown.Convert([]byte(processed), &buf); err != nil {
		return taskText // Fallback to original text
	}

	// Remove wrapping <p> tags for inline content
	result := buf.String()
	result = strings.TrimPrefix(result, "<p>")
	result = strings.TrimSuffix(result, "</p>\n")
	result = strings.TrimSuffix(result, "</p>")

	return result
}

// parseKanbanContentBasic extracts header content and parses kanban boards (basic implementation)
func parseKanbanContentBasic(content string) (string, []KanbanBoard) {
	// Split content by lines
	lines := strings.Split(content, "\n")

	// Regular expressions for headings and tasks - H4 for boards, H5 for columns
	h4Regex := regexp.MustCompile(`^#{4}\s+(.+)$`)
	h5Regex := regexp.MustCompile(`^#{5}\s+(.+)$`)
	taskRegex := regexp.MustCompile(`^\s*[-*+]\s+\[([ xX])\]\s+(.+)$`)

	var boards []KanbanBoard
	var currentBoard *KanbanBoard
	var currentColumn *KanbanColumn
	var headerLines []string
	var foundFirstBoard bool

	for _, line := range lines {
		// Check if this line is a H4 heading (board title)
		if h4Match := h4Regex.FindStringSubmatch(line); h4Match != nil {
			// We found a board heading
			foundFirstBoard = true

			// Save current column to current board if exists
			if currentColumn != nil && currentBoard != nil {
				currentBoard.Columns = append(currentBoard.Columns, *currentColumn)
				currentColumn = nil
			}

			// Save current board if exists
			if currentBoard != nil {
				boards = append(boards, *currentBoard)
			}

			// Start new board
			currentBoard = &KanbanBoard{
				Title:   h4Match[1],
				Columns: []KanbanColumn{},
			}
		} else if foundFirstBoard && currentBoard != nil {
			// Check if this line is a H5 heading (column title)
			if h5Match := h5Regex.FindStringSubmatch(line); h5Match != nil {
				// Save current column to current board if exists
				if currentColumn != nil {
					currentBoard.Columns = append(currentBoard.Columns, *currentColumn)
				}

				// Start new column
				currentColumn = &KanbanColumn{
					Title: h5Match[1],
					Tasks: []KanbanTask{},
				}
			} else if currentColumn != nil {
				// Check if this is a task line
				trimmedLine := strings.TrimSpace(line)
				indent := line[:len(line)-len(trimmedLine)]

				// Calculate indentation level
				indentLevel := len(indent) / 2
				if indentLevel == 0 && len(indent) > 0 {
					indentLevel = 1 // Handle tab indentation
				}

				if taskMatch := taskRegex.FindStringSubmatch(trimmedLine); taskMatch != nil {
					// This is a task line
					isChecked := taskMatch[1] == "x" || taskMatch[1] == "X"
					taskText := taskMatch[2]

					// Process task text for basic markdown formatting
					htmlText := processInlineFormattingBasic(taskText)

					// Add the task to the current column
					currentColumn.Tasks = append(currentColumn.Tasks, KanbanTask{
						Text:        taskText,
						Checked:     isChecked,
						HTMLText:    htmlText,
						IndentLevel: indentLevel,
					})
				}
			}
		} else if !foundFirstBoard {
			// Collect content before the first board heading
			headerLines = append(headerLines, line)
		}
	}

	// Add the last column and board if they exist
	if currentColumn != nil && currentBoard != nil {
		currentBoard.Columns = append(currentBoard.Columns, *currentColumn)
	}
	if currentBoard != nil {
		boards = append(boards, *currentBoard)
	}

	// Process header content
	headerContent := processHeaderContentBasic(strings.Join(headerLines, "\n"))

	return headerContent, boards
}

// processHeaderContentBasic converts header markdown to HTML (basic implementation)
func processHeaderContentBasic(content string) string {
	if content == "" {
		return ""
	}

	// Process basic markdown for the header
	// Process h1 (# Title)
	h1Regex := regexp.MustCompile(`(?m)^#\s+(.+)$`)
	content = h1Regex.ReplaceAllString(content, "<h1>$1</h1>")

	// Process h2 (## Title)
	h2Regex := regexp.MustCompile(`(?m)^##\s+(.+)$`)
	content = h2Regex.ReplaceAllString(content, "<h2>$1</h2>")

	// Process h3 (### Title)
	h3Regex := regexp.MustCompile(`(?m)^###\s+(.+)$`)
	content = h3Regex.ReplaceAllString(content, "<h3>$1</h3>")

	// Process paragraphs (simple approach - treat each line as a paragraph)
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "<h") {
			// Process basic inline formatting
			line = processInlineFormattingBasic(line)
			result = append(result, "<p>"+line+"</p>")
		} else if line != "" {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// processInlineFormattingBasic handles basic markdown formatting for inline text
func processInlineFormattingBasic(text string) string {
	// Process strikethrough text (~~text~~)
	strikePattern := regexp.MustCompile(`~~([^~]+)~~`)
	text = strikePattern.ReplaceAllString(text, "<del>$1</del>")

	// Process bold text (**text** or __text__)
	boldPattern := regexp.MustCompile(`(\*\*|__)([^*_]+)(\*\*|__)`)
	text = boldPattern.ReplaceAllString(text, "<strong>$2</strong>")

	// Process italic text (*text* or _text_)
	italicPattern := regexp.MustCompile(`(\*|_)([^*_]+)(\*|_)`)
	text = italicPattern.ReplaceAllString(text, "<em>$2</em>")

	// Process links [text](url)
	linkPattern := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	text = linkPattern.ReplaceAllString(text, `<a href="$2">$1</a>`)

	// Process inline code (`code`)
	codePattern := regexp.MustCompile("`([^`]+)`")
	text = codePattern.ReplaceAllString(text, "<code>$1</code>")

	return text
}
