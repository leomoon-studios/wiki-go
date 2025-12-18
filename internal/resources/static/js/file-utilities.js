/**
 * File Utilities Module
 * Contains functions for handling file operations in Wiki-Go
 */

// Format file size in human-readable form
function formatFileSize(bytes) {
    if (bytes === undefined || bytes === null) {
        return 'Unknown size';
    }

    if (bytes < 1024) {
        return bytes + ' B';
    } else if (bytes < 1024 * 1024) {
        return (bytes / 1024).toFixed(1) + ' KB';
    } else {
        return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
    }
}

// Get file icon based on type
function getFileIcon(fileType) {
    let icon = '';

    // Default icon if fileType is undefined
    if (!fileType) {
        return '<i class="fa fa-file-o"></i>';
    }

    if (fileType.startsWith('image/')) {
        icon = '<i class="fa fa-file-image-o"></i>';
    } else if (fileType === 'application/pdf') {
        icon = '<i class="fa fa-file-pdf-o"></i>';
    } else if (fileType === 'application/zip') {
        icon = '<i class="fa fa-file-archive-o"></i>';
    } else if (fileType === 'text/plain') {
        icon = '<i class="fa fa-file-text-o"></i>';
    } else if (fileType === 'application/vnd.openxmlformats-officedocument.wordprocessingml.document') {
        icon = '<i class="fa fa-file-word-o"></i>';
    } else if (fileType === 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet') {
        icon = '<i class="fa fa-file-excel-o"></i>';
    } else if (fileType === 'application/vnd.openxmlformats-officedocument.presentationml.presentation') {
        icon = '<i class="fa fa-file-powerpoint-o"></i>';
    } else if (fileType === 'video/mp4' || fileType === 'video/quicktime' || fileType === 'video/webm' || fileType === 'video/avi') {
        icon = '<i class="fa fa-file-video-o"></i>';
    } else {
        icon = '<i class="fa fa-file-o"></i>';
    }

    return icon;
}

// Render files list HTML
function renderFilesList(files, mentionedFiles) {
    // Get filesList element first to prevent reference references
    const filesList = document.querySelector('.files-list');
    if (!filesList) {
        console.error('Files list element not found');
        return;
    }

    // Normalize mentionedFiles to a Set for fast lookup
    let mentionedSet = null;
    if (Array.isArray(mentionedFiles)) {
        mentionedSet = new Set(mentionedFiles.map(f => f.toLowerCase()));
        console.log('Mentioned files set:', mentionedSet);
    }

    console.log("Rendering files list:", files);

    if (!files || files.length === 0) {
        filesList.innerHTML = '<div class="empty-message">' + (window.i18n ? window.i18n.t('attachments.no_files') : 'No files found for this document.') + '</div>';
        return;
    }

    // Clear error message if it exists
    const errorMessage = document.querySelector('.error-message');
    if (errorMessage && errorMessage.style.display === 'block') {
        errorMessage.style.display = 'none';
    }

    const html = files.map(file => {
        // Debug each file in the console
        console.log("Processing file:", file);

        // Check if file is a string (just filename) or object
        let safeFile;
        if (typeof file === 'string') {
            // If file is just a string (filename), create object with defaults
            const filename = file;
            const fileExt = filename.split('.').pop().toLowerCase();

            // Use the globally defined FILE_EXTENSION_MIME_TYPES from base.html template
            let fileType = FILE_EXTENSION_MIME_TYPES[fileExt] || '';

            safeFile = {
                URL: `/api/files/${getCurrentDocPath()}/${filename}`,
                Type: fileType,
                Name: filename,
                Size: 0
            };
        } else {
            // Ensure all properties exist with defaults for missing ones
            safeFile = {
                URL: file.URL || `/api/files/${getCurrentDocPath()}/${file.Name || file.name || 'unknown'}`,
                Type: file.Type || file.type || '',
                Name: file.Name || file.name || 'Unknown file',
                Size: file.Size || file.size || 0
            };
        }

        console.log("Safe file object:", safeFile);

        // Prepare file path for deletion and rename operations
        const docPath = getCurrentDocPath();
        const filename = safeFile.Name;
        const filePath = `${docPath}/${filename}`;

        console.log("File operation path:", filePath);

        // Determine if file is an image
        const isImage = safeFile.Type && safeFile.Type.startsWith('image/');

        // Check if file is mentioned in document
        let highlightClass = '';
        if (mentionedSet && mentionedSet.has(filename.toLowerCase())) {
            highlightClass = ' mentioned-in-doc';
            console.log('File is mentioned:', filename);
        }

        return `
            <div class="file-item${highlightClass}" data-file-url="${safeFile.URL}">
                <input type="checkbox" class="file-select-checkbox" data-filename="${safeFile.Name}">
                <div class="file-info">
                    <div class="file-icon">${getFileIcon(safeFile.Type)}</div>
                    <div class="file-name" data-path="${filePath}" data-current-name="${safeFile.Name}">
                        <span class="name-text">${safeFile.Name}</span>
                        <input type="text" class="name-edit" value="${safeFile.Name}" style="display: none;">
                    </div>
                    <div class="file-size">${formatFileSize(safeFile.Size)}</div>
                </div>
                <div class="file-actions">
                    <button class="insert-file-btn" title="${window.i18n ? window.i18n.t('common.insert') : 'Insert into editor'}" data-url="${safeFile.URL}" data-is-image="${isImage}" data-name="${safeFile.Name}" data-i18n-title="common.insert">
                        <i class="fa fa-plus"></i>
                        <span data-i18n="common.insert">${window.i18n ? window.i18n.t('common.insert') : 'Insert'}</span>
                    </button>
                    <button class="view-file-btn" title="${window.i18n ? window.i18n.t('common.view') : 'View file'}" data-url="${safeFile.URL}" data-i18n-title="common.view">
                        <i class="fa fa-eye"></i>
                        <span data-i18n="common.view">${window.i18n ? window.i18n.t('common.view') : 'View'}</span>
                    </button>
                    <button class="rename-file-btn" title="${window.i18n ? window.i18n.t('common.rename') : 'Rename file'}" data-path="${filePath}" data-name="${safeFile.Name}" data-i18n-title="common.rename">
                        <i class="fa fa-pencil"></i>
                        <span data-i18n="common.rename">${window.i18n ? window.i18n.t('common.rename') : 'Rename'}</span>
                    </button>
                    <button class="delete-file-btn" title="${window.i18n ? window.i18n.t('common.delete') : 'Delete file'}" data-path="${filePath}" data-i18n-title="common.delete">
                        <i class="fa fa-trash"></i>
                        <span data-i18n="common.delete">${window.i18n ? window.i18n.t('common.delete') : 'Delete'}</span>
                    </button>
                </div>
            </div>
        `;
    }).join('');

    filesList.innerHTML = html;

    // Add event listeners for file actions
    filesList.querySelectorAll('.delete-file-btn').forEach(button => {
        button.addEventListener('click', () => {
            const path = button.getAttribute('data-path');
            deleteFile(path);
        });
    });

    // Add event listeners for rename buttons
    filesList.querySelectorAll('.rename-file-btn').forEach(button => {
        button.addEventListener('click', (e) => {
            e.stopPropagation();
            const fileNameElement = button.closest('.file-item').querySelector('.file-name');
            const nameText = fileNameElement.querySelector('.name-text');
            const nameEdit = fileNameElement.querySelector('.name-edit');

            // Get the filename and extension separately
            const fullName = fileNameElement.getAttribute('data-current-name');
            const lastDotIndex = fullName.lastIndexOf('.');
            const hasExtension = lastDotIndex > 0;
            const nameWithoutExt = hasExtension ? fullName.substring(0, lastDotIndex) : fullName;
            const extension = hasExtension ? fullName.substring(lastDotIndex) : '';

            // Show the edit input and hide the text
            nameText.style.display = 'none';
            nameEdit.style.display = 'block';
            nameEdit.focus();

            // Select only the name part without extension
            if (hasExtension) {
                // Select only the name part, not the extension
                nameEdit.setSelectionRange(0, nameWithoutExt.length);
            } else {
                // If no extension, select the whole name
                nameEdit.select();
            }
        });
    });

    // Add event listeners for the edit inputs
    filesList.querySelectorAll('.name-edit').forEach(input => {
        // Save on Enter, cancel on Escape
        input.addEventListener('keydown', async (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                const newName = input.value.trim();
                const fileNameElement = input.closest('.file-name');
                const path = fileNameElement.getAttribute('data-path');
                const currentName = fileNameElement.getAttribute('data-current-name');

                // Check if the current name has an extension
                const lastDotIndex = currentName.lastIndexOf('.');
                const hasExtension = lastDotIndex > 0;
                const extension = hasExtension ? currentName.substring(lastDotIndex) : '';

                // Only add extension if the user hasn't already included one
                let finalNewName = newName;
                if (hasExtension && !newName.includes('.')) {
                    finalNewName = newName + extension;
                }

                if (finalNewName && finalNewName !== currentName) {
                    input.disabled = true;
                    await renameFile(path, finalNewName);
                    input.disabled = false;
                }

                // Hide input, show text
                input.style.display = 'none';
                input.closest('.file-name').querySelector('.name-text').style.display = 'block';
            } else if (e.key === 'Escape') {
                // Cancel on Escape
                input.value = input.closest('.file-name').getAttribute('data-current-name');
                input.style.display = 'none';
                input.closest('.file-name').querySelector('.name-text').style.display = 'block';
            }
        });

        // Remove the blur event listener entirely - only use Enter key to save
        // This prevents double rename requests
    });

    // Add event listeners for view buttons
    filesList.querySelectorAll('.view-file-btn').forEach(button => {
        button.addEventListener('click', () => {
            const url = button.getAttribute('data-url');
            window.open(url, '_blank');
        });
    });

    // Add event listeners for insert buttons
    filesList.querySelectorAll('.insert-file-btn').forEach(button => {
        button.addEventListener('click', () => {
            const url = button.getAttribute('data-url');
            const isImage = button.getAttribute('data-is-image') === 'true';
            const name = button.getAttribute('data-name');
            handleFileInsertion(url, isImage, name);
        });
    });

    // Update file item click behavior
    filesList.querySelectorAll('.file-item').forEach(item => {
        // Remove the click event for opening files - now using the view button instead
        item.style.cursor = 'default';
    });

    // Initialize bulk actions
    initBulkActions(filesList);
}

function initBulkActions(filesList) {
    const selectUnreferencedBtn = document.getElementById('selectUnreferencedBtn');
    const deleteSelectedBtn = document.getElementById('deleteSelectedBtn');
    
    if (!selectUnreferencedBtn || !deleteSelectedBtn) return;

    // Update delete button state when checkboxes change
    const updateDeleteButtonState = () => {
        const checkedCount = filesList.querySelectorAll('.file-select-checkbox:checked').length;
        deleteSelectedBtn.disabled = checkedCount === 0;
        
        if (checkedCount > 0) {
            const text = window.i18n ? window.i18n.t('attachments.delete_selected_count', { count: checkedCount }) : `Delete Selected (${checkedCount})`;
            deleteSelectedBtn.textContent = text.replace('{{count}}', checkedCount);
        } else {
            deleteSelectedBtn.textContent = window.i18n ? window.i18n.t('attachments.delete_selected') : 'Delete Selected';
        }
    };

    // Add change listener to all checkboxes
    filesList.querySelectorAll('.file-select-checkbox').forEach(cb => {
        cb.addEventListener('change', updateDeleteButtonState);
    });

    // Initialize button state
    updateDeleteButtonState();

    // Select Unreferenced handler
    selectUnreferencedBtn.onclick = () => {
        filesList.querySelectorAll('.file-item').forEach(item => {
            const checkbox = item.querySelector('.file-select-checkbox');
            const isReferenced = item.classList.contains('mentioned-in-doc');
            if (checkbox) {
                checkbox.checked = !isReferenced;
            }
        });
        updateDeleteButtonState();
    };

    // Delete Selected handler
    deleteSelectedBtn.onclick = () => {
        const selectedFiles = Array.from(filesList.querySelectorAll('.file-select-checkbox:checked'))
            .map(cb => cb.getAttribute('data-filename'));
        
        if (selectedFiles.length === 0) return;

        const title = window.i18n ? window.i18n.t('attachments.delete_bulk_title') : 'Delete Files';
        let message = window.i18n ? window.i18n.t('attachments.delete_bulk_confirm', { count: selectedFiles.length }) : `Are you sure you want to delete these ${selectedFiles.length} files? This action cannot be undone.`;
        message = message.replace('{{count}}', selectedFiles.length);

        showConfirmDialog(
            title,
            message,
            async (confirmed) => {
                if (!confirmed) return;

                const docPath = getCurrentDocPath();
                let successCount = 0;
                let failCount = 0;

                for (const filename of selectedFiles) {
                    try {
                        const path = `${docPath}/${filename}`;
                        const deleteUrl = `/api/files/delete/${path}`;
                        
                        const response = await fetch(deleteUrl, {
                            method: 'DELETE'
                        });
                        if (response.ok) {
                            successCount++;
                        } else {
                            failCount++;
                        }
                    } catch (e) {
                        failCount++;
                    }
                }

                // Reload list
                await loadDocumentFiles();
                
                // Show result if there were failures
                if (failCount > 0) {
                    const resultTitle = window.i18n ? window.i18n.t('attachments.delete_results_title') : 'Delete Results';
                    let resultMsg = window.i18n ? window.i18n.t('attachments.delete_results_message', { success: successCount, failed: failCount }) : `Deleted ${successCount} files. Failed to delete ${failCount} files.`;
                    resultMsg = resultMsg.replace('{{success}}', successCount).replace('{{failed}}', failCount);
                    showMessageDialog(resultTitle, resultMsg);
                }
            }
        );
    };
}

// Load document files from API
async function loadDocumentFiles() {
    // Get filesList element first
    const filesList = document.querySelector('.files-list');
    if (!filesList) {
        console.error('Files list element not found');
        return;
    }

    try {
        const docPath = getCurrentDocPath();
        console.log("Loading files for document path:", docPath);
        const response = await fetch(`/api/files/list/${docPath}`);

        if (!response.ok) {
            throw new Error('Failed to fetch files');
        }

        const data = await response.json();
        console.log("API response for files:", data);

        if (!data.success) {
            throw new Error(data.message || 'Failed to fetch files');
        }

        // Get markdown content and extract mentioned files
        let markdown = '';
        if (typeof editor !== 'undefined' && editor && editor.getValue) {
            markdown = editor.getValue();
        } else {
            const codeMirrorElement = document.querySelector('.CodeMirror');
            if (codeMirrorElement && codeMirrorElement.CodeMirror) {
                markdown = codeMirrorElement.CodeMirror.getValue();
            } else {
                const textareas = document.querySelectorAll('textarea');
                for (const textarea of textareas) {
                    if (textarea.value && textarea.value.length > 0) {
                        markdown = textarea.value;
                        break;
                    }
                }
            }
        }

        const mentionedFiles = extractMentionedFilenamesFromMarkdown(markdown);
        renderFilesList(data.files || [], mentionedFiles);

        // Also update the file attachments section if it exists
        updateFileAttachments(data.files || []);
    } catch (error) {
        console.error('Error loading files:', error);
        if (filesList) {
            filesList.innerHTML = `<div class="empty-message">Error: ${error.message || 'Failed to load files'}</div>`;
        }
    }
}

// Helper function to update file attachments section
function updateFileAttachments(files) {
    const fileAttachmentsSection = document.querySelector('.file-attachments-section');
    const fileAttachmentsList = document.querySelector('.file-attachments-list');

    if (!fileAttachmentsSection || !fileAttachmentsList) return;

    console.log("Updating file attachments section:", files);

    if (!files || files.length === 0) {
        fileAttachmentsSection.style.display = 'none';
        return;
    }

    // Show the section since we have files
    fileAttachmentsSection.style.display = 'block';

    const html = files.map(file => {
        // Process file similar to the files list
        let safeFile;
        if (typeof file === 'string') {
            // If file is just a string (filename), create object with defaults
            const filename = file;
            const fileExt = filename.split('.').pop().toLowerCase();

            // Use the globally defined FILE_EXTENSION_MIME_TYPES from base.html template
            let fileType = FILE_EXTENSION_MIME_TYPES[fileExt] || '';

            safeFile = {
                URL: `/api/files/${getCurrentDocPath()}/${filename}`,
                Type: fileType,
                Name: filename,
                Size: 0
            };
        } else {
            // Ensure all properties exist with defaults for missing ones
            safeFile = {
                URL: file.URL || `/api/files/${getCurrentDocPath()}/${file.Name || file.name || 'unknown'}`,
                Type: file.Type || file.type || '',
                Name: file.Name || file.name || 'Unknown file',
                Size: file.Size || file.size || 0
            };
        }

        return `
            <a href="${safeFile.URL}" class="attachment-item" target="_blank" title="Open ${safeFile.Name}">
                <div class="attachment-icon">${getFileIcon(safeFile.Type)}</div>
                <div class="attachment-info">
                    <div class="attachment-name">${safeFile.Name}</div>
                    <div class="attachment-size">${formatFileSize(safeFile.Size)}</div>
                </div>
            </a>
        `;
    }).join('');

    fileAttachmentsList.innerHTML = html;
}

// Delete a file
async function deleteFile(path) {
    console.log("Attempting to delete file at path:", path);

    showConfirmDialog(
        window.i18n ? window.i18n.t('delete_file.title') : "Delete File",
        window.i18n ? window.i18n.t('delete_file.confirm_message') : "Are you sure you want to delete this file? This action cannot be undone.",
        async (confirmed) => {
            if (!confirmed) {
                return;
            }

            try {
                // Log the full delete URL for debugging
                const deleteUrl = `/api/files/delete/${path}`;
                console.log("DELETE request to:", deleteUrl);

                const response = await fetch(deleteUrl, {
                    method: 'DELETE'
                });

                console.log("Delete response status:", response.status);

                // Try to parse the response as JSON
                let data;
                try {
                    data = await response.json();
                    console.log("Delete response data:", data);
                } catch (jsonError) {
                    console.error("Error parsing JSON response:", jsonError);
                    // If we can't parse JSON, check if the response is OK
                    if (response.ok) {
                        // The request was successful even if we couldn't parse the JSON
                        loadDocumentFiles();
                        return;
                    } else {
                        // Response wasn't OK and we couldn't parse the JSON
                        throw new Error(`HTTP Error: ${response.status} ${response.statusText}`);
                    }
                }

                if (!data.success) {
                    throw new Error(data.message || 'Failed to delete file');
                }

                // Reload files list after successful deletion
                loadDocumentFiles();
            } catch (error) {
                console.error('Error deleting file:', error);
                window.DialogSystem.showMessageDialog("Delete Failed", `Error: ${error.message || 'Failed to delete file'}. Please check the console for more details.`);
            }
        }
    );
}

// Handle file insertion into the editor
function handleFileInsertion(url, isImage, name) {
    // Use the WikiEditor module from editor.js
    if (window.WikiEditor && window.WikiEditor.insertIntoEditor) {
        // Extract just the filename from the URL if it's a full path
        // The insertIntoEditor function now expects just the filename
        const filename = name;

        if (!window.WikiEditor.insertIntoEditor(url, isImage, filename)) {
            window.DialogSystem.showMessageDialog("Error", "Cannot insert file - editor is not active. Try opening the editor first.");
            return;
        }
        
        // Close the file upload dialog after successful insertion
        // if (window.FileUpload && window.FileUpload.hideFileUploadDialog) {
        //     window.FileUpload.hideFileUploadDialog();
        // }
        
        // Refresh the files list after a short delay to show the link icon for newly inserted files
        if (document.querySelector('.file-upload-dialog.active')) {
            // Only refresh if dialog is still open (shouldn't be, but just in case)
            loadAndHighlightFilesTab();
        }
        
    } else {
        window.DialogSystem.showMessageDialog("Error", "Cannot insert file - editor interface not available.");
        return;
    }
}

// Rename a file
async function renameFile(path, newName) {
    console.log("Attempting to rename file at path:", path, "to", newName);

    try {
        const response = await fetch('/api/files/rename', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                currentPath: path,
                newName: newName
            })
        });

        console.log("Rename response status:", response.status);

        // Check if we got a 404 (Not Found) status, but try to reload files anyway since
        // the file might have been renamed successfully despite the error
        if (response.status === 404) {
            console.log("Got 404 status, but checking if rename succeeded anyway");
            // Reload files list to see if rename actually succeeded
            await loadDocumentFiles();

            // Wait a moment to check if the renamed file appears in the list
            setTimeout(() => {
                // Find if any file in the list has the new name
                const filesList = document.querySelector('.files-list');
                if (filesList && filesList.querySelector(`.file-name[data-current-name="${newName}"]`)) {
                    console.log("File with new name found despite 404 error, rename likely succeeded");
                    // Show success message
                    if (window.DialogSystem && window.DialogSystem.showToast) {
                        window.DialogSystem.showToast(
                            window.i18n ? window.i18n.t('file_renamed.success') : "File renamed successfully",
                            "success"
                        );
                    }
                    return true;
                } else {
                    // If we can't find the file with the new name, show the error
                    throw new Error("File not found or rename failed");
                }
            }, 500);

            return true;
        }

        const data = await response.json();
        console.log("Rename response data:", data);

        if (!data.success) {
            console.error("Server reported error:", data.message);
            throw new Error(data.message || 'Failed to rename file');
        }

        // Show success message
        if (window.DialogSystem && window.DialogSystem.showToast) {
            window.DialogSystem.showToast(
                window.i18n ? window.i18n.t('file_renamed.success') : "File renamed successfully",
                "success"
            );
        } else {
            console.log("File renamed successfully (toast not available)");
        }

        // Reload files list after successful rename
        loadDocumentFiles();
        return true;
    } catch (error) {
        console.error('Error renaming file:', error);

        // Check if files list already contains a file with the new name,
        // which would indicate the rename succeeded despite an error response
        const filesList = document.querySelector('.files-list');
        if (filesList && filesList.querySelector(`.file-name[data-current-name="${newName}"]`)) {
            console.log("File with new name found in DOM, rename likely succeeded despite error");
            return true;
        }

        // Only show error dialog if we're certain the rename failed
        if (window.DialogSystem && window.DialogSystem.showMessageDialog) {
            window.DialogSystem.showMessageDialog(
                window.i18n ? window.i18n.t('file_renamed.error_title') : "Rename Failed",
                `${window.i18n ? window.i18n.t('file_renamed.error_message') : "Error"}: ${error.message || 'Failed to rename file'}.`
            );
        } else {
            alert("Error renaming file: " + (error.message || "Unknown error"));
        }

        return false;
    }
}

// Helper to extract mentioned filenames from markdown content
function extractMentionedFilenamesFromMarkdown(markdown) {
    if (!markdown) return [];
    
    // Updated regex to capture filenames with spaces: matches [text](filename with spaces.ext) and ![alt](filename with spaces.ext)
    const linkRegex = /!?\[[^\]]*\]\(([^)]+)\)/g;
    const matches = [];
    let match;
    
    console.log('Extracting from markdown:', markdown.substring(0, 500));
    
    while ((match = linkRegex.exec(markdown)) !== null) {
        const url = match[1].trim(); // Trim whitespace
        console.log('Found markdown link:', url);
        
        if (url && !url.startsWith('http') && !url.includes('://')) {
            // If it's just a filename or relative path
            const parts = url.split('/');
            const filename = parts[parts.length - 1];
            console.log('Extracted filename:', filename);
            matches.push(filename);
        }
    }
    
    console.log('All extracted filenames:', matches);
    return matches;
}

// Use the global getCurrentDocPath function from utilities.js

// Export functions
window.FileUtilities = {
    formatFileSize,
    getFileIcon,
    renderFilesList,
    loadDocumentFiles,
    deleteFile,
    renameFile,
    handleFileInsertion,
    updateFileAttachments,
    extractMentionedFilenamesFromMarkdown
};