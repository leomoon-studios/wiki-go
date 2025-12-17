/**
 * Lazy Loader Module
 * Conditionally loads heavy libraries only when needed
 * Saves 2-3MB bandwidth on pages without code/math/diagrams
 */

(function() {
    'use strict';

    // Track what's been loaded to avoid duplicates
    const loadedLibraries = {
        prism: false,
        mathjax: false,
        mermaid: false
    };

    // Get version from global context if available
    function getVersion() {
        // Extract version from any existing script tag with ?= parameter
        const scripts = document.querySelectorAll('script[src*="?="]');
        for (let script of scripts) {
            const match = script.src.match(/\?=(.+)$/);
            if (match) return match[1];
        }
        return 'dev-' + Date.now(); // Fallback
    }

    /**
     * Load a script dynamically
     */
    function loadScript(src, defer = true) {
        return new Promise((resolve, reject) => {
            // Check if script already exists
            if (document.querySelector(`script[src="${src}"]`)) {
                resolve();
                return;
            }

            const script = document.createElement('script');
            script.src = src;
            if (defer) script.defer = true;
            script.onload = resolve;
            script.onerror = reject;
            document.head.appendChild(script);
        });
    }

    /**
     * Check if page has code blocks
     */
    function hasCodeBlocks() {
        return document.querySelector('pre code') !== null;
    }

    /**
     * Check if page has math formulas
     */
    function hasMathFormulas() {
        // Check for inline math ($...$) or display math ($$...$$)
        // Also check for MathJax class markers
        const mathMarkers = document.querySelector('.math, .katex, [class*="math"], [class*="katex"]');
        if (mathMarkers) return true;

        // Check for $ or $$ in text content (basic check)
        const content = document.querySelector('.markdown-content');
        if (!content) return false;
        
        const text = content.textContent || '';
        return text.includes('$') || text.includes('\\(') || text.includes('\\[');
    }

    /**
     * Check if page has Mermaid diagrams
     */
    function hasMermaidDiagrams() {
        return document.querySelector('.mermaid') !== null;
    }

    /**
     * Load Prism (syntax highlighting)
     */
    async function loadPrism() {
        if (loadedLibraries.prism) return;
        
        console.log('Lazy loading Prism syntax highlighting...');
        const version = getVersion();
        
        try {
            await loadScript('/static/libs/prism-1.30.0/prism.min.js');
            await loadScript(`/static/js/prism-init.js?=${version}`);
            loadedLibraries.prism = true;
            console.log('Prism loaded successfully');
            
            // Trigger Prism highlighting after loading
            if (typeof Prism !== 'undefined') {
                Prism.highlightAll();
            }
        } catch (error) {
            console.error('Failed to load Prism:', error);
        }
    }

    /**
     * Load MathJax (math equations)
     */
    async function loadMathJax() {
        if (loadedLibraries.mathjax) return;
        
        console.log('Lazy loading MathJax...');
        const version = getVersion();
        
        try {
            await loadScript(`/static/js/mathjax-init.js?=${version}`);
            await loadScript('/static/libs/mathjax-3.2.2/tex-mml-chtml.js');
            loadedLibraries.mathjax = true;
            console.log('MathJax loaded successfully');
            
            // MathJax auto-initializes and processes the page when loaded
        } catch (error) {
            console.error('Failed to load MathJax:', error);
        }
    }

    /**
     * Load Mermaid (diagrams)
     */
    async function loadMermaid() {
        if (loadedLibraries.mermaid) return;
        
        console.log('Lazy loading Mermaid...');
        const version = getVersion();
        
        try {
            await loadScript('/static/libs/mermaid-11.12.1/mermaid.min.js');
            await loadScript(`/static/js/mermaid-init.js?=${version}`);
            loadedLibraries.mermaid = true;
            console.log('Mermaid loaded successfully');
            
            // Initialize Mermaid for main document after loading
            if (window.MermaidHandler && typeof window.MermaidHandler.initMain === 'function') {
                window.MermaidHandler.initMain();
            }
        } catch (error) {
            console.error('Failed to load Mermaid:', error);
        }
    }

    /**
     * Main initialization - check content and load necessary libraries
     */
    function init() {
        // Check what content exists and load accordingly
        const checks = [];

        if (hasCodeBlocks()) {
            console.log('Code blocks detected, loading Prism...');
            checks.push(loadPrism());
        }

        if (hasMathFormulas()) {
            console.log('Math formulas detected, loading MathJax...');
            checks.push(loadMathJax());
        }

        if (hasMermaidDiagrams()) {
            console.log('Mermaid diagrams detected, loading Mermaid...');
            checks.push(loadMermaid());
        }

        if (checks.length === 0) {
            console.log('No heavy libraries needed on this page');
        }

        return Promise.all(checks);
    }

    /**
     * Force load specific libraries (for dynamic content like version preview)
     */
    function forceLoad(libraryName) {
        switch(libraryName) {
            case 'prism':
                return loadPrism();
            case 'mathjax':
                return loadMathJax();
            case 'mermaid':
                return loadMermaid();
            default:
                console.warn('Unknown library:', libraryName);
                return Promise.resolve();
        }
    }

    // Export to global scope
    window.LazyLoader = {
        init,
        forceLoad,
        hasCodeBlocks,
        hasMathFormulas,
        hasMermaidDiagrams,
        loadPrism,
        loadMathJax,
        loadMermaid
    };

    // Auto-initialize when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
