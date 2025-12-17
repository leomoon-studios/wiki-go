/**
 * Mermaid diagram rendering initialization
 * Handles both main document and version preview rendering
 */

// Create a global namespace for all mermaid functionality
window.MermaidHandler = {
  // Initialize mermaid for the main document
  initMain: function() {
    if (typeof mermaid === 'undefined') {
      console.error('Mermaid library not loaded');
      return;
    }

    try {
      // Get current theme
      const isDarkMode = document.documentElement.getAttribute('data-theme') === 'dark';

      // Store original sources before mermaid processes them
      this._storeOriginalSources();

      // Configure mermaid for main document
      mermaid.initialize({
        startOnLoad: true,
        theme: isDarkMode ? 'dark' : 'default',
        securityLevel: 'strict',
        maxTextSize: 15000,
        fontFamily: 'system-ui, -apple-system, sans-serif'
      });

      console.log('Mermaid initialized for main document');

      // Set up theme change listeners
      this._setupThemeChangeListeners();
    } catch (error) {
      console.error('Failed to initialize Mermaid:', error);
    }
  },

  // Special handling for version preview
  initVersionPreview: function(container) {
    if (!container || typeof mermaid === 'undefined') {
      console.error('Invalid container or Mermaid not loaded');
      return;
    }

    try {
      console.log('Initializing Mermaid for version preview');

      // Use the exact same initialization as the main document
      const isDarkMode = document.documentElement.getAttribute('data-theme') === 'dark';

      // Initialize with same settings as main document
      mermaid.initialize({
        startOnLoad: false, // Don't start on load - we'll process diagrams manually
        theme: isDarkMode ? 'dark' : 'default',
        securityLevel: 'strict',
        maxTextSize: 15000,
        fontFamily: 'system-ui, -apple-system, sans-serif'
      });

      // Find all mermaid diagrams in the container
      const diagrams = container.querySelectorAll('.mermaid');
      if (diagrams.length === 0) {
        console.log('No mermaid diagrams found in preview');
        return;
      }

      // Process each diagram
      diagrams.forEach((diagram, index) => {
        // Create a unique ID for this rendering
        const id = `mermaid-preview-${index}-${Date.now()}`;

        try {
          // Get content and run mermaid renderer
          const source = diagram.textContent || diagram.innerText;
          if (!source || source.trim() === '') {
            console.warn('Empty diagram source found');
            return;
          }

          // Render this specific diagram
          mermaid.render(id, source)
            .then(result => {
              // Insert the rendered SVG
              diagram.innerHTML = result.svg;
              console.log(`Diagram ${index} rendered successfully`);
            })
            .catch(err => {
              console.error(`Failed to render diagram ${index}:`, err);
              diagram.innerHTML = `<div style="color:red;border:1px solid red;padding:10px;">
                Diagram rendering error: ${err.message}
              </div>`;
            });
        } catch (err) {
          console.error(`Error processing diagram ${index}:`, err);
          diagram.innerHTML = `<div style="color:red;border:1px solid red;padding:10px;">
            Error: ${err.message}
          </div>`;
        }
      });
    } catch (error) {
      console.error('Error initializing for version preview:', error);
    }
  },

  // Set up theme change listeners
  _setupThemeChangeListeners: function() {
    // Listen for theme changes and re-render mermaid diagrams
    document.addEventListener('themeChanged', (event) => {
      if (!window.isPrintActive) {
        this._rerenderDiagrams(event.detail.theme);
      }
    });
  },

  // Re-render all mermaid diagrams with new theme
  _rerenderDiagrams: function(newTheme) {
    if (typeof mermaid === 'undefined') {
      console.error('Mermaid library not loaded');
      return;
    }

    try {
      // Update mermaid configuration with new theme
      mermaid.initialize({
        startOnLoad: false,
        theme: newTheme === 'dark' ? 'dark' : 'default',
        securityLevel: 'strict',
        maxTextSize: 15000,
        fontFamily: 'system-ui, -apple-system, sans-serif'
      });

      // Find all mermaid diagrams and re-render them
      const diagrams = document.querySelectorAll('.mermaid');
      
      // Batch process diagrams for better performance
      const renderPromises = [];
      
      diagrams.forEach((diagram, index) => {
        const source = diagram.dataset.mermaidSource;
        if (!source || source.trim() === '') {
          console.warn('No source found for diagram', index);
          return;
        }

        // Create unique ID for re-rendering
        const id = `mermaid-${index}-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;

        // Add loading class for CSS animations
        diagram.classList.add('mermaid-rerendering');

        // Render with new theme and collect promises
        const renderPromise = mermaid.render(id, source)
          .then(result => {
            diagram.innerHTML = result.svg;
            diagram.classList.remove('mermaid-rerendering');
            console.log(`Diagram ${index} re-rendered with ${newTheme} theme`);
          })
          .catch(err => {
            console.error(`Failed to re-render diagram ${index}:`, err);
            diagram.innerHTML = `<div style="color:red;border:1px solid red;padding:10px;">
              Re-rendering error: ${err.message}
            </div>`;
            diagram.classList.remove('mermaid-rerendering');
          });
          
        renderPromises.push(renderPromise);
      });

      // Wait for all diagrams to finish rendering
      Promise.allSettled(renderPromises).then(() => {
        console.log(`Completed re-rendering ${diagrams.length} mermaid diagrams with ${newTheme} theme`);
      });

    } catch (error) {
      console.error('Error re-rendering mermaid diagrams:', error);
    }
  },

  // Store original sources when diagrams are first rendered
  _storeOriginalSources: function() {
    const diagrams = document.querySelectorAll('.mermaid');
    diagrams.forEach((diagram) => {
      if (!diagram.dataset.mermaidSource) {
        // Clean and store the source
        const source = (diagram.textContent || diagram.innerText).trim();
        if (source) {
          diagram.dataset.mermaidSource = source;
        }
      }
    });
  }
};

// Initialize main document on page load
// Note: When lazy-loaded, this will be called manually by lazy-loader.js
// When loaded directly (edit mode, etc), this will auto-initialize
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', function() {
    // Only auto-initialize if we have mermaid diagrams
    if (document.querySelector('.mermaid')) {
      window.MermaidHandler.initMain();
    }
  });
} else {
  // DOM already loaded, check and initialize if needed
  if (document.querySelector('.mermaid')) {
    window.MermaidHandler.initMain();
  }
}
