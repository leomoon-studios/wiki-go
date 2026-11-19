// kanban-persistence.js - Save/load operations
class KanbanPersistenceManager {
  constructor(core) {
    this.core = core;
    this.docPath = core.docPath;

    // Track processed headers to avoid duplicates
    this.processedHeaders = new Set();
    this.processedBoards = new Set();

    // Maps to store original formatting
    this.originalSectionHeaders = new Map();
    this.originalBoardHeaders = new Map();
    this.originalTaskFormatting = new Map();

    // Save state tracking
    this.isSaving = false;
    this.saveQueue = [];
  }

  init() {
    console.log('Initializing KanbanPersistenceManager');

    // Pre-load original markdown for better performance
    this.preloadOriginalMarkdown();

    console.log('KanbanPersistenceManager initialized');
  }

  /**
   * Pre-load the original markdown to cache formatting information
   */
  async preloadOriginalMarkdown() {
    try {
      const response = await fetch(wikiURL(`/api/source/${this.docPath}`));
      if (response.ok) {
        const markdown = await response.text();
        this.cacheOriginalFormatting(markdown);
        console.log('Original markdown cached for formatting preservation');
      }
    } catch (error) {
      console.warn('Could not preload original markdown:', error);
    }
  }

  /**
   * Cache original formatting from markdown
   */
  cacheOriginalFormatting(markdown) {
    const lines = markdown.split(/\r?\n/);

    // Find frontmatter end
    const frontmatterEndIndex = lines.findIndex((line, index) =>
      index > 0 && line.trim() === '---' && lines[0].trim() === '---'
    );

    if (frontmatterEndIndex > 0) {
      this.collectOriginalFormatting(lines, frontmatterEndIndex, this.originalBoardHeaders, this.originalSectionHeaders, this.originalTaskFormatting);
    }
  }

  /**
   * Save kanban changes to the server
   */
  async saveKanbanChanges() {
    // Prevent multiple simultaneous saves
    if (this.isSaving) {
      console.log('Save already in progress, queuing request');
      return new Promise((resolve, reject) => {
        this.saveQueue.push({ resolve, reject });
      });
    }

    this.isSaving = true;

    try {
      console.log('Starting kanban save operation');

      // 1. Fetch current markdown
      const srcResp = await fetch(wikiURL(`/api/source/${this.docPath}`));
      if (!srcResp.ok) throw new Error('source fetch failed');
      const markdown = await srcResp.text();

      // 2. Build updated markdown based on the current DOM structure
      const updatedMarkdown = this.buildUpdatedMarkdown(markdown);

      // 3. Save updated markdown
      const saveResp = await fetch(wikiURL(`/api/save/${this.docPath}`), {
        method: 'POST',
        headers: { 'Content-Type': 'text/markdown' },
        body: updatedMarkdown,
      });

      if (!saveResp.ok) throw new Error('save failed');

      // 4. After successful save, update task tracking
      this.updateTaskTrackingAfterSave();

      // Success - no visual feedback needed for columns

      console.log('Kanban save completed successfully');

      // Process any queued save requests
      this.processQueuedSaves(null);

    } catch (err) {
      console.error('Error saving kanban changes:', err);

      // Show error indicator
      if (this.core.columnManager) {
        this.core.columnManager.showColumnStatus('error', 'error', 3000);
      }

      // Process any queued save requests with error
      this.processQueuedSaves(err);

      throw err;
    } finally {
      this.isSaving = false;
    }
  }

  /**
   * Process queued save requests
   */
  processQueuedSaves(error) {
    const queue = [...this.saveQueue];
    this.saveQueue = [];

    queue.forEach(({ resolve, reject }) => {
      if (error) {
        reject(error);
      } else {
        resolve();
      }
    });
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
   * Build updated markdown based on the current DOM structure
   */
  buildUpdatedMarkdown(originalMarkdown) {
    console.log('Building updated markdown');

    // Split the original markdown into lines
    const lines = originalMarkdown.split(/\r?\n/);

    // Extract frontmatter and content before the first kanban section
    const frontmatterEndIndex = lines.findIndex((line, index) =>
      index > 0 && line.trim() === '---' && lines[0].trim() === '---'
    );

    // Reset processed boards for this save operation
    this.processedBoards.clear();

    // Get column information from the column manager
    const renamedColumns = this.core.columnManager ? this.core.columnManager.getRenamedColumns() : new Map();
    const allBoardColumnCounts = this.core.columnManager ? this.core.columnManager.getAllBoardColumnCounts() : new Map();

    // Collect original formatting if not already cached
    if (this.originalSectionHeaders.size === 0) {
      this.collectOriginalFormatting(lines, frontmatterEndIndex, this.originalBoardHeaders, this.originalSectionHeaders, this.originalTaskFormatting);
    }

    // Iterate through original markdown and replace kanban sections as we encounter them
    const updatedLines = [];
    const kanbanContainers = document.querySelectorAll('.kanban-container');
    let containerIndex = 0;

    let i = 0;

    while (i < lines.length) {
      const line = lines[i];

      // Check if this is a board header (H4)
      const boardMatch = line.match(/^####\s+(.+)$/);
      if (boardMatch) {
        const boardTitle = boardMatch[1].trim();

        // Check if we have a corresponding kanban container in the DOM
        const matchingContainer = this.findMatchingKanbanContainer(kanbanContainers, boardTitle, containerIndex);

        if (matchingContainer) {
          // Replace this kanban section with the updated version from the DOM
          console.log(`Replacing kanban board: ${boardTitle}`);

          // Process this kanban container and add its content
          this.processKanbanContainer(matchingContainer.container, updatedLines, renamedColumns, allBoardColumnCounts);

          // Skip the original kanban content in the markdown
          i = this.skipOriginalKanbanContent(lines, i);
          containerIndex = matchingContainer.index + 1;
        } else {
          // This is a kanban board that no longer exists in the DOM, skip it
          console.log(`Skipping removed kanban board: ${boardTitle}`);
          i = this.skipOriginalKanbanContent(lines, i);
        }
      } else {
        // This is not a kanban board header, preserve the line as-is
        updatedLines.push(line);
        i++;
      }
    }

    // Add any remaining kanban containers that weren't found in the original markdown
    // (these would be newly created boards)
    for (let j = containerIndex; j < kanbanContainers.length; j++) {
      console.log(`Adding new kanban container ${j + 1}`);
      updatedLines.push(''); // Add empty line before new board
      this.processKanbanContainer(kanbanContainers[j], updatedLines, renamedColumns, allBoardColumnCounts);
    }

    // Log the first few lines of the updated markdown for debugging
    console.log('Updated markdown first few lines:', updatedLines.slice(0, Math.min(10, updatedLines.length)).join('\n'));

    return updatedLines.join('\n');
  }

  /**
   * Find a matching kanban container in the DOM for a given board title
   */
  findMatchingKanbanContainer(containers, boardTitle, startIndex) {
    // First, try to find an exact match by title
    for (let i = startIndex; i < containers.length; i++) {
      const container = containers[i];
      const boardTitleElement = container.querySelector('.kanban-board-title');
      const containerBoardTitle = boardTitleElement ? boardTitleElement.textContent.trim() : '';

      if (containerBoardTitle === boardTitle) {
        return { container, index: i };
      }
    }

    // If no exact match found and we're looking for a non-empty title, return null
    if (boardTitle.trim() !== '') {
      return null;
    }

    // If we're looking for an empty title (first board might not have a title),
    // return the next available container without a title
    for (let i = startIndex; i < containers.length; i++) {
      const container = containers[i];
      const boardTitleElement = container.querySelector('.kanban-board-title');
      const containerBoardTitle = boardTitleElement ? boardTitleElement.textContent.trim() : '';

      if (containerBoardTitle === '') {
        return { container, index: i };
      }
    }

    return null;
  }

  /**
   * Skip the original kanban content in the markdown starting from the board header
   */
  skipOriginalKanbanContent(lines, startIndex) {
    let i = startIndex + 1; // Start after the board header

    while (i < lines.length) {
      const line = lines[i];

      // If we encounter another H4 header, we've reached the next board or section
      if (line.match(/^####\s+/)) {
        break;
      }

      // If we encounter non-kanban content (not H5 header, not task, not empty line)
      const isH5Header = line.match(/^#####\s+/);
      const isTaskLine = line.match(/^\s*[-*+]\s+\[([ xX])\]\s+/);
      const isEmpty = line.trim() === '';

      if (!isH5Header && !isTaskLine && !isEmpty) {
        // This is regular content after the kanban section, don't skip it
        break;
      }

      i++;
    }

    return i;
  }

  /**
   * Process a single kanban container
   */
  processKanbanContainer(container, updatedLines, renamedColumns, allBoardColumnCounts) {
    // Get or create board ID
    const boardId = this.getBoardId(container);

    // Reset processed headers for this specific board
    const boardProcessedHeaders = new Set();

    // Get board title if it exists
    const boardTitleElement = container.querySelector('.kanban-board-title');
    let boardTitle = boardTitleElement ? boardTitleElement.textContent.trim() : '';

    // Add the board title if it exists
    if (boardTitle) {
      const boardKey = boardTitle.toLowerCase();

      // Skip if already processed
      if (this.processedBoards.has(boardKey)) {
        console.log(`Board ${boardTitle} already processed, skipping`);
        return;
      }

      // Mark as processed
      this.processedBoards.add(boardKey);

      // Use original board header format if available
      const originalBoardHeader = this.originalBoardHeaders.get(boardKey) || boardTitle;
      updatedLines.push(`#### ${originalBoardHeader}`);
      updatedLines.push(''); // Add empty line after board title
    }

    // Get board-specific column counts
    const boardColumnCounts = allBoardColumnCounts.get(boardId) || new Map();

    // Process each column in this container
    const columns = container.querySelectorAll('.kanban-column');
    columns.forEach((column, columnIndex) => {
      this.processKanbanColumn(column, updatedLines, renamedColumns, boardColumnCounts, boardProcessedHeaders, boardId);
    });

    // Add empty line after each board
    updatedLines.push('');
  }

  /**
   * Get or create board ID for a kanban container
   */
  getBoardId(container) {
    let boardId = container.getAttribute('data-board-id');
    if (!boardId) {
      const boardTitle = container.querySelector('.kanban-board-title');
      const title = boardTitle ? boardTitle.textContent.trim() : '';
      boardId = title || `board-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
      container.setAttribute('data-board-id', boardId);
    }
    return boardId;
  }

  /**
   * Process a single kanban column
   */
  processKanbanColumn(column, updatedLines, renamedColumns, boardColumnCounts, boardProcessedHeaders, boardId) {
    // Get the column header text without status indicators
    const header = column.querySelector('.kanban-column-header');
    if (!header) return;

    // Extract the column title from the title span
    const titleSpan = header.querySelector('.column-title');
    let headerText = titleSpan ? titleSpan.textContent.trim() : header.textContent.trim();

    // Clean up any status indicators
    headerText = this.cleanHeaderText(headerText);

    console.log(`Processing column: ${headerText} in board: ${boardId}`);

    // Check if this column was renamed
    const originalTitle = header.getAttribute('data-original-title');
    if (originalTitle && originalTitle !== headerText) {
      console.log(`Column was renamed from "${originalTitle}" to "${headerText}"`);
      // Mark the original title as processed for this board only
      boardProcessedHeaders.add(`${originalTitle.toLowerCase()}-1`);
    }

    // Handle duplicate column names within this board
    const headerKey = headerText.toLowerCase();

    // Count how many times this column name appears in this board
    let currentCount = 1;
    const allColumnsInBoard = column.closest('.kanban-container').querySelectorAll('.kanban-column');
    let foundSelf = false;

    allColumnsInBoard.forEach(col => {
      const colHeader = col.querySelector('.column-title');
      if (colHeader) {
        const colText = this.cleanHeaderText(colHeader.textContent.trim()).toLowerCase();
        if (colText === headerKey) {
          if (col === column) {
            foundSelf = true;
          } else if (!foundSelf) {
            currentCount++;
          }
        }
      }
    });

    // Create unique markdown header for duplicates within this board
    let markdownHeaderText = headerText;
    if (currentCount > 1) {
      markdownHeaderText = `${headerText} (${currentCount})`;
      console.log(`Handling duplicate column name in board ${boardId}: "${headerText}" -> "${markdownHeaderText}"`);
    }

    // Create board-specific processed key
    const processedKey = `${headerKey}-${currentCount}`;

    // Skip if already processed in this board
    if (boardProcessedHeaders.has(processedKey)) {
      console.log(`Column ${headerText} (${currentCount}) already processed in board ${boardId}, skipping`);
      return;
    }

    // Mark as processed for this board
    boardProcessedHeaders.add(processedKey);

    // Use original header format if available, otherwise use our generated one
    const originalHeader = this.originalSectionHeaders.get(headerKey) || markdownHeaderText;
    updatedLines.push(`##### ${originalHeader}`);

    // Process each task in this column
    this.processColumnTasks(column, updatedLines);

    // Add empty line after each column
    updatedLines.push('');
  }

  /**
   * Clean header text of status indicators and symbols
   */
  cleanHeaderText(headerText) {
    return headerText
      .replace(/\s*saving…\s*/g, '')
      .replace(/\s*saved\s*/g, '')
      .replace(/\s*error\s*/g, '')
      .replace(/\s*\(Saving\.\.\.\)\s*/g, '')
      .replace(/\s*\(Saved\)\s*/g, '')
      .replace(/\s*\(Error saving\)\s*/g, '')
      .replace(/\+/g, ''); // Remove + signs from add button
  }

  /**
   * Process all tasks in a column and add them to the updated markdown
   */
  processColumnTasks(column, updatedLines) {
    const tasks = column.querySelectorAll('.task-list-item-container');
    console.log(`Processing ${tasks.length} tasks in column`);

    if (tasks.length === 0) return;

    // Process tasks in the exact DOM order they appear
    // This preserves the hierarchy and ordering after drag and drop
    tasks.forEach((task, index) => {
      const taskLine = this.processSingleTask(task, index, tasks.length);
      if (taskLine) {
        updatedLines.push(taskLine);
      }
    });
  }

  /**
   * Process a single task and return its markdown line
   */
  processSingleTask(task, index, totalTasks) {
    const taskTextElement = task.querySelector('.task-text');
    if (!taskTextElement) return null;

    // Get the displayed text content
    const taskText = taskTextElement.textContent.trim();

    // Get the task ID
    const taskId = task.getAttribute('data-task-id');

    // Check if we have the original markdown stored as a data attribute
    // This is crucial for preserving formatting
    const originalMarkdown = task.getAttribute('data-original-markdown');

    const checkbox = task.querySelector('.task-checkbox');
    const isChecked = checkbox && checkbox.checked;
    const indentLevel = parseInt(task.getAttribute('data-indent-level') || '0');
    const indent = '  '.repeat(indentLevel);

    // Try to find original formatting from the markdown file
    const plainTextKey = taskText.replace(/\s+/g, ' ').trim();
    const originalFormat = this.originalTaskFormatting.get(plainTextKey);

    // Log task info for debugging
    if (index === 0 || index === totalTasks - 1 || indentLevel > 0) {
      console.log(`Task ${index + 1}/${totalTasks}: "${taskText}" (indent: ${indentLevel}, checked: ${isChecked}, id: ${taskId || 'none'})`);
    }

    // Priority for saving:
    // 1. Use data-original-markdown if available (user edited or created)
    // 2. Use original formatting from the markdown file
    // 3. Fallback to plain text
    let taskLine;

    if (originalMarkdown) {
      // Use the stored original markdown - this preserves user edits
      taskLine = `${indent}- [${isChecked ? 'x' : ' '}] ${originalMarkdown}`;
    } else if (originalFormat) {
      // Use original formatting but update check status and indentation
      taskLine = `${indent}${originalFormat.listMarker} [${isChecked ? 'x' : ' '}] ${originalFormat.taskContent}`;
    } else {
      // Fallback to basic formatting
      taskLine = `${indent}- [${isChecked ? 'x' : ' '}] ${taskText}`;
    }

    // Add task ID as a comment if available
    if (taskId) {
      // Check if the line already has a task ID comment
      if (!taskLine.includes('<!-- task-id:')) {
        taskLine += ` <!-- task-id: ${taskId} -->`;
      }
    }

    return taskLine;
  }

  /**
   * Collect original formatting from the markdown
   */
  collectOriginalFormatting(lines, startIndex, boardHeaders, sectionHeaders, taskFormatting) {
    let i = startIndex + 1;
    let currentBoard = null;

    while (i < lines.length) {
      // Check for H4 board headers
      const boardMatch = lines[i].match(/^####\s+(.+)$/);
      if (boardMatch) {
        const boardTitle = boardMatch[1].trim();
        const cleanBoardTitle = this.cleanHeaderText(boardTitle);
        currentBoard = cleanBoardTitle.toLowerCase();
        boardHeaders.set(currentBoard, boardTitle);
        i++;
        continue;
      }

      // Check for H5 column headers
      const headerMatch = lines[i].match(/^#####\s+(.+)$/);
      if (headerMatch) {
        const sectionTitle = headerMatch[1].trim();
        // Clean the title of any status indicators and + signs
        const cleanTitle = this.cleanHeaderText(sectionTitle);

        // Check if the next lines contain task items
        let hasTaskItems = false;
        let j = i + 1;
        while (j < lines.length && !lines[j].match(/^#{4,5}\s+/)) {
          const taskMatch = lines[j].match(/^(\s*)([-*+])\s+\[([ xX])\]\s+(.+)$/);
          if (taskMatch) {
            hasTaskItems = true;
            this.processOriginalTaskFormatting(taskMatch, taskFormatting, j, i);
          }
          j++;
        }

        if (hasTaskItems) {
          // This is a kanban section, store the original header
          sectionHeaders.set(cleanTitle.toLowerCase(), sectionTitle);
        }
      }
      i++;
    }
  }

  /**
   * Process original task formatting from regex match
   */
  processOriginalTaskFormatting(taskMatch, taskFormatting, lineIndex, sectionStart) {
    const indent = taskMatch[1];
    const listMarker = taskMatch[2];
    const checkStatus = taskMatch[3];
    const taskContent = taskMatch[4]; // This is the raw markdown

    // Check if there's a task ID in the content
    let taskId = null;
    const idMatch = taskContent.match(/<!--\s*task-id:\s*([a-zA-Z0-9_]+)\s*-->/);
    if (idMatch) {
      taskId = idMatch[1];
    }

    // Get content without comments
    const contentWithoutComments = taskContent.split('<!--')[0].trim();

    // Store the raw markdown content with its formatting intact
    taskFormatting.set(contentWithoutComments, {
      indent,
      listMarker,
      checkStatus,
      taskContent: contentWithoutComments,
      raw: taskContent, // Store the raw markdown
      taskId // Store the task ID if found
    });

    // Also store a plain text version for better matching
    const plainTextKey = this.normalizeTaskContent(contentWithoutComments);

    if (plainTextKey !== contentWithoutComments) {
      taskFormatting.set(plainTextKey, {
        indent,
        listMarker,
        checkStatus,
        taskContent: contentWithoutComments,
        raw: taskContent,
        taskId
      });
    }

    // Log task IDs for debugging (only first few)
    if (taskId && lineIndex < sectionStart + 5) {
      console.log(`Found task with ID: ${taskId}, content: "${contentWithoutComments.substring(0, 30)}${contentWithoutComments.length > 30 ? '...' : ''}", indent: ${indent.length / 2}`);
    }
  }

  /**
   * Normalize task content for better matching
   */
  normalizeTaskContent(content) {
    return content
      .replace(/\*\*([^*]+)\*\*/g, '$1') // Remove bold
      .replace(/__([^_]+)__/g, '$1')     // Remove bold with underscore
      .replace(/\*([^*]+)\*/g, '$1')     // Remove italic
      .replace(/_([^_]+)_/g, '$1')       // Remove italic with underscore
      .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1') // Remove links
      .replace(/`([^`]+)`/g, '$1')       // Remove code
      .trim();
  }

  /**
   * Get original task formatting for a given task
   */
  getOriginalTaskFormatting(taskText) {
    const plainTextKey = taskText.replace(/\s+/g, ' ').trim();
    return this.originalTaskFormatting.get(plainTextKey) ||
           this.originalTaskFormatting.get(this.normalizeTaskContent(plainTextKey));
  }

  /**
   * Get original section header for a given title
   */
  getOriginalSectionHeader(title) {
    return this.originalSectionHeaders.get(title.toLowerCase());
  }

  /**
   * Clear cached formatting (useful for refresh)
   */
  clearCache() {
    this.originalSectionHeaders.clear();
    this.originalTaskFormatting.clear();
    this.processedHeaders.clear();
    this.processedBoards.clear();
    console.log('Persistence cache cleared');
  }

  /**
   * Force refresh of cached formatting
   */
  async refreshCache() {
    this.clearCache();
    await this.preloadOriginalMarkdown();
  }

  /**
   * Check if a save operation is currently in progress
   */
  isSaveInProgress() {
    return this.isSaving;
  }

  /**
   * Get the current save queue length
   */
  getSaveQueueLength() {
    return this.saveQueue.length;
  }

  /**
   * Cleanup and destroy the persistence manager
   */
  destroy() {
    console.log('Destroying KanbanPersistenceManager');

    // Clear all cached data
    this.clearCache();

    // Clear save queue
    this.saveQueue = [];

    // Reset save state
    this.isSaving = false;

    // Clear core reference
    this.core = null;
  }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
  module.exports = KanbanPersistenceManager;
}