// Base path support for wikis served from a URL subdirectory (server.base_path).
// Wiki-relative paths such as "/api/search" or "/docs/page" become
// "/wiki/api/search" when the wiki is served from https://example.com/wiki/.
(function() {
    'use strict';

    const meta = document.querySelector('meta[name="wiki-base-path"]');
    const base = ((meta && meta.content) || '').replace(/\/+$/, '');

    // Add the base path to a wiki-relative URL. Other values are returned unchanged.
    function wikiURL(path) {
        if (typeof path !== 'string' || !path.startsWith('/') || path.startsWith('//')) return path;
        return base + path;
    }

    // Return a browser pathname relative to the wiki root, e.g. "/docs/page".
    function wikiPathname(pathname = window.location.pathname) {
        if (!base) return pathname;
        if (pathname === base) return '/';
        return pathname.startsWith(base + '/') ? pathname.slice(base.length) : pathname;
    }

    window.WIKI_BASE_PATH = base;
    window.wikiURL = wikiURL;
    window.wikiPathname = wikiPathname;
})();
