package i18n

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"sync"
	"wiki-go/internal/config"
)

// Regular expression to match placeholders like {{allowedTypes}}
var placeholderRegex = regexp.MustCompile(`\{\{(\w+)\}\}`)

// TranslationManager handles loading and retrieving translations
type TranslationManager struct {
	translations map[string]map[string]string
	mutex        sync.RWMutex
	defaultLang  string
	config       *config.Config
}

// NewTranslationManager creates a new translation manager
func NewTranslationManager(cfg *config.Config) *TranslationManager {
	return &TranslationManager{
		translations: make(map[string]map[string]string),
		defaultLang:  cfg.Wiki.Language,
		config:       cfg,
	}
}

// LoadTranslations loads translation files from the given directory
func (tm *TranslationManager) LoadTranslations(rootDir string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// Path to static langs directory
	staticLangsDir := filepath.Join(rootDir, "static", "langs")

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(staticLangsDir, 0755); err != nil {
		return err
	}

	// Read directory
	files, err := os.ReadDir(staticLangsDir)
	if err != nil {
		return err
	}

	// Load translation files
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		langCode := file.Name()[:len(file.Name())-5] // Remove .json extension
		filePath := filepath.Join(staticLangsDir, file.Name())

		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Warning: Failed to read translation file %s: %v", filePath, err)
			continue
		}

		var translations map[string]string
		if err := json.Unmarshal(data, &translations); err != nil {
			log.Printf("Warning: Failed to parse translation file %s: %v", filePath, err)
			continue
		}

		tm.translations[langCode] = translations
		log.Printf("Loaded %d translations for language %s", len(translations), langCode)
	}

	return nil
}

// Translate returns the translation for the given key
func (tm *TranslationManager) Translate(key string, langOverride ...string) string {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	// Get current language from config
	lang := tm.config.Wiki.Language

	// Override with specified language if provided
	if len(langOverride) > 0 && langOverride[0] != "" {
		lang = langOverride[0]
	}

	// Try to get translation in the requested language
	var translation string
	var found bool

	if translations, ok := tm.translations[lang]; ok {
		if translation, found = translations[key]; found {
			// Process placeholders
			translation = tm.processPlaceholders(translation)
			return translation
		}
	}

	// Fall back to default language
	if lang != tm.defaultLang {
		if translations, ok := tm.translations[tm.defaultLang]; ok {
			if translation, found = translations[key]; found {
				// Process placeholders
				translation = tm.processPlaceholders(translation)
				return translation
			}
		}
	}

	// Return the key as a last resort
	return key
}

// processPlaceholders replaces placeholders in translation strings
func (tm *TranslationManager) processPlaceholders(translation string) string {
	return placeholderRegex.ReplaceAllStringFunc(translation, func(placeholder string) string {
		// Extract the placeholder name without the {{ }}
		name := placeholder[2 : len(placeholder)-2]

		// Handle specific placeholders
		switch name {
		case "allowedTypes":
			return config.GetAllowedExtensionsDisplayText()
		case "maxFileSize":
			return config.GetMaxUploadSizeFormatted(tm.config)
		default:
			return placeholder // Keep the placeholder if not recognized
		}
	})
}

// GetAvailableLanguages returns a list of available language codes
func (tm *TranslationManager) GetAvailableLanguages() []string {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	languages := make([]string, 0, len(tm.translations))
	for lang := range tm.translations {
		languages = append(languages, lang)
	}

	// Sort languages alphabetically by language code
	sort.Strings(languages)

	return languages
}

// Global instance
var defaultManager *TranslationManager
var once sync.Once

// Initialize initializes the global translation manager
func Initialize(cfg *config.Config) error {
	var err error
	once.Do(func() {
		defaultManager = NewTranslationManager(cfg)

		// First, copy language files from internal to static directory
		if err = CopyLangsToStaticDir(cfg.Wiki.RootDir); err != nil {
			log.Printf("Warning: Failed to copy language files to static directory: %v", err)
		}

		// Then, load translations from static directory
		if err = defaultManager.LoadTranslations(cfg.Wiki.RootDir); err != nil {
			log.Printf("Warning: Failed to load translations: %v", err)
		}
	})
	return err
}

// Translate is a convenience function using the default manager
func Translate(key string, langOverride ...string) string {
	if defaultManager == nil {
		return key
	}
	return defaultManager.Translate(key, langOverride...)
}

// GetAvailableLanguages is a convenience function using the default manager
func GetAvailableLanguages() []string {
	if defaultManager == nil {
		return []string{}
	}
	return defaultManager.GetAvailableLanguages()
}
