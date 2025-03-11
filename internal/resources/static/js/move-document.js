// Move Document functionality
document.addEventListener('DOMContentLoaded', function() {
    const moveDocButton = document.querySelector('.move-document');
    const moveDocDialog = document.querySelector('.move-document-dialog');

    if (!moveDocButton || !moveDocDialog) {
        return; // Exit if elements don't exist
    }

    const moveDocForm = document.getElementById('moveDocumentForm');
    const closeMoveDocDialog = moveDocDialog.querySelector('.close-dialog');
    const cancelMoveDocButton = moveDocDialog.querySelector('.cancel-dialog');
    const moveDocErrorMessage = moveDocDialog.querySelector('.error-message');
    const moveSourcePathInput = document.getElementById('moveSourcePath');
    const moveTargetPathInput = document.getElementById('moveTargetPath');

    // Hide move button for homepage
    function updateMoveButtonVisibility() {
        const currentPath = getCurrentDocPath();
        if (currentPath === '' || currentPath === '/' || currentPath.toLowerCase() === 'homepage') {
            moveDocButton.style.display = 'none';
        } else {
            moveDocButton.style.display = '';
        }
    }

    // Update visibility when editor is loaded
    document.addEventListener('editor-loaded', updateMoveButtonVisibility);
    updateMoveButtonVisibility();

    // Show dialog when clicking Move Document button
    moveDocButton.addEventListener('click', function() {
        showMoveDocDialog();
    });

    // Function to show the move dialog
    function showMoveDocDialog() {
        // First check if user is an admin
        if (typeof checkIfUserIsAdmin === 'function') {
            checkIfUserIsAdmin().then(isAdmin => {
                if (!isAdmin) {
                    if (typeof showAdminOnlyError === 'function') {
                        showAdminOnlyError();
                    } else {
                        alert('This feature is only available to administrators.');
                    }
                    return;
                }

                showMoveDialogInternal();
            });
        } else {
            showMoveDialogInternal();
        }
    }

    // Internal function to show the dialog after admin check
    function showMoveDialogInternal() {
        // Don't show for homepage
        const currentPath = getCurrentDocPath();
        if (currentPath === '' || currentPath === '/' || currentPath.toLowerCase() === 'homepage') {
            showMessageDialog('Cannot Move Homepage', 'The homepage cannot be moved or renamed.');
            return;
        }

        moveDocDialog.classList.add('active');
        moveDocErrorMessage.style.display = 'none';
        moveDocForm.reset();

        // Pre-populate source path with current path
        moveSourcePathInput.value = currentPath;

        // Pre-populate target path with the current path for editing
        moveTargetPathInput.value = currentPath;

        // Focus on target path field
        setTimeout(() => {
            moveTargetPathInput.focus();
            moveTargetPathInput.select();
        }, 100);
    }

    // Define global hide function
    window.hideMoveDocDialog = function() {
        moveDocDialog.classList.remove('active');
    };

    // Close dialog when clicking close button or cancel
    if (closeMoveDocDialog) {
        closeMoveDocDialog.addEventListener('click', window.hideMoveDocDialog);
    }

    if (cancelMoveDocButton) {
        cancelMoveDocButton.addEventListener('click', window.hideMoveDocDialog);
    }

    // Handle form submission
    moveDocForm.addEventListener('submit', async function(e) {
        e.preventDefault();

        // Get values
        const sourcePath = moveSourcePathInput.value.trim();
        const targetPath = moveTargetPathInput.value.trim();

        // Validate
        if (!sourcePath) {
            moveDocErrorMessage.textContent = 'Source path is required';
            moveDocErrorMessage.style.display = 'block';
            return;
        }

        if (!targetPath) {
            moveDocErrorMessage.textContent = 'New path is required';
            moveDocErrorMessage.style.display = 'block';
            return;
        }

        // Extract the new slug from the target path
        const pathParts = targetPath.split('/');
        const newSlug = pathParts.pop();
        const newPath = pathParts.join('/');

        try {
            const response = await fetch('/api/document/move', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    sourcePath: sourcePath,
                    targetPath: newPath,
                    newSlug: newSlug
                })
            });

            if (response.ok) {
                const result = await response.json();
                if (result.success) {
                    // Redirect to the new document location
                    window.location.href = '/' + result.newPath;
                } else {
                    moveDocErrorMessage.textContent = result.message || 'Failed to move document';
                    moveDocErrorMessage.style.display = 'block';
                }
            } else {
                const errorData = await response.json().catch(() => null);
                if (errorData && errorData.message) {
                    // Check for specific error messages
                    if (errorData.message.includes("already exists")) {
                        moveDocErrorMessage.textContent = window.i18n ? window.i18n.t('move.target_exists') : 'Target already exists';
                    } else {
                        moveDocErrorMessage.textContent = errorData.message;
                    }
                } else {
                    moveDocErrorMessage.textContent = 'Failed to move document';
                }
                moveDocErrorMessage.style.display = 'block';
                console.error('Document move error:', errorData);
            }
        } catch (error) {
            console.error('Error moving document:', error);
            moveDocErrorMessage.textContent = 'An error occurred. Please try again.';
            moveDocErrorMessage.style.display = 'block';
        }
    });
});