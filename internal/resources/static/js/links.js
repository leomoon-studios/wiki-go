/**
 * Links Document Interactive Functionality
 * Handles copy links, edit/delete actions, and add link dialog
 */

(function() {
    'use strict';

    // Module state
    let isInitialized = false;

    // Initialize links functionality when DOM is loaded
    document.addEventListener('DOMContentLoaded', function() {
        // Only initialize on links documents
        if (!document.querySelector('.links-container')) {
            return;
        }

        initLinksInteractivity();
        isInitialized = true;
    });

    /**
     * Initialize all links interactive features
     */
    function initLinksInteractivity() {
        console.log('Initializing links document interactivity');
        
        // Get user role and add role class to body for CSS visibility
        const userRole = document.querySelector('meta[name="user-role"]')?.content || 'viewer';
        document.body.classList.add(`role-${userRole}`);
        console.log('Links init - Role:', userRole);
        
        // Initialize search and filtering functionality
        initSearchAndFiltering();
        
        // Add action buttons to all link items dynamically (like kanban tasks)
        addActionButtonsToLinks();
        
        // Set up floating add link button (admin/editor only)
        setupAddLinkButton();
    }

    /**
     * Add action buttons to all link items dynamically (like kanban tasks)
     */
    function addActionButtonsToLinks() {
        const linkItems = document.querySelectorAll('.link-item');
        
        linkItems.forEach(linkItem => {
            // Check if action buttons already exist
            if (linkItem.querySelector('.link-action-buttons')) {
                return;
            }

            // Get link data from the item
            const linkElement = linkItem.querySelector('.link-title a[href]');
            const titleElement = linkItem.querySelector('.link-title a');
            const descriptionElement = linkItem.querySelector('.link-description');
            
            if (!linkElement || !titleElement) {
                console.error('Could not find link data for action buttons');
                return;
            }

            const url = linkElement.getAttribute('href');
            const title = titleElement.textContent;
            const description = descriptionElement ? descriptionElement.textContent : '';
            
            // Find category from the nearest category header
            const categoryHeader = linkItem.closest('.links-category')?.querySelector('.links-category-header');
            const category = categoryHeader ? categoryHeader.textContent.replace(/\s*\(\d+\)\s*$/, '').trim() : '';

            // Create action buttons container
            const actionButtons = document.createElement('div');
            actionButtons.className = 'link-action-buttons';

            // Add copy button (available to all users)
            const copyBtn = document.createElement('button');
            copyBtn.className = 'link-action-btn link-copy-btn';
            copyBtn.title = 'Copy link';
            copyBtn.innerHTML = '<i class="fa fa-copy"></i>';
            copyBtn.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                copyLinkToClipboard(url);
            });

            // Add edit button (admin/editor only)
            const editBtn = document.createElement('button');
            editBtn.className = 'link-action-btn link-edit-btn editor-admin-only';
            editBtn.title = 'Edit link';
            editBtn.innerHTML = '<i class="fa fa-pencil"></i>';
            editBtn.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                editLink(title, url, description, category);
            });

            // Add delete button (admin/editor only)
            const deleteBtn = document.createElement('button');
            deleteBtn.className = 'link-action-btn link-delete-btn editor-admin-only';
            deleteBtn.title = 'Delete link';
            deleteBtn.innerHTML = '<i class="fa fa-trash"></i>';
            deleteBtn.addEventListener('click', (e) => {
                e.preventDefault();
                e.stopPropagation();
                deleteLink(url, title, description, category);
            });

            // Add buttons to container
            actionButtons.appendChild(copyBtn);
            actionButtons.appendChild(editBtn);
            actionButtons.appendChild(deleteBtn);

            // Add container to link item
            linkItem.appendChild(actionButtons);
        });
        
        console.log(`Added action buttons to ${linkItems.length} links`);
    }

    /**
     * Set up copy link functionality for all copy buttons
     */
    function setupCopyLinkButtons() {
        // This function is now handled by addActionButtonsToLinks()
        // Keeping for backward compatibility
    }

    /**
     * Set up action buttons for edit/delete functionality  
     */
    function setupActionButtons() {
        // This function is now handled by addActionButtonsToLinks()
        // Keeping for backward compatibility
    }

    /**
     * Handle copying a link URL to clipboard
     */
    async function handleCopyLink(event) {
        event.preventDefault();
        event.stopPropagation();
        
        // Find the link URL from the nearest link item
        const linkItem = event.target.closest('.link-item');
        if (!linkItem) {
            console.error('Could not find link item for copy button');
            return;
        }
        
        const linkElement = linkItem.querySelector('.link-title a[href]');
        if (!linkElement) {
            console.error('Could not find link URL');
            return;
        }
        
        const url = linkElement.getAttribute('href');
        
        try {
            // Try modern clipboard API first
            await navigator.clipboard.writeText(url);
            showCopyFeedback(event.target, true, url);
        } catch (err) {
            console.warn('Clipboard API failed, trying fallback method:', err);
            
            // Fallback for older browsers or insecure contexts
            const success = copyToClipboardFallback(url);
            showCopyFeedback(event.target, success, url);
        }
    }

    /**
     * Fallback clipboard method for older browsers
     */
    function copyToClipboardFallback(text) {
        try {
            const textarea = document.createElement('textarea');
            textarea.value = text;
            textarea.setAttribute('readonly', '');
            textarea.style.position = 'absolute';
            textarea.style.left = '-9999px';
            document.body.appendChild(textarea);
            
            textarea.select();
            const success = document.execCommand('copy');
            
            document.body.removeChild(textarea);
            return success;
        } catch (err) {
            console.error('Fallback clipboard method failed:', err);
            return false;
        }
    }

    /**
     * Show visual feedback after copy operation
     */
    function showCopyFeedback(button, success, url) {
        const originalIcon = button.innerHTML;
        
        if (success) {
            button.innerHTML = '<i class="fa fa-check"></i>';
            button.classList.add('copied');
            
            // Reset button after 2 seconds
            setTimeout(() => {
                button.innerHTML = originalIcon;
                button.classList.remove('copied');
            }, 2000);
        } else {
            button.innerHTML = '<i class="fa fa-times"></i>';
            button.classList.add('copy-failed');
            
            // Reset button after 2 seconds
            setTimeout(() => {
                button.innerHTML = originalIcon;
                button.classList.remove('copy-failed');
            }, 2000);
        }
    }

    /**
     * Set up edit and delete action buttons
     */
    function setupActionButtons() {
        // Set up edit buttons
        const editButtons = document.querySelectorAll('.link-edit-btn');
        editButtons.forEach(button => {
            button.addEventListener('click', handleEditLink);
        });
        
        // Set up delete buttons
        const deleteButtons = document.querySelectorAll('.link-delete-btn');
        deleteButtons.forEach(button => {
            button.addEventListener('click', handleDeleteLink);
        });
        
        console.log(`Set up ${editButtons.length} edit buttons and ${deleteButtons.length} delete buttons`);
    }

    /**
     * Handle edit link button click
     */
    function handleEditLink(event) {
        event.preventDefault();
        event.stopPropagation();
        
        const linkItem = event.target.closest('.link-item');
        if (!linkItem) {
            console.error('Could not find link item for edit button');
            return;
        }
        
        // Extract link data from the DOM
        const linkData = extractLinkDataFromDOM(linkItem);
        if (!linkData) {
            console.error('Could not extract link data');
            return;
        }
        
        // Show edit dialog with pre-filled data
        showAddLinkDialog(linkData);
    }

    /**
     * Handle delete link button click
     */
    function handleDeleteLink(event) {
        event.preventDefault();
        event.stopPropagation();
        
        const linkItem = event.target.closest('.link-item');
        if (!linkItem) {
            console.error('Could not find link item for delete button');
            return;
        }
        
        const linkData = extractLinkDataFromDOM(linkItem);
        if (!linkData) {
            console.error('Could not extract link data');
            return;
        }
        
        // Show confirmation dialog
        const title = 'Delete Link';
        const message = `Are you sure you want to delete "${linkData.title}"?`;
        
        if (window.DialogSystem && window.DialogSystem.showConfirmDialog) {
            window.DialogSystem.showConfirmDialog(title, message, (confirmed) => {
                if (confirmed) {
                    deleteLinkFromDocument(linkData);
                }
            });
        } else {
            // Fallback to native confirm
            if (confirm(`${title}\n\n${message}`)) {
                deleteLinkFromDocument(linkData);
            }
        }
    }

    /**
     * Extract link data from DOM element
     */
    function extractLinkDataFromDOM(linkItem) {
        const linkElement = linkItem.querySelector('.link-title a[href]');
        const descriptionElement = linkItem.querySelector('.link-description');
        
        if (!linkElement) {
            return null;
        }
        
        // Find category by looking at the parent category container
        let category = 'Uncategorized';
        
        // First, try to get the category from the data attribute
        const categoryContainer = linkItem.closest('.links-category');
        if (categoryContainer) {
            const dataCategory = categoryContainer.getAttribute('data-category');
            if (dataCategory) {
                category = dataCategory;
            } else {
                // Fallback: look for the category header within the same container
                const categoryHeader = categoryContainer.querySelector('.links-category-header');
                if (categoryHeader) {
                    const categoryText = categoryHeader.textContent.trim();
                    
                    // Remove count from category name (e.g., "Work Tools (3)" -> "Work Tools")
                    category = categoryText.replace(/\s*\(\d+\)\s*$/, '').trim();
                }
            }
        } else {
            // Try to get category from link item's own data attribute
            const itemCategory = linkItem.getAttribute('data-category');
            if (itemCategory) {
                category = itemCategory;
            }
        }
        
        return {
            title: linkElement.textContent.trim(),
            url: linkElement.getAttribute('href'),
            description: descriptionElement ? descriptionElement.textContent.trim() : '',
            category: category
        };
    }

    /**
     * Set up floating add link button
     */
    function setupAddLinkButton() {
        const addButton = document.querySelector('.floating-add-link-btn');
        if (addButton) {
            addButton.addEventListener('click', () => {
                showAddLinkDialog();
            });
            console.log('Set up floating add link button');
        }
    }

    /**
     * Show add/edit link dialog
     */
    function showAddLinkDialog(editData = null) {
        const dialog = document.querySelector('.add-link-dialog');
        if (!dialog) {
            console.error('Add link dialog not found');
            return;
        }

        const form = dialog.querySelector('#addLinkForm');
        const errorMessage = dialog.querySelector('.error-message');
        const dialogTitle = dialog.querySelector('.dialog-title');
        const submitButton = dialog.querySelector('.dialog-button.primary');
        
        // Set dialog title and button text based on mode
        const isEdit = editData && editData.url;
        if (dialogTitle) {
            dialogTitle.textContent = isEdit ? 
                (window.i18n ? window.i18n.t('links.edit_link_title') : 'Edit Link') :
                (window.i18n ? window.i18n.t('links.add_link_title') : 'Add Link');
        }
        if (submitButton) {
            submitButton.textContent = isEdit ?
                (window.i18n ? window.i18n.t('links.update_button') : 'Update Link') :
                (window.i18n ? window.i18n.t('links.add_button') : 'Add Link');
        }

        // Reset form and error message
        form.reset();
        errorMessage.style.display = 'none';
        errorMessage.textContent = '';

        // Populate category dropdown
        populateCategoryDropdown();

        // If editing, populate form with existing data
        if (isEdit) {
            document.getElementById('linkUrl').value = editData.url || '';
            document.getElementById('linkTitle').value = editData.title || '';
            document.getElementById('linkDescription').value = editData.description || '';
            
            // Set category if provided
            if (editData.category) {
                const categorySelect = document.getElementById('linkCategory');
                categorySelect.value = editData.category;
            }
        }

        // Set up form submission handler
        form.onsubmit = function(e) {
            e.preventDefault();
            handleLinkFormSubmission(isEdit, editData);
        };

        // Set up close handlers
        const closeButton = dialog.querySelector('.close-dialog');
        const cancelButton = dialog.querySelector('.cancel-dialog');
        
        closeButton.onclick = cancelButton.onclick = function() {
            hideAddLinkDialog();
        };

        // Set up URL metadata fetching
        setupUrlMetadataFetching();

        // Set up category selection logic
        setupCategorySelection();

        // Show dialog
        dialog.classList.add('active');

        // Focus on URL field if adding, title field if editing
        setTimeout(() => {
            if (isEdit) {
                document.getElementById('linkTitle').focus();
            } else {
                document.getElementById('linkUrl').focus();
            }
        }, 100);
    }

    /**
     * Hide the add link dialog
     */
    function hideAddLinkDialog() {
        const dialog = document.querySelector('.add-link-dialog');
        if (dialog) {
            dialog.classList.remove('active');
        }
    }

    /**
     * Populate the category dropdown with existing categories
     */
    function populateCategoryDropdown() {
        const categorySelect = document.getElementById('linkCategory');
        if (!categorySelect) return;

        // Clear existing options except the first placeholder
        const firstOption = categorySelect.querySelector('option[value=""]');
        categorySelect.innerHTML = '';
        if (firstOption) {
            categorySelect.appendChild(firstOption);
        }

        // Get existing categories
        const existingCategories = getExistingCategories();
        
        // Add existing categories
        existingCategories.forEach(category => {
            const option = document.createElement('option');
            option.value = category;
            option.textContent = category;
            categorySelect.appendChild(option);
        });

        // Add "Create New Category" option
        const newCategoryOption = document.createElement('option');
        newCategoryOption.value = '__new__';
        newCategoryOption.textContent = window.i18n ? window.i18n.t('links.new_category') : 'Create New Category';
        categorySelect.appendChild(newCategoryOption);
    }

    /**
     * Set up category selection logic
     */
    function setupCategorySelection() {
        const categorySelect = document.getElementById('linkCategory');
        const newCategoryWrapper = document.querySelector('.new-category-wrapper');
        
        if (!categorySelect || !newCategoryWrapper) return;

        categorySelect.onchange = function() {
            if (this.value === '__new__') {
                newCategoryWrapper.style.display = 'block';
                document.getElementById('newCategoryInput').focus();
            } else {
                newCategoryWrapper.style.display = 'none';
            }
        };
    }

    /**
     * Set up URL metadata fetching
     */
    function setupUrlMetadataFetching() {
        const fetchButton = document.querySelector('.fetch-details-btn');
        if (!fetchButton) return;

        // Add click handler to the fetch button
        fetchButton.addEventListener('click', async function() {
            const urlInput = document.getElementById('linkUrl');
            const url = urlInput ? urlInput.value.trim() : '';
            
            if (!url || !isValidUrl(url)) {
                return;
            }

            await fetchUrlMetadata(url);
        });
    }

    /**
     * Validate if a string is a valid URL
     */
    function isValidUrl(string) {
        try {
            const url = new URL(string);
            return url.protocol === 'http:' || url.protocol === 'https:';
        } catch (_) {
            return false;
        }
    }

    /**
     * Fetch metadata for a URL
     */
    async function fetchUrlMetadata(url) {
        try {
            // Validate URL format
            new URL(url);
            
            const titleInput = document.getElementById('linkTitle');
            const descriptionInput = document.getElementById('linkDescription');
            const fetchButton = document.querySelector('.fetch-details-btn');
            
            if (!titleInput) {
                console.warn('Title input not found');
                return;
            }

            console.log('Starting metadata fetch for:', url);

            // Capture original button content before any changes
            const originalButtonContent = '<i class="fa fa-download"></i> ' + 
                (window.i18n ? window.i18n.t('links.fetch_details') : 'Fetch Link Details');

            // Show loading state
            const originalTitlePlaceholder = titleInput.placeholder;
            const originalDescPlaceholder = descriptionInput ? descriptionInput.placeholder : '';
            
            titleInput.placeholder = window.i18n ? window.i18n.t('links.fetching_metadata') : 'Fetching page information...';
            titleInput.disabled = true;
            
            if (descriptionInput) {
                descriptionInput.placeholder = 'Fetching description...';
                descriptionInput.disabled = true;
            }

            if (fetchButton) {
                fetchButton.disabled = true;
                fetchButton.innerHTML = '<i class="fa fa-spinner fa-spin"></i> Fetching...';
            }

            // Try to fetch page metadata
            const metadata = await fetchPageMetadata(url);
            
            console.log('Metadata fetch result:', metadata);
            
            if (metadata) {
                // Set title if fetched successfully
                if (metadata.title) {
                    console.log('Setting title:', metadata.title);
                    titleInput.value = metadata.title;
                }
                
                // Set description if fetched successfully
                if (metadata.description && descriptionInput) {
                    console.log('Setting description:', metadata.description);
                    descriptionInput.value = metadata.description;
                }

                // Show success feedback
                if (fetchButton) {
                    fetchButton.innerHTML = '<i class="fa fa-check"></i> Fetched!';
                    setTimeout(() => {
                        fetchButton.innerHTML = originalButtonContent;
                    }, 2000);
                }
            } else {
                console.log('No metadata found, using fallback');
                // Fallback: extract domain as title
                const urlObj = new URL(url);
                const domain = urlObj.hostname.replace(/^www\./, '');
                const fallbackTitle = domain.split('.')[0];
                
                const finalTitle = fallbackTitle.charAt(0).toUpperCase() + fallbackTitle.slice(1);
                console.log('Setting fallback title:', finalTitle);
                titleInput.value = finalTitle;

                // Show partial success feedback
                if (fetchButton) {
                    fetchButton.innerHTML = '<i class="fa fa-exclamation-triangle"></i> Partial';
                    setTimeout(() => {
                        fetchButton.innerHTML = originalButtonContent;
                    }, 2000);
                }
            }
            
        } catch (error) {
            console.warn('Error fetching metadata:', url, error);
            
            // Show error feedback
            const fetchButton = document.querySelector('.fetch-details-btn');
            if (fetchButton) {
                fetchButton.innerHTML = '<i class="fa fa-times"></i> Failed';
                setTimeout(() => {
                    const originalContent = '<i class="fa fa-download"></i> ' + 
                        (window.i18n ? window.i18n.t('links.fetch_details') : 'Fetch Link Details');
                    fetchButton.innerHTML = originalContent;
                }, 2000);
            }
        } finally {
            // Restore input state
            const titleInput = document.getElementById('linkTitle');
            const descriptionInput = document.getElementById('linkDescription');
            const fetchButton = document.querySelector('.fetch-details-btn');
            
            if (titleInput) {
                titleInput.disabled = false;
                titleInput.placeholder = window.i18n ? window.i18n.t('links.title_placeholder') : 'Link title';
            }
            
            if (descriptionInput) {
                descriptionInput.disabled = false;
                descriptionInput.placeholder = window.i18n ? window.i18n.t('links.description_placeholder') : 'Brief description of the link';
            }

            if (fetchButton) {
                fetchButton.disabled = false;
            }
            
            console.log('Metadata fetch completed for:', url);
        }
    }

    /**
     * Fetch page metadata using server-side endpoint
     */
    async function fetchPageMetadata(url) {
        try {
            console.log('Fetching metadata from server for:', url);
            
            const response = await fetch('/api/links/fetch-metadata', {
                method: 'POST',
                headers: {
                    'Accept': 'application/json',
                    'Content-Type': 'application/json'
                },
                credentials: 'same-origin',
                body: JSON.stringify({
                    url: url
                })
            });
            
            if (!response.ok) {
                console.warn('Server metadata fetch failed:', response.status, response.statusText);
                return null;
            }
            
            const data = await response.json();
            console.log('Server metadata response:', data);
            
            if (data.success && (data.title || data.description)) {
                return {
                    title: data.title || '',
                    description: data.description || ''
                };
            }
            
            return null;
            
        } catch (error) {
            console.warn('Server metadata fetch error:', error);
            return null;
        }
    }

    /**
     * Handle link form submission
     */
    async function handleLinkFormSubmission(isEdit, originalData) {
        const form = document.getElementById('addLinkForm');
        const errorMessage = document.querySelector('.add-link-dialog .error-message');
        const submitButton = form.querySelector('.dialog-button.primary');
        
        try {
            // Get form data
            const formData = new FormData(form);
            const url = formData.get('linkUrl').trim();
            const title = formData.get('linkTitle').trim();
            const description = formData.get('linkDescription').trim();
            const categorySelect = document.getElementById('linkCategory');
            const newCategoryInput = document.getElementById('newCategoryInput');
            
            let category = categorySelect.value;
            if (category === '__new__') {
                category = newCategoryInput.value.trim();
            }

            // Validate required fields
            if (!url) {
                throw new Error(window.i18n ? window.i18n.t('links.invalid_url') : 'Please enter a valid URL');
            }

            // Validate URL format
            try {
                new URL(url);
            } catch {
                throw new Error(window.i18n ? window.i18n.t('links.invalid_url') : 'Please enter a valid URL');
            }

            // Use URL as title if title is empty
            const finalTitle = title || new URL(url).hostname;

            // Show loading state
            const originalButtonText = submitButton.textContent;
            submitButton.textContent = window.i18n ? window.i18n.t('common.sending') : 'Sending...';
            submitButton.disabled = true;

            // 1. Fetch current document
            const srcResp = await fetch(`/api/source/${getCurrentDocumentPath()}`);
            if (!srcResp.ok) throw new Error('Failed to fetch current document');
            const currentMarkdown = await srcResp.text();

            // 2. Update the markdown content
            const updatedMarkdown = isEdit ? 
                updateLinkInMarkdown(currentMarkdown, originalData.url, finalTitle, url, description, category) :
                addLinkToMarkdown(currentMarkdown, finalTitle, url, description, category);

            // 3. Save the updated markdown
            const saveResp = await fetch(`/api/save/${getCurrentDocumentPath()}`, {
                method: 'POST',
                headers: { 'Content-Type': 'text/markdown' },
                body: updatedMarkdown
            });

            if (!saveResp.ok) {
                throw new Error('Failed to save document');
            }

            // Success - hide dialog and reload page
            hideAddLinkDialog();
            window.location.reload();

        } catch (error) {
            console.error('Link submission error:', error);
            errorMessage.textContent = error.message;
            errorMessage.style.display = 'block';
        } finally {
            // Restore button state
            submitButton.disabled = false;
            const isEditMode = isEdit && originalData;
            submitButton.textContent = isEditMode ?
                (window.i18n ? window.i18n.t('links.update_button') : 'Update Link') :
                (window.i18n ? window.i18n.t('links.add_button') : 'Add Link');
        }
    }

    /**
     * Get the current document path for API calls
     */
    function getCurrentDocumentPath() {
        const currentPath = window.location.pathname;
        // Remove leading slash and return empty string for root
        return currentPath === '/' ? '' : currentPath.replace(/^\//, '');
    }

    /**
     * Add a link to the markdown content
     */
    function addLinkToMarkdown(markdown, title, url, description, category) {
        const lines = markdown.split('\n');
        const categoryName = category || 'Uncategorized';
        const dateString = new Date().toISOString().split('T')[0]; // YYYY-MM-DD format
        
        // Create the link line
        let linkLine = `- [${title}](${url})`;
        if (description) {
            linkLine += ` - ${description}`;
        }
        linkLine += ` | ${dateString}`;

        // Find or create the category section
        let categoryLineIndex = -1;
        let insertIndex = -1;

        // Look for existing category
        for (let i = 0; i < lines.length; i++) {
            const line = lines[i].trim();
            if (line.startsWith('## ') && line.substring(3).trim() === categoryName) {
                categoryLineIndex = i;
                break;
            }
        }

        if (categoryLineIndex !== -1) {
            // Category exists, find where to insert the link
            insertIndex = categoryLineIndex + 1;
            // Skip empty lines after category header
            while (insertIndex < lines.length && lines[insertIndex].trim() === '') {
                insertIndex++;
            }
            // Find the end of this category (next ## header or end of file)
            while (insertIndex < lines.length && 
                   !lines[insertIndex].trim().startsWith('## ') &&
                   lines[insertIndex].trim() !== '') {
                insertIndex++;
            }
        } else {
            // Category doesn't exist, add it at the end
            // Find the end of the document content (before any trailing whitespace)
            insertIndex = lines.length;
            while (insertIndex > 0 && lines[insertIndex - 1].trim() === '') {
                insertIndex--;
            }
            
            // Add category header
            if (insertIndex < lines.length) {
                lines.splice(insertIndex, 0, '');
            }
            lines.splice(insertIndex + (insertIndex < lines.length ? 1 : 0), 0, `## ${categoryName}`);
            insertIndex = insertIndex + (insertIndex < lines.length ? 2 : 1);
        }

        // Insert the link
        lines.splice(insertIndex, 0, linkLine);
        
        return lines.join('\n');
    }

    /**
     * Update an existing link in the markdown content
     */
    function updateLinkInMarkdown(markdown, oldUrl, newTitle, newUrl, newDescription, newCategory) {
        const lines = markdown.split('\n');
        const newCategoryName = newCategory || 'Uncategorized';
        const dateString = new Date().toISOString().split('T')[0]; // YYYY-MM-DD format
        
        // Create the new link line
        let newLinkLine = `- [${newTitle}](${newUrl})`;
        if (newDescription) {
            newLinkLine += ` - ${newDescription}`;
        }
        newLinkLine += ` | ${dateString}`;

        // Find the existing link with exact URL matching
        let foundLineIndex = -1;
        let currentCategory = '';

        for (let i = 0; i < lines.length; i++) {
            const line = lines[i].trim();
            
            // Track current category - any ## header
            if (line.startsWith('## ')) {
                currentCategory = line.replace(/^##\s*/, '').trim();
            }
            
            // Check if this line contains a link with exact URL match
            const linkMatch = line.match(/\[([^\]]*)\]\(([^)]+)\)/);
            if (linkMatch) {
                const [, linkTitle, linkUrl] = linkMatch;
                
                // Exact URL match (not just includes)
                if (linkUrl === oldUrl) {
                    foundLineIndex = i;
                    break;
                }
            }
        }

        if (foundLineIndex === -1) {
            throw new Error('Link not found in document');
        }

        // Remove the old link
        lines.splice(foundLineIndex, 1);

        // If category changed, add to new category
        if (currentCategory !== newCategoryName) {
            return addLinkToMarkdown(lines.join('\n'), newTitle, newUrl, newDescription, newCategory);
        } else {
            // Same category, insert at the same position
            lines.splice(foundLineIndex, 0, newLinkLine);
            return lines.join('\n');
        }
    }

    /**
     * Delete a link from the markdown content
     */
    function deleteLinkFromMarkdown(markdown, linkData) {
        const lines = markdown.split('\n');
        let currentCategory = '';
        let targetCategory = linkData.category || 'Uncategorized';
        
        // Find and remove the specific link
        for (let i = 0; i < lines.length; i++) {
            const line = lines[i].trim();
            
            // Track current category - any ## header
            if (line.startsWith('##')) {
                // Extract category name from header (remove ## and any extra whitespace)
                currentCategory = line.replace(/^##\s*/, '').trim();
                continue;
            }
            
            // Check if this line contains a link
            const linkMatch = line.match(/\[([^\]]*)\]\(([^)]+)\)/);
            if (linkMatch) {
                const [, linkTitle, linkUrl] = linkMatch;
                
                // Exact URL match (not just includes)
                if (linkUrl === linkData.url) {
                    // Verify we're in the correct category
                    if (currentCategory === targetCategory) {
                        // Additional verification: check if title matches (if provided)
                        if (linkData.title && linkTitle !== linkData.title) {
                            continue; // Wrong title, keep looking
                        }
                        
                        // Found the exact link to delete
                        lines.splice(i, 1);
                        return lines.join('\n');
                    }
                }
            }
        }

        // If we didn't find exact match with category, try fallback approach
        // This handles cases where category detection might fail
        for (let i = 0; i < lines.length; i++) {
            const line = lines[i].trim();
            const linkMatch = line.match(/\[([^\]]*)\]\(([^)]+)\)/);
            if (linkMatch) {
                const [, linkTitle, linkUrl] = linkMatch;
                
                // Exact URL and title match as fallback
                if (linkUrl === linkData.url && (!linkData.title || linkTitle === linkData.title)) {
                    lines.splice(i, 1);
                    break;
                }
            }
        }

        return lines.join('\n');
    }

    /**
     * Delete a link from the document
     */
    async function deleteLinkFromDocument(linkData) {
        try {
            // 1. Fetch current document
            const srcResp = await fetch(`/api/source/${getCurrentDocumentPath()}`);
            if (!srcResp.ok) throw new Error('Failed to fetch current document');
            const currentMarkdown = await srcResp.text();

            // 2. Remove the link from markdown using full link data
            const updatedMarkdown = deleteLinkFromMarkdown(currentMarkdown, linkData);

            // 3. Save the updated markdown
            const saveResp = await fetch(`/api/save/${getCurrentDocumentPath()}`, {
                method: 'POST',
                headers: { 'Content-Type': 'text/markdown' },
                body: updatedMarkdown
            });

            if (!saveResp.ok) {
                throw new Error('Failed to save document');
            }

            // Success - reload page immediately
            window.location.reload();

        } catch (error) {
            console.error('Delete link error:', error);
            showToast(`Failed to delete link: ${error.message}`, 'error');
        }
    }

    /**
     * Show toast notification
     */
    function showToast(message, type = 'success') {
        // Create toast element if it doesn't exist
        let toast = document.getElementById('links-toast');
        if (!toast) {
            toast = document.createElement('div');
            toast.id = 'links-toast';
            toast.className = 'toast';
            document.body.appendChild(toast);
        }
        
        // Set message and type
        toast.textContent = message;
        toast.className = `toast ${type}`;
        
        // Show toast
        toast.classList.add('show');
        
        // Hide after 3 seconds
        setTimeout(() => {
            toast.classList.remove('show');
        }, 3000);
    }

    /**
     * Get list of existing categories from the document
     */
    function getExistingCategories() {
        const categories = [];
        
        // First, try to get all categories from the hidden data element (includes empty categories)
        const allCategoriesElement = document.getElementById('allCategories');
        if (allCategoriesElement) {
            const categoriesData = allCategoriesElement.getAttribute('data-categories');
            if (categoriesData) {
                const categoryList = categoriesData.split(',').filter(cat => cat.trim() !== '');
                categoryList.forEach(category => {
                    const trimmedCategory = category.trim();
                    if (trimmedCategory && !categories.includes(trimmedCategory)) {
                        categories.push(trimmedCategory);
                    }
                });
            }
        }
        
        // Fallback: get categories from visible headers (only non-empty categories)
        if (categories.length === 0) {
            const categoryHeaders = document.querySelectorAll('.links-category-header');
            categoryHeaders.forEach(header => {
                const categoryText = header.textContent.trim();
                // Remove count from category name (e.g., "Work Tools (3)" -> "Work Tools")
                const category = categoryText.replace(/\s*\(\d+\)\s*$/, '').trim();
                if (category && !categories.includes(category)) {
                    categories.push(category);
                }
            });
        }
        
        return categories;
    }

    // Expose functions globally for use by other modules or inline handlers
    window.LinksInteractivity = {
        init: initLinksInteractivity,
        copyLink: handleCopyLink,
        editLink: handleEditLink,
        deleteLink: handleDeleteLink,
        showAddLinkDialog: showAddLinkDialog,
        hideAddLinkDialog: hideAddLinkDialog,
        getExistingCategories: getExistingCategories,
        showToast: showToast
    };

    // Make functions available globally for backward compatibility
    window.copyLinkToClipboard = function(url) {
        // Check if modern clipboard API is available
        if (navigator.clipboard && navigator.clipboard.writeText) {
            navigator.clipboard.writeText(url).then(() => {
                // Silent copy - no toast notification
            }).catch(() => {
                copyToClipboardFallback(url);
                // Silent copy - no toast notification
            });
        } else {
            // Use fallback method
            copyToClipboardFallback(url);
            // Silent copy - no toast notification
        }
    };

    /**
     * Initialize search and filtering functionality
     */
    function initSearchAndFiltering() {
        const searchInput = document.getElementById('linksSearch');
        const categoryFilter = document.getElementById('categoryFilter');
        const sortFilter = document.getElementById('sortFilter');
        const searchClear = document.getElementById('searchClear');
        const clearFilters = document.getElementById('clearFilters');
        const searchResultsInfo = document.getElementById('searchResultsInfo');
        const resultsCount = searchResultsInfo?.querySelector('.results-count');
        const noResults = document.getElementById('noResults');
        
        if (!searchInput || !categoryFilter || !sortFilter) {
            console.log('Search/filter elements not found, skipping initialization');
            return;
        }

        let allLinks = Array.from(document.querySelectorAll('.link-item'));
        let filteredLinks = [...allLinks];

        // Store original order for default sorting
        const originalOrder = allLinks.map((link, index) => ({ link, index }));

        // Search functionality
        searchInput.addEventListener('input', function() {
            const query = this.value.trim();
            searchClear.hidden = !query;
            applyFilters();
        });

        // Category filter
        categoryFilter.addEventListener('change', applyFilters);

        // Sort filter
        sortFilter.addEventListener('change', applyFilters);

        // Clear search button
        if (searchClear) {
            searchClear.addEventListener('click', function() {
                searchInput.value = '';
                searchClear.hidden = true;
                applyFilters();
                searchInput.focus();
            });
        }

        // Clear all filters button
        if (clearFilters) {
            clearFilters.addEventListener('click', function() {
                searchInput.value = '';
                categoryFilter.value = '';
                sortFilter.value = 'default';
                searchClear.hidden = true;
                applyFilters();
            });
        }

        /**
         * Apply current search and filter settings
         */
        function applyFilters() {
            const searchQuery = searchInput.value.toLowerCase().trim();
            const categoryValue = categoryFilter.value;
            const sortValue = sortFilter.value;

            // Start with all links
            filteredLinks = [...allLinks];

            // Apply search filter
            if (searchQuery) {
                filteredLinks = filteredLinks.filter(link => {
                    const title = link.dataset.title?.toLowerCase() || '';
                    const description = link.dataset.description?.toLowerCase() || '';
                    const url = link.dataset.url?.toLowerCase() || '';
                    
                    return title.includes(searchQuery) || 
                           description.includes(searchQuery) || 
                           url.includes(searchQuery);
                });
            }

            // Apply category filter
            if (categoryValue) {
                filteredLinks = filteredLinks.filter(link => 
                    link.dataset.category === categoryValue
                );
            }

            // Apply sorting
            applySorting(sortValue);

            // Update display
            updateDisplay();
            updateResultsInfo();
        }

        /**
         * Apply sorting to filtered links
         */
        function applySorting(sortValue) {
            switch (sortValue) {
                case 'title-asc':
                    filteredLinks.sort((a, b) => 
                        (a.dataset.title || '').localeCompare(b.dataset.title || '')
                    );
                    break;
                case 'title-desc':
                    filteredLinks.sort((a, b) => 
                        (b.dataset.title || '').localeCompare(a.dataset.title || '')
                    );
                    break;
                case 'date-newest':
                    filteredLinks.sort((a, b) => {
                        const dateA = parseInt(a.dataset.date) || 0;
                        const dateB = parseInt(b.dataset.date) || 0;
                        return dateB - dateA;
                    });
                    break;
                case 'date-oldest':
                    filteredLinks.sort((a, b) => {
                        const dateA = parseInt(a.dataset.date) || 0;
                        const dateB = parseInt(b.dataset.date) || 0;
                        return dateA - dateB;
                    });
                    break;
                case 'category':
                    filteredLinks.sort((a, b) => {
                        const catA = a.dataset.category || '';
                        const catB = b.dataset.category || '';
                        if (catA !== catB) {
                            return catA.localeCompare(catB);
                        }
                        return (a.dataset.title || '').localeCompare(b.dataset.title || '');
                    });
                    break;
                case 'default':
                default:
                    // Restore original order
                    filteredLinks.sort((a, b) => {
                        const indexA = originalOrder.find(item => item.link === a)?.index || 0;
                        const indexB = originalOrder.find(item => item.link === b)?.index || 0;
                        return indexA - indexB;
                    });
                    break;
            }
        }

        /**
         * Update the display of links and categories
         */
        function updateDisplay() {
            // Hide all links first
            allLinks.forEach(link => {
                link.classList.add('hidden');
            });

            // Hide all categories first
            const categories = document.querySelectorAll('.links-category');
            categories.forEach(category => {
                category.classList.add('hidden');
            });

            if (filteredLinks.length === 0) {
                // Show no results message
                if (noResults) {
                    noResults.hidden = false;
                }
                return;
            }

            // Hide no results message
            if (noResults) {
                noResults.hidden = true;
            }

            // Group filtered links by category and maintain sort order
            const linksByCategory = {};
            filteredLinks.forEach(link => {
                const category = link.dataset.category || 'General';
                if (!linksByCategory[category]) {
                    linksByCategory[category] = [];
                }
                linksByCategory[category].push(link);
            });

            // Show links and their categories in sorted order
            Object.keys(linksByCategory).forEach(categoryName => {
                const links = linksByCategory[categoryName];
                
                // Find the category element
                const categoryElement = Array.from(document.querySelectorAll('[data-category]'))
                    .find(element => element.dataset.category === categoryName);
                if (categoryElement) {
                    categoryElement.classList.remove('hidden');
                    
                    // Update category count
                    const countSpan = categoryElement.querySelector('.section-count');
                    if (countSpan) {
                        countSpan.textContent = `(${links.length})`;
                    }
                    
                    // Get the category header (first child should be the h2)
                    const categoryHeader = categoryElement.querySelector('h2');
                    
                    // Remove all link items from category (but keep header)
                    const existingLinks = categoryElement.querySelectorAll('.link-item');
                    existingLinks.forEach(link => link.remove());
                    
                    // Re-add links in sorted order
                    links.forEach(link => {
                        link.classList.remove('hidden');
                        categoryElement.appendChild(link);
                    });
                }
            });
        }

        /**
         * Update the search results information
         */
        function updateResultsInfo() {
            const hasFilters = searchInput.value.trim() || categoryFilter.value || sortFilter.value !== 'default';
            
            if (hasFilters) {
                const count = filteredLinks.length;
                const totalLinks = parseInt(document.querySelector('.links-container')?.dataset.totalLinks) || allLinks.length;
                
                if (resultsCount) {
                    resultsCount.textContent = count === 1 ? '1 link found' : `${count} links found`;
                    if (count < totalLinks) {
                        resultsCount.textContent += ` (${totalLinks} total)`;
                    }
                }
                
                if (searchResultsInfo) {
                    searchResultsInfo.hidden = false;
                }
            } else {
                if (searchResultsInfo) {
                    searchResultsInfo.hidden = true;
                }
            }
        }

        // Initialize display
        updateDisplay();
        updateResultsInfo();
    }

    // Global functions for backward compatibility
    window.editLink = function(title, url, description, category) {
        showAddLinkDialog({ title, url, description, category });
    };

    window.deleteLink = function(url, title, description, category) {
        // Try to build more complete link data if available
        const linkData = { 
            url, 
            title: title || url,
            description: description || '',
            category: category || ''
        };
        
        const dialogTitle = 'Delete Link';
        const message = `Are you sure you want to delete this link?`;
        
        if (window.DialogSystem && window.DialogSystem.showConfirmDialog) {
            window.DialogSystem.showConfirmDialog(dialogTitle, message, (confirmed) => {
                if (confirmed) {
                    deleteLinkFromDocument(linkData);
                }
            });
        } else {
            if (confirm(`${dialogTitle}\n\n${message}`)) {
                deleteLinkFromDocument(linkData);
            }
        }
    };

    window.showAddLinkDialog = showAddLinkDialog;
    window.hideAddLinkDialog = hideAddLinkDialog;

})();
