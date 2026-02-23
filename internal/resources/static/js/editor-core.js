/**
 * Editor Core Module
 * Handles main editor functionality, initialization, and content management
 */

// Global editor variables
let editor = null;
let originalContent = '';

// Define custom CodeMirror modes
if (typeof CodeMirror !== 'undefined') {
    // Define a custom overlay mode for highlight syntax (==text==)
    CodeMirror.defineMode("markdown-highlight-overlay", function(config, parserConfig) {
        // Improved regex that handles edge cases better
        const highlightRegex = /==(?:[^=]|=(?!=))+?==/;

        // Helper function to check if a line number is inside a code block
        function isInCodeBlock(cm, lineNo) {
            let inCodeBlock = false;
            for (let i = 0; i < lineNo; i++) {
                let line = cm.getLine(i);
                if (line !== undefined) {
                    line = line.trim();
                    // Strip blockquote prefix(es)
                    while (line.startsWith('>')) {
                        line = line.substring(1).trim();
                    }
                    if (line.startsWith('```') || line.startsWith('~~~')) {
                        inCodeBlock = !inCodeBlock;
                    }
                }
            }
            return inCodeBlock;
        }

        return {
            token: function(stream, state) {
                // Get the CodeMirror instance from the stream
                const cm = stream.lineOracle && stream.lineOracle.doc && stream.lineOracle.doc.cm;
                
                // Check if current line starts/is a code block marker
                if (stream.sol()) {
                    let line = stream.string.trim();
                    // Strip blockquote prefix(es)
                    while (line.startsWith('>')) {
                        line = line.substring(1).trim();
                    }
                    
                    if (line.startsWith('```') || line.startsWith('~~~')) {
                        stream.skipToEnd();
                        return null;
                    }
                }
                
                // Check if we're inside a code block by scanning previous lines
                if (cm) {
                    const lineNo = stream.lineOracle.line;
                    let inCodeBlock = false;
                    
                    for (let i = 0; i <= lineNo; i++) {
                        let line = cm.getLine(i);
                        if (line !== undefined) {
                            line = line.trim();
                            // Strip blockquote prefix(es)
                            while (line.startsWith('>')) {
                                line = line.substring(1).trim();
                            }
                            if (line.startsWith('```') || line.startsWith('~~~')) {
                                inCodeBlock = !inCodeBlock;
                            }
                        }
                        // If we're on the current line and it's a code block marker, we're not "in" it yet
                        if (i === lineNo && (line.startsWith('```') || line.startsWith('~~~'))) {
                            break;
                        }
                    }
                    
                    if (inCodeBlock) {
                        stream.skipToEnd();
                        return null;
                    }
                }
                
                // Check if current position is inside inline code (backticks)
                const lineUpToCursor = stream.string.substring(0, stream.pos);
                const backtickCount = (lineUpToCursor.match(/`/g) || []).length;
                const inInlineCode = backtickCount % 2 === 1;
                
                if (inInlineCode) {
                    // Skip until closing backtick or end of line
                    while (stream.next() != null) {
                        if (stream.peek() === '`') break;
                    }
                    return null;
                }
                
                // Look for the start of a highlight marker
                if (stream.match(/==/)) {
                    // Check if we have a complete highlight pattern
                    const line = stream.string.slice(stream.pos - 2);
                    const match = line.match(highlightRegex);

                    if (match && match.index === 0) {
                        // We found a complete highlight pattern
                        // Move the stream position to the end of the match
                        stream.pos += match[0].length - 2;
                        return "highlight";
                    }

                    // If not a complete highlight, just return null
                    return null;
                }

                // Skip until we find a potential highlight marker or backtick
                while (stream.next() != null && !stream.match(/==/, false) && stream.peek() !== '`') {}
                return null;
            }
        };
    });

    // Define a custom overlay mode for strikethrough syntax (~~text~~)
    // We use a custom overlay instead of CodeMirror's built-in strikethrough
    // because the built-in one bleeds across lines and shows strikethrough
    // even when the closing ~~ is missing.
    CodeMirror.defineMode("markdown-strikethrough-overlay", function(config, parserConfig) {
        const strikethroughRegex = /~~(?:[^~]|~(?!~))+?~~/;

        return {
            token: function(stream, state) {
                // Get the CodeMirror instance from the stream
                const cm = stream.lineOracle && stream.lineOracle.doc && stream.lineOracle.doc.cm;

                // Check if current line starts/is a code block marker
                if (stream.sol()) {
                    let line = stream.string.trim();
                    while (line.startsWith('>')) {
                        line = line.substring(1).trim();
                    }
                    if (line.startsWith('```') || line.startsWith('~~~')) {
                        stream.skipToEnd();
                        return null;
                    }
                }

                // Check if we're inside a code block by scanning previous lines
                if (cm) {
                    const lineNo = stream.lineOracle.line;
                    let inCodeBlock = false;

                    for (let i = 0; i <= lineNo; i++) {
                        let line = cm.getLine(i);
                        if (line !== undefined) {
                            line = line.trim();
                            while (line.startsWith('>')) {
                                line = line.substring(1).trim();
                            }
                            if (line.startsWith('```') || line.startsWith('~~~')) {
                                inCodeBlock = !inCodeBlock;
                            }
                        }
                        if (i === lineNo && (line.startsWith('```') || line.startsWith('~~~'))) {
                            break;
                        }
                    }

                    if (inCodeBlock) {
                        stream.skipToEnd();
                        return null;
                    }
                }

                // Check if current position is inside inline code (backticks)
                const lineUpToCursor = stream.string.substring(0, stream.pos);
                const backtickCount = (lineUpToCursor.match(/`/g) || []).length;
                const inInlineCode = backtickCount % 2 === 1;

                if (inInlineCode) {
                    while (stream.next() != null) {
                        if (stream.peek() === '`') break;
                    }
                    return null;
                }

                // Look for the start of a strikethrough marker
                if (stream.match(/~~/)) {
                    // Check if we have a complete strikethrough pattern on this line
                    const line = stream.string.slice(stream.pos - 2);
                    const match = line.match(strikethroughRegex);

                    if (match && match.index === 0) {
                        // We found a complete strikethrough pattern
                        stream.pos += match[0].length - 2;
                        return "strikethrough";
                    }

                    // No closing ~~ found, don't style it
                    return null;
                }

                // Skip until we find a potential strikethrough marker or backtick
                while (stream.next() != null && !stream.match(/~~/, false) && stream.peek() !== '`') {}
                return null;
            }
        };
    });

    // Create a custom markdown mode that properly handles frontmatter
    CodeMirror.defineMode("markdown-with-frontmatter", function(config) {
        // Create the base markdown mode
        const markdownMode = CodeMirror.getMode(config, {
            name: "markdown",
            highlightFormatting: true,
            strikethrough: false,
            fencedCodeBlockHighlighting: true,
            taskLists: true
        });

        return {
            startState: function() {
                return {
                    markdownState: CodeMirror.startState(markdownMode),
                    inFrontmatter: false,
                    firstLine: true
                };
            },

            copyState: function(state) {
                return {
                    markdownState: CodeMirror.copyState(markdownMode, state.markdownState),
                    inFrontmatter: state.inFrontmatter,
                    firstLine: state.firstLine
                };
            },

            token: function(stream, state) {
                // Handle frontmatter start
                if (state.firstLine && stream.sol() && stream.match(/---/)) {
                    state.inFrontmatter = true;
                    state.firstLine = false;
                    return "frontmatter-delimiter";
                }

                // No longer on first line
                if (state.firstLine && stream.sol()) {
                    state.firstLine = false;
                }

                // Handle frontmatter end
                if (state.inFrontmatter && stream.sol() && stream.match(/---/)) {
                    state.inFrontmatter = false;
                    return "frontmatter-delimiter";
                }

                // Inside frontmatter
                if (state.inFrontmatter) {
                    // Skip to end of line for frontmatter content
                    stream.skipToEnd();
                    return "frontmatter";
                }

                // Default to markdown mode outside frontmatter
                return markdownMode.token(stream, state.markdownState);
            },

            blankLine: function(state) {
                if (state.inFrontmatter) {
                    return "frontmatter";
                }
                if (markdownMode.blankLine) {
                    return markdownMode.blankLine(state.markdownState);
                }
                return null;
            },

            indent: function(state, textAfter) {
                if (state.inFrontmatter) return 0;
                if (markdownMode.indent) {
                    return markdownMode.indent(state.markdownState, textAfter);
                }
                return CodeMirror.Pass;
            },

            innerMode: function(state) {
                if (state.inFrontmatter) return {state: state, mode: this};
                return {state: state.markdownState, mode: markdownMode};
            }
        };
    });

    // Create a simple frontmatter detector overlay that doesn't use state
    CodeMirror.defineMode("frontmatter-detector", function() {
        return {
            token: function(stream, state) {
                // Check for frontmatter delimiters at the beginning of lines
                if (stream.sol() && stream.match(/---/)) {
                    return "frontmatter-delimiter";
                }

                // Skip to next position
                stream.next();
                return null;
            }
        };
    });
}

// Helper functions for text manipulation
function wrapText(cm, prefix = '', suffix = '', placeholder = '') {
    const selection = cm.getSelection();
    if (selection) {
        cm.replaceSelection(`${prefix}${selection}${suffix}`);
    } else {
        const cursor = cm.getCursor();
        cm.replaceRange(prefix + placeholder + suffix, cursor);
        // Place cursor between prefix and suffix so the user can start typing.
        cm.setCursor(cursor.line, cursor.ch + prefix.length);
    }
    // Ensure the editor keeps focus.
    cm.focus();
}

// Main editor loading function
async function loadEditor(mainContent, editorContainer, viewToolbar, editToolbar) {
    try {
        const isHomepage = window.location.pathname === '/';
        const apiPath = isHomepage ? '/api/source/' : `/api/source${window.location.pathname}`;

        const response = await fetch(apiPath);
        if (!response.ok) throw new Error('Failed to fetch content');

        const markdown = await response.text();

        // Store original content for change detection
        originalContent = markdown;

        // Show editor and switch toolbars
        mainContent.classList.add('editing');
        editorContainer.classList.add('active');
        viewToolbar.style.display = 'none';
        editToolbar.style.display = 'flex';

        // Clear the editor container
        editorContainer.innerHTML = '';

        // Create a container for our editor components (toolbar, editor, preview, statusbar)
        const editorLayout = document.createElement('div');
        editorLayout.className = 'editor-layout';
        editorContainer.appendChild(editorLayout);

        // 1. Create toolbar at the top
        const toolbar = window.EditorToolbar.createToolbar(editorLayout);

        // 2. Create an editor area container to hold both editor and preview
        const editorArea = document.createElement('div');
        editorArea.className = 'editor-area';
        editorLayout.appendChild(editorArea);

        // 3. Create editor wrapper inside the editor area
        const editorWrapper = document.createElement('div');
        editorWrapper.className = 'custom-editor-wrapper';
        editorArea.appendChild(editorWrapper);

        // 4. Create textarea for CodeMirror
        const textarea = document.createElement('textarea');
        textarea.id = 'markdown-editor';
        editorWrapper.appendChild(textarea);

        // 5. Create preview element inside the editor area
        const previewElement = window.EditorPreview.createPreview(editorArea);

        // 6. Create statusbar at the bottom
        const statusbar = createStatusbar(editorLayout);

        // Reset the preview mode
        previewElement.classList.remove('editor-preview-active');

        // Initialize CodeMirror
        if (!editor) {
            // Check if we're on mobile
            const isMobile = window.innerWidth <= 768;

            editor = CodeMirror.fromTextArea(textarea, {
                mode: 'markdown',
                theme: 'default', // Always use default theme and customize it
                lineNumbers: false, // Line numbers off by default
                lineWrapping: true,
                autofocus: true,
                tabSize: 2,
                indentWithTabs: false,
                // Disable styleActiveLine completely on mobile to prevent the highlighting issue
                styleActiveLine: isMobile ? false : {
                    nonEmpty: true
                },
                extraKeys: {
                    // Tab and Enter handling is now done directly in markdown-table-editor.js
                    // by overriding the CodeMirror commands
                    // All other shortcuts are handled centrally in keyboard-shortcuts.js
                },
                placeholder: 'Write your markdown here...',
                spellcheck: true,
                gutters: ["CodeMirror-linenumbers"]
            });

            // Apply our custom markdown mode with frontmatter support
            editor.setOption("mode", "markdown-with-frontmatter");
            editor.addOverlay("markdown-highlight-overlay");
            editor.addOverlay("markdown-strikethrough-overlay");

            // Apply custom styling to the editor
            const isDarkMode = document.documentElement.getAttribute('data-theme') === 'dark';

            // Apply custom styling to tone down the syntax highlighting colors
            const editorElement = document.querySelector('.CodeMirror');
            if (editorElement) {
                window.EditorThemes.updateCodeMirrorTheme(isDarkMode ? 'dark' : 'light');
            }

            // Set editor height
            editor.setSize(null, "calc(100vh - 380px)");

            // Set up events for toolbar interactions
            window.EditorToolbar.setupToolbarActions(toolbar);

            // Set up events for statusbar updates
            editor.on('cursorActivity', () => updateStatusbar(statusbar));
            editor.on('change', () => {
                updateStatusbar(statusbar);

                // Update preview if in split or preview mode
                const editorMode = window.EditorPreview.getEditorMode();
                if (editorMode !== 'edit') {
                    const update = () => window.EditorPreview.updatePreview(editor.getValue());
                    if (editorMode === 'split') {
                        window.EditorPreview.setDebounceTimer(update);
                    } else {
                        update();
                    }
                }

                // Set up beforeunload handler when changes occur
                setupBeforeUnloadHandler();
            });

            // Set initial state of the wordwrap button based on editor settings
            const wrapButton = document.querySelector('.toggle-wordwrap-button');
            if (wrapButton && editor.getOption('lineWrapping')) {
                // Use the shared getShortcut function from EditorToolbar
                const getShortcut = window.EditorToolbar && window.EditorToolbar.getShortcut ? 
                    window.EditorToolbar.getShortcut : (mac, other) => mac || other;
                
                const shortcut = getShortcut('Option+Z', 'Alt+Z');
                
                wrapButton.classList.add('active');
                wrapButton.title = `Disable Word Wrap (${shortcut})`;
            }
        }

        // Set initial content
        editor.setValue(markdown);
        // Clear history so the loaded content is the initial undo boundary
        if (editor && typeof editor.clearHistory === 'function') {
            editor.clearHistory();
        }

        // Make sure editor is visible (not in preview mode)
        document.querySelector('.CodeMirror').style.display = 'block';

        // Force a refresh with multiple attempts to ensure editor renders properly
        refreshEditor(statusbar);

    } catch (error) {
        console.error('Error:', error);
        alert('Failed to load content for editing');
    }
}

// Function to create statusbar
function createStatusbar(container) {
    const statusbar = document.createElement('div');
    statusbar.className = 'editor-statusbar';

    // Add status elements (lines, words, cursor position)
    const linesStatus = document.createElement('span');
    linesStatus.className = 'lines';
    linesStatus.textContent = '0';

    const wordsStatus = document.createElement('span');
    wordsStatus.className = 'words';
    wordsStatus.textContent = '0';

    const cursorStatus = document.createElement('span');
    cursorStatus.className = 'cursor';
    cursorStatus.textContent = '0:0';

    statusbar.appendChild(linesStatus);
    statusbar.appendChild(wordsStatus);
    statusbar.appendChild(cursorStatus);

    container.appendChild(statusbar);
    return statusbar;
}

// Function to update statusbar
function updateStatusbar(statusbar) {
    if (!editor || !statusbar) return;

    const content = editor.getValue();
    const lines = content.split('\n').length;
    const words = content.trim() ? content.trim().split(/\s+/).length : 0;
    const cursor = editor.getCursor();

    statusbar.querySelector('.lines').textContent = lines;
    statusbar.querySelector('.words').textContent = words;
    statusbar.querySelector('.cursor').textContent = `${cursor.line + 1}:${cursor.ch}`;
}

// Helper function to ensure editor gets properly refreshed
function refreshEditor(statusbar) {
    if (!editor) return;

    // Immediate refresh - apply all refreshes at once
    editor.refresh();
    editor.focus();
    if (statusbar) updateStatusbar(statusbar);

    // Secondary quick refresh to ensure rendering
    requestAnimationFrame(() => {
        if (editor) {
            editor.refresh();
            editor.focus();
            if (statusbar) updateStatusbar(statusbar);
        }
    });
}

// Exit edit mode
function exitEditMode(mainContent, editorContainer, viewToolbar, editToolbar) {
    // Remove beforeunload handler first
    removeBeforeUnloadHandler();

    // Clear any pending preview debounce timer and reset editor mode
    if (window.EditorPreview) {
        window.EditorPreview.clearDebounceTimer();
        window.EditorPreview.setEditorMode('edit');
    }

    // Make sure to update UI state classes first
    if (mainContent) mainContent.classList.remove('editing');
    if (editorContainer) editorContainer.classList.remove('active');

    // Update toolbar visibility
    if (viewToolbar) viewToolbar.style.display = 'flex';
    if (editToolbar) editToolbar.style.display = 'none';

    // Reset original content
    originalContent = '';

    // Completely destroy the editor instance
    if (editor) {
        try {
            editor.toTextArea();
        } catch(e) {
            console.warn('Error while destroying editor:', e);
        }
        editor = null;
    }

    // Remove preview element
    window.EditorPreview.cleanup();

    // Remove pickers
    window.EditorPickers.cleanup();
}

// Content management functions
function getEditorContent() {
    return editor ? editor.getValue() : '';
}

function insertIntoEditor(url, isImage, name) {
    // Check if editor exists and is initialized
    if (!editor) {
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

    // Insert markdown at cursor position
    editor.replaceSelection(markdown);

    // Focus the editor
    editor.focus();

    return true;
}

function insertRawContent(content) {
    // Check if editor exists and is initialized
    if (!editor) {
        return false;
    }

    // Insert content at cursor position
    editor.replaceSelection(content);

    // Focus the editor
    editor.focus();

    return true;
}

function isEditorActive() {
    return !!editor;
}

// Function to check if document has unsaved changes
function hasUnsavedChanges() {
    if (!editor) return false;

    // Compare current content with original content
    const currentContent = editor.getValue();
    if (currentContent !== originalContent) {
        return true;
    }

    return false;
}

// Add function to handle beforeunload event
function setupBeforeUnloadHandler() {
    // Only add the handler if it doesn't already exist
    if (!window._beforeUnloadHandler) {
        window._beforeUnloadHandler = function(e) {
            if (hasUnsavedChanges()) {
                // Standard way to show a confirmation dialog when leaving page
                e.preventDefault();
                // Custom message (note: most modern browsers no longer show this message, just a generic one)
                const leaveMsg = window.i18n ? window.i18n.t('editor.unsaved_changes_leave') : 'You have unsaved changes. Are you sure you want to leave?';
                e.returnValue = leaveMsg;
                return leaveMsg;
            }
        };

        // Add the event listener
        window.addEventListener('beforeunload', window._beforeUnloadHandler);
    }
}

// Function to remove the beforeunload handler
function removeBeforeUnloadHandler() {
    if (window._beforeUnloadHandler) {
        window.removeEventListener('beforeunload', window._beforeUnloadHandler);
        window._beforeUnloadHandler = null;
    }
}

// Function to toggle word wrap
function toggleWordWrap() {
    if (!editor) return;

    // Get current line wrapping state
    const currentWrapping = editor.getOption('lineWrapping');

    // Toggle line wrapping
    editor.setOption('lineWrapping', !currentWrapping);

    // Update button appearance
    const wrapButton = document.querySelector('.toggle-wordwrap-button');
    if (wrapButton) {
        // Use the shared getShortcut function from EditorToolbar
        const getShortcut = window.EditorToolbar && window.EditorToolbar.getShortcut ? 
            window.EditorToolbar.getShortcut : (mac, other) => mac || other;
        
        const shortcut = getShortcut('Option+Z', 'Alt+Z');
        
        if (!currentWrapping) {
            // Line wrapping was turned on
            wrapButton.classList.add('active');
            wrapButton.title = `Disable Word Wrap (${shortcut})`;
        } else {
            // Line wrapping was turned off
            wrapButton.classList.remove('active');
            wrapButton.title = `Enable Word Wrap (${shortcut})`;
        }
    }

    // Refresh editor to apply changes
    editor.refresh();
}

// Function to toggle line numbers
function toggleLineNumbers() {
    if (!editor) return;

    // Get current line numbers state
    const currentLineNumbers = editor.getOption('lineNumbers');

    // Toggle line numbers
    editor.setOption('lineNumbers', !currentLineNumbers);

    // Update button appearance
    const lineNumbersButton = document.querySelector('.toggle-linenumbers-button');
    if (lineNumbersButton) {
        // Use the shared getShortcut function from EditorToolbar
        const getShortcut = window.EditorToolbar && window.EditorToolbar.getShortcut ? 
            window.EditorToolbar.getShortcut : (mac, other) => mac || other;
        
        const shortcut = getShortcut('Option+N', 'Alt+N');
        
        if (!currentLineNumbers) {
            // Line numbers were turned on
            lineNumbersButton.classList.add('active');
            lineNumbersButton.title = `Hide Line Numbers (${shortcut})`;
        } else {
            // Line numbers were turned off
            lineNumbersButton.classList.remove('active');
            lineNumbersButton.title = `Show Line Numbers (${shortcut})`;
        }
    }

    // Refresh editor to apply changes
    editor.refresh();
}

// Variable to track autocapitalize state
let autocapitalizeEnabled = false;

// Function to toggle autocapitalize
function toggleAutocapitalize() {
    if (!editor) return;

    // Toggle autocapitalize state
    autocapitalizeEnabled = !autocapitalizeEnabled;
    
    // Update button appearance
    const autocapitalizeButton = document.querySelector('.toggle-autocapitalize-button');
    if (autocapitalizeButton) {
        // Use the shared getShortcut function from EditorToolbar
        const getShortcut = window.EditorToolbar && window.EditorToolbar.getShortcut ? 
            window.EditorToolbar.getShortcut : (mac, other) => mac || other;
        
        const shortcut = getShortcut('Option+C', 'Alt+C');
        
        if (autocapitalizeEnabled) {
            // Autocapitalize was turned on
            autocapitalizeButton.classList.add('active');
            autocapitalizeButton.title = `Disable Auto-Capitalize (${shortcut})`;
        } else {
            // Autocapitalize was turned off
            autocapitalizeButton.classList.remove('active');
            autocapitalizeButton.title = `Enable Auto-Capitalize (${shortcut})`;
        }
    }
    
    // If enabled, set up the event handler for autocapitalize
    if (autocapitalizeEnabled) {
        // Only add if not already added
        if (!editor._autocapitalizeHandler) {
            editor._autocapitalizeHandler = function(cm, change) {
                // Only process if autocapitalize is enabled
                if (!autocapitalizeEnabled) return;
                
                // Only process single character insertions and newlines
                if (change.origin === '+input' && (change.text.length === 1 || (change.text.length === 2 && change.text[0] === ''))) {
                    const doc = cm.getDoc();
                    const cursor = doc.getCursor();
                    const line = doc.getLine(cursor.line);
                    
                    // Check if we're at the beginning of a line or after punctuation followed by space
                    if (cursor.ch === 1 || 
                        (cursor.ch >= 2 && 
                         /[.!?]\s$/.test(line.substring(cursor.ch - 2, cursor.ch)))) {
                        
                        // Get the character just inserted
                        const char = line.charAt(cursor.ch - 1);
                        
                        // If it's a lowercase letter, capitalize it
                        if (/[a-z]/.test(char)) {
                            const uppercase = char.toUpperCase();
                            doc.replaceRange(uppercase, 
                                {line: cursor.line, ch: cursor.ch - 1}, 
                                {line: cursor.line, ch: cursor.ch});
                        }
                    }
                }
            };
            
            // Add the change event handler
            editor.on('change', editor._autocapitalizeHandler);
        }
    } else {
        // Remove the event handler if it exists
        if (editor._autocapitalizeHandler) {
            editor.off('change', editor._autocapitalizeHandler);
            editor._autocapitalizeHandler = null;
        }
    }
}

// Export the module
window.EditorCore = {
    // Editor management
    loadEditor,
    exitEditMode,
    refreshEditor,

    // Content functions
    getEditorContent,
    insertIntoEditor,
    insertRawContent,
    isEditorActive,
    hasUnsavedChanges,

        // Utility functions
    wrapText,
    toggleWordWrap,
    toggleLineNumbers,
    toggleAutocapitalize,

    // Getters
    getEditor: () => editor,
    getOriginalContent: () => originalContent,
    setOriginalContent: (content) => { originalContent = content; }
};