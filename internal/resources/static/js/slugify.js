// Slug generator for Wiki-Go
// Adds automatic transliteration for document titles in any language
(function() {
    'use strict';

    // Initialize when DOM is loaded
    document.addEventListener('DOMContentLoaded', function() {
        initSlugGenerator();
    });

    // Initialize slug generator for new document form
    function initSlugGenerator() {
        const titleInput = document.getElementById('docTitle');
        const slugInput = document.getElementById('docSlug');

        if (titleInput && slugInput) {
            console.log('Slugify: Title/slug inputs found - attaching listener');

            titleInput.addEventListener('input', function() {
                const title = this.value.trim();

                if (!title) {
                    slugInput.value = '';
                    return;
                }

                // Show "generating..." text
                slugInput.value = 'Generating...';

                // Call server API to generate slug with proper transliteration
                fetch(wikiURL('/api/utils/slugify'), {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        text: title,
                        lang: document.documentElement.lang || 'en'
                    })
                })
                .then(response => response.json())
                .then(data => {
                    if (data.slug) {
                        slugInput.value = data.slug;
                    } else {
                        slugInput.value = '';
                    }
                })
                .catch(error => {
                    console.error('Slug API error:', error);
                    slugInput.value = '';
                });
            });
        }
    }

    // Expose public API (if needed in other modules)
    window.Slugify = {
        init: initSlugGenerator
    };
})();