// kanban-ui.js - UI interactions and visual feedback
class KanbanUIManager {
  constructor(core) {
    this.core = core;
    this.initialized = false;

    // UI state tracking
    this.activeInputs = new Set();
    this.statusTimers = new Map();
  }

  init() {
    console.log('Initializing KanbanUIManager');

    // Add CSS class to indicate drag and drop is enabled
    document.body.classList.add('kanban-drag-enabled');

    this.initialized = true;
    console.log('KanbanUIManager initialized');
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
   * Get a unique ID for a column for timer tracking
   */
  getColumnId(column) {
    const header = column.querySelector('.kanban-column-header .column-title');
    return header ? header.textContent.trim() : Math.random().toString(36).substr(2, 9);
  }

  /**
   * Create a new task input field in a column
   */
  createNewTaskInput(column, onComplete, onCancel) {
    const columnContent = column.querySelector('.kanban-column-content');
    const taskList = columnContent.querySelector('.task-list');

    // Check if there's already an input field
    if (columnContent.querySelector('.new-task-input-container')) {
      return null; // Don't add another input if one already exists
    }

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

    // Track this input
    this.activeInputs.add(input);

    // Focus the input
    input.focus();

    // Handle input events
    input.addEventListener('keydown', (e) => {
      if (e.key === 'Enter') {
        e.preventDefault();
        const taskText = input.value.trim();

        if (taskText) {
          // Call completion callback
          if (onComplete) {
            onComplete(taskText, taskList);
          }
        }

        // Remove input
        this.removeInput(inputContainer, input);
      } else if (e.key === 'Escape') {
        // Cancel on escape
        if (onCancel) {
          onCancel();
        }
        this.removeInput(inputContainer, input);
      }
    });

    // Also handle blur event to cancel
    input.addEventListener('blur', () => {
      // Small delay to allow for Enter key processing
      setTimeout(() => {
        if (inputContainer.parentNode) {
          if (onCancel) {
            onCancel();
          }
          this.removeInput(inputContainer, input);
        }
      }, 200);
    });

    return { input, container: inputContainer };
  }

  /**
   * Create a column rename input field
   */
  createColumnRenameInput(columnHeader, columnTitle, onComplete, onCancel) {
    // Store the original title text
    const originalTitle = columnTitle.textContent;

    // Create input field
    const input = document.createElement('input');
    input.type = 'text';
    input.className = 'column-rename-input';
    input.value = originalTitle;
    input.style.width = '100%';
    input.style.padding = '4px';
    input.style.border = '1px solid var(--border-color)';
    input.style.borderRadius = '4px';
    input.style.backgroundColor = 'var(--bg-color)';
    input.style.color = 'var(--text-color)';

    // Hide the title and show input
    columnTitle.style.display = 'none';
    columnHeader.insertBefore(input, columnTitle);

    // Track this input
    this.activeInputs.add(input);

    // Focus the input
    input.focus();
    input.select();

    // Store the original column title as a data attribute to prevent duplication
    columnHeader.setAttribute('data-original-title', originalTitle);

    // Handle input events
    input.addEventListener('keydown', (e) => {
      if (e.key === 'Enter') {
        e.preventDefault();
        const newTitle = input.value.trim();

        if (newTitle && newTitle !== originalTitle) {
          // Check if the new name already exists
          const isDuplicate = this.checkForDuplicateColumnName(newTitle, columnTitle);

          // If this is a duplicate, mark it as such for internal tracking
          if (isDuplicate) {
            columnHeader.setAttribute('data-is-duplicate', 'true');
            console.log(`Column name "${newTitle}" already exists, but will be allowed as duplicate`);
          } else {
            columnHeader.removeAttribute('data-is-duplicate');
          }

          // Update the column title
          columnTitle.textContent = newTitle;

          // Call completion callback
          if (onComplete) {
            onComplete(newTitle, originalTitle);
          }
        }

        // Restore display
        this.restoreColumnTitle(columnTitle, input);
      } else if (e.key === 'Escape') {
        // Cancel on escape
        if (onCancel) {
          onCancel();
        }
        this.restoreColumnTitle(columnTitle, input);
      }
    });

    // Also handle blur event to cancel
    input.addEventListener('blur', () => {
      // Small delay to allow for Enter key processing
      setTimeout(() => {
        if (input.parentNode) {
          if (onCancel) {
            onCancel();
          }
          this.restoreColumnTitle(columnTitle, input);
        }
      }, 200);
    });

    return input;
  }

  /**
   * Create a task rename input field
   */
  createTaskRenameInput(taskItem, taskTextElement, onComplete, onCancel) {
    // Check if we have the original markdown stored as a data attribute
    let originalMarkdown = taskItem.getAttribute('data-original-markdown');

    // If no stored original markdown, fall back to the displayed text
    if (!originalMarkdown) {
      originalMarkdown = taskTextElement.textContent;
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

    // Track this input
    this.activeInputs.add(input);

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

      // Remove from tracking
      this.activeInputs.delete(input);
    };

    // Handle input events
    input.addEventListener('keydown', (e) => {
      if (e.key === 'Enter') {
        e.preventDefault();
        const newText = input.value.trim();

        if (newText && newText !== originalMarkdown) {
          // Store the new markdown as a data attribute for reference
          taskItem.setAttribute('data-original-markdown', newText);

          // Update task text with processed markdown
          taskTextElement.innerHTML = this.processMarkdown(newText);

          // Call completion callback
          if (onComplete) {
            onComplete(newText, originalMarkdown);
          }
        }

        finishEditing();
      } else if (e.key === 'Escape') {
        // Cancel on escape
        if (onCancel) {
          onCancel();
        }
        finishEditing();
      }
    });

    // Also handle blur event to cancel
    input.addEventListener('blur', () => {
      // Small delay to allow for Enter key processing
      setTimeout(() => {
        if (onCancel) {
          onCancel();
        }
        finishEditing();
      }, 200);
    });

    // Prevent drag events from bubbling up from the input
    input.addEventListener('mousedown', (e) => {
      e.stopPropagation();
    });

    input.addEventListener('dragstart', (e) => {
      e.preventDefault();
      e.stopPropagation();
    });

    return input;
  }

  /**
   * Check if a column name already exists
   */
  checkForDuplicateColumnName(newTitle, excludeColumn) {
    const existingColumns = document.querySelectorAll('.kanban-column-header .column-title');
    let isDuplicate = false;

    existingColumns.forEach(col => {
      if (col !== excludeColumn && col.textContent.trim().toLowerCase() === newTitle.toLowerCase()) {
        isDuplicate = true;
      }
    });

    return isDuplicate;
  }

  /**
   * Restore column title display
   */
  restoreColumnTitle(columnTitle, input) {
    columnTitle.style.display = '';
    if (input.parentNode) {
      input.remove();
    }
    this.activeInputs.delete(input);
  }

  /**
   * Remove an input field and clean up tracking
   */
  removeInput(container, input) {
    if (container.parentNode) {
      container.remove();
    }
    this.activeInputs.delete(input);
  }

  /**
   * Show a dialog with the given title and message
   */
  showDialog(title, message, onConfirm, onCancel) {
    // Use the DialogSystem if available
    if (window.DialogSystem && window.DialogSystem.showConfirmDialog) {
      window.DialogSystem.showConfirmDialog(title, message, (confirmed) => {
        if (confirmed && onConfirm) {
          onConfirm();
        } else if (!confirmed && onCancel) {
          onCancel();
        }
      });
    } else {
      // Fallback to native confirm
      const confirmed = confirm(`${title}\n\n${message}`);
      if (confirmed && onConfirm) {
        onConfirm();
      } else if (!confirmed && onCancel) {
        onCancel();
      }
    }
  }

  /**
   * Show/hide the add column dialog
   * @param {boolean} show - Whether to show or hide the dialog
   */
  showAddColumnDialog(show = true) {
    const addColumnDialog = document.querySelector('.add-column-dialog');
    const columnNameInput = document.querySelector('#columnName');

    if (!addColumnDialog) {
      console.error('Add column dialog not found');
      return;
    }

    if (show) {
      // Reset input
      if (columnNameInput) {
        columnNameInput.value = '';
      }

      addColumnDialog.classList.add('active');

      // Focus input after a short delay
      setTimeout(() => {
        if (columnNameInput) {
          columnNameInput.focus();
        }
      }, 100);
    } else {
      addColumnDialog.classList.remove('active');
    }
  }

  /**
   * Clean up all active inputs
   */
  cleanupActiveInputs() {
    this.activeInputs.forEach(input => {
      if (input.parentNode) {
        // Trigger blur to clean up properly
        input.blur();
      }
    });
    this.activeInputs.clear();
  }

  /**
   * Clean up all visual feedback from dragging
   */
  cleanupDragVisuals() {
    // Remove all drag-over indicators
    document.querySelectorAll('.drag-over').forEach(el => {
      el.classList.remove('drag-over');
    });

    // Remove drop indicator
    this.removeDropIndicator();

    // Remove task drop targets
    this.removeDropHighlights();
  }

  /**
   * Get or create the drop indicator element
   */
  getOrCreateDropIndicator() {
    let indicator = document.querySelector('.drop-indicator');

    if (!indicator) {
      indicator = document.createElement('div');
      indicator.className = 'drop-indicator';

      // Add a visual cue to make the drop position more obvious
      const arrowDiv = document.createElement('div');
      arrowDiv.className = 'drop-indicator-arrow';
      indicator.appendChild(arrowDiv);
    } else {
      // Remove from current position
      if (indicator.parentNode) {
        indicator.parentNode.removeChild(indicator);
      }
    }

    return indicator;
  }

  /**
   * Remove all drop highlights
   */
  removeDropHighlights() {
    document.querySelectorAll('.task-drop-target').forEach(el => {
      el.classList.remove('task-drop-target');
    });
  }

  /**
   * Remove the drop indicator
   */
  removeDropIndicator() {
    const dropIndicator = document.querySelector('.drop-indicator');
    if (dropIndicator && dropIndicator.parentNode) {
      dropIndicator.parentNode.removeChild(dropIndicator);
    }
  }

  /**
   * Add visual feedback for task drop target
   */
  addTaskDropTarget(task) {
    task.classList.add('task-drop-target');
  }

  /**
   * Remove visual feedback for task drop target
   */
  removeTaskDropTarget(task) {
    task.classList.remove('task-drop-target');
  }

  /**
   * Add visual feedback for column drag over
   */
  addColumnDragOver(column) {
    column.classList.add('drag-over');
  }

  /**
   * Remove visual feedback for column drag over
   */
  removeColumnDragOver(column) {
    column.classList.remove('drag-over');
  }

  /**
   * Add visual feedback for dragging item
   */
  addDraggingFeedback(item) {
    item.classList.add('dragging');
  }

  /**
   * Remove visual feedback for dragging item
   */
  removeDraggingFeedback(item) {
    item.classList.remove('dragging');
  }

  /**
   * Clear all status timers
   */
  clearStatusTimers() {
    this.statusTimers.forEach(timer => {
      clearTimeout(timer);
    });
    this.statusTimers.clear();
  }

  /**
   * Check if the UI manager is initialized
   */
  isInitialized() {
    return this.initialized;
  }

  /**
   * Get the number of active inputs
   */
  getActiveInputCount() {
    return this.activeInputs.size;
  }

  /**
   * Check if there are any active inputs
   */
  hasActiveInputs() {
    return this.activeInputs.size > 0;
  }

  /**
   * Cleanup and destroy the UI manager
   */
  destroy() {
    if (!this.initialized) {
      return;
    }

    console.log('Destroying KanbanUIManager');

    // Clean up all visual feedback
    this.cleanupDragVisuals();

    // Clean up all active inputs
    this.cleanupActiveInputs();

    // Clear all status timers
    this.clearStatusTimers();

    // Remove CSS class
    document.body.classList.remove('kanban-drag-enabled');

    // Clear core reference
    this.core = null;

    this.initialized = false;
    console.log('KanbanUIManager destroyed');
  }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
  module.exports = KanbanUIManager;
}
