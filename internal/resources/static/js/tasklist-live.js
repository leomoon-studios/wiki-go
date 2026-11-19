// Task List Live Editing
// This script enables live checkbox toggling for admins & editors.
// It relies on the existing /api/source and /api/save endpoints.
// Note: Kanban tasks are handled by kanban-tasks.js instead.
// Permissions and checkbox enabling are handled by tasklist-permissions.js
(function () {
  document.addEventListener('DOMContentLoaded', () => {
    // Check permissions using shared module
    if (!window.TaskListPermissions || !window.TaskListPermissions.isTaskEditingAllowed()) {
      console.log('tasklist-live.js: Task editing not allowed, aborting');
      return;
    }

    const docPath = window.TaskListPermissions.getDocPath();
    const container = document.querySelector('.markdown-content');
    if (!container) return;

    // Helper function to check if a checkbox is in a kanban context
    const isKanbanCheckbox = (checkbox) => {
      return checkbox.closest('.kanban-column-content') !== null;
    };

    // Index only non-kanban checkboxes for sequential task matching
    // Note: Checkbox enabling is handled by tasklist-permissions.js
    const allCheckboxes = [...container.querySelectorAll('input[type="checkbox"]')]
      .filter(cb => !isKanbanCheckbox(cb));

    allCheckboxes.forEach((cb, idx) => {
      cb.dataset.cbIndex = idx;
    });

    console.log(`tasklist-live.js: Managing ${allCheckboxes.length} regular task checkboxes (excluding kanban)`);

    // Helper to enable / disable all list checkboxes (only regular tasks)
    const toggleAll = (disabled) => { allCheckboxes.forEach(cb => cb.disabled = disabled); };

    // Ensure every task-list <li> has a persistent .save-state span
    // Skip kanban tasks as they have their own save-state management
    container.querySelectorAll('li').forEach(li => {
      if (!li.querySelector('input[type="checkbox"]')) return; // not a task item
      if (isKanbanCheckbox(li.querySelector('input[type="checkbox"]'))) return; // skip kanban tasks
      if (!li.querySelector('.save-state')) {
        const span = document.createElement('span');
        span.className = 'save-state';
        const firstUL = li.querySelector('ul');
        if (firstUL) {
          li.insertBefore(span, firstUL);
        } else {
          li.appendChild(span);
        }
      }
    });

    const showState = (li, text, cssClass) => {
      const s = li.querySelector('.save-state');
      if (!s) return;
      s.textContent = text;
      s.classList.remove('saved','error');
      if(cssClass) s.classList.add(cssClass);
      
      // Auto-hide after 3 seconds for errors
      setTimeout(() => {
        s.textContent = '';
        s.classList.remove('saved','error');
      }, 3000);
    };

    container.addEventListener('click', async (e) => {
      const target = e.target;
      if (!(target instanceof HTMLInputElement) || target.type !== 'checkbox') return;

      // Skip kanban checkboxes - they are handled by kanban-tasks.js
      if (isKanbanCheckbox(target)) {
        console.log('tasklist-live.js: Skipping kanban checkbox, handled by kanban-tasks.js');
        return;
      }

      e.preventDefault();
      const li = target.closest('li');
      if (!li) return;

      // Disable all checkboxes during save
      toggleAll(true);

      try {
        // 1. Fetch current markdown
        const srcResp = await fetch(wikiURL(`/api/source/${docPath}`));
        if (!srcResp.ok) throw new Error('source fetch failed');
        const markdown = await srcResp.text();

        // 2. Build updated markdown by toggling the first matching task line.
        const lines = markdown.split(/\r?\n/);
        const desiredChecked = !target.checked;
        // Determine checkbox index among all rendered checkboxes
        const cbIndex = Number(target.dataset.cbIndex);

        let currentIdx = -1;
        const updatedLines = lines.map((line) => {
          const match = line.match(/^(\s*[-*+]\s+)\[(.| )\]\s+(.*)$/);
          if (!match) return line; // not a task list line
          currentIdx += 1;
          if (currentIdx === cbIndex) {
            const [, prefix, , restText] = match;
            const newMark = desiredChecked ? 'x' : ' ';
            return `${prefix}[${newMark}] ${restText}`;
          }
          return line;
        });
        if (cbIndex > currentIdx) throw new Error('checkbox index exceeds markdown task list count');

        // 3. Save updated markdown
        const saveResp = await fetch(wikiURL(`/api/save/${docPath}`), {
          method: 'POST',
          headers: { 'Content-Type': 'text/markdown' },
          body: updatedLines.join('\n'),
        });
        if (!saveResp.ok) throw new Error('save failed');

        // 4. Update UI checkbox state
        target.checked = desiredChecked;
        // Success - no visual feedback needed
      } catch (err) {
        console.error(err);
        showState(li,'error','error');
      } finally {
        toggleAll(false);
      }
    });
  });
})();