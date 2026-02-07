package hosting

import (
	"bytes"
	"encoding/json"
	"strings"
)

// analyticsScript is the minimal tracking snippet injected into HTML pages
// It sends a pageview beacon to the admin subdomain's /track endpoint
// The script extracts the base domain and constructs the admin URL dynamically
// For root domains (zyt.app), uses full hostname. For subdomains (tetris.zyt.app), strips subdomain.
const analyticsScript = `<script>(function(){
var h=location.hostname;
var s=h.split('.').slice(1).join('.');
var d=(s&&s.includes('.'))?s:h;
var p=location.port&&location.port!=='80'&&location.port!=='443'?':'+location.port:'';
var u=location.protocol+'//admin.'+d+p+'/track';
navigator.sendBeacon(u,JSON.stringify({h:h,p:location.pathname,e:'pageview'}))
})();</script>`

// InjectAnalytics injects the analytics tracking script into HTML content
// It inserts the script right before the closing </body> tag
// Returns the original content unchanged if injection is disabled or body tag not found
func InjectAnalytics(content []byte, siteID string) []byte {
	if isAnalyticsDisabled(siteID) {
		return content
	}

	// Find </body> tag (case-insensitive)
	lower := bytes.ToLower(content)
	idx := bytes.LastIndex(lower, []byte("</body>"))
	if idx == -1 {
		// No body tag found, return original
		return content
	}

	// Insert script before </body>
	result := make([]byte, 0, len(content)+len(analyticsScript))
	result = append(result, content[:idx]...)
	result = append(result, analyticsScript...)
	result = append(result, content[idx:]...)

	return result
}

// isAnalyticsDisabled checks if analytics is disabled for the site via manifest.json
// Looks for: { "analytics": { "enabled": false } }
func isAnalyticsDisabled(siteID string) bool {
	if fs == nil {
		return false
	}

	// Read manifest.json from VFS
	file, err := fs.ReadFile(siteID, "manifest.json")
	if err != nil {
		return false // No manifest = analytics enabled by default
	}
	defer file.Content.Close()

	// Parse manifest
	var manifest struct {
		Analytics struct {
			Enabled *bool `json:"enabled"`
		} `json:"analytics"`
	}

	decoder := json.NewDecoder(file.Content)
	if err := decoder.Decode(&manifest); err != nil {
		return false // Parse error = analytics enabled by default
	}

	// If analytics.enabled is explicitly set to false, disable analytics
	if manifest.Analytics.Enabled != nil && !*manifest.Analytics.Enabled {
		return true
	}

	return false
}

// ShouldInjectAnalytics returns true if the path is an HTML file that should have analytics injected
func ShouldInjectAnalytics(path string) bool {
	// Only inject into HTML files
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, ".html") || strings.HasSuffix(lower, ".htm")
}
