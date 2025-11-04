/**
 * Editor Module - Main Coordinator
 * Orchestrates all editor functionality through modular components
 * 
 * This file has been refactored into multiple modules:
 * - editor-core.js: Core editor functionality and content management
 * - editor-toolbar.js: Toolbar creation and actions
 * - editor-pickers.js: Emoji, document, table, and anchor pickers
 * - editor-preview.js: Preview functionality
 * - editor-themes.js: Theme management and mobile handling
 */

// Ensure modules are loaded before using them
function ensureModulesLoaded() {
    const modules = ['EditorCore', 'EditorToolbar', 'EditorPickers', 'EditorPreview', 'EditorThemes'];
    for (const module of modules) {
        if (!window[module]) {
            console.warn(`Editor module ${module} not loaded yet`);
            return false;
        }
    }
    return true;
}

// Legacy compatibility functions - delegate to new modules
function loadEditor(mainContent, editorContainer, viewToolbar, editToolbar) {
    if (!ensureModulesLoaded()) {
        console.error('Editor modules not loaded');
        return;
    }
    return window.EditorCore.loadEditor(mainContent, editorContainer, viewToolbar, editToolbar);
}

function exitEditMode(mainContent, editorContainer, viewToolbar, editToolbar) {
    if (!window.EditorCore) return;
    return window.EditorCore.exitEditMode(mainContent, editorContainer, viewToolbar, editToolbar);
}

function getEditorContent() {
    if (!window.EditorCore) return '';
    return window.EditorCore.getEditorContent();
}

function insertIntoEditor(url, isImage, name) {
    if (!window.EditorCore) return false;
    return window.EditorCore.insertIntoEditor(url, isImage, name);
}

function insertRawContent(content) {
    if (!window.EditorCore) return false;
    return window.EditorCore.insertRawContent(content);
}

function isEditorActive() {
    if (!window.EditorCore) return false;
    return window.EditorCore.isEditorActive();
}

function hasUnsavedChanges() {
    if (!window.EditorCore) return false;
    return window.EditorCore.hasUnsavedChanges();
}

// Toolbar functions
function createToolbar(container) {
    if (!window.EditorToolbar) return null;
    return window.EditorToolbar.createToolbar(container);
}

function setupToolbarActions(toolbar) {
    if (!window.EditorToolbar) return;
    return window.EditorToolbar.setupToolbarActions(toolbar);
}

// Preview functions
function createPreview(container) {
    if (!window.EditorPreview) return null;
    return window.EditorPreview.createPreview(container);
}

function togglePreview() {
    if (!window.EditorPreview) return;
    return window.EditorPreview.togglePreview();
}

function updatePreview(content) {
    if (!window.EditorPreview) return;
    return window.EditorPreview.updatePreview(content);
}

// Picker functions
function showEmojiPicker(button) {
    if (!window.EditorPickers) return;
    return window.EditorPickers.showEmojiPicker(button);
}

function showDocPicker(button) {
    if (!window.EditorPickers) return;
    return window.EditorPickers.showDocPicker(button);
}

function showTablePicker(button) {
    if (!window.EditorPickers) return;
    return window.EditorPickers.showTablePicker(button);
}

function showAnchorPicker(button) {
    if (!window.EditorPickers) return;
    return window.EditorPickers.showAnchorPicker(button);
}

// Dialog functions for unsaved changes
function showUnsavedChangesDialog(saveCallback, discardCallback) {
    // Create a custom dialog if it doesn't exist
    let dialog = document.querySelector('.unsaved-changes-dialog');

    if (!dialog) {
        // Create dialog from scratch
        dialog = document.createElement('div');
        dialog.className = 'unsaved-changes-dialog user-confirmation-dialog';
        dialog.setAttribute('dir', 'auto');

        const container = document.createElement('div');
        container.className = 'dialog-container';

        const closeBtn = document.createElement('button');
        closeBtn.className = 'close-dialog';
        closeBtn.setAttribute('aria-label', 'Close confirmation dialog');
        closeBtn.innerHTML = '<i class="fa fa-times"></i>';

        const title = document.createElement('h2');
        title.className = 'dialog-title confirm-title';
        title.textContent = window.i18n ? window.i18n.t('editor.unsaved_changes') : 'Unsaved Changes';

        const content = document.createElement('p');
        content.className = 'dialog-message confirm-content';
        content.textContent = window.i18n ? window.i18n.t('editor.unsaved_changes_save') : 'You have unsaved changes. Do you want to save them before exiting?';

        const actions = document.createElement('div');
        actions.className = 'form-actions';

        const saveBtn = document.createElement('button');
        saveBtn.type = 'button';
        saveBtn.className = 'dialog-button confirm-save primary';
        saveBtn.textContent = window.i18n ? window.i18n.t('common.yes') : 'Save';

        const discardBtn = document.createElement('button');
        discardBtn.type = 'button';
        discardBtn.className = 'dialog-button confirm-discard';
        discardBtn.textContent = window.i18n ? window.i18n.t('common.no') : "Don't Save";

        const stayBtn = document.createElement('button');
        stayBtn.type = 'button';
        stayBtn.className = 'dialog-button confirm-stay';
        stayBtn.textContent = window.i18n ? window.i18n.t('common.cancel') : 'Cancel';

        // Assemble the dialog
        actions.appendChild(saveBtn);
        actions.appendChild(discardBtn);
        actions.appendChild(stayBtn);

        container.appendChild(closeBtn);
        container.appendChild(title);
        container.appendChild(content);
        container.appendChild(actions);

        dialog.appendChild(container);

        // Add to document
        document.body.appendChild(dialog);

        // Set up event listeners
        saveBtn.addEventListener('click', function() {
            hideUnsavedChangesDialog();
            if (typeof saveCallback === 'function') {
                saveCallback();
            }
        });

        discardBtn.addEventListener('click', function() {
            hideUnsavedChangesDialog();
            if (typeof discardCallback === 'function') {
                discardCallback();
            }
        });

        stayBtn.addEventListener('click', function() {
            hideUnsavedChangesDialog();
            // No callback - just stay in edit mode
        });

        closeBtn.addEventListener('click', function() {
            hideUnsavedChangesDialog();
            // No callback - just close the dialog and stay in edit mode
        });
    } else {
        // Update existing dialog text
        const title = dialog.querySelector('.confirm-title');
        const content = dialog.querySelector('.confirm-content');

        if (title) {
            title.textContent = window.i18n ? window.i18n.t('editor.unsaved_changes') : 'Unsaved Changes';
        }

        if (content) {
            content.textContent = window.i18n ? window.i18n.t('editor.unsaved_changes_save') : 'You have unsaved changes. Do you want to save them before exiting?';
        }

        // Update button event listeners
        const saveBtn = dialog.querySelector('.confirm-save');
        const discardBtn = dialog.querySelector('.confirm-discard');

        if (saveBtn) {
            const newSaveBtn = saveBtn.cloneNode(true);
            saveBtn.parentNode.replaceChild(newSaveBtn, saveBtn);
            newSaveBtn.addEventListener('click', function() {
                hideUnsavedChangesDialog();
                if (typeof saveCallback === 'function') {
                    saveCallback();
                }
            });
        }

        if (discardBtn) {
            const newDiscardBtn = discardBtn.cloneNode(true);
            discardBtn.parentNode.replaceChild(newDiscardBtn, discardBtn);
            newDiscardBtn.addEventListener('click', function() {
                hideUnsavedChangesDialog();
                if (typeof discardCallback === 'function') {
                    discardCallback();
                }
            });
        }
    }

    // Show the dialog
    dialog.classList.add('active');
}

// Function to hide the unsaved changes dialog
function hideUnsavedChangesDialog() {
    const dialog = document.querySelector('.unsaved-changes-dialog');
    if (dialog) {
        dialog.classList.remove('active');
    }
}

// Function to initialize edit controls
function initializeEditControls() {
    const editPageButton = document.querySelector('.edit-page');
    const saveButton = document.querySelector('.save-changes');
    const cancelButton = document.querySelector('.cancel-edit');
    const mainContent = document.querySelector('.content');
    const editorContainer = document.querySelector('.editor-container');
    const viewToolbar = document.querySelector('.view-toolbar');
    const editToolbar = document.querySelector('.edit-toolbar');
    const markdownContent = document.querySelector('.markdown-content');

    // Check if we're in edit mode (?mode=edit in URL)
    const urlParams = new URLSearchParams(window.location.search);
    const isEditMode = urlParams.get('mode') === 'edit';
    
    // Set up save and cancel buttons FIRST (needed in edit mode)
    // Save button functionality
    if (saveButton) {
        saveButton.addEventListener('click', async function() {
            try {
                const isHomepage = window.location.pathname === '/';
                const apiPath = isHomepage ? '/api/save/' : `/api/save${window.location.pathname}`;

                const content = getEditorContent();

                const response = await fetch(apiPath, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'text/plain',
                    },
                    body: content
                });

                if (!response.ok) throw new Error('Failed to save content');

                // Update originalContent to match what was just saved
                if (window.EditorCore) {
                    window.EditorCore.setOriginalContent(content);
                }

                // Navigate back to view mode (remove ?mode=edit)
                const url = new URL(window.location);
                url.searchParams.delete('mode');
                window.location.href = url.pathname + url.search;

            } catch (error) {
                console.error('Error:', error);
                alert('Failed to save changes');
            }
        });
    }

    // Cancel button functionality
    if (cancelButton) {
        cancelButton.addEventListener('click', function() {
            // Check if there are unsaved changes
            if (hasUnsavedChanges()) {
                // Show custom unsaved changes dialog
                showUnsavedChangesDialog(
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
            } else {
                // No unsaved changes, navigate to view mode
                const url = new URL(window.location);
                url.searchParams.delete('mode');
                window.location.href = url.pathname + url.search;
            }
        });
    }
    
    // If in edit mode, auto-load the editor
    if (isEditMode && mainContent && editorContainer && viewToolbar && editToolbar) {
        loadEditor(mainContent, editorContainer, viewToolbar, editToolbar);
        return; // Exit early, don't set up edit button handler
    }

    // Auto-enter edit mode if content is empty
    if (markdownContent && editPageButton) {
        const contentText = markdownContent.textContent.trim();
        const h1Only = markdownContent.children.length === 1 &&
                      markdownContent.children[0].tagName === 'H1';

        if (h1Only) {
            editPageButton.click();
        }
    }

    // Update edit button functionality - navigate to ?mode=edit
    if (editPageButton) {
        editPageButton.addEventListener('click', async function() {
            // Check authentication first
            const authResponse = await fetch('/api/check-auth');
            if (authResponse.status === 401) {
                // Show login dialog
                window.Auth.showLoginDialog(() => {
                    // After successful login, check role and navigate to edit mode
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
            
            // Check if user has editor/admin role
            const canEdit = await window.Auth.checkUserRole('editor');
            if (!canEdit) {
                window.Auth.showPermissionError('editor');
                return;
            }
            
            // Navigate to edit mode
            const url = new URL(window.location);
            url.searchParams.set('mode', 'edit');
            window.location.href = url.toString();
        });
    }
}

// Export the functions for global access
window.WikiEditor = {
    // Main functions
    loadEditor,
    exitEditMode,
    getEditorContent,
    insertIntoEditor,
    insertRawContent,
    isEditorActive,
    hasUnsavedChanges,
    
    // Dialog functions
    showUnsavedChangesDialog,
    hideUnsavedChangesDialog,
    
    // Initialization
    initializeEditControls
};

// Backward compatibility - ensure these functions are available globally
window.loadEditor = loadEditor;
window.exitEditMode = exitEditMode;
window.getEditorContent = getEditorContent;
window.insertIntoEditor = insertIntoEditor;
window.insertRawContent = insertRawContent;
window.isEditorActive = isEditorActive;
window.hasUnsavedChanges = hasUnsavedChanges;
window.showUnsavedChangesDialog = showUnsavedChangesDialog;
window.hideUnsavedChangesDialog = hideUnsavedChangesDialog;