/**
 * Editor Toolbar Module
 * Handles toolbar creation, button actions, and formatting commands
 */

// Formatting action functions
function addHighlight(cm) {
    window.EditorCore.wrapText(cm, '==', '==');
}

function addSuperscript(cm) {
    window.EditorCore.wrapText(cm, '^', '^');
}

function addSubscript(cm) {
    window.EditorCore.wrapText(cm, '~', '~');
}

function addRecentEdits(cm) {
    var cursor = cm.getCursor();
    cm.replaceRange(":::stats recent=5:::\n", cursor);
    // Ensure editor retains focus
    cm.focus();
}

function addTotal(cm) {
    var cursor = cm.getCursor();
    cm.replaceRange(":::stats count=*:::\n", cursor);
    // Ensure editor retains focus
    cm.focus();
}

// Function to insert TOC shortcode
function insertTOC() {
    const editor = window.EditorCore.getEditor();
    if (!editor) return;

    // Insert [toc] shortcode at cursor position with a newline after it
    const cursor = editor.getCursor();
    editor.replaceRange("[toc]\n", cursor);

    // Ensure editor retains focus
    editor.focus();
}

// Function to insert kanban frontmatter and board template
function insertKanbanFrontmatter(cm) {
    const content = cm.getValue();
    let needsFrontmatter = false;
    let frontmatterToAdd = '';

    // Check if the document needs frontmatter
    if (!content.startsWith('---\n')) {
        needsFrontmatter = true;
        frontmatterToAdd = `---
layout: kanban
---

`;
    } else {
        // Check if existing frontmatter has kanban layout
        const frontmatterMatch = content.match(/^---\n([\s\S]*?)\n---/);
        if (frontmatterMatch) {
            const frontmatterContent = frontmatterMatch[1];
            if (!frontmatterContent.includes('layout: kanban')) {
                // Alert user that they need to manually add layout: kanban
                alert('Please add "layout: kanban" to your existing frontmatter to enable kanban functionality.');
                return;
            }
        }
    }

    // Determine the next kanban board number
    let boardNumber = 1;
    const kanbanMatches = content.match(/^#### Kanban \d+/gm);
    if (kanbanMatches && kanbanMatches.length > 0) {
        // Find the highest existing kanban number
        const numbers = kanbanMatches.map(match => {
            const num = match.match(/#### Kanban (\d+)/);
            return num ? parseInt(num[1]) : 0;
        });
        boardNumber = Math.max(...numbers) + 1;
    }

    // Create the kanban board template
    const kanbanBoardTemplate = `#### Kanban ${boardNumber}

##### To Do
- [ ] Task 1

##### In Progress

##### Done

`;

    // Get cursor position
    const cursor = cm.getCursor();

    // If we need to add frontmatter, add it at the beginning
    if (needsFrontmatter) {
        cm.replaceRange(frontmatterToAdd, {line: 0, ch: 0});

        // Adjust cursor position to account for added frontmatter
        const frontmatterLines = frontmatterToAdd.split('\n').length - 1;
        const newCursorLine = cursor.line + frontmatterLines;
        const newCursor = {line: newCursorLine, ch: cursor.ch};

        // Insert kanban board at the adjusted cursor position
        cm.replaceRange(kanbanBoardTemplate, newCursor);

        // Position cursor after the inserted content
        const templateLines = kanbanBoardTemplate.split('\n').length - 1;
        cm.setCursor({line: newCursorLine + templateLines, ch: 0});
    } else {
        // Just insert the kanban board at cursor position
        cm.replaceRange(kanbanBoardTemplate, cursor);

        // Position cursor after the inserted content
        const templateLines = kanbanBoardTemplate.split('\n').length - 1;
        cm.setCursor({line: cursor.line + templateLines, ch: 0});
    }

    // Focus the editor
    cm.focus();
}

// Create custom toolbar
function createToolbar(container) {
    const toolbar = document.createElement('div');
    toolbar.className = 'editor-toolbar custom-toolbar';

    // Simple platform detection
    const isMac = /Mac|iPhone|iPad|iPod/.test(navigator.userAgent);
    const getShortcut = (mac, other) => isMac ? mac : other;

    // Define toolbar buttons with dynamic shortcuts
    const buttons = [
        { icon: 'fa-header', action: 'heading', title: `Heading (${getShortcut('Cmd+H', 'Ctrl+H')})` },
        { icon: 'fa-bold', action: 'bold', title: `Bold (${getShortcut('Cmd+B', 'Ctrl+B')})` },
        { icon: 'fa-italic', action: 'italic', title: `Italic (${getShortcut('Cmd+I', 'Ctrl+I')})` },
        { icon: 'fa-paint-brush', action: 'highlight', title: 'Highlight Text' },
        { icon: 'fa-strikethrough', action: 'strikethrough', title: 'Strikethrough' },
        { icon: 'fa-subscript', action: 'subscript', title: 'Add Subscript' },
        { icon: 'fa-superscript', action: 'superscript', title: 'Add Superscript' },
        { type: 'separator' },
        { icon: 'fa-code', action: 'code', title: `Code (${getShortcut('Cmd+/', 'Ctrl+/')})` },
        { icon: 'fa-quote-left', action: 'quote', title: `Quote (${getShortcut('Cmd+K', 'Ctrl+K')})` },
        { icon: 'fa-list-ul', action: 'unordered-list', title: 'Unordered List' },
        { icon: 'fa-list-ol', action: 'ordered-list', title: 'Ordered List' },
        { type: 'separator' },
        { icon: 'fa-picture-o', action: 'image', title: 'Image' },
        { icon: 'fa-link', action: 'link', title: 'Link' },
        { icon: 'fa-anchor', action: 'anchor-link', title: 'Link to Heading' },
        { icon: 'fa-file-text-o', action: 'doc-link', title: 'Link to Document' },
        { icon: 'fa-smile-o', action: 'emoji', title: 'Insert Emoji' },
        { type: 'separator' },
        { icon: 'fa-th', action: 'table', title: 'Insert Custom Table', id: 'insert-table' },
        { icon: 'fa-align-center', action: 'format-table', title: 'Format Markdown Table' },
        { icon: 'fa-plus-square', action: 'insert-row', title: 'Insert Row', style: 'position: relative;' },
        { icon: 'fa-plus-square', action: 'insert-column', title: 'Insert Column', style: 'position: relative;' },
        { icon: 'fa-minus-square', action: 'delete-row', title: 'Delete Row', style: 'position: relative;' },
        { icon: 'fa-minus-square', action: 'delete-column', title: 'Delete Column', style: 'position: relative;' },
        { icon: 'fa-arrow-up', action: 'move-row-up', title: `Move Row Up (${getShortcut('Option+Up Arrow', 'Alt+Up Arrow')})`, style: 'position: relative;' },
        { icon: 'fa-arrow-down', action: 'move-row-down', title: `Move Row Down (${getShortcut('Option+Down Arrow', 'Alt+Down Arrow')})`, style: 'position: relative;' },
        { icon: 'fa-arrow-left', action: 'move-column-left', title: `Move Column Left (${getShortcut('Option+Left Arrow', 'Alt+Left Arrow')})`, style: 'position: relative;' },
        { icon: 'fa-arrow-right', action: 'move-column-right', title: `Move Column Right (${getShortcut('Option+Right Arrow', 'Alt+Right Arrow')})`, style: 'position: relative;' },
        { icon: 'fa-align-left', action: 'align-column-left', title: `Align Column Left (${getShortcut('Cmd+Shift+Left Arrow', 'Ctrl+Shift+Left Arrow')})`, style: 'position: relative;' },
        { icon: 'fa-align-center', action: 'align-column-center', title: `Align Column Center (${getShortcut('Cmd+Shift+Up Arrow', 'Ctrl+Shift+Up Arrow')})`, style: 'position: relative;' },
        { icon: 'fa-align-right', action: 'align-column-right', title: `Align Column Right (${getShortcut('Cmd+Shift+Right Arrow', 'Ctrl+Shift+Right Arrow')})`, style: 'position: relative;' },
        { icon: 'fa-align-justify', action: 'align-column-none', title: `Remove Column Alignment (${getShortcut('Cmd+Shift+Down Arrow', 'Ctrl+Shift+Down Arrow')})`, style: 'position: relative;' },
        { type: 'separator' },
        { icon: 'fa-ellipsis-h', action: 'horizontal-rule', title: 'Horizontal Rule' },
        { type: 'separator' },
        { icon: 'fa-text-width', action: 'toggle-wordwrap', title: `Toggle Word Wrap (${getShortcut('Option+Z', 'Alt+Z')})`, id: 'toggle-wordwrap' },
        { icon: 'fa-list-ol', action: 'toggle-linenumbers', title: `Show Line Numbers (${getShortcut('Option+N', 'Alt+N')})`, id: 'toggle-linenumbers' },
        { icon: 'fa-columns', action: 'toggle-split', title: `Toggle Split View (${getShortcut('Cmd+Shift+S', 'Ctrl+Shift+S')})`, id: 'toggle-split' },
        { icon: 'fa-font', action: 'toggle-autocapitalize', title: `Enable Auto-Capitalize (${getShortcut('Option+C', 'Alt+C')})`, id: 'toggle-autocapitalize' },
        { type: 'separator' },
        { icon: 'fa-list-alt', action: 'insert-toc', title: 'Insert Table of Contents' },
        { icon: 'fa-clock-o', action: 'recent-edits', title: 'Insert Recent Edits' },
        { icon: 'fa-book', action: 'total', title: 'Insert Total Number of Documents' },
        { icon: 'fa-tasks', action: 'kanban', title: 'Insert Kanban Board Template' },
        { type: 'separator' },
        { icon: 'fa-undo', action: 'undo', title: 'Undo' },
        { icon: 'fa-repeat', action: 'redo', title: 'Redo' },
        { type: 'separator' },
        { icon: 'fa-eye', action: 'preview', title: `Toggle Preview (${getShortcut('Cmd+Shift+P', 'Ctrl+Shift+P')})`, id: 'toggle-preview' }
    ];

    buttons.forEach(button => {
        if (button.type === 'separator') {
            const separator = document.createElement('i');
            separator.className = 'separator';
            toolbar.appendChild(separator);
        } else {
            const btn = document.createElement('button');
            if (button.id) {
                btn.id = button.id;
            }
            btn.className = `toolbar-button ${button.action}-button`;
            btn.dataset.action = button.action;
            btn.title = button.title;

            // Apply inline style if provided
            if (button.style) {
                btn.setAttribute('style', button.style);
            }

            const icon = document.createElement('i');
            icon.className = `fa ${button.icon}`;
            btn.appendChild(icon);

            toolbar.appendChild(btn);
        }
    });

    container.appendChild(toolbar);
    return toolbar;
}

// Add event listeners to toolbar buttons
function setupToolbarActions(toolbar) {
    const editor = window.EditorCore.getEditor();
    if (!editor || !toolbar) return;

    if (toolbar._delegateSetup) return; // Prevent double-initialisation

    // Executes the requested action.
    const exec = (action, button) => {
        switch (action) {
            case 'bold':
                window.EditorCore.wrapText(editor, '**', '**');
                break;
            case 'italic':
                window.EditorCore.wrapText(editor, '*', '*');
                break;
            case 'strikethrough':
                window.EditorCore.wrapText(editor, '~~', '~~');
                break;
            case 'highlight':
                addHighlight(editor);
                break;
            case 'subscript':
                addSubscript(editor);
                break;
            case 'superscript':
                addSuperscript(editor);
                break;
            case 'heading': {
                const cursor = editor.getCursor();
                const line = editor.getLine(cursor.line);
                const headingMatch = line.match(/^(#+)\s/);

                if (headingMatch) {
                    const prefix = headingMatch[1];
                    const contentStart = prefix.length + 1; // +1 for space
                    const content = line.substring(contentStart);

                    if (prefix.length < 6) {
                        editor.replaceRange('#' + prefix + ' ' + content,
                            { line: cursor.line, ch: 0 },
                            { line: cursor.line, ch: line.length });
                    } else {
                        editor.replaceRange(content,
                            { line: cursor.line, ch: 0 },
                            { line: cursor.line, ch: line.length });
                    }
                } else {
                    editor.replaceRange('# ' + line,
                        { line: cursor.line, ch: 0 },
                        { line: cursor.line, ch: line.length });
                }
                editor.focus();
                break;
            }
            case 'code': {
                const sel = editor.getSelection();
                if (sel) {
                    if (sel.indexOf('\n') !== -1) {
                        editor.replaceSelection("```\n" + sel + "\n```");
                    } else {
                        editor.replaceSelection("`" + sel + "`");
                    }
                } else {
                    const cursor = editor.getCursor();
                    editor.replaceRange("``", cursor);
                    editor.setCursor(cursor.line, cursor.ch + 1);
                }
                editor.focus();
                break;
            }
            case 'quote': {
                const sel = editor.getSelection();
                if (sel) {
                    const lines = sel.split('\n');
                    editor.replaceSelection(lines.map(l => `> ${l}`).join('\n'));
                } else {
                    const cursor = editor.getCursor();
                    editor.replaceRange('> ', cursor);
                }
                editor.focus();
                break;
            }
            case 'unordered-list': {
                const sel = editor.getSelection();
                if (sel) {
                    const lines = sel.split('\n');
                    editor.replaceSelection(lines.map(l => `- ${l}`).join('\n'));
                } else {
                    const cursor = editor.getCursor();
                    editor.replaceRange('- ', cursor);
                }
                editor.focus();
                break;
            }
            case 'ordered-list': {
                const sel = editor.getSelection();
                if (sel) {
                    const lines = sel.split('\n');
                    editor.replaceSelection(lines.map((l, i) => `${i + 1}. ${l}`).join('\n'));
                } else {
                    const cursor = editor.getCursor();
                    editor.replaceRange('1. ', cursor);
                }
                editor.focus();
                break;
            }
            case 'link': {
                const sel = editor.getSelection();
                if (sel) {
                    window.EditorCore.wrapText(editor, '[', '](url)');
                    const cursor = editor.getCursor();
                    editor.setCursor(cursor.line, cursor.ch - 1);
                } else {
                    const start = editor.getCursor();
                    window.EditorCore.wrapText(editor, '[', '](url)', 'text');
                    editor.setSelection({ line: start.line, ch: start.ch + 1 },
                                       { line: start.line, ch: start.ch + 5 });
                }
                break;
            }
            case 'image': {
                const sel = editor.getSelection();
                if (sel) {
                    window.EditorCore.wrapText(editor, '![' , '](url)');
                    const cursor = editor.getCursor();
                    editor.setCursor(cursor.line, cursor.ch - 1);
                } else {
                    const start = editor.getCursor();
                    window.EditorCore.wrapText(editor, '![' , '](url)', 'alt text');
                    editor.setSelection({ line: start.line, ch: start.ch + 2 },
                                       { line: start.line, ch: start.ch + 10 });
                }
                break;
            }
            case 'table':
                window.EditorPickers.showTablePicker(button);
                break;
            case 'format-table':
                editor.execCommand('markdownTableFormat');
                editor.focus();
                break;
            case 'insert-row':
                editor.execCommand('markdownTableInsertRow');
                editor.focus();
                break;
            case 'insert-column':
                editor.execCommand('markdownTableInsertColumn');
                editor.focus();
                break;
            case 'delete-row':
                editor.execCommand('markdownTableDeleteRow');
                editor.focus();
                break;
            case 'delete-column':
                editor.execCommand('markdownTableDeleteColumn');
                editor.focus();
                break;
            case 'move-row-up':
                console.log('[DEBUG-DETAIL] Move Row Up button clicked');
                if (typeof CodeMirror.commands.directMoveRowUp === 'function') {
                    console.log('[DEBUG-DETAIL] Calling directMoveRowUp directly from toolbar button');
                    CodeMirror.commands.directMoveRowUp(editor);
                } else {
                    console.log('[DEBUG-DETAIL] Using execCommand for markdownTableMoveRowUp');
                    editor.execCommand('markdownTableMoveRowUp');
                }
                editor.focus();
                break;
            case 'move-row-down':
                console.log('[DEBUG-DETAIL] Move Row Down button clicked');
                if (typeof CodeMirror.commands.directMoveRowDown === 'function') {
                    console.log('[DEBUG-DETAIL] Calling directMoveRowDown directly from toolbar button');
                    CodeMirror.commands.directMoveRowDown(editor);
                } else {
                    console.log('[DEBUG-DETAIL] Using execCommand for markdownTableMoveRowDown');
                    editor.execCommand('markdownTableMoveRowDown');
                }
                editor.focus();
                break;
            case 'move-column-left':
                console.log('[DEBUG-DETAIL] Move Column Left button clicked');
                if (typeof CodeMirror.commands.directMoveColumnLeft === 'function') {
                    console.log('[DEBUG-DETAIL] Calling directMoveColumnLeft directly from toolbar button');
                    CodeMirror.commands.directMoveColumnLeft(editor);
                } else {
                    console.log('[DEBUG-DETAIL] Using execCommand for markdownTableMoveColumnLeft');
                    editor.execCommand('markdownTableMoveColumnLeft');
                }
                editor.focus();
                break;
            case 'move-column-right':
                console.log('[DEBUG-DETAIL] Move Column Right button clicked');
                if (typeof CodeMirror.commands.directMoveColumnRight === 'function') {
                    console.log('[DEBUG-DETAIL] Calling directMoveColumnRight directly from toolbar button');
                    CodeMirror.commands.directMoveColumnRight(editor);
                } else {
                    console.log('[DEBUG-DETAIL] Using execCommand for markdownTableMoveColumnRight');
                    editor.execCommand('markdownTableMoveColumnRight');
                }
                editor.focus();
                break;
            case 'horizontal-rule': {
                const cursor = editor.getCursor();
                editor.replaceRange('\n---\n', cursor);
                editor.focus();
                break;
            }
            case 'recent-edits':
                addRecentEdits(editor);
                break;
            case 'total':
                addTotal(editor);
                break;
            case 'undo':
                if (editor.historySize && editor.historySize().undo > 0) {
                    editor.undo();
                    editor.focus();
                }
                break;
            case 'redo':
                if (editor.historySize && editor.historySize().redo > 0) {
                    editor.redo();
                    editor.focus();
                }
                break;
            case 'preview':
                window.EditorPreview.togglePreview();
                break;
            case 'emoji':
                window.EditorPickers.showEmojiPicker(button);
                break;
            case 'doc-link':
                window.EditorPickers.showDocPicker(button);
                break;
            case 'anchor-link':
                window.EditorPickers.showAnchorPicker(button);
                break;
            case 'align-column-left':
                editor.execCommand('markdownTableAlignColumnLeft');
                editor.focus();
                break;
            case 'align-column-center':
                editor.execCommand('markdownTableAlignColumnCenter');
                editor.focus();
                break;
            case 'align-column-right':
                editor.execCommand('markdownTableAlignColumnRight');
                editor.focus();
                break;
            case 'align-column-none':
                editor.execCommand('markdownTableAlignColumnNone');
                editor.focus();
                break;
            case 'toggle-wordwrap':
                window.EditorCore.toggleWordWrap();
                break;
            case 'toggle-linenumbers':
                window.EditorCore.toggleLineNumbers();
                break;
            case 'toggle-split':
                window.EditorPreview.toggleSplit();
                break;
            case 'toggle-autocapitalize':
                window.EditorCore.toggleAutocapitalize();
                break;
            case 'insert-toc':
                insertTOC();
                break;
            case 'kanban':
                insertKanbanFrontmatter(editor);
                break;
            default:
                break;
        }
    };

    toolbar.addEventListener('click', (e) => {
        const btn = e.target.closest('button');
        if (!btn) return;
        const action = btn.dataset.action;
        if (action) exec(action, btn);
    });

    toolbar._delegateSetup = true;
}

// Export the module
window.EditorToolbar = {
    createToolbar,
    setupToolbarActions,
    
    // Action functions
    addHighlight,
    addSuperscript,
    addSubscript,
    addRecentEdits,
    addTotal,
    insertTOC,
    insertKanbanFrontmatter,
    
    // Utility functions
    getShortcut: (mac, other) => {
        const isMac = /Mac|iPhone|iPad|iPod/.test(navigator.userAgent);
        return isMac ? mac : other;
    }
};