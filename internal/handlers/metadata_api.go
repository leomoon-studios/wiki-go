package handlers

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"wiki-go/internal/logger"
)

// MetadataRequest represents the request structure for metadata fetching
type MetadataRequest struct {
	URL string `json:"url"`
}

// MetadataResponse represents the response structure for metadata fetching
type MetadataResponse struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
}

// FetchMetadataHandler handles POST /api/links/fetch-metadata requests
func FetchMetadataHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests (like all other API endpoints)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON request body
	var req MetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}

	// Get URL from JSON payload
	targetURL := req.URL
	if targetURL == "" {
		respondWithError(w, "URL field is required", http.StatusBadRequest)
		return
	}

	// Validate the initial outbound destination before creating a request.
	if _, err := validateOutboundMetadataURL(r.Context(), targetURL, net.DefaultResolver); err != nil {
		respondWithError(w, "URL destination is not allowed", http.StatusBadRequest)
		return
	}

	// Fetch metadata
	metadata, err := fetchURLMetadata(targetURL)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Failed to fetch metadata: %v", err), http.StatusInternalServerError)
		return
	}

	// Return successful response
	response := MetadataResponse{
		Title:       metadata.Title,
		Description: metadata.Description,
		URL:         targetURL,
		Success:     true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// URLMetadata represents extracted metadata from a webpage
type URLMetadata struct {
	Title       string
	Description string
}

// fetchURLMetadata fetches and parses HTML metadata from a URL
func fetchURLMetadata(targetURL string) (*URLMetadata, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 2 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow up to 5 redirects
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	// Create request with proper headers
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set realistic browser headers (but don't request compression for easier parsing)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Charset", "utf-8,*;q=0.1")
	// Removed Accept-Encoding to avoid compression issues
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	// Perform request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	// Handle gzip decompression if needed
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		logger.Debug("Response is gzipped, decompressing...")
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %v", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	// Read response body (limit to 1MB to prevent abuse)
	body, err := io.ReadAll(io.LimitReader(reader, 1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Convert to UTF-8 if needed
	htmlContent := convertToUTF8(body, resp.Header.Get("Content-Type"))

	// Parse HTML metadata
	metadata := parseHTMLMetadata(htmlContent)

	// If no title found, generate fallback from URL
	if metadata.Title == "" {
		metadata.Title = generateFallbackTitle(targetURL)
	}

	return metadata, nil
}

// convertToUTF8 converts the HTML content to UTF-8 if it's in a different encoding
func convertToUTF8(body []byte, contentType string) string {
	// First, try to detect charset from Content-Type header
	charset := extractCharsetFromContentType(contentType)

	// If no charset in header, try to detect from HTML meta tags
	if charset == "" {
		charset = extractCharsetFromHTML(string(body))
	}

	// If still no charset, assume UTF-8
	if charset == "" {
		charset = "utf-8"
	}

	// Normalize charset name
	charset = strings.ToLower(strings.TrimSpace(charset))

	// If it's already UTF-8, just return as string
	if charset == "utf-8" || charset == "utf8" {
		return string(body)
	}

	// Handle common non-UTF-8 encodings manually
	switch charset {
	case "windows-1251", "cp1251":
		return convertWindows1251ToUTF8(body)
	case "iso-8859-1", "latin1":
		return convertISO88591ToUTF8(body)
	case "windows-1252", "cp1252":
		return convertWindows1252ToUTF8(body)
	default:
		// For unknown encodings, try as UTF-8 and hope for the best
		return string(body)
	}
}

// extractCharsetFromContentType extracts charset from Content-Type header
func extractCharsetFromContentType(contentType string) string {
	if contentType == "" {
		return ""
	}

	// Look for charset= in Content-Type header
	charsetRegex := regexp.MustCompile(`(?i)charset\s*=\s*([^;\s]+)`)
	matches := charsetRegex.FindStringSubmatch(contentType)
	if len(matches) > 1 {
		return strings.Trim(matches[1], `"'`)
	}

	return ""
}

// extractCharsetFromHTML extracts charset from HTML meta tags
func extractCharsetFromHTML(html string) string {
	// Try to find charset in meta tag - look at first 2KB only for performance
	searchArea := html
	if len(html) > 2048 {
		searchArea = html[:2048]
	}

	// Pattern 1: <meta charset="windows-1251">
	charsetRegex1 := regexp.MustCompile(`(?i)<meta[^>]*charset\s*=\s*["\']?([^"\'\s>]+)`)
	if matches := charsetRegex1.FindStringSubmatch(searchArea); len(matches) > 1 {
		return matches[1]
	}

	// Pattern 2: <meta http-equiv="Content-Type" content="text/html; charset=windows-1251">
	charsetRegex2 := regexp.MustCompile(`(?i)<meta[^>]*http-equiv\s*=\s*["\']?content-type["\']?[^>]*content\s*=\s*["\'][^"\']*charset\s*=\s*([^"\'\s;]+)`)
	if matches := charsetRegex2.FindStringSubmatch(searchArea); len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// convertWindows1251ToUTF8 converts Windows-1251 (Cyrillic) bytes to UTF-8 string
func convertWindows1251ToUTF8(data []byte) string {
	// Windows-1251 to Unicode mapping for characters 128-255
	win1251Table := []rune{
		// 128-159: Windows-1251 specific characters
		0x0402, 0x0403, 0x201A, 0x0453, 0x201E, 0x2026, 0x2020, 0x2021,
		0x20AC, 0x2030, 0x0409, 0x2039, 0x040A, 0x040C, 0x040B, 0x040F,
		0x0452, 0x2018, 0x2019, 0x201C, 0x201D, 0x2022, 0x2013, 0x2014,
		0x00A0, 0x2122, 0x0459, 0x203A, 0x045A, 0x045C, 0x045B, 0x045F,
		// 160-255: Standard Windows-1251 Cyrillic range
		0x00A0, 0x040E, 0x045E, 0x0408, 0x00A4, 0x0490, 0x00A6, 0x00A7,
		0x0401, 0x00A9, 0x0404, 0x00AB, 0x00AC, 0x00AD, 0x00AE, 0x0407,
		0x00B0, 0x00B1, 0x0406, 0x0456, 0x0491, 0x00B5, 0x00B6, 0x00B7,
		0x0451, 0x2116, 0x0454, 0x00BB, 0x0458, 0x0405, 0x0455, 0x0457,
		0x0410, 0x0411, 0x0412, 0x0413, 0x0414, 0x0415, 0x0416, 0x0417,
		0x0418, 0x0419, 0x041A, 0x041B, 0x041C, 0x041D, 0x041E, 0x041F,
		0x0420, 0x0421, 0x0422, 0x0423, 0x0424, 0x0425, 0x0426, 0x0427,
		0x0428, 0x0429, 0x042A, 0x042B, 0x042C, 0x042D, 0x042E, 0x042F,
		0x0430, 0x0431, 0x0432, 0x0433, 0x0434, 0x0435, 0x0436, 0x0437,
		0x0438, 0x0439, 0x043A, 0x043B, 0x043C, 0x043D, 0x043E, 0x043F,
		0x0440, 0x0441, 0x0442, 0x0443, 0x0444, 0x0445, 0x0446, 0x0447,
		0x0448, 0x0449, 0x044A, 0x044B, 0x044C, 0x044D, 0x044E, 0x044F,
	}

	var result strings.Builder
	result.Grow(len(data) * 2) // Pre-allocate for efficiency

	for _, b := range data {
		if b < 128 {
			// ASCII characters (0-127) are the same in Windows-1251 and UTF-8
			result.WriteByte(b)
		} else {
			// Characters 128-255 need conversion
			unicode := win1251Table[b-128]
			result.WriteRune(unicode)
		}
	}

	return result.String()
}

// convertISO88591ToUTF8 converts ISO-8859-1 (Latin-1) bytes to UTF-8 string
func convertISO88591ToUTF8(data []byte) string {
	// ISO-8859-1 is a subset of Unicode, so conversion is direct
	var result strings.Builder
	result.Grow(len(data))

	for _, b := range data {
		result.WriteRune(rune(b))
	}

	return result.String()
}

// convertWindows1252ToUTF8 converts Windows-1252 bytes to UTF-8 string
func convertWindows1252ToUTF8(data []byte) string {
	// Windows-1252 to Unicode mapping for characters 128-159
	win1252Table := []rune{
		0x20AC, 0x0081, 0x201A, 0x0192, 0x201E, 0x2026, 0x2020, 0x2021,
		0x02C6, 0x2030, 0x0160, 0x2039, 0x0152, 0x008D, 0x017D, 0x008F,
		0x0090, 0x2018, 0x2019, 0x201C, 0x201D, 0x2022, 0x2013, 0x2014,
		0x02DC, 0x2122, 0x0161, 0x203A, 0x0153, 0x009D, 0x017E, 0x0178,
	}

	var result strings.Builder
	result.Grow(len(data))

	for _, b := range data {
		if b < 128 {
			// ASCII characters
			result.WriteByte(b)
		} else if b < 160 {
			// Windows-1252 specific range (128-159)
			unicode := win1252Table[b-128]
			result.WriteRune(unicode)
		} else {
			// Characters 160-255 are the same as ISO-8859-1
			result.WriteRune(rune(b))
		}
	}

	return result.String()
}

// parseHTMLMetadata extracts title and description from HTML content
func parseHTMLMetadata(html string) *URLMetadata {
	metadata := &URLMetadata{}

	// Extract title (try multiple methods)
	metadata.Title = extractMetaTitle(html)

	// Extract description (try multiple methods)
	metadata.Description = extractMetaDescription(html)

	// Clean up extracted data
	metadata.Title = cleanMetaContent(metadata.Title)
	metadata.Description = cleanMetaContent(metadata.Description)

	// Limit description length
	if len(metadata.Description) > 200 {
		metadata.Description = metadata.Description[:197] + "..."
	}

	return metadata
}

// extractMetaTitle tries multiple methods to extract the page title
func extractMetaTitle(html string) string {
	// PRIORITY 1: Try Open Graph title first
	if title := extractMetaProperty(html, "og:title"); title != "" {
		return title
	}

	// PRIORITY 2: Try Twitter Card title
	if title := extractMetaName(html, "twitter:title"); title != "" {
		return title
	}

	// PRIORITY 3: FALLBACK - Try regular title tag
	titleRegex := regexp.MustCompile(`(?i)<title[^>]*>([^<]*)</title>`)
	if matches := titleRegex.FindStringSubmatch(html); len(matches) > 1 {
		title := strings.TrimSpace(matches[1])
		return title
	}

	return ""
}

// extractMetaDescription tries multiple methods to extract the page description
func extractMetaDescription(html string) string {
	// PRIORITY 1: Try Open Graph description first
	if desc := extractMetaProperty(html, "og:description"); desc != "" {
		return desc
	}

	// PRIORITY 2: Try Twitter Card description
	if desc := extractMetaName(html, "twitter:description"); desc != "" {
		return desc
	}

	// PRIORITY 3: FALLBACK - Try regular meta description
	if desc := extractMetaName(html, "description"); desc != "" {
		return desc
	}

	return ""
}

// extractMetaProperty extracts content from meta property tags (e.g., og:title)
func extractMetaProperty(html, property string) string {
	// Check if the property exists in HTML at all
	if !strings.Contains(strings.ToLower(html), strings.ToLower(property)) {
		return ""
	}

	// SIMPLE APPROACH: Just look for the exact pattern
	// <meta property="og:title" content="The Furniture Center" />
	simplePattern := fmt.Sprintf(`<meta property="%s" content="([^"]*)"`, regexp.QuoteMeta(property))
	simpleRegex := regexp.MustCompile(simplePattern)
	if matches := simpleRegex.FindStringSubmatch(html); len(matches) > 1 {
		return matches[1]
	}

	// Try case-insensitive version
	ciPattern := fmt.Sprintf(`(?i)<meta property="%s" content="([^"]*)"`, regexp.QuoteMeta(property))
	ciRegex := regexp.MustCompile(ciPattern)
	if matches := ciRegex.FindStringSubmatch(html); len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// extractMetaName extracts content from meta name tags (e.g., description)
func extractMetaName(html, name string) string {
	// Check if the name exists in HTML at all
	if !strings.Contains(strings.ToLower(html), strings.ToLower(name)) {
		return ""
	}

	// Try multiple case variations since some sites use "Description" instead of "description"
	nameVariations := []string{
		name,                  // exact case (e.g., "description")
		strings.Title(name),   // title case (e.g., "Description")
		strings.ToUpper(name), // upper case (e.g., "DESCRIPTION")
	}

	for _, nameVar := range nameVariations {
		// More flexible pattern that handles attributes in any order and additional attributes
		// Matches: <meta name="Description" content="..." /> or <meta id="..." name="Description" content="..." />
		flexiblePattern := fmt.Sprintf(`(?i)<meta[^>]*name\s*=\s*["\']%s["\'][^>]*content\s*=\s*["\']([^"\']*)["\']`, regexp.QuoteMeta(nameVar))
		flexibleRegex := regexp.MustCompile(flexiblePattern)
		if matches := flexibleRegex.FindStringSubmatch(html); len(matches) > 1 {
			return matches[1]
		}

		// Also try content before name (some sites have different attribute order)
		altOrderPattern := fmt.Sprintf(`(?i)<meta[^>]*content\s*=\s*["\']([^"\']*)["\'][^>]*name\s*=\s*["\']%s["\']`, regexp.QuoteMeta(nameVar))
		altOrderRegex := regexp.MustCompile(altOrderPattern)
		if matches := altOrderRegex.FindStringSubmatch(html); len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

// cleanMetaContent cleans up extracted meta content
func cleanMetaContent(content string) string {
	if content == "" {
		return ""
	}

	// Decode HTML entities
	content = strings.ReplaceAll(content, "&amp;", "&")
	content = strings.ReplaceAll(content, "&lt;", "<")
	content = strings.ReplaceAll(content, "&gt;", ">")
	content = strings.ReplaceAll(content, "&quot;", "\"")
	content = strings.ReplaceAll(content, "&#39;", "'")
	content = strings.ReplaceAll(content, "&nbsp;", " ")

	// Trim whitespace and normalize spaces
	content = strings.TrimSpace(content)
	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")

	return content
}

// generateFallbackTitle creates a fallback title from the URL
func generateFallbackTitle(targetURL string) string {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return "Unknown"
	}

	// Try to extract from path first
	if parsedURL.Path != "" && parsedURL.Path != "/" {
		pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
		if len(pathParts) > 0 && pathParts[len(pathParts)-1] != "" {
			lastPart := pathParts[len(pathParts)-1]
			// Clean up file extensions and URL patterns
			lastPart = regexp.MustCompile(`\.(html|htm|php|asp|aspx|jsp)$`).ReplaceAllString(lastPart, "")
			lastPart = strings.ReplaceAll(lastPart, "-", " ")
			lastPart = strings.ReplaceAll(lastPart, "_", " ")

			// Capitalize words
			words := strings.Fields(lastPart)
			for i, word := range words {
				if len(word) > 0 {
					words[i] = strings.ToUpper(word[:1]) + word[1:]
				}
			}

			if title := strings.Join(words, " "); len(title) > 2 {
				return title
			}
		}
	}

	// Fallback to domain name
	hostname := parsedURL.Hostname()
	hostname = strings.TrimPrefix(hostname, "www.")

	if hostname != "" {
		domainParts := strings.Split(hostname, ".")
		if len(domainParts) > 0 && domainParts[0] != "" {
			return strings.ToUpper(domainParts[0][:1]) + domainParts[0][1:]
		}
	}

	return "Unknown"
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, message string, statusCode int) {
	response := MetadataResponse{
		Success: false,
		Error:   message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
