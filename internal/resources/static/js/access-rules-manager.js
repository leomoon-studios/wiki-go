// Access Rules Management Module
document.addEventListener('DOMContentLoaded', function() {
    'use strict';

    // Elements
    const addRuleBtn = document.getElementById('addRuleBtn');
    const accessRuleDialog = document.querySelector('.access-rule-dialog');
    const accessRuleForm = document.getElementById('accessRuleForm');
    const accessRulesList = document.getElementById('accessRulesList');
    const ruleTemplate = document.getElementById('access-rule-template');
    const closeButtons = accessRuleDialog ? accessRuleDialog.querySelectorAll('.close-dialog, .cancel-dialog') : [];
    
    // Form Elements
    const ruleIndexInput = document.getElementById('ruleIndex');
    const selectedFolderPath = document.getElementById('selectedFolderPath');
    const folderTree = document.getElementById('folderTree');
    const matchTypeInputs = document.getElementsByName('matchType');
    const accessLevelInputs = document.getElementsByName('accessLevel');
    const groupsContainer = document.getElementById('groupsContainer');
    const selectedGroups = document.getElementById('selectedGroups');
    const groupInput = document.getElementById('groupInput');
    const addGroupBtn = document.getElementById('addGroupBtn');
    const ruleDescription = document.getElementById('ruleDescription');

    let currentRules = [];
    let currentGroups = [];

    // Initialize
    if (accessRulesList) {
        loadAccessRules();
    }

    // Event Listeners
    if (addRuleBtn) {
        addRuleBtn.addEventListener('click', () => showRuleForm());
    }

    closeButtons.forEach(button => {
        button.addEventListener('click', hideRuleForm);
    });

    // Background click closing disabled as per request
    /* if (accessRuleDialog) {
        accessRuleDialog.addEventListener('click', (e) => {
            if (e.target === accessRuleDialog) hideRuleForm();
        });
    } */

    if (accessRuleForm) {
        accessRuleForm.addEventListener('submit', handleFormSubmit);
    }

    if (addGroupBtn) {
        addGroupBtn.addEventListener('click', addGroup);
    }

    if (groupInput) {
        groupInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                addGroup();
            }
        });
    }

    // Access Level Change Handler
    Array.from(accessLevelInputs).forEach(input => {
        input.addEventListener('change', updateGroupsVisibility);
    });

    // Functions
    async function loadAccessRules() {
        try {
            const response = await fetch('/api/access-rules');
            if (response.ok) {
                const data = await response.json();
                currentRules = data.rules || [];
                renderRulesList();
            } else {
                console.error('Failed to load access rules');
                currentRules = [];
                renderRulesList();
            }
        } catch (error) {
            console.error('Error loading rules:', error);
        }
    }

    function renderRulesList() {
        if (!accessRulesList || !ruleTemplate) return;
        
        accessRulesList.innerHTML = '';
        
        if (currentRules.length === 0) {
            accessRulesList.innerHTML = `<div class="empty-message" data-i18n="access.no_rules">${window.i18n ? window.i18n.t('access.no_rules') : 'No access rules defined'}</div>`;
            return;
        }

        currentRules.forEach((rule, index) => {
            const clone = ruleTemplate.content.cloneNode(true);
            
            // Set Icon
            const iconDiv = clone.querySelector('.rule-icon');
            if (rule.access === 'public') iconDiv.innerHTML = '<i class="fa fa-globe"></i>';
            else if (rule.access === 'private') iconDiv.innerHTML = '<i class="fa fa-lock"></i>';
            else iconDiv.innerHTML = '<i class="fa fa-shield"></i>';

            // Set Content
            const titleEl = clone.querySelector('.rule-title');
            const subtitleEl = clone.querySelector('.rule-subtitle');

            // Parse pattern for display
            let basePath = rule.pattern;
            let matchIcon = 'fa-file-text-o';
            let matchKey = 'access.match_exact';
            let matchDefault = 'This document only';
            
            if (basePath.endsWith('/**')) {
                basePath = basePath.substring(0, basePath.length - 3) || '/';
                matchIcon = 'fa-sitemap';
                matchKey = 'access.match_recursive';
                matchDefault = 'This document and all sub documents';
            } else if (basePath.endsWith('/*')) {
                basePath = basePath.substring(0, basePath.length - 2) || '/';
                matchIcon = 'fa-folder-open-o';
                matchKey = 'access.match_children';
                matchDefault = 'Direct children only';
            }

            const matchText = window.i18n ? window.i18n.t(matchKey) : matchDefault;
            const matchHtml = `<span class="match-type" title="${matchText}"><i class="fa ${matchIcon}"></i></span>`;

            if (rule.description) {
                titleEl.textContent = rule.description;
                titleEl.style.fontFamily = 'inherit';
                
                subtitleEl.innerHTML = `${matchHtml} <span class="rule-path">${basePath}</span>`;
                subtitleEl.style.display = 'block';
            } else {
                titleEl.textContent = basePath;
                titleEl.style.fontFamily = 'monospace';
                
                subtitleEl.innerHTML = `${matchHtml} <span class="match-text" data-i18n="${matchKey}">${matchText}</span>`;
                subtitleEl.style.display = 'block';
            }
            
            const badge = clone.querySelector('.rule-access-badge');
            badge.textContent = rule.access;
            badge.className = `rule-access-badge access-${rule.access}`;

            // Groups
            const groupsDiv = clone.querySelector('.rule-groups');
            if (rule.groups && rule.groups.length > 0) {
                rule.groups.forEach(group => {
                    const tag = document.createElement('span');
                    tag.className = 'group-tag';
                    tag.textContent = group;
                    groupsDiv.appendChild(tag);
                });
            }

            // Actions
            clone.querySelector('.move-up').onclick = () => moveRule(index, -1);
            clone.querySelector('.move-down').onclick = () => moveRule(index, 1);
            clone.querySelector('.edit-rule').onclick = () => showRuleForm(index);
            clone.querySelector('.delete-rule').onclick = () => deleteRule(index);

            // Disable move buttons appropriately
            if (index === 0) {
                const btn = clone.querySelector('.move-up');
                btn.disabled = true;
                btn.style.visibility = 'hidden';
            }
            if (index === currentRules.length - 1) {
                const btn = clone.querySelector('.move-down');
                btn.disabled = true;
                btn.style.display = 'none';
            }

            accessRulesList.appendChild(clone);
        });
    }

    async function showRuleForm(index = -1) {
        if (!accessRuleDialog) return;

        // Reset state
        currentGroups = [];
        selectedFolder = '/';
        ruleIndexInput.value = index;

        // Determine selectedFolder based on index (Edit Mode) BEFORE populating tree
        if (index >= 0) {
            const rule = currentRules[index];
            let pattern = rule.pattern;
            
            if (pattern.endsWith('/**')) {
                selectedFolder = pattern.substring(0, pattern.length - 3) || '/';
            } else if (pattern.endsWith('/*')) {
                selectedFolder = pattern.substring(0, pattern.length - 2) || '/';
            } else {
                selectedFolder = pattern;
            }
        }
        
        // Populate Folder Tree (Mock for now, or fetch)
        await populateFolderTree();

        if (index >= 0) {
            // Edit Mode
            const rule = currentRules[index];
            
            // Parse pattern to set match type
            let pattern = rule.pattern;
            let matchType = 'exact';
            
            if (pattern.endsWith('/**')) {
                matchType = 'recursive';
            } else if (pattern.endsWith('/*')) {
                matchType = 'children';
            }

            selectedFolderPath.textContent = selectedFolder;
            
            // Set Match Type
            const matchTypeInput = document.querySelector(`input[name="matchType"][value="${matchType}"]`);
            if (matchTypeInput) matchTypeInput.checked = true;
            
            // Set Access Level
            const accessLevelInput = document.querySelector(`input[name="accessLevel"][value="${rule.access}"]`);
            if (accessLevelInput) accessLevelInput.checked = true;
            
            // Set Groups
            currentGroups = [...(rule.groups || [])];
            
            // Set Description
            ruleDescription.value = rule.description || '';
        } else {
            // Add Mode
            accessRuleForm.reset();
            selectedFolderPath.textContent = '/';
            document.querySelector('input[name="matchType"][value="recursive"]').checked = true;
            document.querySelector('input[name="accessLevel"][value="restricted"]').checked = true;
            ruleDescription.value = '';
        }

        renderGroups();
        updateGroupsVisibility();
        accessRuleDialog.classList.add('active');
    }

    function hideRuleForm() {
        if (accessRuleDialog) accessRuleDialog.classList.remove('active');
    }

    function updateGroupsVisibility() {
        const accessLevelInput = document.querySelector('input[name="accessLevel"]:checked');
        if (!accessLevelInput) return;
        
        const accessLevel = accessLevelInput.value;
        if (accessLevel === 'restricted') {
            groupsContainer.style.display = 'block';
        } else {
            groupsContainer.style.display = 'none';
        }
    }

    function addGroup() {
        const group = groupInput.value.trim();
        if (group && !currentGroups.includes(group)) {
            currentGroups.push(group);
            renderGroups();
            groupInput.value = '';
        }
    }

    function removeGroup(group) {
        currentGroups = currentGroups.filter(g => g !== group);
        renderGroups();
    }

    function renderGroups() {
        selectedGroups.innerHTML = '';
        currentGroups.forEach(group => {
            const tag = document.createElement('div');
            tag.className = 'group-tag-removable';
            tag.innerHTML = `
                <span>${group}</span>
                <span class="remove-group" data-group="${group}">&times;</span>
            `;
            tag.querySelector('.remove-group').onclick = () => removeGroup(group);
            selectedGroups.appendChild(tag);
        });
    }

    async function handleFormSubmit(e) {
        e.preventDefault();
        
        const index = parseInt(ruleIndexInput.value);
        const matchType = document.querySelector('input[name="matchType"]:checked').value;
        const accessLevel = document.querySelector('input[name="accessLevel"]:checked').value;
        
        // Construct Pattern
        let pattern = selectedFolder;
        if (pattern === '/') pattern = ''; // Root handling
        
        if (matchType === 'recursive') {
            pattern += '/**';
        } else if (matchType === 'children') {
            pattern += '/*';
        }
        
        if (pattern === '') pattern = '/'; // Root exact match?

        // Validation
        if (accessLevel === 'restricted' && currentGroups.length === 0) {
            window.showMessageDialog(
                window.i18n ? window.i18n.t('access.validation_error') : 'Validation Error',
                window.i18n ? window.i18n.t('access.error_no_groups') : 'Please add at least one group for restricted access.'
            );
            return;
        }

        const rule = {
            pattern: pattern,
            access: accessLevel,
            groups: accessLevel === 'restricted' ? currentGroups : [],
            description: ruleDescription.value || selectedFolder || (window.i18n ? window.i18n.t('access.root') : 'Root')
        };

        try {
            let response;
            if (index >= 0) {
                // Update
                response = await fetch(`/api/access-rules/${index}`, {
                    method: 'PUT',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify(rule)
                });
            } else {
                // Create
                response = await fetch('/api/access-rules', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify(rule)
                });
            }

            if (response.ok) {
                hideRuleForm();
                loadAccessRules();
            } else {
                const errorData = await response.json().catch(() => ({}));
                const errorTitle = window.i18n ? window.i18n.t('common.error') : 'Error';
                const errorPrefix = window.i18n ? window.i18n.t('access.error_save_failed') : 'Failed to save rule';
                const unknownError = window.i18n ? window.i18n.t('common.unknown_error') : 'Unknown error';
                
                window.showMessageDialog(errorTitle, errorPrefix + ': ' + (errorData.message || unknownError));
            }
        } catch (error) {
            console.error('Error saving rule:', error);
            const errorTitle = window.i18n ? window.i18n.t('common.error') : 'Error';
            const errorPrefix = window.i18n ? window.i18n.t('access.error_save') : 'Error saving rule';
            
            window.showMessageDialog(errorTitle, errorPrefix + ': ' + error.message);
        }
    }

    function deleteRule(index) {
        const deleteTitle = window.i18n ? window.i18n.t('access.delete_title') : 'Delete Rule';
        const deleteConfirm = window.i18n ? window.i18n.t('access.delete_confirm') : 'Are you sure you want to delete this rule?';
        
        window.showConfirmDialog(deleteTitle, deleteConfirm, async (confirmed) => {
            if (!confirmed) return;
            
            try {
                const response = await fetch(`/api/access-rules/${index}`, {
                    method: 'DELETE'
                });
                
                if (response.ok) {
                    loadAccessRules();
                } else {
                    window.showMessageDialog(
                        window.i18n ? window.i18n.t('common.error') : 'Error',
                        window.i18n ? window.i18n.t('access.error_delete_failed') : 'Failed to delete rule'
                    );
                }
            } catch (error) {
                console.error('Error deleting rule:', error);
                window.showMessageDialog(
                    window.i18n ? window.i18n.t('common.error') : 'Error',
                    window.i18n ? window.i18n.t('access.error_delete') : 'Error deleting rule'
                );
            }
        });
    }

    async function moveRule(index, direction) {
        const newIndex = index + direction;
        if (newIndex < 0 || newIndex >= currentRules.length) return;

        // Create new order of indices
        const indices = currentRules.map((_, i) => i);
        // Swap indices
        const temp = indices[index];
        indices[index] = indices[newIndex];
        indices[newIndex] = temp;

        try {
            const response = await fetch('/api/access-rules/reorder', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({ indices: indices })
            });

            if (response.ok) {
                loadAccessRules();
            } else {
                console.error('Failed to reorder rules');
            }
        } catch (error) {
            console.error('Error reordering rules:', error);
        }
    }

    async function populateFolderTree() {
        folderTree.innerHTML = `<div class="loading">${window.i18n ? window.i18n.t('access.loading_folders') : 'Loading folders...'}</div>`;
        
        try {
            const response = await fetch('/api/folders');
            if (response.ok) {
                const data = await response.json();
                folderTree.innerHTML = '';
                
                if (data.folders && data.folders.length > 0) {
                    data.folders.forEach(folder => {
                        addFolderToTree(folder.path, folder.name, folder.level);
                    });
                } else {
                    folderTree.innerHTML = `<div class="empty-message">${window.i18n ? window.i18n.t('access.no_folders') : 'No folders found'}</div>`;
                }
            } else {
                folderTree.innerHTML = `<div class="error">${window.i18n ? window.i18n.t('access.error_load_folders') : 'Failed to load folders'}</div>`;
            }
        } catch (error) {
            console.error('Error loading folders:', error);
            folderTree.innerHTML = `<div class="error">${window.i18n ? window.i18n.t('access.error_load_folders') : 'Failed to load folders'}</div>`;
        }
    }

    function addFolderToTree(path, name, level) {
        const item = document.createElement('div');
        item.className = 'folder-tree-item';
        item.style.paddingLeft = `${level * 20 + 10}px`;
        if (path === selectedFolder) item.classList.add('selected');
        
        item.innerHTML = `
            <i class="fa fa-folder folder-icon"></i>
            <span class="folder-name">${name}</span>
        `;
        
        item.onclick = () => {
            document.querySelectorAll('.folder-tree-item').forEach(el => el.classList.remove('selected'));
            item.classList.add('selected');
            selectedFolder = path;
            selectedFolderPath.textContent = path;
        };
        
        folderTree.appendChild(item);
    }
});
