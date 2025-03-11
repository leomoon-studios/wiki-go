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

        // Close search results on ESC key
        document.addEventListener('keydown', function(e) {
            if (e.key === 'Escape' && searchResults.classList.contains('active')) {
                searchResults.classList.remove('active');
                searchBox.value = '';
            }
        });
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
            displaySearchResults(results, query);
        } catch (error) {
            console.error('Search error:', error);
            searchResultsContent.innerHTML = '<div class="empty-message">An error occurred while searching. Please try again.</div>';
        }
    }

    /**
     * Display search results in the UI
     * @param {Array} results - Search results from the API
     * @param {string} query - The original search query
     */
    function displaySearchResults(results, query) {
        searchResults.classList.add('active');

        if (!results.length) {
            searchResultsContent.innerHTML = '<div class="empty-message">No results found.</div>';
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
        const pattern = new RegExp(`(${terms.join('|')})`, 'gi');

        const html = results.map(result => {
            // Highlight matches in title and excerpt
            const highlightedTitle = result.title.replace(pattern, '<span class="search-result-highlight">$1</span>');
            const highlightedExcerpt = result.excerpt.replace(pattern, '<span class="search-result-highlight">$1</span>');

            return `
                <div class="search-result-item">
                    <a href="${result.path}" class="search-result-title">${highlightedTitle}</a>
                    <div class="search-result-path">${result.path}</div>
                    <div class="search-result-excerpt">${highlightedExcerpt}</div>
                </div>
            `;
        }).join('');

        searchResultsContent.innerHTML = html;
    }

    // Public API
    return {
        init: init,
        // Export additional methods that might be needed by other modules
        performSearch: performSearch
    };
})();

// Initialize search when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    WikiSearch.init();
});