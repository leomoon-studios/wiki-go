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
    let selectedFolder = '/';

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
                // If API not ready, use empty list or mock
                console.log('API not ready yet');
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
            accessRulesList.innerHTML = '<div class="empty-message">No access rules defined</div>';
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

            if (rule.description) {
                titleEl.textContent = rule.description;
                // Description is normal text
                
                subtitleEl.textContent = rule.pattern;
                subtitleEl.style.fontFamily = 'monospace';
            } else {
                titleEl.textContent = rule.pattern;
                titleEl.style.fontFamily = 'monospace';
                
                subtitleEl.textContent = '';
                subtitleEl.style.display = 'none';
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
            if (index === 0) clone.querySelector('.move-up').disabled = true;
            if (index === currentRules.length - 1) clone.querySelector('.move-down').disabled = true;

            accessRulesList.appendChild(clone);
        });
    }

    async function showRuleForm(index = -1) {
        if (!accessRuleDialog) return;

        // Reset state
        currentGroups = [];
        selectedFolder = '/';
        ruleIndexInput.value = index;
        
        // Populate Folder Tree (Mock for now, or fetch)
        await populateFolderTree();

        if (index >= 0) {
            // Edit Mode
            const rule = currentRules[index];
            
            // Parse pattern to set folder and match type
            let pattern = rule.pattern;
            let matchType = 'exact';
            
            if (pattern.endsWith('/**')) {
                matchType = 'recursive';
                selectedFolder = pattern.substring(0, pattern.length - 3) || '/';
            } else if (pattern.endsWith('/*')) {
                matchType = 'children';
                selectedFolder = pattern.substring(0, pattern.length - 2) || '/';
            } else {
                selectedFolder = pattern;
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
            alert('Please add at least one group for restricted access.');
            return;
        }

        const rule = {
            pattern: pattern,
            access: accessLevel,
            groups: accessLevel === 'restricted' ? currentGroups : [],
            description: ruleDescription.value || selectedFolder || 'Root'
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
                // Fallback for UI testing if API fails
                console.warn('API call failed, updating UI locally for testing');
                if (index >= 0) currentRules[index] = rule;
                else currentRules.push(rule);
                renderRulesList();
                hideRuleForm();
            }
        } catch (error) {
            console.error('Error saving rule:', error);
            // For now, just update local state to simulate
            if (index >= 0) currentRules[index] = rule;
            else currentRules.push(rule);
            renderRulesList();
            hideRuleForm();
        }
    }

    async function deleteRule(index) {
        if (!confirm('Are you sure you want to delete this rule?')) return;
        
        try {
            const response = await fetch(`/api/access-rules/${index}`, {
                method: 'DELETE'
            });
            
            if (response.ok) {
                loadAccessRules();
            } else {
                // Simulate
                currentRules.splice(index, 1);
                renderRulesList();
            }
        } catch (error) {
            console.error('Error deleting rule:', error);
            // Simulate
            currentRules.splice(index, 1);
            renderRulesList();
        }
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
        folderTree.innerHTML = '<div class="loading">Loading folders...</div>';
        
        try {
            // Mock folders for demonstration until API is ready
            folderTree.innerHTML = '';
            
            // Add Root
            addFolderToTree('/', '/', 0);
            
            // Mock folders
            addFolderToTree('/finance', 'finance', 0);
            addFolderToTree('/finance/reports', 'reports', 1);
            addFolderToTree('/internal', 'internal', 0);
            addFolderToTree('/public', 'public', 0);

        } catch (error) {
            folderTree.innerHTML = '<div class="error">Failed to load folders</div>';
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
