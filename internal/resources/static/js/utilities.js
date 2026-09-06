/**
 * Utility functions for the Wiki-Go application
 */

/**
 * Get the current document path from the URL
 * @returns {string} The standardized document path
 */
function getCurrentDocPath() {
    // Log the raw path for debugging
    console.log("Raw pathname for path processing:", window.location.pathname);

    const isHomepage = window.location.pathname === '/';
    if (isHomepage) {
        console.log("Using homepage path");
        return 'pages/home';
    }

    // For versions, we need to keep the full path structure
    let path = window.location.pathname;

    // Remove leading slash
    if (path.startsWith('/')) {
        path = path.substring(1);
    }

    // Remove trailing slash if it exists
    if (path.endsWith('/')) {
        path = path.substring(0, path.length - 1);
    }

    // If .md exists in the path, remove it (some implementations add .md to URLs)
    if (path.endsWith('.md')) {
        path = path.substring(0, path.length - 3);
    }

    console.log("Processed document path:", path);
    return path || 'pages/home';
}

// Make function available globally
window.getCurrentDocPath = getCurrentDocPath;

/**
 * DOM helpers for rendering untrusted text without parsing it as markup.
 * Server-rendered Markdown is deliberately not accepted by these helpers.
 */
window.WikiDOM = {
    clear(element) {
        element.replaceChildren();
    },

    element(tagName, className, text) {
        const element = document.createElement(tagName);
        if (className) element.className = className;
        if (text !== undefined && text !== null) element.textContent = String(text);
        return element;
    },

    message(container, className, text) {
        const message = this.element('div', className, text);
        container.replaceChildren(message);
        return message;
    },

    icon(classNames) {
        const icon = document.createElement('i');
        icon.className = classNames;
        icon.setAttribute('aria-hidden', 'true');
        return icon;
    },

    localURL(value, fallback = '#') {
        if (typeof value !== 'string') return fallback;
        try {
            const parsed = new URL(value, window.location.origin);
            if (parsed.origin !== window.location.origin || !parsed.pathname.startsWith('/')) return fallback;
            return parsed.pathname + parsed.search + parsed.hash;
        } catch (_) {
            return fallback;
        }
    },

    markdownURL(value, image = false) {
        const url = String(value ?? '').trim();
        if (!url || /[\u0000-\u001f\u007f]/.test(url) || url.startsWith('//')) return null;
        const scheme = url.match(/^([a-z][a-z0-9+.-]*):/i);
        if (!scheme) return url;
        const allowed = image ? ['http', 'https'] : ['http', 'https', 'mailto'];
        return allowed.includes(scheme[1].toLowerCase()) ? url : null;
    },

    appendHighlightedText(container, value, pattern) {
        const text = String(value ?? '');
        if (!pattern || pattern.source === '()') {
            container.textContent = text;
            return;
        }

        pattern.lastIndex = 0;
        let lastIndex = 0;
        let match;
        while ((match = pattern.exec(text)) !== null) {
            container.appendChild(document.createTextNode(text.slice(lastIndex, match.index)));
            container.appendChild(this.element('span', 'search-result-highlight', match[0]));
            lastIndex = match.index + match[0].length;
            if (match[0].length === 0) pattern.lastIndex++;
        }
        container.appendChild(document.createTextNode(text.slice(lastIndex)));
    }
};
