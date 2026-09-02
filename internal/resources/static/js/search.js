/**
 * search.js - Handles search functionality for Wiki-Go
 */

// Search module using IIFE to avoid polluting global namespace
const WikiSearch = (function() {
    // DOM elements
    let searchBox;
    let searchResults;
    let searchResultsContent;
    let searchClose;

    // Variables
    let searchTimeout;

    /**
     * Initialize search functionality
     */
    function init() {
        // Get DOM elements
        searchBox = document.querySelector('.search-box');
        searchResults = document.querySelector('.search-results');
        searchResultsContent = document.querySelector('.search-results-content');
        searchClose = document.querySelector('.search-close');

        // Add event listeners
        bindEvents();
    }

    /**
     * Bind all event listeners
     */
    function bindEvents() {
        // Search input handler
        searchBox.addEventListener('input', function(e) {
            clearTimeout(searchTimeout);

            const query = e.target.value.trim();
            if (!query) {
                searchResults.classList.remove('active');
                return;
            }

            searchTimeout = setTimeout(() => {
                performSearch(query);
            }, 300);
        });

        // Close search results
        searchClose.addEventListener('click', function() {
            searchResults.classList.remove('active');
            searchBox.value = '';
        });

        // Escape key is now handled by keyboard-shortcuts.js
    }

    /**
     * Perform search query against the API
     * @param {string} query - The search query
     */
    async function performSearch(query) {
        try {
            const response = await fetch('/api/search', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ query })
            });

            if (!response.ok) {
                throw new Error('Search failed');
            }

            const results = await response.json();

            // Add console log to debug the response
            console.log('Search results:', results);

            // Make sure results is always an array
            const resultsArray = Array.isArray(results) ? results : [];

            // Display results (or no results message if empty)
            displaySearchResults(resultsArray, query);
        } catch (error) {
            console.error('Search error:', error);
            WikiDOM.message(searchResultsContent, 'empty-message', 'An error occurred while searching. Please try again.');
        }
    }

    /**
     * Display search results in the UI
     * @param {Array} results - Search results from the API
     * @param {string} query - The original search query
     */
    function displaySearchResults(results, query) {
        searchResults.classList.add('active');

        // Make sure results is an array and check if it's empty
        if (!Array.isArray(results) || results.length === 0) {
            WikiDOM.message(searchResultsContent, 'empty-message', window.i18n ? window.i18n.t('search.no_results') : 'No results found.');
            return;
        }

        // Create regex pattern for highlighting
        const terms = [];
        let match;
        const quotedRegex = /"([^"]+)"/g;
        const remainingTerms = query.split(/\s+/)
            .filter(term => term && !term.startsWith('NOT') && !term.includes('"'));

        // Extract quoted phrases first
        while ((match = quotedRegex.exec(query)) !== null) {
            terms.push(match[1].replace(/[.*+?^${}()|[\]\\]/g, '\\$&'));
        }

        // Add remaining individual terms
        terms.push(...remainingTerms.map(term => term.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')));

        // Create pattern that matches whole phrases and individual terms
        const pattern = terms.length > 0 ? new RegExp(`(${terms.join('|')})`, 'gi') : null;
        const fragment = document.createDocumentFragment();

        results.forEach(result => {
            const item = WikiDOM.element('div', 'search-result-item');
            const title = WikiDOM.element('a', 'search-result-title');
            const resultPath = String(result?.path ?? '');
            title.href = WikiDOM.localURL(resultPath);
            WikiDOM.appendHighlightedText(title, result?.title, pattern);

            const pathElement = WikiDOM.element('div', 'search-result-path', resultPath);
            const excerpt = WikiDOM.element('div', 'search-result-excerpt');
            WikiDOM.appendHighlightedText(excerpt, result?.excerpt, pattern);

            item.append(title, pathElement, excerpt);
            fragment.appendChild(item);
        });

        searchResultsContent.replaceChildren(fragment);
    }

    // Hide search results function for keyboard shortcuts
    function hideSearchResults() {
        if (searchResults) {
            searchResults.classList.remove('active');
            if (searchBox) {
                searchBox.value = '';
            }
        }
    }

    // Public API
    return {
        init: init,
        // Export additional methods that might be needed by other modules
        performSearch: performSearch,
        hideSearchResults: hideSearchResults
    };
})();

// Create window.Search for compatibility with keyboard-shortcuts.js
window.Search = {
    hideSearchResults: function() {
        const searchResults = document.querySelector('.search-results');
        const searchBox = document.querySelector('.search-box');
        if (searchResults) {
            searchResults.classList.remove('active');
            if (searchBox) {
                searchBox.value = '';
            }
        }
    }
};

// Initialize search when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    WikiSearch.init();
});
