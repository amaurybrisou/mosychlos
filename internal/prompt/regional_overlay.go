package prompt

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/models"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// loadRegionalOverlay loads regional template overlay with fallback strategy
func (rm *regionalManager) loadRegionalOverlay(
	analysisType models.AnalysisType,
	country, language string,
) (string, error) {
	// Template resolution paths with fallback strategy
	cacheKey := fmt.Sprintf("%s_%s_%s_overlay", analysisType, country, language)

	// Check cache first
	if cached, exists := rm.overlayCache[cacheKey]; exists {
		var buf bytes.Buffer
		if err := cached.Execute(&buf, nil); err == nil {
			return buf.String(), nil
		}
	}

	// Resolution hierarchy with fallback
	paths := []string{
		filepath.Join(rm.configDir, "templates", string(analysisType), "regional", "overlays",
			fmt.Sprintf("%s_%s_overlay.tmpl", country, language)),
		filepath.Join(rm.configDir, "templates", string(analysisType), "regional", "overlays",
			fmt.Sprintf("%s_overlay.tmpl", country)),
		"", // No overlay - graceful fallback to base template only
	}

	for _, path := range paths {
		if path == "" {
			return "", nil // No regional overlay - use base template
		}

		if content, err := rm.fs.ReadFile(path); err == nil {
			// Cache the template for future use
			// tmpl, parseErr := template.New("overlay").Parse(string(content))
			tmpl, parseErr := template.
				New("overlay").
				Funcs(rm.tmplFuncMap()).
				Option("missingkey=zero").
				Parse(string(content))
			if parseErr == nil {
				rm.overlayCache[cacheKey] = tmpl
			}
			return string(content), nil
		}
	}

	return "", nil // Graceful fallback - no regional overlay needed
}

// tmplFuncMap provides helpers usable inside overlays (join, default, etc.).
func (rm *regionalManager) tmplFuncMap() template.FuncMap {
	return template.FuncMap{
		// join slice of strings with a separator
		"join": func(ss []string, sep string) string {
			return strings.Join(ss, sep)
		},
		// default: if v is "empty", return dflt, else v
		"default": func(v any, dflt any) any {
			switch x := v.(type) {
			case nil:
				return dflt
			case string:
				if strings.TrimSpace(x) == "" {
					return dflt
				}
				return x
			case []string:
				if len(x) == 0 {
					return dflt
				}
				return x
			default:
				// zero-ish detection for common numeric types could be added if needed
				return v
			}
		},
		// coalesce: first non-empty (string) value
		"coalesce": func(vals ...string) string {
			for _, v := range vals {
				if strings.TrimSpace(v) != "" {
					return v
				}
			}
			return ""
		},
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": cases.Title(language.Und).String, // note: deprecated behavior; use cases are simple here
		"trim":  strings.TrimSpace,
		"now":   time.Now,
		"iso8601": func(t time.Time) string {
			// keep local offset (e.g., Europe/Paris) if the caller provided a localized time
			// _, offset := t.Zone()
			// format with numeric offset like +02:00
			return t.Format("2006-01-02T15:04:05-07:00")
		},
		// dict: handy to build small maps inside templates
		"dict": func(kv ...any) (map[string]any, error) {
			if len(kv)%2 != 0 {
				return nil, fmt.Errorf("dict expects even number of args")
			}
			m := make(map[string]any, len(kv)/2)
			for i := 0; i < len(kv); i += 2 {
				k, ok := kv[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				m[k] = kv[i+1]
			}
			return m, nil
		},
	}
}
