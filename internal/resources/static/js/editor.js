// EasyMDE Editor Module
let easyMDE = null;

// Helper functions for editor toolbar actions
function addSubscript(editor) {
    var cm = editor.codemirror;
    var selection = cm.getSelection();
    if (selection) {
        cm.replaceSelection("~" + selection + "~");
    } else {
        var cursor = cm.getCursor();
        cm.replaceRange("~~", cursor);
        cm.setCursor(cursor.line, cursor.ch + 1);
    }
}

function addSuperscript(editor) {
    var cm = editor.codemirror;
    var selection = cm.getSelection();
    if (selection) {
        cm.replaceSelection("^" + selection + "^");
    } else {
        var cursor = cm.getCursor();
        cm.replaceRange("^^", cursor);
        cm.setCursor(cursor.line, cursor.ch + 1);
    }
}

function addRecentEdits(editor) {
    var cm = editor.codemirror;
    var cursor = cm.getCursor();
    cm.replaceRange(":::stats recent=5:::\n", cursor);
}

function addTotal(editor) {
    var cm = editor.codemirror;
    var cursor = cm.getCursor();
    cm.replaceRange(":::stats count=all:::\n", cursor);
}

// Table picker functionality
function createTablePicker(editor) {
    const picker = document.createElement('div');
    picker.className = 'table-picker';
    picker.style.display = 'none';

    for (let i = 0; i < 10; i++) {
        for (let j = 0; j < 10; j++) {
            const cell = document.createElement('div');
            cell.className = 'cell';
            cell.dataset.row = i + 1;
            cell.dataset.col = j + 1;
            cell.addEventListener('mouseover', () => highlightCells(cell));
            cell.addEventListener('click', () => insertTable(editor, i + 1, j + 1));
            picker.appendChild(cell);
        }
        picker.appendChild(document.createElement('br'));
    }

    document.body.appendChild(picker);
    return picker;
}

function highlightCells(cell) {
    const row = parseInt(cell.dataset.row);
    const col = parseInt(cell.dataset.col);
    const cells = document.querySelectorAll('.table-picker .cell');
    cells.forEach(c => {
        if (parseInt(c.dataset.row) <= row && parseInt(c.dataset.col) <= col) {
            c.classList.add('active');
        } else {
            c.classList.remove('active');
        }
    });
}

function insertTable(editor, rows, cols) {
    const cm = editor.codemirror;
    let table = '\n';
    // Header
    table += '| ' + Array(cols).fill('Header').join(' | ') + ' |\n';
    // Separator
    table += '| ' + Array(cols).fill('---').join(' | ') + ' |\n';
    // Rows
    for (let i = 0; i < rows; i++) {
        table += '| ' + Array(cols).fill('Cell').join(' | ') + ' |\n';
    }
    cm.replaceSelection(table);
    document.querySelector('.table-picker').style.display = 'none';
}

// Main editor functions
async function loadEditor(mainContent, editorContainer, viewToolbar, editToolbar) {
    try {
        const isHomepage = window.location.pathname === '/';
        const apiPath = isHomepage ? '/api/source/' : `/api/source${window.location.pathname}`;

        const response = await fetch(apiPath);
        if (!response.ok) throw new Error('Failed to fetch content');

        const markdown = await response.text();

        // Show editor and switch toolbars
        mainContent.classList.add('editing');
        editorContainer.classList.add('active');
        viewToolbar.style.display = 'none';
        editToolbar.style.display = 'flex';

        // Initialize EasyMDE if not already initialized
        if (!easyMDE && document.getElementById('markdown-editor')) {
            easyMDE = new EasyMDE({
                element: document.getElementById('markdown-editor'),
                autoDownloadFontAwesome: false,
                spellChecker: false,
                autofocus: true,
                showIcons: ["code", "table"],
                toolbar: [
                    "bold", "italic", "strikethrough", "heading", "|",
                    "code", "quote", "unordered-list", "ordered-list", "|",
                    'link', 'image', '|',
                    {
                        name: "insert-table",
                        action: function(editor) {
                            const picker = document.querySelector('.table-picker') || createTablePicker(editor);
                            const button = editor.toolbarElements['insert-table'];
                            const rect = button.getBoundingClientRect();
                            picker.style.top = `${rect.bottom + window.scrollY}px`;
                            picker.style.left = `${rect.left + window.scrollX}px`;

                            // Toggle the display of the picker
                            if (picker.style.display === 'none' || picker.style.display === '') {
                                picker.style.display = 'block';

                                // Add a one-time event listener to close when clicking outside
                                const closeHandler = function(e) {
                                    if (!picker.contains(e.target) &&
                                        !e.target.closest('.fa-th') &&
                                        !e.target.classList.contains('fa-th') &&
                                        e.target.name !== 'insert-table') {
                                        picker.style.display = 'none';
                                        document.removeEventListener('click', closeHandler);
                                    }
                                };

                                // Use setTimeout to avoid the current click event from immediately closing the picker
                                setTimeout(() => {
                                    document.addEventListener('click', closeHandler);
                                }, 0);
                            } else {
                                picker.style.display = 'none';
                            }
                        },
                        className: "fa fa-th",
                        title: "Insert Custom Table",
                    },
                    "horizontal-rule", "|",
                    {
                        name: "subscript",
                        action: addSubscript,
                        className: "fa fa-subscript",
                        title: "Add Subscript",
                    },
                    {
                        name: "superscript",
                        action: addSuperscript,
                        className: "fa fa-superscript",
                        title: "Add Superscript",
                    },
                    {
                        name: "recent-edits",
                        action: addRecentEdits,
                        className: "fa fa-clock-o",
                        title: "Insert Recent Edits",
                    },
                    {
                        name: "total",
                        action: addTotal,
                        className: "fa fa-book",
                        title: "Insert Total Number of Documents",
                    },
                    "|", "preview", "|", "undo", "redo"
                ],
                status: ['lines', 'words', 'cursor'],
                minHeight: '300px',
                initialValue: markdown,
                autoRefresh: true,
            });
        } else if (easyMDE) {
            // Set editor content and force refresh
            easyMDE.value(markdown);
        }

        // Force a refresh of the editor
        setTimeout(() => {
            if (easyMDE) {
                easyMDE.codemirror.refresh();
                easyMDE.codemirror.focus();
            }
        }, 0);
    } catch (error) {
        console.error('Error:', error);
        alert('Failed to load content for editing');
    }
}

function exitEditMode(mainContent, editorContainer, viewToolbar, editToolbar) {
    mainContent.classList.remove('editing');
    editorContainer.classList.remove('active');
    viewToolbar.style.display = 'flex';
    editToolbar.style.display = 'none';

    // Completely destroy the editor instance
    if (easyMDE) {
        easyMDE.toTextArea();
        easyMDE = null;
    }
}

function getEditorContent() {
    return easyMDE ? easyMDE.value() : '';
}

function insertIntoEditor(url, isImage, name) {
    // Check if editor exists and is initialized
    if (!easyMDE) {
        return false;
    }

    // Check if it's an MP4 file
    const isVideo = name.toLowerCase().endsWith('.mp4');

    // Extract just the filename from the URL
    const filename = name;

    // Create markdown code based on file type
    let markdown = '';
    if (isVideo) {
        // For MP4 files, use code block syntax with just the filename
        markdown = "```mp4\n" + filename + "\n```\n\n";
    } else if (isImage) {
        // For images, use image markdown with just the filename
        markdown = `![${name}](${filename})\n\n`;
    } else {
        // For other files, use link markdown with just the filename
        markdown = `[${name}](${filename})\n\n`;
    }

    // Get current cursor position
    const currentPosition = easyMDE.codemirror.getCursor();

    // Insert markdown at cursor position
    easyMDE.codemirror.replaceRange(markdown, currentPosition);

    // Focus the editor
    easyMDE.codemirror.focus();

    return true;
}

// Function to insert raw content into the editor
function insertRawContent(content) {
    // Check if editor exists and is initialized
    if (!easyMDE) {
        return false;
    }

    // Get current cursor position
    const currentPosition = easyMDE.codemirror.getCursor();

    // Insert content at cursor position
    easyMDE.codemirror.replaceRange(content, currentPosition);

    // Focus the editor
    easyMDE.codemirror.focus();

    return true;
}

// Function to check if editor is active
function isEditorActive() {
    return !!easyMDE;
}

// Export the functions
window.WikiEditor = {
    loadEditor,
    exitEditMode,
    getEditorContent,
    insertIntoEditor,
    insertRawContent,
    isEditorActive
};