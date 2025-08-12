package prompt

import (
	"fmt"
	"text/template"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// loadTemplates loads and parses all prompt templates from embedded files
func (m *manager) loadTemplates() error {
	// Create template functions
	funcMap := template.FuncMap{
		"mul": func(a, b float64) float64 { return a * b },
		"printf": func(format string, args ...interface{}) string {
			return fmt.Sprintf(format, args...)
		},
	}

	// Map analysis types to template file names
	templateFiles := map[models.AnalysisType]string{
		models.AnalysisRisk:         "templates/portfolio/risk.tmpl",
		models.AnalysisAllocation:   "templates/portfolio/allocation.tmpl",
		models.AnalysisPerformance:  "templates/portfolio/performance.tmpl",
		models.AnalysisCompliance:   "templates/portfolio/compliance.tmpl",
		models.AnalysisReallocation: "templates/portfolio/reallocation.tmpl",
	}

	for analysisType, fileName := range templateFiles {
		// Read template content from embedded filesystem
		content, err := templateFS.ReadFile(fileName)
		if err != nil {
			return fmt.Errorf("failed to read template file %s: %w", fileName, err)
		}

		// Parse template with functions
		tmpl, err := template.New(string(analysisType)).Funcs(funcMap).Parse(string(content))
		if err != nil {
			return fmt.Errorf("failed to parse %s template: %w", analysisType, err)
		}

		m.templates[analysisType] = tmpl
	}

	return nil
}
