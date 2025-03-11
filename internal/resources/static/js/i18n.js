/**
 * Client-side internationalization support
 */

// Global translations object
let translations = {};
let currentLanguage = document.documentElement.lang || 'en';

/**
 * Initialize translations
 */
async function initTranslations() {
    try {
        // Get current language from HTML lang attribute
        currentLanguage = document.documentElement.lang || 'en';

        // Load translations for current language
        await loadTranslations(currentLanguage);

        // Apply translations to all elements with data-i18n attribute
        applyTranslations();

        // Set up a MutationObserver to catch dynamically added elements
        setupMutationObserver();

        console.log(`Translations initialized for language: ${currentLanguage}`);
    } catch (error) {
        console.error('Failed to initialize translations:', error);
    }
}

/**
 * Set up a MutationObserver to catch dynamically added elements with data-i18n attributes
 */
function setupMutationObserver() {
    // Create a new observer
    const observer = new MutationObserver((mutations) => {
        let needsTranslation = false;

        // Check if any mutations added elements with data-i18n
        mutations.forEach(mutation => {
            if (mutation.type === 'childList') {
                mutation.addedNodes.forEach(node => {
                    if (node.nodeType === 1) { // Element node
                        // Check if the added node has data-i18n
                        if (node.hasAttribute && node.hasAttribute('data-i18n')) {
                            needsTranslation = true;
                        }

                        // Check children for data-i18n
                        const elementsWithI18n = node.querySelectorAll ?
                            node.querySelectorAll('[data-i18n]') : [];
                        if (elementsWithI18n.length > 0) {
                            needsTranslation = true;
                        }
                    }
                });
            }
        });

        // Apply translations if needed
        if (needsTranslation) {
            applyTranslations();
        }
    });

    // Start observing the document with the configured parameters
    observer.observe(document.body, {
        childList: true,
        subtree: true
    });
}

// Load translations for the specified language
async function loadLanguageData(lang) {
    try {
        const response = await fetch(`/static/langs/${lang}.json`);
        if (!response.ok) {
            throw new Error(`Failed to load translations for ${lang}: ${response.status}`);
        }
        return await response.json();
    } catch (error) {
        console.error(`Error loading translations for ${lang}:`, error);
        return {};
    }
}

// Global cache for loaded language data
const languageCache = {};

/**
 * Load translations for the specified language
 * @param {string} lang - Language code
 */
async function loadTranslations(lang) {
    try {
        if (!languageCache[lang]) {
            languageCache[lang] = await loadLanguageData(lang);
        }
        translations = languageCache[lang];
    } catch (error) {
        console.error(`Error loading translations for ${lang}:`, error);
        // Fallback to empty translations
        translations = {};
    }
}

/**
 * Get translation for a key
 * @param {string} key - Translation key
 * @param {string} [lang] - Optional language override
 * @returns {string} - Translated text or key if not found
 */
function t(key, lang) {
    // If a specific language is requested and it's not the current language
    if (lang && lang !== currentLanguage) {
        // If we have this language cached, use it
        if (languageCache[lang]) {
            return languageCache[lang][key] || key;
        }
        // Otherwise, we don't have this language loaded, just return the current translation or key
    }

    // Use current language translations
    return translations[key] || key;
}

/**
 * Apply translations to all elements with data-i18n attribute
 */
function applyTranslations() {
    document.querySelectorAll('[data-i18n]').forEach(element => {
        const key = element.getAttribute('data-i18n');
        if (key) {
            element.textContent = t(key);
        }
    });
}

/**
 * Translate a specific element and its children
 * @param {HTMLElement} element - The element to translate
 */
function translateElement(element) {
    if (element.hasAttribute && element.hasAttribute('data-i18n')) {
        const key = element.getAttribute('data-i18n');
        if (key) {
            element.textContent = t(key);
        }
    }

    // Translate children
    const children = element.querySelectorAll ? element.querySelectorAll('[data-i18n]') : [];
    children.forEach(child => {
        const key = child.getAttribute('data-i18n');
        if (key) {
            child.textContent = t(key);
        }
    });
}

// Initialize translations when DOM is loaded
document.addEventListener('DOMContentLoaded', initTranslations);

// Export functions for use in other scripts
window.i18n = {
    t,
    loadTranslations,
    applyTranslations,
    translateElement,
    getCurrentLanguage: () => currentLanguage,
    setCurrentLanguage: async (lang) => {
        currentLanguage = lang;
        await loadTranslations(lang);
        applyTranslations();
    }
};