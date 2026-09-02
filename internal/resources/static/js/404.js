// JavaScript dedicated to the custom 404 page.

(function () {
    // Ensure we are actually on a 404 page by checking for the create button
    document.addEventListener('DOMContentLoaded', () => {
        const btn = document.getElementById('create-missing-page');
        if (!btn) return;

        const fullPath = btn.dataset.currentPath || window.location.pathname;

        btn.addEventListener('click', () => {
            const cleaned = fullPath.replace(/^\/+|\/+$/g, '');
            const parts = cleaned.split('/');
            const docName = parts.pop() || 'page';
            const dirParts = parts; // may be empty array

            const slugify = async (text) => {
                const res = await fetch('/api/utils/slugify', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ text })
                });
                const json = await res.json();
                return json.slug || text;
            };

            // First slugify document name
            slugify(docName).then(async (docSlug) => {
                // Slugify each parent part in parallel
                const parentSlugs = await Promise.all(dirParts.map(p => slugify(p)));
                const dirSlugPath = parentSlugs.filter(Boolean).join('/');

                const docPath = dirSlugPath ? dirSlugPath + '/' + docSlug : docSlug;

                // Prettify title from original docName
                const prettyTitle = docName.replace(/[-_]+/g, ' ')
                    .split(' ').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ');

                return fetch('/api/document/create', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        path: docPath,
                        title: prettyTitle,
                        content: `# ${prettyTitle}\n\nContent goes here...`
                    })
                }).then(r => r.json())
                  .then(res => {
                      if (res.success) {
                          window.location.href = '/' + docPath;
                      } else {
                          alert('Error creating document: ' + res.message);
                      }
                  });
            }).catch(err => {
                console.error('Error slugifying:', err);
                alert('Error creating document. Please try again.');
            });
        });

        // Remove Edit button to prevent editing on 404 page
        const removeEdit = () => {
            const e = document.querySelector('.edit-page');
            if (e) e.remove();
        };
        removeEdit();

        // Ensure toolbar updates don't re-add the button
        if (window.Auth && typeof window.Auth.updateToolbarButtons === 'function') {
            const origUpdate = window.Auth.updateToolbarButtons;
            window.Auth.updateToolbarButtons = async function (...args) {
                await origUpdate.apply(this, args);
                removeEdit();
            };
        }
    });
})();
