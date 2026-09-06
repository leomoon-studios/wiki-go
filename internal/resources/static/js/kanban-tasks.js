// Kanban Task Management
class KanbanTaskManager {
  constructor(core) {
    this.core = core;
    this.taskIdMap = new Map();
    this.taskCheckboxHandler = null;
    this.initialized = false;
  }

  /**
   * Initialize task management
   */
  init() {
    if (this.initialized) {
      console.log('Task manager already initialized');
      return;
    }

    // Check permissions using shared module
    if (!window.TaskListPermissions || !window.TaskListPermissions.isTaskEditingAllowed()) {
      console.log('KanbanTaskManager: Task editing not allowed, aborting initialization');
      return;
    }

    console.log('Initializing task manager');

    // Create task checkbox handler
    this.taskCheckboxHandler = this.createTaskCheckboxHandler();

    // Setup all existing task items
    this.setupTaskItems();

    // Setup add task buttons
    this.setupAddTaskButtons();

    this.initialized = true;
    console.log('Task manager initialized');
  }

  /**
   * Setup all task items for management
   */
  setupTaskItems() {
    const allTaskItems = document.querySelectorAll('.kanban-column-content .task-list-item-container');
    console.log('Task items found:', allTaskItems.length);

    // Track which task IDs have already been assigned to prevent duplicates
    const assignedTaskIds = new Set();

    // Fetch the original markdown to preserve formatting
    fetch(`/api/source/${this.core.getDocPath()}`)
      .then(response => response.text())
      .then(markdown => {
        // Extract task content and IDs from the markdown
        const taskInfo = this.extractTasksFromMarkdown(markdown);

        // Set up each task item
        allTaskItems.forEach((item, index) => {
          // Initialize indentation classes
          const indentLevel = parseInt(item.getAttribute('data-indent-level') || '0');
          item.classList.add(`indent-${indentLevel}`);

          // Remove any inline margin-left that might have been set by the server
          if (item.style.marginLeft) {
            item.style.removeProperty('margin-left');
          }

          // Make item draggable
          item.setAttribute('draggable', 'true');

          // Try to match with original markdown and assign task ID
          const taskTextElement = item.querySelector('.task-text');
          if (taskTextElement) {
            const displayedText = taskTextElement.textContent.trim();

            // Try to find a matching task in the markdown that hasn't been assigned yet
            const matchingTask = this.findMatchingTask(taskInfo, displayedText, indentLevel, assignedTaskIds);

            if (matchingTask) {
              // Use existing task ID if found
              item.setAttribute('data-task-id', matchingTask.id);
              item.setAttribute('data-original-markdown', matchingTask.content);

              // Store in the task ID map
              this.taskIdMap.set(matchingTask.id, item);

              // Mark this task ID as assigned
              assignedTaskIds.add(matchingTask.id);

              console.log(`Assigned existing ID ${matchingTask.id} to task "${displayedText.substring(0, 20)}${displayedText.length > 20 ? '...' : ''}"`);
            } else {
              // Generate a new ID for this task
              const newId = this.generateTaskId();
              item.setAttribute('data-task-id', newId);

              // Store in the task ID map
              this.taskIdMap.set(newId, item);

              console.log(`Generated new ID ${newId} for task "${displayedText.substring(0, 20)}${displayedText.length > 20 ? '...' : ''}"`);
            }
          }

          // Add drag handle
          this.addDragHandle(item);

          // Setup checkbox click handler
          this.setupCheckboxForTask(item);

          if (index === 0) {
            console.log('First task setup complete:', item.outerHTML);
          }
        });
      })
      .catch(error => {
        console.error('Error fetching original markdown:', error);

        // Set up basic functionality even if we couldn't fetch the markdown
        allTaskItems.forEach((item, index) => {
          // Initialize indentation classes
          const indentLevel = parseInt(item.getAttribute('data-indent-level') || '0');
          item.classList.add(`indent-${indentLevel}`);

          // Remove any inline margin-left that might have been set by the server
          if (item.style.marginLeft) {
            item.style.removeProperty('margin-left');
          }

          // Make item draggable
          item.setAttribute('draggable', 'true');

          // Generate a new ID for this task
          const newId = this.generateTaskId();
          item.setAttribute('data-task-id', newId);

          // Store in the task ID map
          this.taskIdMap.set(newId, item);

          // Add drag handle
          this.addDragHandle(item);

          // Setup checkbox click handler
          this.setupCheckboxForTask(item);
        });
      });
  }

  /**
   * Generate a unique ID for a task
   */
  generateTaskId() {
    return 'task_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
  }

  /**
   * Create a handler for task checkboxes
   */
  createTaskCheckboxHandler() {
    // Helper to show state in the save-state span - only for errors
    const showState = (li, text, cssClass) => {
      const s = li.querySelector('.save-state');
      if (!s) return;
      
      // Only show error states
      if (cssClass !== 'error') return;
      
      s.textContent = text;
      s.classList.remove('saved','error');
      s.classList.add(cssClass);
      
      // Auto-hide after 3 seconds for errors
      setTimeout(() => {
        s.textContent = '';
        s.classList.remove('saved','error');
      }, 3000);
    };

    // Return the handler function
    return async (e) => {
      const target = e.target;
      if (!(target instanceof HTMLInputElement) || target.type !== 'checkbox') return;

      // Always prevent default to ensure we handle it ourselves
      e.preventDefault();
      e.stopPropagation();

      const li = target.closest('li');
      if (!li) return;

      // Get the task ID
      const taskId = li.getAttribute('data-task-id');
      if (!taskId) {
        console.error('Task has no ID, cannot toggle');
        return;
      }

      // Get the indentation level
      const indentLevel = parseInt(li.getAttribute('data-indent-level') || '0');
      console.log(`Handling checkbox click for task ID ${taskId} at indent level ${indentLevel}`);

      // Disable all checkboxes during save
      const allCheckboxes = document.querySelectorAll('.task-checkbox');
      allCheckboxes.forEach(cb => cb.disabled = true);

      try {
        // 1. Fetch current markdown
        const srcResp = await fetch(`/api/source/${this.core.getDocPath()}`);
        if (!srcResp.ok) throw new Error('source fetch failed');
        const markdown = await srcResp.text();

        // 2. Build updated markdown
        const updatedMarkdown = this.toggleTaskInMarkdown(markdown, li, taskId);

        // 3. Save updated markdown
        const saveResp = await fetch(`/api/save/${this.core.getDocPath()}`, {
          method: 'POST',
          headers: { 'Content-Type': 'text/markdown' },
          body: updatedMarkdown,
        });

        if (!saveResp.ok) throw new Error('save failed');

        // 4. Update UI checkbox state
        target.checked = !target.checked;
        // Success - no visual feedback needed

        // 5. Clear the moved flag since we've successfully saved the task
        if (li.hasAttribute('data-was-moved')) {
          li.removeAttribute('data-was-moved');
          console.log(`Cleared 'was-moved' flag for task ${taskId}`);
        }

        // 6. If this was a new task, clear the new flag
        if (li.hasAttribute('data-is-new')) {
          li.removeAttribute('data-is-new');
          console.log(`Cleared 'is-new' flag for task ${taskId}`);
        }
      } catch (err) {
        console.error('Error toggling task checkbox:', err);
        showState(li, 'error', 'error');
      } finally {
        // Re-enable all checkboxes
        allCheckboxes.forEach(cb => cb.disabled = false);
      }
    };
  }

  /**
   * Toggle a task checkbox in the markdown content using task ID
   */
  toggleTaskInMarkdown(markdown, taskItem, taskId) {
    const lines = markdown.split(/\r?\n/);
    const desiredChecked = !taskItem.querySelector('.task-checkbox').checked;

    // Get the task text and indentation level
    const taskText = taskItem.querySelector('.task-text').textContent.trim();
    const indentLevel = parseInt(taskItem.getAttribute('data-indent-level') || '0');

    console.log(`Toggling task: "${taskText}" (ID: ${taskId}, indent: ${indentLevel}, desired state: ${desiredChecked ? 'checked' : 'unchecked'})`);

    // Check if this is a newly created task that hasn't been saved yet
    const isNewTask = taskItem.hasAttribute('data-is-new') && taskItem.getAttribute('data-is-new') === 'true';

    // For new tasks, we need to find them by content since they don't have IDs in the markdown yet
    if (isNewTask) {
      console.log('This is a new task, finding by content');
      return this.toggleNewTaskByContent(lines, taskText, indentLevel, desiredChecked, taskId);
    }

    // For existing tasks, try to find them by ID comment
    return this.toggleExistingTaskById(lines, taskText, indentLevel, desiredChecked, taskId, taskItem);
  }

  /**
   * Toggle a new task by content matching
   */
  toggleNewTaskByContent(lines, taskText, indentLevel, desiredChecked, taskId) {
    let found = false;
    const updatedLines = lines.map(line => {
      if (found) return line;

      const match = line.match(/^(\s*)([-*+])\s+\[(.| )\]\s+(.*)$/);
      if (!match) return line;

      const [, indent, listMarker, currentCheck, content] = match;
      const lineIndentLevel = Math.floor(indent.length / 2);

      // Extract content without comments
      const contentWithoutComments = content.split('<!--')[0].trim();

      // Normalize content for comparison
      const normalizedContent = this.normalizeTaskContent(contentWithoutComments);
      const normalizedTaskText = this.normalizeTaskContent(taskText);

      // For new tasks, we need to match both content and indentation level
      if (normalizedContent === normalizedTaskText && lineIndentLevel === indentLevel) {
        found = true;
        const newMark = desiredChecked ? 'x' : ' ';

        console.log(`Found new task to toggle at indent level ${lineIndentLevel}`);

        // Add task ID comment for future reference
        return `${indent}${listMarker} [${newMark}] ${contentWithoutComments} <!-- task-id: ${taskId} -->`;
      }

      return line;
    });

    if (!found) {
      console.warn('Could not find new task to toggle');
    }

    return updatedLines.join('\n');
  }

  /**
   * Toggle an existing task by ID
   */
  toggleExistingTaskById(lines, taskText, indentLevel, desiredChecked, taskId, taskItem) {
    console.log('Looking for task by ID in markdown');

    // Check if this task was recently moved
    const wasTaskMoved = taskItem.hasAttribute('data-was-moved') && taskItem.getAttribute('data-was-moved') === 'true';

    if (wasTaskMoved) {
      console.log('Task was moved, using extra care when toggling');
    }

    // Try different patterns for finding the task ID in the markdown
    const idPatterns = [
      `<!-- task-id: ${taskId} -->`,
      `<!--task-id: ${taskId}-->`,
      `<!-- task-id:${taskId} -->`,
      `<!--task-id:${taskId}-->`
    ];

    // First, try to find the exact task by ID and indentation level
    let foundExact = false;
    let foundAny = false;
    let exactMatchLine = -1;
    let anyMatchLine = -1;

    // First pass: find all potential matches
    lines.forEach((line, index) => {
      if (foundExact) return;

      // Check if this line contains our task ID in any format
      const hasTaskId = idPatterns.some(pattern => line.includes(pattern));

      if (hasTaskId) {
        foundAny = true;
        anyMatchLine = index;

        // Extract the task parts
        const match = line.match(/^(\s*)([-*+])\s+\[(.| )\]\s+(.*)$/);
        if (!match) return;

        const [, indent, listMarker, currentCheck, content] = match;

        // Calculate the line's indentation level
        const lineIndentLevel = Math.floor(indent.length / 2);

        // If indentation level matches, this is our exact match
        if (lineIndentLevel === indentLevel) {
          foundExact = true;
          exactMatchLine = index;
          console.log(`Found exact match for task ID ${taskId} at line ${index} with indent level ${lineIndentLevel}`);
        }
      }
    });

    // Now update the appropriate line
    if (foundExact || foundAny) {
      const lineIndex = foundExact ? exactMatchLine : anyMatchLine;
      const line = lines[lineIndex];

      // Extract the task parts
      const match = line.match(/^(\s*)([-*+])\s+\[(.| )\]\s+(.*)$/);
      if (!match) {
        console.warn('Found task ID but line format is unexpected:', line);
        return lines.join('\n');
      }

      const [, indent, listMarker, currentCheck, content] = match;

      // Calculate the line's indentation level
      const lineIndentLevel = Math.floor(indent.length / 2);

      // Split content and comment
      const parts = content.split('<!--');
      const mainContent = parts[0].trim();
      const comment = parts.length > 1 ? '<!--' + parts.slice(1).join('<!--') : '';

      // Toggle the checkbox
      const newMark = desiredChecked ? 'x' : ' ';

      // Log the change with indentation info
      console.log(`Toggling task by ID at line ${lineIndex}, indent level ${lineIndentLevel}, from [${currentCheck}] to [${newMark}]`);

      // Update the line
      lines[lineIndex] = `${indent}${listMarker} [${newMark}] ${mainContent} ${comment}`;

      return lines.join('\n');
    }

    console.warn('Could not find task by ID, falling back to content matching');

    // Fall back to content matching with indentation level
    return this.fallbackToggleByContent(lines, taskText, indentLevel, desiredChecked, taskId, wasTaskMoved);
  }

  /**
   * Fallback method to toggle a task by content and indentation level
   */
  fallbackToggleByContent(lines, taskText, indentLevel, desiredChecked, taskId, wasTaskMoved) {
    console.log(`Fallback: toggling task by content: "${taskText}" at indent level ${indentLevel}, desired state: ${desiredChecked ? 'checked' : 'unchecked'}`);

    // Normalize the task text
    const normalizedTaskText = this.normalizeTaskContent(taskText);

    // Collect all potential matches
    const potentialMatches = [];

    lines.forEach((line, index) => {
      const match = line.match(/^(\s*)([-*+])\s+\[(.| )\]\s+(.*)$/);
      if (!match) return;

      const [, indent, listMarker, currentCheck, content] = match;
      const lineIndentLevel = Math.floor(indent.length / 2);

      // Extract content without comments
      const contentWithoutComments = content.split('<!--')[0].trim();

      // Normalize content
      const normalizedContent = this.normalizeTaskContent(contentWithoutComments);

      // If the task was moved, we need to be more strict about matching
      let isMatch;

      if (wasTaskMoved) {
        // For moved tasks, require exact text match and correct indentation
        isMatch = normalizedContent === normalizedTaskText && lineIndentLevel === indentLevel;
      } else {
        // For non-moved tasks, be more flexible with content but strict with indentation
        isMatch = (normalizedContent === normalizedTaskText ||
                  normalizedContent.includes(normalizedTaskText) ||
                  normalizedTaskText.includes(normalizedContent)) &&
                  lineIndentLevel === indentLevel;
      }

      if (isMatch) {
        potentialMatches.push({
          index,
          line,
          indentLevel: lineIndentLevel,
          content,
          contentWithoutComments,
          indent,
          listMarker,
          exactMatch: normalizedContent === normalizedTaskText && lineIndentLevel === indentLevel,
          currentCheck
        });
      }
    });

    // Log all potential matches for debugging
    console.log(`Found ${potentialMatches.length} potential matches`);
    potentialMatches.forEach((match, i) => {
      console.log(`Match ${i+1}: "${match.contentWithoutComments}" (indent: ${match.indentLevel}, exact: ${match.exactMatch}, current: [${match.currentCheck}])`);
    });

    // If we have matches, use the one with the correct indentation level
    if (potentialMatches.length > 0) {
      // First try exact match with both content and indent level
      const exactMatches = potentialMatches.filter(match => match.exactMatch);

      let matchToUse;

      if (exactMatches.length > 0) {
        console.log('Using exact content and indent match');
        matchToUse = exactMatches[0];
      } else {
        // Next try to find a match with the correct indent level
        const indentMatches = potentialMatches.filter(match => match.indentLevel === indentLevel);

        if (indentMatches.length > 0) {
          console.log('Using indent-level match');
          matchToUse = indentMatches[0];
        } else {
          // Use closest match by indentation
          console.log('No exact matches, using closest match by indent');
          potentialMatches.sort((a, b) =>
            Math.abs(a.indentLevel - indentLevel) - Math.abs(b.indentLevel - indentLevel)
          );
          matchToUse = potentialMatches[0];
        }
      }

      // Create updated lines
      const updatedLines = [...lines];
      const index = matchToUse.index;

      // Extract content and comments
      const parts = matchToUse.content.split('<!--');
      const mainContent = parts[0].trim();
      const comment = parts.length > 1 ? '<!--' + parts.slice(1).join('<!--') : '';

      // Toggle the checkbox and add task ID if not present
      const newMark = desiredChecked ? 'x' : ' ';
      const hasTaskId = comment.includes('task-id:');

      console.log(`Toggling checkbox from [${matchToUse.currentCheck}] to [${newMark}] at indent level ${matchToUse.indentLevel}`);

      if (hasTaskId) {
        updatedLines[index] = `${matchToUse.indent}${matchToUse.listMarker} [${newMark}] ${mainContent} ${comment}`;
      } else {
        updatedLines[index] = `${matchToUse.indent}${matchToUse.listMarker} [${newMark}] ${mainContent} <!-- task-id: ${taskId} -->`;
      }

      console.log(`Updated line ${index}: ${updatedLines[index]}`);
      return updatedLines.join('\n');
    }

    console.warn('Could not find any matching task');
    return lines.join('\n');
  }

  /**
   * Normalize task content for comparison
   */
  normalizeTaskContent(content) {
    return content
      .replace(/\*\*([^*]+)\*\*/g, '$1') // Remove bold
      .replace(/__([^_]+)__/g, '$1')     // Remove bold with underscore
      .replace(/\*([^*]+)\*/g, '$1')     // Remove italic
      .replace(/_([^_]+)_/g, '$1')       // Remove italic with underscore
      .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1') // Remove links
      .replace(/`([^`]+)`/g, '$1')       // Remove code
      .replace(/~~([^~]+)~~/g, '$1')     // Remove strikethrough
      .trim();
  }

  /**
   * Find a matching task from the extracted task info
   */
  findMatchingTask(taskInfo, displayedText, indentLevel, assignedTaskIds) {
    // Normalize the displayed text
    const normalizedText = this.normalizeTaskContent(displayedText);

    // First try to find an exact match with the same indentation level that hasn't been assigned yet
    const exactMatches = taskInfo.filter(task => {
      const normalizedTaskContent = task.normalizedContent;
      return normalizedTaskContent === normalizedText &&
             task.indentLevel === indentLevel &&
             !assignedTaskIds.has(task.id);
    });

    if (exactMatches.length > 0) {
      return exactMatches[0];
    }

    // If no exact match, try a more flexible match that hasn't been assigned yet
    const flexibleMatches = taskInfo.filter(task => {
      const normalizedTaskContent = task.normalizedContent;
      return (normalizedTaskContent.includes(normalizedText) ||
              normalizedText.includes(normalizedTaskContent)) &&
             task.indentLevel === indentLevel &&
             !assignedTaskIds.has(task.id);
    });

    if (flexibleMatches.length > 0) {
      return flexibleMatches[0];
    }

    // If still no match, try ignoring indentation but still check if not assigned
    const contentMatches = taskInfo.filter(task => {
      const normalizedTaskContent = task.normalizedContent;
      return (normalizedTaskContent === normalizedText ||
             normalizedTaskContent.includes(normalizedText) ||
             normalizedText.includes(normalizedTaskContent)) &&
             !assignedTaskIds.has(task.id);
    });

    if (contentMatches.length > 0) {
      // Sort by closest indentation level
      contentMatches.sort((a, b) =>
        Math.abs(a.indentLevel - indentLevel) - Math.abs(b.indentLevel - indentLevel)
      );
      return contentMatches[0];
    }

    return null;
  }

  /**
   * Extract tasks from markdown content with their IDs
   */
  extractTasksFromMarkdown(markdown) {
    const tasks = [];
    const lines = markdown.split(/\r?\n/);

    // Regular expression to match task items with optional task ID comment
    const taskRegex = /^(\s*)([-*+])\s+\[([ xX])\]\s+(.+?)(?:\s+<!--\s*task-id:\s*([a-zA-Z0-9_]+)\s*-->)?$/;

    lines.forEach((line, index) => {
      const match = line.match(taskRegex);
      if (match) {
        const indent = match[1];
        const listMarker = match[2];
        const isChecked = match[3] !== ' ';
        const content = match[4];

        // Extract task ID if present, or generate a new one
        let taskId;

        // Check if there's a task ID in the comment
        if (match[5]) {
          taskId = match[5];
        } else {
          // Check if there's a task ID in a different format
          const idMatch = line.match(/<!--\s*task-id:\s*([a-zA-Z0-9_]+)\s*-->/);
          if (idMatch) {
            taskId = idMatch[1];
          } else {
            // Generate a new ID
            taskId = this.generateTaskId();
          }
        }

        // Calculate indentation level
        const indentLevel = Math.floor(indent.length / 2);

        // Create a normalized version of the content for comparison
        // Remove any HTML comments first
        const contentWithoutComments = content.split('<!--')[0].trim();
        const normalizedContent = this.normalizeTaskContent(contentWithoutComments);

        tasks.push({
          index,
          indent,
          listMarker,
          isChecked,
          content: contentWithoutComments,
          rawContent: content,
          normalizedContent,
          id: taskId,
          indentLevel,
          line
        });

        // Log for debugging
        if (index < 5 || index > lines.length - 5) {
          console.log(`Extracted task: "${contentWithoutComments.substring(0, 30)}${contentWithoutComments.length > 30 ? '...' : ''}" (ID: ${taskId}, indent: ${indentLevel})`);
        }
      }
    });

    console.log(`Extracted ${tasks.length} tasks from markdown`);
    return tasks;
  }

  /**
   * Add a drag handle to a task item
   */
  addDragHandle(item) {
    const taskItem = item.querySelector('.task-list-item');
    if (taskItem) {
      // Check if action buttons already exist
      if (taskItem.querySelector('.task-action-buttons')) {
        return;
      }

      // Create action buttons container
      const actionButtons = document.createElement('div');
      actionButtons.className = 'task-action-buttons editor-admin-only';

      // Add rename button
      const renameBtn = document.createElement('button');
      renameBtn.className = 'task-action-btn task-rename-btn';
      renameBtn.title = window.i18n ? window.i18n.t('kanban.rename_task') : 'Rename task';
      renameBtn.innerHTML = '<i class="fa fa-pencil"></i>'; // Edit/pencil icon
      renameBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        this.renameTask(item);
      });

      // Add delete button
      const deleteBtn = document.createElement('button');
      deleteBtn.className = 'task-action-btn task-delete-btn';
      deleteBtn.title = window.i18n ? window.i18n.t('kanban.delete_task') : 'Delete task';
      deleteBtn.innerHTML = '<i class="fa fa-times"></i>'; // Times/cross icon
      deleteBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        this.deleteTask(item);
      });

      // Add drag handle
      const dragHandle = document.createElement('span');
      dragHandle.className = 'task-drag-handle';

      // Add buttons to container
      actionButtons.appendChild(renameBtn);
      actionButtons.appendChild(deleteBtn);
      actionButtons.appendChild(dragHandle);

      // Add container to task item
      taskItem.appendChild(actionButtons);
    }
  }

  /**
   * Convert processed HTML back to markdown for editing
   */
  convertHtmlToMarkdown(html) {
    // Create a temporary div to work with the HTML
    const tempDiv = document.createElement('div');
    tempDiv.innerHTML = html;

    // Convert HTML elements back to markdown
    let markdown = tempDiv.innerHTML;

    // Convert HTML tags back to markdown syntax
    markdown = markdown.replace(/<strong>(.*?)<\/strong>/g, '**$1**');
    markdown = markdown.replace(/<em>(.*?)<\/em>/g, '*$1*');
    markdown = markdown.replace(/<mark>(.*?)<\/mark>/g, '==$1==');
    markdown = markdown.replace(/<code>(.*?)<\/code>/g, '`$1`');
    markdown = markdown.replace(/<a href="([^"]*)">(.*?)<\/a>/g, '[$2]($1)');
    markdown = markdown.replace(/<del>(.*?)<\/del>/g, '~~$1~~'); // Add strikethrough

    // Get the text content and clean up any remaining HTML
    tempDiv.innerHTML = markdown;
    return tempDiv.textContent || tempDiv.innerText || '';
  }

  /**
   * Rename a task
   */
  renameTask(taskItem) {
    const taskTextElement = taskItem.querySelector('.task-text');
    if (!taskTextElement) return;

    // Check if we have the original markdown stored as a data attribute
    let originalMarkdown = taskItem.getAttribute('data-original-markdown');

    // If no stored original markdown, try to convert the displayed HTML back to markdown
    if (!originalMarkdown) {
      const htmlContent = taskTextElement.innerHTML;
      originalMarkdown = this.convertHtmlToMarkdown(htmlContent);

      // If that still doesn't work, fall back to plain text
      if (!originalMarkdown) {
        originalMarkdown = taskTextElement.textContent;
      }
    }

    const taskListItem = taskItem.querySelector('.task-list-item');

    // Disable dragging while editing
    taskItem.setAttribute('draggable', 'false');
    taskItem.classList.add('editing');

    // Create input field
    const input = document.createElement('input');
    input.type = 'text';
    input.className = 'task-rename-input';
    input.value = originalMarkdown; // Use the original markdown for editing

    // Hide the task text and show input
    taskTextElement.style.display = 'none';
    taskListItem.insertBefore(input, taskTextElement.nextSibling);

    // Focus the input
    input.focus();
    input.select();

    // Function to finish editing and restore dragging
    const finishEditing = () => {
      // Re-enable dragging
      taskItem.setAttribute('draggable', 'true');
      taskItem.classList.remove('editing');

      // Restore display
      taskTextElement.style.display = '';
      if (input.parentNode) {
        input.remove();
      }
    };

    // Handle input events
    input.addEventListener('keydown', e => {
      if (e.key === 'Enter') {
        e.preventDefault();
        const newText = input.value.trim();

        if (newText && newText !== originalMarkdown) {
          // Store the new markdown as a data attribute for reference
          taskItem.setAttribute('data-original-markdown', newText);

          // Update task text with processed markdown
          taskTextElement.innerHTML = this.processMarkdown(newText);

          // Save changes
          this.core.getPersistenceManager().saveKanbanChanges();
        }

        finishEditing();
      } else if (e.key === 'Escape') {
        // Cancel on escape
        finishEditing();
      }
    });

    // Also handle blur event to cancel
    input.addEventListener('blur', () => {
      // Small delay to allow for Enter key processing
      setTimeout(() => {
        finishEditing();
      }, 200);
    });

    // Prevent drag events from bubbling up from the input
    input.addEventListener('mousedown', e => {
      e.stopPropagation();
    });

    input.addEventListener('dragstart', e => {
      e.preventDefault();
      e.stopPropagation();
    });
  }

  /**
   * Delete a task
   */
  deleteTask(taskItem) {
    // Ask for confirmation using the DialogSystem
    window.DialogSystem.showConfirmDialog(
      window.i18n ? window.i18n.t('kanban.delete_task_title') : "Delete Task",
      window.i18n ? window.i18n.t('kanban.delete_task_confirm') : "Are you sure you want to delete this task?",
      (confirmed) => {
        if (!confirmed) return;

        // Find all descendants
        const descendants = this.core.getDragHandler().findAllDescendants(taskItem);

        // Remove descendants first
        descendants.forEach(descendant => {
          descendant.remove();
        });

        // Remove the task
        taskItem.remove();

        // Save changes
        this.core.getPersistenceManager().saveKanbanChanges();
      }
    );
  }

  /**
   * Create a new task with the given text
   */
  createNewTask(taskText, taskList) {
    // Generate a unique ID for this task
    const taskId = this.generateTaskId();

    // Create task container
    const taskContainer = document.createElement('li');
    taskContainer.className = 'task-list-item-container indent-0';
    taskContainer.setAttribute('data-indent-level', '0');
    taskContainer.setAttribute('data-task-id', taskId);
    taskContainer.setAttribute('data-is-new', 'true');
    taskContainer.style.listStyleType = 'none';
    taskContainer.setAttribute('draggable', 'true');

    // Store the original markdown text as a data attribute
    taskContainer.setAttribute('data-original-markdown', taskText);

    // Process markdown in task text
    const processedText = this.processMarkdown(taskText);

    // Create task content
    taskContainer.innerHTML = `
      <span class="task-list-item">
        <input type="checkbox" class="task-checkbox">
        <span class="task-text">${processedText}</span>
        <span class="save-state"></span>
      </span>
    `;

    // Store in the task ID map
    this.taskIdMap.set(taskId, taskContainer);

    // Add drag handle and action buttons
    this.addDragHandle(taskContainer);

    // Setup drag events for the new task
    this.core.getDragHandler().setupDragForNewTask(taskContainer);

    // Setup checkbox for newly created task
    this.setupCheckboxForTask(taskContainer);

    // Add to the beginning of the list
    if (taskList.firstChild) {
      taskList.insertBefore(taskContainer, taskList.firstChild);
    } else {
      taskList.appendChild(taskContainer);
    }

    console.log('New task created:', taskText, 'with ID:', taskId);
    return taskContainer;
  }

  /**
   * Setup add task buttons
   */
  setupAddTaskButtons() {
    const addTaskButtons = document.querySelectorAll('.add-task-btn');
    console.log('Add task buttons found:', addTaskButtons.length);

    addTaskButtons.forEach(button => {
      button.addEventListener('click', e => {
        e.preventDefault();

        // Find the column content where we need to add the task
        const column = button.closest('.kanban-column');
        const columnContent = column.querySelector('.kanban-column-content');
        const taskList = columnContent.querySelector('.task-list');

        // Check if there's already an input field
        if (columnContent.querySelector('.new-task-input-container')) {
          return; // Don't add another input if one already exists
        }

        this.showNewTaskInput(columnContent, taskList);
      });
    });
  }

  /**
   * Show new task input field
   */
  showNewTaskInput(columnContent, taskList) {
    // Create input container
    const inputContainer = document.createElement('div');
    inputContainer.className = 'new-task-input-container';

    // Create input field
    const input = document.createElement('input');
    input.type = 'text';
    input.className = 'new-task-input';
    input.placeholder = window.i18n ? window.i18n.t('kanban.enter_task_name') : 'Enter task name';

    // Add input to container
    inputContainer.appendChild(input);

    // Insert at the top of the column
    columnContent.insertBefore(inputContainer, columnContent.firstChild);

    // Focus the input
    input.focus();

    // Handle input events
    input.addEventListener('keydown', e => {
      if (e.key === 'Enter') {
        e.preventDefault();
        const taskText = input.value.trim();

        if (taskText) {
          // Create new task
          this.createNewTask(taskText, taskList);

          // Save changes
          this.core.getPersistenceManager().saveKanbanChanges();

          // Remove input
          inputContainer.remove();
        }
      } else if (e.key === 'Escape') {
        // Cancel on escape
        inputContainer.remove();
      }
    });

    // Also handle blur event to cancel
    input.addEventListener('blur', () => {
      // Small delay to allow for Enter key processing
      setTimeout(() => {
        if (inputContainer.parentNode) {
          inputContainer.remove();
        }
      }, 200);
    });
  }

  /**
   * Setup checkbox for a task
   */
  setupCheckboxForTask(taskContainer) {
    const checkbox = taskContainer.querySelector('input[type="checkbox"]');
    if (!checkbox) return;

    // Add click handler for the checkbox
    checkbox.addEventListener('click', this.taskCheckboxHandler);
  }

  /**
   * Process markdown formatting in text
   */
  processMarkdown(text) {
    // Escape HTML first to prevent XSS
    let processed = this.escapeHtml(text);

    // Process bold text (**text** or __text__)
    processed = processed.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
    processed = processed.replace(/__([^_]+)__/g, '<strong>$1</strong>');

    // Process italic text (*text* or _text_)
    processed = processed.replace(/\*([^*]+)\*/g, '<em>$1</em>');
    processed = processed.replace(/_([^_]+)_/g, '<em>$1</em>');

    // Process highlight text (==text==)
    processed = processed.replace(/==([^=]+)==/g, '<mark>$1</mark>');

    // Process images before links because image syntax contains link syntax.
    processed = processed.replace(/!\[([^\]]*)\]\(([^)]+)\)/g, function(match, alt, url) {
      let safeURL = WikiDOM.markdownURL(url, true);
      if (!safeURL) return alt;
      if (!/^([/]|https?:)/i.test(safeURL)) {
        const docPath = (typeof getCurrentDocPath === 'function') ? getCurrentDocPath() : '';
        safeURL = '/api/files/' + docPath + '/' + safeURL;
      }
      return '<img src="' + safeURL + '" alt="' + alt + '">';
    });

    // Process links [text](url), allowing only browser-safe schemes.
    processed = processed.replace(/\[([^\]]+)\]\(([^)]+)\)/g, function(match, label, url) {
      const safeURL = WikiDOM.markdownURL(url);
      return safeURL ? '<a href="' + safeURL + '">' + label + '</a>' : label;
    });

    // Process inline code (`code`)
    processed = processed.replace(/`([^`]+)`/g, '<code>$1</code>');

    // Process strikethrough text (~~text~~)
    processed = processed.replace(/~~([^~]+)~~/g, '<del>$1</del>');

    return processed;
  }

  /**
   * Escape HTML special characters to prevent XSS
   */
  escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }

  /**
   * Update task tracking after a successful save
   */
  updateTaskTrackingAfterSave() {
    // Clear the "new" and "moved" flags from all tasks
    document.querySelectorAll('.task-list-item-container').forEach(task => {
      if (task.hasAttribute('data-is-new')) {
        task.removeAttribute('data-is-new');
      }
      if (task.hasAttribute('data-was-moved')) {
        task.removeAttribute('data-was-moved');
      }
    });

    console.log('Task tracking updated after save');
  }

  /**
   * Get task by ID
   */
  getTaskById(taskId) {
    return this.taskIdMap.get(taskId);
  }

  /**
   * Get all task IDs
   */
  getAllTaskIds() {
    return Array.from(this.taskIdMap.keys());
  }

  /**
   * Clean up task management
   */
  destroy() {
    this.taskIdMap.clear();
    this.taskCheckboxHandler = null;
    this.initialized = false;
    console.log('Task manager destroyed');
  }
}
