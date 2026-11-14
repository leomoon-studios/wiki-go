// Search from the sidebar opens the bookmarkable results page after a short pause.
(function() {
    let timer;
    let request;
    let sequence = 0;

    function urlFor(path, query, key = 'q') {
        const url = new URL(path, window.location.origin);
        url.searchParams.set(key, query);
        return url.pathname + url.search;
    }

    function renderResults(container, results, query) {
        if (!Array.isArray(results) || results.length === 0) {
            WikiDOM.message(container, 'empty-message', window.i18n ? window.i18n.t('search.no_results') : 'No results found.');
            return;
        }
        const terms = query.match(/"[^"]+"|\S+/g) || [];
        const pattern = terms.length ? new RegExp('(' + terms.map(term => term.replace(/^"|"$/g, '').replace(/[.*+?^${}()|[\]\\]/g, '\\$&')).join('|') + ')', 'gi') : null;
        const fragment = document.createDocumentFragment();
        results.forEach(result => {
            const item = WikiDOM.element('article', 'search-result-item');
            const title = WikiDOM.element('a', 'search-result-title');
            const path = String(result?.path ?? '');
            title.href = WikiDOM.localURL(path);
            WikiDOM.appendHighlightedText(title, result?.title, pattern);
            const pathElement = WikiDOM.element('div', 'search-result-path', path);
            const excerpt = WikiDOM.element('div', 'search-result-excerpt');
            WikiDOM.appendHighlightedText(excerpt, result?.excerpt, pattern);
            item.append(title, pathElement, excerpt);
            fragment.appendChild(item);
        });
        container.replaceChildren(fragment);
    }

    document.addEventListener('DOMContentLoaded', function() {
        const box = document.querySelector('.search-box');
        const form = document.querySelector('.search-form');
        const page = document.querySelector('.search-page');
        const pageInput = document.querySelector('#search-page-query');
        const pageForm = document.querySelector('.search-page-form');
        const results = document.querySelector('.search-page-results');
        if (!box || !form) return;

        function cancelPending() {
            clearTimeout(timer);
            sequence++;
            if (request) request.abort();
        }

        async function updatePage(query, historyMode) {
            cancelPending();
            const current = sequence;
            const destination = urlFor('/', query, 'search');
            if (historyMode && destination !== window.location.pathname + window.location.search) {
                window.history[historyMode](null, '', destination);
            }
            box.value = query;
            pageInput.value = query;
            const error = document.querySelector('.search-page-error');
            if (error) error.remove();
            if (!query.trim()) {
                results.replaceChildren();
                return;
            }
            request = new AbortController();
            try {
                const response = await fetch(urlFor('/api/search', query), {signal: request.signal});
                if (!response.ok) throw new Error('Search failed');
                const matches = await response.json();
                if (current === sequence) renderResults(results, matches, query);
            } catch (error) {
                if (current === sequence && error.name !== 'AbortError') {
                    WikiDOM.message(results, 'empty-message', 'An error occurred while searching. Please try again.');
                }
            }
        }

        function schedule(input) {
            cancelPending();
            const query = input.value;
            if (page) {
                if (input === box) pageInput.value = query;
                else box.value = query;
            }
            timer = setTimeout(function() {
                if (page) updatePage(query, 'replaceState');
                else if (query.trim()) window.location.assign(urlFor('/', query, 'search'));
            }, 300);
        }

        box.addEventListener('input', () => schedule(box));
        form.addEventListener('submit', function(event) {
            cancelPending();
            if (page) {
                event.preventDefault();
                updatePage(box.value, 'pushState');
            }
        });
        if (page && pageForm && pageInput && results) {
            pageInput.addEventListener('input', () => schedule(pageInput));
            pageForm.addEventListener('submit', function(event) {
                event.preventDefault();
                updatePage(pageInput.value, 'pushState');
            });
            window.addEventListener('popstate', function() {
                updatePage(new URL(window.location.href).searchParams.get('search') || '', null);
            });
        }
    });
})();
