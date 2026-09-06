// Document Management Module for Wiki-Go
// Handles document creation and deletion
(function() {
    'use strict';

    // Module state variables
    let newDocButton;
    let newDocDialog;
    let newDocForm;
    let closeNewDocDialog;
    let cancelNewDocButton;
    let docTitleInput;
    let docPathInput;
    let docSlugInput;
    let docTypeInput;
    let newDocErrorMessage;

    let deleteButton;
    let confirmationDialog;
    let deleteConfirmBtn;
    let cancelDeleteBtn;

    // Path autocomplete variables
    let pathAutocomplete;
    let pathSuggestions = [];
    let selectedPathIndex = -1;

    // Initialize module when DOM is loaded
    document.addEventListener('DOMContentLoaded', function() {
        // New Document functionality
        newDocButton = document.querySelector('.new-document');
        newDocDialog = document.querySelector('.new-document-dialog');
        newDocForm = document.getElementById('newDocumentForm');
        closeNewDocDialog = newDocDialog?.querySelector('.close-dialog');
        cancelNewDocButton = newDocDialog?.querySelector('.cancel-new-doc');
        docTitleInput = document.getElementById('docTitle');
        docPathInput = document.getElementById('docPath');
        docSlugInput = document.getElementById('docSlug');
        docTypeInput = document.getElementById('docType');
        newDocErrorMessage = newDocDialog?.querySelector('.error-message');

        // Delete Document functionality
        deleteButton = document.querySelector('.delete-document');
        confirmationDialog = document.querySelector('.confirmation-dialog');
        deleteConfirmBtn = document.querySelector('.confirmation-dialog .delete-confirm');
        cancelDeleteBtn = document.querySelector('.confirmation-dialog .cancel-delete');

        // Initialize components if they exist
        if (docTitleInput && docSlugInput) {
            // Note: Slug generation is now handled by slugify.js
            // No event listeners needed here
        }

        // Initialize new document button
        if (newDocButton) {
            newDocButton.addEventListener('click', async function() {
                try {
                    // Check if user is authenticated
                    const authResponse = await fetch('/api/check-auth');
                    if (authResponse.status === 401) {
                        // Show login dialog
                        window.Auth.showLoginDialog(() => {
                            // After login, check if user has editor or admin role
                            window.Auth.checkUserRole('editor').then(canEdit => {
                                if (canEdit) {
                                    showNewDocDialog();
                                    // Update toolbar buttons after login
                                    window.Auth.updateToolbarButtons();
                                } else {
                                    window.Auth.showPermissionError('editor');
                                }
                            });
                        });
                        return;
                    }

                    // User is authenticated, check if user has editor or admin role
                    const canEdit = await window.Auth.checkUserRole('editor');
                    if (canEdit) {
                        showNewDocDialog();
                    } else {
                        window.Auth.showPermissionError('editor');
                    }
                } catch (error) {
                    console.error('Error:', error);
                    alert('Failed to check authentication status');
                }
            });
        }

        // Close dialog when clicking close button or cancel
        if (closeNewDocDialog) {
            closeNewDocDialog.addEventListener('click', hideNewDocDialog);
        }

        if (cancelNewDocButton) {
            cancelNewDocButton.addEventListener('click', hideNewDocDialog);
        }

        // Escape key is now handled by keyboard-shortcuts.js

        // Handle new document form submission
        if (newDocForm) {
            newDocForm.addEventListener('submit', async function(e) {
                e.preventDefault();

                // Get values
                const title = docTitleInput.value.trim();
                let path = docPathInput.value.trim();
                let slug = docSlugInput.value.trim();
                const type = docTypeInput.value;

                // Validate - only title is required
                if (!title) {
                    newDocErrorMessage.textContent = 'Title is required';
                    newDocErrorMessage.style.display = 'block';
                    return;
                }

                // If slug is empty, use title for server-side transliteration
                if (!slug) {
                    slug = title;
                }

                // Remove leading and trailing slashes from path if it's not empty
                if (path) {
                    path = path.replace(/^\/|\/$/g, '');
                }

                // Combine path and slug
                // If path is empty, just use the slug (creates document at root level)
                const fullPath = path ? `${path}/${slug}` : slug;

                try {
                    const response = await fetch('/api/document/create', {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                        },
                        body: JSON.stringify({
                            title: title,
                            path: fullPath,
                            type: type
                        })
                    });

                    if (response.ok) {
                        const result = await response.json();
                        // Redirect to the new document
                        window.location.href = result.url;
                    } else {
                        const errorData = await response.json().catch(() => null);
                        if (errorData && errorData.message) {
                            let errorMsg = errorData.message;
                            // Check if it's a "Document already exists" error
                            if (errorMsg.includes("Document already exists")) {
                                errorMsg = window.i18n ? window.i18n.t('new_doc.already_exists') : 'Document already exists';
                            } else if (errorData.error) {
                                // Add error details if available
                                errorMsg += `: ${errorData.error}`;
                            }
                            newDocErrorMessage.textContent = errorMsg;
                        } else {
                            newDocErrorMessage.textContent = 'Failed to create document';
                        }
                        newDocErrorMessage.style.display = 'block';
                        console.error('Document creation error:', errorData);
                    }
                } catch (error) {
                    console.error('Error creating document:', error);
                    newDocErrorMessage.textContent = 'An error occurred. Please try again.';
                    newDocErrorMessage.style.display = 'block';
                }
            });
        }

        // Initialize delete document button
        if (deleteButton) {
            deleteButton.addEventListener('click', function() {
                showDeleteConfirmDialog();
            });
        }

        // Initialize delete confirmation buttons
        if (cancelDeleteBtn) {
            cancelDeleteBtn.addEventListener('click', function() {
                hideConfirmationDialog();
            });
        }

        if (deleteConfirmBtn) {
            deleteConfirmBtn.addEventListener('click', handleDocumentDeletion);
        }

        // Escape key is now handled by keyboard-shortcuts.js
    });

    // Show new document dialog
    function showNewDocDialog() {
        newDocDialog.classList.add('active');
        newDocErrorMessage.style.display = 'none';
        newDocForm.reset();

        // Pre-populate path with current path - make new doc a child of current doc
        const currentPath = window.location.pathname;
        if (currentPath && currentPath !== '/') {
            // Remove leading and trailing slashes
            let path = currentPath.replace(/^\/|\/$/g, '');
            // Use the full current path as the parent directory for the new document
            docPathInput.value = path;
        }

        // Initialize path autocomplete if not already done
        if (!pathAutocomplete) {
            initPathAutocomplete();
        }

        // Focus on title field
        setTimeout(() => {
            docTitleInput.focus();
        }, 100);
    }

    // Hide new document dialog
    function hideNewDocDialog() {
        newDocDialog.classList.remove('active');
        // Hide path autocomplete when dialog closes
        hidePathAutocomplete();
    }

    // Show delete confirmation dialog
    function showDeleteConfirmDialog() {
        confirmationDialog.classList.add('active');
    }

    // Hide delete confirmation dialog
    function hideConfirmationDialog() {
        confirmationDialog.classList.remove('active');
    }

    // Handle document deletion
    async function handleDocumentDeletion() {
        try {
            const isHomepage = window.location.pathname === '/';

            // Don't allow deleting the homepage
            if (isHomepage) {
                alert('The homepage cannot be deleted.');
                hideConfirmationDialog();
                return;
            }

            const apiPath = `/api/document${window.location.pathname}`;

            const response = await fetch(apiPath, {
                method: 'DELETE',
            });

            if (!response.ok) {
                const errorData = await response.json().catch(() => null);
                throw new Error(errorData?.message || 'Failed to delete document');
            }

            // Handle successful deletion
            // Redirect to parent directory
            const pathParts = window.location.pathname.split('/').filter(Boolean);
            pathParts.pop(); // Remove the last part (document name)
            const parentPath = pathParts.length > 0 ? '/' + pathParts.join('/') : '/';

            window.location.href = parentPath;
        } catch (error) {
            console.error('Error deleting document:', error);
            alert(error.message || 'Failed to delete document');
            hideConfirmationDialog();
        }
    }

    // ===== PATH AUTOCOMPLETE FUNCTIONALITY =====

    // Fetch folders at a specific level
    async function fetchFoldersAtLevel(basePath = '') {
        try {
            const response = await fetch('/api/documents/list');
            if (!response.ok) throw new Error('Failed to fetch documents');
            
            const data = await response.json();
            if (!data.success || !data.documents) return [];

            const folders = new Set();
            
            data.documents.forEach(doc => {
                if (!doc.path) return;
                
                const path = doc.path.replace(/^\/|\/$/g, ''); // Remove leading/trailing slashes
                if (!path) return;
                
                if (basePath) {
                    // Show folders that start with basePath
                    if (path.startsWith(basePath + '/')) {
                        const remainingPath = path.substring(basePath.length + 1);
                        const nextFolder = remainingPath.split('/')[0];
                        if (nextFolder) folders.add(basePath + '/' + nextFolder);
                    }
                } else {
                    // Show top-level folders
                    const topFolder = path.split('/')[0];
                    if (topFolder) folders.add(topFolder);
                }
            });
            
            return Array.from(folders).sort();
        } catch (error) {
            console.error('Error fetching folders:', error);
            return [];
        }
    }

    // Show path suggestions
    async function showPathSuggestions(query) {
        // Parse query to determine base path and filter text
        let basePath = '';
        let filterText = '';
        
        if (query.includes('/')) {
            const lastSlashIndex = query.lastIndexOf('/');
            basePath = query.substring(0, lastSlashIndex);
            filterText = query.substring(lastSlashIndex + 1);
        } else {
            filterText = query;
        }

        // Fetch folders at current level
        const availableFolders = await fetchFoldersAtLevel(basePath);
        let filtered = [];
        
        if (filterText) {
            // Find folders that start with filter text
            const matchingFolders = availableFolders.filter(path => {
                const folderName = path.split('/').pop();
                return folderName.toLowerCase().startsWith(filterText.toLowerCase());
            });
            
            if (matchingFolders.length === 1) {
                const exactMatch = matchingFolders[0];
                const folderName = exactMatch.split('/').pop();
                const isExactMatch = folderName.toLowerCase() === filterText.toLowerCase();
                
                if (isExactMatch) {
                    // Show only children for exact match
                    filtered = await fetchFoldersAtLevel(exactMatch);
                } else {
                    // Show folder and its children for partial match
                    const children = await fetchFoldersAtLevel(exactMatch);
                    filtered = [exactMatch, ...children].sort((a, b) => {
                        if (a === exactMatch) return -1;
                        if (b === exactMatch) return 1;
                        return a.localeCompare(b);
                    });
                }
            } else {
                // Multiple matches - show only matching folders
                filtered = matchingFolders.sort((a, b) => {
                    const aName = a.split('/').pop().toLowerCase();
                    const bName = b.split('/').pop().toLowerCase();
                    const filterLower = filterText.toLowerCase();
                    
                    // Exact matches first, then by length, then alphabetical
                    if (aName === filterLower && bName !== filterLower) return -1;
                    if (bName === filterLower && aName !== filterLower) return 1;
                    if (aName.length !== bName.length) return aName.length - bName.length;
                    return a.localeCompare(b);
                });
            }
        } else {
            // Show all folders at current level
            filtered = availableFolders;
        }

        if (filtered.length === 0) {
            hidePathAutocomplete();
            return;
        }

        // Create dropdown items
        WikiDOM.clear(pathAutocomplete);
        selectedPathIndex = -1;

        filtered.forEach((path, index) => {
            const item = document.createElement('div');
            item.className = 'path-autocomplete-item';
            
            const folderName = path.split('/').pop();
            const displayText = String(folderName ?? '');

            // Highlight matching text with text nodes so folder names cannot become markup.
            if (filterText) {
                const regex = new RegExp(filterText.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'gi');
                let lastIndex = 0;
                let match;
                while ((match = regex.exec(displayText)) !== null) {
                    item.appendChild(document.createTextNode(displayText.slice(lastIndex, match.index)));
                    item.appendChild(WikiDOM.element('strong', '', match[0]));
                    lastIndex = match.index + match[0].length;
                }
                item.appendChild(document.createTextNode(displayText.slice(lastIndex)));
            } else {
                item.textContent = displayText;
            }
            
            item.dataset.fullPath = path;
            
            item.addEventListener('mouseenter', () => {
                clearSelection();
                item.classList.add('selected');
                selectedPathIndex = index;
            });
            
            item.addEventListener('click', () => selectPath(path));
            pathAutocomplete.appendChild(item);
        });

        // Position dropdown
        const inputRect = docPathInput.getBoundingClientRect();
        const containerRect = newDocDialog.querySelector('.dialog-container').getBoundingClientRect();
        
        pathAutocomplete.style.top = `${inputRect.bottom - containerRect.top + 2}px`;
        pathAutocomplete.style.left = `${inputRect.left - containerRect.left}px`;
        pathAutocomplete.style.width = `${inputRect.width}px`;
        pathAutocomplete.style.display = 'block';
    }

    // Hide path autocomplete
    function hidePathAutocomplete() {
        if (pathAutocomplete) {
            pathAutocomplete.style.display = 'none';
        }
        selectedPathIndex = -1;
    }

    // Clear selection styling
    function clearSelection() {
        const items = pathAutocomplete.querySelectorAll('.path-autocomplete-item');
        items.forEach(item => item.classList.remove('selected'));
    }

    // Select a path
    function selectPath(path) {
        docPathInput.value = path;
        hidePathAutocomplete();
        docPathInput.focus();
    }

    // Handle keyboard navigation
    function handlePathKeydown(e) {
        const items = pathAutocomplete.querySelectorAll('.path-autocomplete-item');
        
        if (pathAutocomplete.style.display !== 'block' || items.length === 0) {
            return;
        }

        switch (e.key) {
            case 'ArrowDown':
                e.preventDefault();
                selectedPathIndex = Math.min(selectedPathIndex + 1, items.length - 1);
                updateSelection(items);
                break;
                
            case 'ArrowUp':
                e.preventDefault();
                selectedPathIndex = Math.max(selectedPathIndex - 1, -1);
                updateSelection(items);
                break;
                
            case 'Enter':
                if (selectedPathIndex >= 0 && items[selectedPathIndex]) {
                    e.preventDefault();
                    const selectedPath = items[selectedPathIndex].dataset.fullPath;
                    selectPath(selectedPath);
                }
                break;
                
            case 'Escape':
                e.preventDefault();
                hidePathAutocomplete();
                break;
        }
    }

    // Update visual selection
    function updateSelection(items) {
        clearSelection();
        if (selectedPathIndex >= 0 && items[selectedPathIndex]) {
            items[selectedPathIndex].classList.add('selected');
            items[selectedPathIndex].scrollIntoView({ block: 'nearest' });
        }
    }

    // Initialize path autocomplete
    async function initPathAutocomplete() {
        if (!docPathInput) return;

        // Create autocomplete element
        pathAutocomplete = document.createElement('div');
        pathAutocomplete.className = 'path-autocomplete';
        
        // Add to dialog container
        const dialogContainer = newDocDialog.querySelector('.dialog-container');
        if (dialogContainer) {
            dialogContainer.style.position = 'relative';
            dialogContainer.appendChild(pathAutocomplete);
        }

        // Event listeners
        docPathInput.addEventListener('input', (e) => showPathSuggestions(e.target.value));
        docPathInput.addEventListener('focus', () => showPathSuggestions(docPathInput.value));
        docPathInput.addEventListener('keydown', handlePathKeydown);

        // Hide on outside click
        document.addEventListener('click', (e) => {
            if (!docPathInput.contains(e.target) && !pathAutocomplete.contains(e.target)) {
                hidePathAutocomplete();
            }
        });
    }

    // Expose public API
    window.DocumentManager = {
        showNewDocDialog: showNewDocDialog,
        hideNewDocDialog: hideNewDocDialog,
        showDeleteConfirmDialog: showDeleteConfirmDialog,
        hideConfirmationDialog: hideConfirmationDialog
    };
})();
