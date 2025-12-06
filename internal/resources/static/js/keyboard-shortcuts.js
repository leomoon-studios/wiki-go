/**
 * Keyboard Shortcuts Module
 * Centralizes keyboard shortcut handling for the entire application
 */

// Platform detection
const isMac = (() => {
    // Modern API with fallback to userAgent
    if (navigator.userAgentData) {
        return navigator.userAgentData.platform === 'macOS';
    }
    // Fallback for older browsers
    return /Mac|iPhone|iPad|iPod/.test(navigator.userAgent);
})();

// Centralized shortcut definitions
const SHORTCUT_DEFINITIONS = {
    // Global shortcuts
    'enterEditMode': {
        mac: { cmd: true, key: 'e' },
        other: { ctrl: true, key: 'e' },
        description: 'Enter edit mode'
    },
    'saveDocument': {
        mac: { cmd: true, key: 's' },
        other: { ctrl: true, key: 's' },
        description: 'Save document when in edit mode'
    },
    'focusSearch': {
        mac: { cmd: true, shift: true, key: 'f' },
        other: { ctrl: true, shift: true, key: 'f' },
        description: 'Focus the search box'
    },

    // Formatting shortcuts
    'formatBold': {
        mac: { cmd: true, key: 'b' },
        other: { ctrl: true, key: 'b' },
        description: 'Toggle bold formatting'
    },
    'formatItalic': {
        mac: { cmd: true, key: 'i' },
        other: { ctrl: true, key: 'i' },
        description: 'Toggle italic formatting'
    },
    'formatHeading': {
        mac: { cmd: true, key: 'h' },
        other: { ctrl: true, key: 'h' },
        description: 'Toggle/cycle heading levels'
    },
    'formatQuote': {
        mac: { cmd: true, key: 'k' },
        other: { ctrl: true, key: 'k' },
        description: 'Toggle block quote'
    },
    'formatCode': {
        mac: { cmd: true, key: '/' },
        other: { ctrl: true, key: '/' },
        description: 'Toggle code formatting'
    },

    // Editor shortcuts
    'togglePreview': {
        mac: { cmd: true, shift: true, key: 'p' },
        other: { ctrl: true, shift: true, key: 'p' },
        description: 'Preview Toggle'
    },
    'toggleWordWrap': {
        mac: { option: true, key: 'z' },
        other: { alt: true, key: 'z' },
        description: 'Toggle word wrap'
    },
    'toggleLineNumbers': {
        mac: { option: true, key: 'n' },
        other: { alt: true, key: 'n' },
        description: 'Toggle line numbers'
    },
    'toggleAutocapitalize': {
        mac: { option: true, key: 'c' },
        other: { alt: true, key: 'c' },
        description: 'Toggle auto-capitalize'
    },
    
    // Table shortcuts
    'tableEscape': {
        mac: { cmd: true, key: 'Enter' },
        other: { ctrl: true, key: 'Enter' },
        description: 'Exit table at current position'
    },
    'tableMoveLeft': {
        mac: { cmd: true, key: 'ArrowLeft' },
        other: { ctrl: true, key: 'ArrowLeft' },
        description: 'Move to cell on the left'
    },
    'tableMoveRight': {
        mac: { cmd: true, key: 'ArrowRight' },
        other: { ctrl: true, key: 'ArrowRight' },
        description: 'Move to cell on the right'
    },
    'tableMoveUp': {
        mac: { cmd: true, key: 'ArrowUp' },
        other: { ctrl: true, key: 'ArrowUp' },
        description: 'Move to cell above'
    },
    'tableMoveDown': {
        mac: { cmd: true, key: 'ArrowDown' },
        other: { ctrl: true, key: 'ArrowDown' },
        description: 'Move to cell below'
    },
    'tableAlignLeft': {
        mac: { cmd: true, shift: true, key: 'ArrowLeft' },
        other: { ctrl: true, shift: true, key: 'ArrowLeft' },
        description: 'Align column left'
    },
    'tableAlignRight': {
        mac: { cmd: true, shift: true, key: 'ArrowRight' },
        other: { ctrl: true, shift: true, key: 'ArrowRight' },
        description: 'Align column right'
    },
    'tableAlignCenter': {
        mac: { cmd: true, shift: true, key: 'ArrowUp' },
        other: { ctrl: true, shift: true, key: 'ArrowUp' },
        description: 'Align column center'
    },
    'tableAlignNone': {
        mac: { cmd: true, shift: true, key: 'ArrowDown' },
        other: { ctrl: true, shift: true, key: 'ArrowDown' },
        description: 'Remove column alignment'
    },
    'tableMoveRowUp': {
        mac: { option: true, key: 'ArrowUp' },
        other: { alt: true, key: 'ArrowUp' },
        description: 'Move row up'
    },
    'tableMoveRowDown': {
        mac: { option: true, key: 'ArrowDown' },
        other: { alt: true, key: 'ArrowDown' },
        description: 'Move row down'
    },
    'tableMoveColumnLeft': {
        mac: { option: true, key: 'ArrowLeft' },
        other: { alt: true, key: 'ArrowLeft' },
        description: 'Move column left'
    },
    'tableMoveColumnRight': {
        mac: { option: true, key: 'ArrowRight' },
        other: { alt: true, key: 'ArrowRight' },
        description: 'Move column right'
    }
};

// Track editor related elements
let mainContent;
let editorContainer;
let viewToolbar;
let editToolbar;
let editPageButton;
let saveButton;

// Helper function to check if shortcut matches
function matchesShortcut(event, shortcut) {
    const platform = isMac ? 'mac' : 'other';
    const platformShortcut = shortcut[platform];

    if (!platformShortcut) return false;

    // Handle special key mappings
    let eventKey = event.key;
    if (eventKey === 'ArrowLeft') eventKey = 'ArrowLeft';
    else if (eventKey === 'ArrowRight') eventKey = 'ArrowRight';
    else if (eventKey === 'ArrowUp') eventKey = 'ArrowUp';
    else if (eventKey === 'ArrowDown') eventKey = 'ArrowDown';
    else if (event.code === 'Slash') eventKey = '/';
    else eventKey = eventKey.toLowerCase();

    const targetKey = platformShortcut.key.toLowerCase();

    return (!platformShortcut.ctrl || event.ctrlKey) &&
           (!platformShortcut.cmd || event.metaKey) &&
           (!platformShortcut.meta || event.metaKey) &&
           (!platformShortcut.alt || event.altKey) &&
           (!platformShortcut.option || event.altKey) &&
           (!platformShortcut.shift || event.shiftKey) &&
           (eventKey === targetKey || eventKey === platformShortcut.key);
}

// Helper function to get shortcut action from event
function getShortcutAction(event) {
    for (const [actionName, shortcut] of Object.entries(SHORTCUT_DEFINITIONS)) {
        if (matchesShortcut(event, shortcut)) {
            return actionName;
        }
    }
    return null;
}

// Initialize keyboard shortcuts
function initKeyboardShortcuts() {
    // Get DOM references
    mainContent = document.querySelector('.content');
    editorContainer = document.querySelector('.editor-container');
    viewToolbar = document.querySelector('.view-toolbar');
    editToolbar = document.querySelector('.edit-toolbar');
    editPageButton = document.querySelector('.edit-page');
    saveButton = document.querySelector('.save-changes');

    // Add global keyboard event listener
    document.addEventListener('keydown', handleKeyDown);

    // Register formatting commands when CodeMirror is available
    if (window.CodeMirror) {
        registerFormattingCommands();
        registerAllCodeMirrorShortcuts();
    }
}

// Register all CodeMirror shortcuts
function registerAllCodeMirrorShortcuts() {
    // Make sure we're working with the default keyMap
    if (!CodeMirror.keyMap.default) {
        CodeMirror.keyMap.default = {};
    }

    // Platform-aware key registration
    if (isMac) {
        // Mac shortcuts use Cmd
        CodeMirror.keyMap.default['Cmd-B'] = 'formatBold';
        CodeMirror.keyMap.default['Cmd-I'] = 'formatItalic';
        CodeMirror.keyMap.default['Cmd-H'] = 'formatHeading';
        CodeMirror.keyMap.default['Cmd-K'] = 'formatQuote';
        CodeMirror.keyMap.default['Cmd-/'] = 'formatCode';
        CodeMirror.keyMap.default['Cmd-Shift-P'] = 'togglePreview';
        CodeMirror.keyMap.default['Alt-Z'] = 'toggleWordWrap';
        CodeMirror.keyMap.default['Alt-N'] = 'toggleLineNumbers';
        CodeMirror.keyMap.default['Alt-C'] = 'toggleAutocapitalize';

        // Table shortcuts for Mac
        CodeMirror.keyMap.default['Cmd-Enter'] = 'tableEscape';
        CodeMirror.keyMap.default['Cmd-Left'] = 'tableMoveLeft';
        CodeMirror.keyMap.default['Cmd-Right'] = 'tableMoveRight';
        CodeMirror.keyMap.default['Cmd-Up'] = 'tableMoveUp';
        CodeMirror.keyMap.default['Cmd-Down'] = 'tableMoveDown';
        CodeMirror.keyMap.default['Shift-Cmd-Left'] = 'tableAlignLeft';
        CodeMirror.keyMap.default['Shift-Cmd-Right'] = 'tableAlignRight';
        CodeMirror.keyMap.default['Shift-Cmd-Up'] = 'tableAlignCenter';
        CodeMirror.keyMap.default['Shift-Cmd-Down'] = 'tableAlignNone';
        CodeMirror.keyMap.default['Alt-Up'] = 'markdownTableMoveRowUp';
        CodeMirror.keyMap.default['Alt-Down'] = 'markdownTableMoveRowDown';
        CodeMirror.keyMap.default['Alt-Left'] = 'markdownTableMoveColumnLeft';
        CodeMirror.keyMap.default['Alt-Right'] = 'markdownTableMoveColumnRight';
    } else {
        // Windows/Linux shortcuts use Ctrl
        CodeMirror.keyMap.default['Ctrl-B'] = 'formatBold';
        CodeMirror.keyMap.default['Ctrl-I'] = 'formatItalic';
        CodeMirror.keyMap.default['Ctrl-H'] = 'formatHeading';
        CodeMirror.keyMap.default['Ctrl-K'] = 'formatQuote';
        CodeMirror.keyMap.default['Ctrl-/'] = 'formatCode';
        CodeMirror.keyMap.default['Ctrl-Shift-P'] = 'togglePreview';
        CodeMirror.keyMap.default['Alt-Z'] = 'toggleWordWrap';
        CodeMirror.keyMap.default['Alt-N'] = 'toggleLineNumbers';
        CodeMirror.keyMap.default['Alt-C'] = 'toggleAutocapitalize';

        // Table shortcuts for Windows/Linux
        CodeMirror.keyMap.default['Ctrl-Enter'] = 'tableEscape';
        CodeMirror.keyMap.default['Ctrl-Left'] = 'tableMoveLeft';
        CodeMirror.keyMap.default['Ctrl-Right'] = 'tableMoveRight';
        CodeMirror.keyMap.default['Ctrl-Up'] = 'tableMoveUp';
        CodeMirror.keyMap.default['Ctrl-Down'] = 'tableMoveDown';
        CodeMirror.keyMap.default['Shift-Ctrl-Left'] = 'tableAlignLeft';
        CodeMirror.keyMap.default['Shift-Ctrl-Right'] = 'tableAlignRight';
        CodeMirror.keyMap.default['Shift-Ctrl-Up'] = 'tableAlignCenter';
        CodeMirror.keyMap.default['Shift-Ctrl-Down'] = 'tableAlignNone';
        CodeMirror.keyMap.default['Alt-Up'] = 'markdownTableMoveRowUp';
        CodeMirror.keyMap.default['Alt-Down'] = 'markdownTableMoveRowDown';
        CodeMirror.keyMap.default['Alt-Left'] = 'markdownTableMoveColumnLeft';
        CodeMirror.keyMap.default['Alt-Right'] = 'markdownTableMoveColumnRight';
    }
}

// Register text formatting commands and shortcuts with CodeMirror
function registerFormattingCommands() {
    // Define the formatting commands
    // Bold formatting
    CodeMirror.commands.formatBold = function(cm) {
        const selection = cm.getSelection();
        if (selection) {
            cm.replaceSelection(`**${selection}**`);
            // Move cursor after the bold text
            const cursor = cm.getCursor();
            cm.setCursor(cursor);
        } else {
            const cursor = cm.getCursor();
            cm.replaceRange('**', cursor);
            cm.setCursor({ line: cursor.line, ch: cursor.ch });
            cm.replaceRange('**', cursor);
            cm.setCursor({ line: cursor.line, ch: cursor.ch });
        }
        cm.focus();
    };

    // Italic formatting
    CodeMirror.commands.formatItalic = function(cm) {
        const selection = cm.getSelection();
        if (selection) {
            cm.replaceSelection(`*${selection}*`);
            // Move cursor after the italic text
            const cursor = cm.getCursor();
            cm.setCursor(cursor);
        } else {
            const cursor = cm.getCursor();
            cm.replaceRange('*', cursor);
            cm.setCursor({ line: cursor.line, ch: cursor.ch });
            cm.replaceRange('*', cursor);
            cm.setCursor({ line: cursor.line, ch: cursor.ch });
        }
        cm.focus();
    };

    // Heading formatting
    CodeMirror.commands.formatHeading = function(cm) {
        const cursor = cm.getCursor();
        const line = cm.getLine(cursor.line);
        const lineWithoutIndent = line.trimStart();
        const indent = line.length - lineWithoutIndent.length;

        // Check for existing heading level
        let headingLevel = 0;
        for (let i = indent; i < line.length; i++) {
            if (line[i] === '#') {
                headingLevel++;
            } else if (line[i] !== ' ') {
                break;
            }
        }

        if (headingLevel > 0 && headingLevel < 6) {
            // Increment heading level (# -> ## -> ### -> ...)
            const newPrefix = ' '.repeat(indent) + '#'.repeat(headingLevel + 1) + ' ';
            const content = line.substring(indent + headingLevel + (line[indent + headingLevel] === ' ' ? 1 : 0)).trim();

            cm.replaceRange(newPrefix + content,
                { line: cursor.line, ch: 0 },
                { line: cursor.line, ch: line.length });

            cm.setCursor({ line: cursor.line, ch: newPrefix.length + content.length });

        } else if (headingLevel >= 6) {
            // Reset to no heading (removing the heading markers)
            const content = line.substring(indent + headingLevel + (line[indent + headingLevel] === ' ' ? 1 : 0)).trim();
            cm.replaceRange(' '.repeat(indent) + content,
                { line: cursor.line, ch: 0 },
                { line: cursor.line, ch: line.length });

            cm.setCursor({ line: cursor.line, ch: indent + content.length });

        } else {
            // Add heading level 1
            const newPrefix = ' '.repeat(indent) + '# ';
            const content = lineWithoutIndent;

            cm.replaceRange(newPrefix + content,
                { line: cursor.line, ch: 0 },
                { line: cursor.line, ch: line.length });

            cm.setCursor({ line: cursor.line, ch: newPrefix.length + content.length });
        }
        cm.focus();
    };

    // Quote formatting
    CodeMirror.commands.formatQuote = function(cm) {
        const cursor = cm.getCursor();
        const line = cm.getLine(cursor.line);
        const trimmedLine = line.trimStart();
        const indentSize = line.length - trimmedLine.length;

        if (trimmedLine.startsWith('>')) {
            // Line is already a quote, remove the quote marker
            const quoteContent = trimmedLine.substring(1).trimStart();
            cm.replaceRange(' '.repeat(indentSize) + quoteContent,
                { line: cursor.line, ch: 0 },
                { line: cursor.line, ch: line.length });

            cm.setCursor({ line: cursor.line, ch: indentSize + quoteContent.length });
        } else {
            // Convert line to a quote
            const newPrefix = ' '.repeat(indentSize) + '> ';
            cm.replaceRange(newPrefix + trimmedLine,
                { line: cursor.line, ch: 0 },
                { line: cursor.line, ch: line.length });

            cm.setCursor({ line: cursor.line, ch: newPrefix.length + trimmedLine.length });
        }
        cm.focus();
    };

    // Smart code formatting
    CodeMirror.commands.formatCode = function(cm) {
        const selection = cm.getSelection();
        const cursor = cm.getCursor();

        // Check if there's a selection and if it spans multiple lines
        if (selection) {
            const hasMultipleLines = selection.includes('\n');

            if (hasMultipleLines) {
                // Multi-line selection: Use code block (```) formatting
                // Check if already in a code block
                const lines = selection.split('\n');
                const isCodeBlock = lines.length >= 2 &&
                                lines[0].trim().startsWith('```') &&
                                lines[lines.length - 1].trim() === '```';

                if (isCodeBlock) {
                    // Remove code block formatting
                    const contentLines = lines.slice(1, lines.length - 1);
                    cm.replaceSelection(contentLines.join('\n'));
                } else {
                    // Add code block formatting
                    cm.replaceSelection('```\n' + selection + '\n```');
                }
            } else {
                // Single-line selection: Use inline code (`) formatting
                // Check if already in inline code
                const isInlineCode = selection.startsWith('`') && selection.endsWith('`');

                if (isInlineCode) {
                    // Remove inline code formatting
                    cm.replaceSelection(selection.substring(1, selection.length - 1));
                } else {
                    // Add inline code formatting
                    cm.replaceSelection('`' + selection + '`');
                }
            }
        } else {
            // No selection: Get the current line and add code block formatting (```)
            const line = cursor.line;
            const lineText = cm.getLine(line);

            // Check if we need to remove existing code block
            const prevLine = line > 0 ? cm.getLine(line - 1) : "";
            const nextLine = line < cm.lineCount() - 1 ? cm.getLine(line + 1) : "";
            const isInCodeBlock = prevLine.trim() === '```' && nextLine.trim() === '```';

            if (isInCodeBlock) {
                // Remove the code block (remove the lines before and after)
                cm.replaceRange('',
                    { line: line + 1, ch: 0 },
                    { line: line + 2, ch: 0 });
                cm.replaceRange('',
                    { line: line - 1, ch: 0 },
                    { line: line, ch: 0 });
            } else {
                // Add a code block around the current line
                // Insert the closing ``` first to avoid line number changes
                cm.replaceRange('\n```',
                    { line: line, ch: lineText.length });

                // Then insert the opening ```
                cm.replaceRange('```\n',
                    { line: line, ch: 0 });

                // Position cursor at the same position on the line (now line+1)
                cm.setCursor({ line: line + 1, ch: cursor.ch });
            }
        }

        cm.focus();
    };

    // Register toggle preview command
    CodeMirror.commands.togglePreview = function(cm) {
        if (window.EditorPreview && typeof window.EditorPreview.togglePreview === 'function') {
            window.EditorPreview.togglePreview();
        }
    };

    // Register toggle word wrap command
    CodeMirror.commands.toggleWordWrap = function(cm) {
        if (window.EditorCore && typeof window.EditorCore.toggleWordWrap === 'function') {
            window.EditorCore.toggleWordWrap();
        }
    };

    // Register toggle line numbers command
    CodeMirror.commands.toggleLineNumbers = function(cm) {
        if (window.EditorCore && typeof window.EditorCore.toggleLineNumbers === 'function') {
            window.EditorCore.toggleLineNumbers();
        }
    };
    
    // Register toggle autocapitalize command
    CodeMirror.commands.toggleAutocapitalize = function(cm) {
        if (window.EditorCore && typeof window.EditorCore.toggleAutocapitalize === 'function') {
            window.EditorCore.toggleAutocapitalize();
        }
    };
}

// Handle all keyboard events
function handleKeyDown(e) {
    // Get the shortcut action for this event
    const action = getShortcutAction(e);

    if (action) {
        // Handle global shortcuts
        switch (action) {
            case 'enterEditMode':
                e.preventDefault();
                if (editPageButton && !mainContent.classList.contains('editing')) {
                    // Check auth before navigating
                    fetch('/api/check-auth').then(authResponse => {
                        if (authResponse.status === 401) {
                            // Show login dialog
                            window.Auth.showLoginDialog(() => {
                                window.Auth.checkUserRole('editor').then(canEdit => {
                                    if (canEdit) {
                                        const url = new URL(window.location);
                                        url.searchParams.set('mode', 'edit');
                                        window.location.href = url.toString();
                                    } else {
                                        window.Auth.showPermissionError('editor');
                                    }
                                });
                            });
                            return;
                        }
                        
                        // Check role and navigate
                        window.Auth.checkUserRole('editor').then(canEdit => {
                            if (canEdit) {
                                const url = new URL(window.location);
                                url.searchParams.set('mode', 'edit');
                                window.location.href = url.toString();
                            } else {
                                window.Auth.showPermissionError('editor');
                            }
                        });
                    });
                }
                return;

            case 'saveDocument':
                e.preventDefault();
                if (mainContent && mainContent.classList.contains('editing') && saveButton) {
                    saveButton.click();
                }
                return;

            case 'focusSearch':
                e.preventDefault();
                const searchBox = document.querySelector('.search-box');
                if (searchBox) {
                    searchBox.focus();
                }
                return;

            case 'togglePreview':
                e.preventDefault();
                if (mainContent && mainContent.classList.contains('editing')) {
                    if (window.EditorPreview && typeof window.EditorPreview.togglePreview === 'function') {
                        window.EditorPreview.togglePreview();
                    }
                }
                return;
        }
    }

    // Handle non-shortcut keys
    if (e.key === 'Escape') {
        handleEscapeKey(e);
    }
}

// Handle Escape key for various scenarios
function handleEscapeKey(e) {
    // Check the state of dialogs and UI
    const versionHistoryDialog = document.querySelector('.version-history-dialog');
    const isVersionHistoryDialogOpen = versionHistoryDialog && versionHistoryDialog.classList.contains('active');
    const isFileUploadDialogOpen = document.querySelector('.file-upload-dialog')?.classList.contains('active');
    const isLoginDialogOpen = document.querySelector('.login-dialog')?.classList.contains('active');
    const isMessageDialogOpen = document.querySelector('.message-dialog')?.classList.contains('active');
    const isDeleteConfirmDialogOpen = document.querySelector('.confirmation-dialog')?.classList.contains('active');
    const isUserConfirmDialogOpen = document.querySelector('.user-confirmation-dialog')?.classList.contains('active');
    const isNewDocDialogOpen = document.querySelector('.new-document-dialog')?.classList.contains('active');
    const isSettingsDialogOpen = document.querySelector('.settings-dialog')?.classList.contains('active');
    const isAccessRuleDialogOpen = document.querySelector('.access-rule-dialog')?.classList.contains('active');
    const isMoveDocDialogOpen = document.querySelector('.move-document-dialog')?.classList.contains('active');
    const isAddLinkDialogOpen = document.querySelector('.add-link-dialog')?.classList.contains('active');
    const isSearchResultsOpen = document.querySelector('.search-results')?.classList.contains('active');
    const isActionsMenuOpen = document.querySelector('.page-actions-menu')?.classList.contains('active');
    const isEditing = mainContent && mainContent.classList.contains('editing');

    // Handle in priority order (dialogs take precedence over edit mode)
    if (isLoginDialogOpen) {
        // Close login dialog
        window.Auth.hideLoginDialog();
        e.preventDefault();
    } else if (isMessageDialogOpen) {
        // Close message dialog
        window.DialogSystem.hideMessageDialog();
        e.preventDefault();
    } else if (isUserConfirmDialogOpen) {
        // Close user confirmation dialog
        window.DialogSystem.hideConfirmDialog();
        e.preventDefault();
    } else if (isVersionHistoryDialogOpen) {
        // Close version history dialog
        window.VersionHistory.hideVersionHistoryDialog();
        e.preventDefault();
        return; // Exit the event handler completely to prevent exiting edit mode
    } else if (isFileUploadDialogOpen) {
        // Close file upload dialog
        window.FileUpload.hideFileUploadDialog();
        e.preventDefault();
    } else if (isMoveDocDialogOpen) {
        // Close move document dialog
        window.MoveDocument.hideMoveDocDialog();
        e.preventDefault();
    } else if (isDeleteConfirmDialogOpen) {
        // Close delete confirmation dialog
        window.DocumentManager.hideConfirmationDialog();
        e.preventDefault();
    } else if (isNewDocDialogOpen) {
        // Close new document dialog
        window.DocumentManager.hideNewDocDialog();
        e.preventDefault();
    } else if (isAccessRuleDialogOpen) {
        // Close access rule dialog
        const dialog = document.querySelector('.access-rule-dialog');
        if (dialog) {
            dialog.classList.remove('active');
        }
        e.preventDefault();
    } else if (isSettingsDialogOpen) {
        // Close settings dialog
        window.SettingsManager.hideSettingsDialog();
        e.preventDefault();
    } else if (isAddLinkDialogOpen) {
        // Close add link dialog
        if (window.LinksInteractivity && window.LinksInteractivity.hideAddLinkDialog) {
            window.LinksInteractivity.hideAddLinkDialog();
        } else {
            // Fallback to direct DOM manipulation
            const dialog = document.querySelector('.add-link-dialog');
            if (dialog) {
                dialog.classList.remove('active');
            }
        }
        e.preventDefault();
    } else if (isSearchResultsOpen) {
        // Close search results
        window.Search.hideSearchResults();
        e.preventDefault();
    } else if (isActionsMenuOpen) {
        // Close actions menu
        document.querySelector('.page-actions-menu').classList.remove('active');
        e.preventDefault();
    } else if (isEditing) {
        e.preventDefault();
        // Check if there are unsaved changes
        if (window.WikiEditor && window.WikiEditor.hasUnsavedChanges && window.WikiEditor.hasUnsavedChanges()) {
            // Use the custom unsaved changes dialog
            if (window.WikiEditor.showUnsavedChangesDialog) {
                window.WikiEditor.showUnsavedChangesDialog(
                    // Save callback - save then navigate to view mode
                    function() {
                        const saveButton = document.querySelector('.save-changes');
                        if (saveButton) {
                            saveButton.click();
                        }
                    },
                    // Discard callback - navigate to view mode without saving
                    function() {
                        const url = new URL(window.location);
                        url.searchParams.delete('mode');
                        window.location.href = url.pathname + url.search;
                    }
                );
            }
        } else {
            // No unsaved changes, navigate to view mode
            const url = new URL(window.location);
            url.searchParams.delete('mode');
            window.location.href = url.pathname + url.search;
        }
    }
}

// Exit edit mode - use the editor's function directly
function exitEditMode() {
    if (window.WikiEditor && typeof window.WikiEditor.exitEditMode === 'function') {
        window.WikiEditor.exitEditMode(mainContent, editorContainer, viewToolbar, editToolbar);
    }
}

// Export module
window.KeyboardShortcuts = {
    init: initKeyboardShortcuts,
    handleKeyDown,
    handleEscapeKey,
    exitEditMode,
    registerFormattingCommands,
    registerAllCodeMirrorShortcuts  // Export this so it can be called after table editor loads
};

// Initialize keyboard shortcuts when DOM is loaded
document.addEventListener('DOMContentLoaded', initKeyboardShortcuts);