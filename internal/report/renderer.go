package report

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

//go:embed templates/*
var templateFiles embed.FS

// renderCustomerReport renders the customer report in markdown format
func (g *Generator) renderCustomerReport(data *models.CustomerReportData) (string, []string, error) {
	tmpl, err := g.getCustomerTemplate()
	if err != nil {
		return "", nil, fmt.Errorf("failed to get customer template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", nil, fmt.Errorf("failed to execute customer template: %w", err)
	}

	dataSources := g.getDataSourcesUsed(data, nil)
	return buf.String(), dataSources, nil
}

// renderSystemReport renders the system report in markdown format
func (g *Generator) renderSystemReport(data *models.SystemReportData) (string, []string, error) {
	tmpl, err := g.getSystemTemplate()
	if err != nil {
		return "", nil, fmt.Errorf("failed to get system template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", nil, fmt.Errorf("failed to execute system template: %w", err)
	}

	dataSources := g.getDataSourcesUsed(nil, data)
	return buf.String(), dataSources, nil
}

// renderFullReport renders the full report combining customer and system data
func (g *Generator) renderFullReport(customerData *models.CustomerReportData, systemData *models.SystemReportData) (string, []string, error) {
	// Generate customer report
	customerReport, customerDataSources, err := g.renderCustomerReport(customerData)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate customer report: %w", err)
	}

	// Generate system report
	systemReport, systemDataSources, err := g.renderSystemReport(systemData)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate system report: %w", err)
	}

	// Concatenate the reports with a separator
	fullReport := customerReport + "\n\n---\n\n" + systemReport

	// Combine data sources from both reports
	allDataSources := make([]string, 0, len(customerDataSources)+len(systemDataSources))
	sourceSet := make(map[string]bool)

	for _, source := range customerDataSources {
		if !sourceSet[source] {
			allDataSources = append(allDataSources, source)
			sourceSet[source] = true
		}
	}
	for _, source := range systemDataSources {
		if !sourceSet[source] {
			allDataSources = append(allDataSources, source)
			sourceSet[source] = true
		}
	}

	return fullReport, allDataSources, nil
}

// getCustomerTemplate returns the customer report template
func (g *Generator) getCustomerTemplate() (*template.Template, error) {
	funcMap := template.FuncMap{
		"toUpper": strings.ToUpper,
		"multiply": func(a, b float64) float64 {
			return a * b
		},
		"formatDuration": func(d time.Duration) string {
			if d < time.Second {
				return fmt.Sprintf("%dms", d.Milliseconds())
			} else if d < time.Minute {
				return fmt.Sprintf("%.1fs", d.Seconds())
			} else if d < time.Hour {
				return fmt.Sprintf("%.1fm", d.Minutes())
			}
			return fmt.Sprintf("%.1fh", d.Hours())
		},
		"formatNumber": func(n interface{}) string {
			switch v := n.(type) {
			case int:
				return addCommas(fmt.Sprintf("%d", v))
			case int64:
				return addCommas(fmt.Sprintf("%d", v))
			case float64:
				if v == float64(int64(v)) {
					return addCommas(fmt.Sprintf("%.0f", v))
				}
				return addCommas(fmt.Sprintf("%.2f", v))
			default:
				return fmt.Sprintf("%v", v)
			}
		},
		"prettyJSON": func(v interface{}) string {
			jsonBytes, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				return fmt.Sprintf("Error formatting JSON: %v", err)
			}
			return string(jsonBytes)
		},
		"prettyJSONWrapped": func(v interface{}, lineWidth int) string {
			jsonBytes, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				return fmt.Sprintf("Error formatting JSON: %v", err)
			}

			// Split long lines to prevent overflow in PDF
			lines := strings.Split(string(jsonBytes), "\n")
			var wrappedLines []string

			for _, line := range lines {
				if len(line) <= lineWidth {
					wrappedLines = append(wrappedLines, line)
				} else {
					// For long lines, try to break at commas or after colons
					wrapped := wrapJSONLine(line, lineWidth)
					wrappedLines = append(wrappedLines, wrapped...)
				}
			}

			return strings.Join(wrappedLines, "\n")
		},
		"formatInvestmentResearch": func(v interface{}) string {
			return g.formatInvestmentResearchData(v)
		},
		"formatWebSearchData": func(toolName string, arguments string, result string) string {
			return g.formatWebSearchData(toolName, arguments, result)
		},
		"isWebSearchTool": func(toolName string) bool {
			return toolName == bag.WebSearch.String()
		},
	}

	tmplContent, err := templateFiles.ReadFile("templates/customer.md")
	if err != nil {
		return nil, fmt.Errorf("failed to read customer template: %w", err)
	}

	return template.Must(template.New("customer").Funcs(funcMap).Parse(string(tmplContent))), nil
}

// getSystemTemplate returns the system diagnostics template
func (g *Generator) getSystemTemplate() (*template.Template, error) {
	funcMap := template.FuncMap{
		"toUpper": strings.ToUpper,
		"multiply": func(a, b float64) float64 {
			return a * b
		},
		"formatDuration": func(d time.Duration) string {
			if d < time.Second {
				return fmt.Sprintf("%dms", d.Milliseconds())
			} else if d < time.Minute {
				return fmt.Sprintf("%.1fs", d.Seconds())
			} else if d < time.Hour {
				return fmt.Sprintf("%.1fm", d.Minutes())
			}
			return fmt.Sprintf("%.1fh", d.Hours())
		},
		"formatBytes": func(b int64) string {
			if b < 1024 {
				return fmt.Sprintf("%d B", b)
			} else if b < 1024*1024 {
				return fmt.Sprintf("%.1f KB", float64(b)/1024)
			} else if b < 1024*1024*1024 {
				return fmt.Sprintf("%.1f MB", float64(b)/(1024*1024))
			}
			return fmt.Sprintf("%.1f GB", float64(b)/(1024*1024*1024))
		},
		"slice": func(slice any, start, end int) any {
			switch s := slice.(type) {
			case []any:
				if start < 0 || start >= len(s) {
					return []any{}
				}
				if end > len(s) {
					end = len(s)
				}
				return s[start:end]
			case []models.ToolComputation:
				if start < 0 || start >= len(s) {
					return []models.ToolComputation{}
				}
				if end > len(s) {
					end = len(s)
				}
				return s[start:end]
			default:
				return slice
			}
		},
		"min": func(a, b int) int {
			if a < b {
				return a
			}
			return b
		},
		"join": func(slice []string, sep string) string {
			return strings.Join(slice, sep)
		},
		"truncate": func(s string, length int) string {
			if len(s) <= length {
				return s
			}
			if length <= 3 {
				return s[:length]
			}
			return s[:length-3] + "..."
		},
		"append": func(slice any, items ...any) any {
			switch s := slice.(type) {
			case []string:
				for _, item := range items {
					if str, ok := item.(string); ok {
						s = append(s, str)
					}
				}
				return s
			default:
				return slice
			}
		},
		"add": func(a, b int) int {
			return a + b
		},
		"printf": func(format string, args ...interface{}) string {
			return fmt.Sprintf(format, args...)
		},
		"formatNumber": func(n interface{}) string {
			switch v := n.(type) {
			case int:
				return addCommas(fmt.Sprintf("%d", v))
			case int64:
				return addCommas(fmt.Sprintf("%d", v))
			case float64:
				if v == float64(int64(v)) {
					return addCommas(fmt.Sprintf("%.0f", v))
				}
				return addCommas(fmt.Sprintf("%.2f", v))
			default:
				return fmt.Sprintf("%v", v)
			}
		},
		"prettyJSON": func(v any) string {
			jsonBytes, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				return fmt.Sprintf("Error formatting JSON: %v", err)
			}
			return string(jsonBytes)
		},
		"prettyJSONWrapped": func(v any, lineWidth int) string {
			jsonBytes, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				return fmt.Sprintf("Error formatting JSON: %v", err)
			}

			// Split long lines to prevent overflow in PDF
			lines := strings.Split(string(jsonBytes), "\n")
			var wrappedLines []string

			for _, line := range lines {
				if len(line) <= lineWidth {
					wrappedLines = append(wrappedLines, line)
				} else {
					// For long lines, try to break at commas or after colons
					wrapped := wrapJSONLine(line, lineWidth)
					wrappedLines = append(wrappedLines, wrapped...)
				}
			}

			return strings.Join(wrappedLines, "\n")
		},
		"formatWebSearchData": func(toolName string, arguments string, result string) string {
			return g.formatWebSearchData(toolName, arguments, result)
		},
		"isWebSearchTool": func(toolName string) bool {
			return toolName == bag.WebSearch.String()
		},
	}

	tmplContent, err := templateFiles.ReadFile("templates/system.md")
	if err != nil {
		return nil, fmt.Errorf("failed to read system template: %w", err)
	}

	return template.Must(template.New("system").Funcs(funcMap).Parse(string(tmplContent))), nil
}

// addCommas adds comma separators to numeric strings
func addCommas(s string) string {
	// Find the decimal point if it exists
	decimalIndex := strings.Index(s, ".")
	var intPart, decPart string

	if decimalIndex == -1 {
		intPart = s
		decPart = ""
	} else {
		intPart = s[:decimalIndex]
		decPart = s[decimalIndex:]
	}

	// Add commas to the integer part
	if len(intPart) <= 3 {
		return s
	}

	var result strings.Builder
	for i, char := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(char)
	}

	return result.String() + decPart
}

// wrapJSONLine wraps a long JSON line to fit within the specified width
func wrapJSONLine(line string, maxWidth int) []string {
	if len(line) <= maxWidth {
		return []string{line}
	}

	var result []string
	current := ""

	// Get the indentation from the original line
	indent := ""
	for i, char := range line {
		if char == ' ' || char == '\t' {
			indent += string(char)
		} else {
			line = line[i:]
			break
		}
	}

	words := strings.Fields(line)
	for _, word := range words {
		var testLine string
		if current == "" {
			testLine = indent + word
		} else {
			testLine = current + " " + word
		}

		if len(testLine) <= maxWidth {
			current = testLine
		} else {
			if current != "" {
				result = append(result, current)
			}
			current = indent + "  " + word // Add extra indentation for continuation
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}

// formatInvestmentResearchData formats investment research analysis data into readable markdown
func (g *Generator) formatInvestmentResearchData(v interface{}) string {
	// Try to unmarshal as InvestmentResearchResult if it's JSON string
	if jsonStr, ok := v.(string); ok {
		var research models.InvestmentResearchResult
		if err := json.Unmarshal([]byte(jsonStr), &research); err == nil {
			return g.renderInvestmentResearchTemplate(research)
		}
	}

	// Try direct struct conversion
	if research, ok := v.(models.InvestmentResearchResult); ok {
		return g.renderInvestmentResearchTemplate(research)
	}

	// Fallback to pretty JSON
	jsonBytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting data: %v", err)
	}
	return string(jsonBytes)
}

// renderInvestmentResearchTemplate renders investment research data using a dedicated template
func (g *Generator) renderInvestmentResearchTemplate(research models.InvestmentResearchResult) string {
	tmpl, err := g.getInvestmentResearchTemplate()
	if err != nil {
		return fmt.Sprintf("Error loading template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, research); err != nil {
		return fmt.Sprintf("Error executing template: %v", err)
	}

	return buf.String()
}

// formatWebSearchData formats web search tool data into a readable format
func (g *Generator) formatWebSearchData(toolName string, arguments string, result string) string {
	// Parse the result JSON to extract web search data
	var resultData map[string]any
	if err := json.Unmarshal([]byte(result), &resultData); err != nil {
		return fmt.Sprintf("**Web Search Error**: Could not parse result: %v\n\n**Raw Result:** %s", err, result)
	}

	// Check if this is a web search result
	if toolName != bag.WebSearch.String() {
		return result // Not a web search, return as-is
	}

	// Parse arguments to get query details
	var argsData map[string]any
	if err := json.Unmarshal([]byte(arguments), &argsData); err != nil {
		argsData = make(map[string]any) // Use empty map if parsing fails
	}

	return g.renderWebSearchTemplate(argsData, resultData)
}

// renderWebSearchTemplate renders web search data using a dedicated template
func (g *Generator) renderWebSearchTemplate(args map[string]any, result map[string]any) string {
	tmpl, err := g.getWebSearchTemplate()
	if err != nil {
		return fmt.Sprintf("**Web Search Template Error**: %v", err)
	}

	data := map[string]any{
		"Args":   args,
		"Result": result,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Sprintf("**Web Search Render Error**: %v", err)
	}

	return buf.String()
}

// getInvestmentResearchTemplate returns the investment research template
func (g *Generator) getInvestmentResearchTemplate() (*template.Template, error) {
	funcMap := template.FuncMap{
		"toUpper": strings.ToUpper,
		"formatFloat": func(f float64, decimals int) string {
			format := fmt.Sprintf("%%.%df", decimals)
			return fmt.Sprintf(format, f)
		},
		"formatPercent": func(f float64) string {
			return fmt.Sprintf("%.1f%%", f)
		},
		"join":      strings.Join,
		"hasPrefix": strings.HasPrefix,
		"title":     strings.Title,
		"formatWebSearchData": func(toolName string, arguments string, result string) string {
			return g.formatWebSearchData(toolName, arguments, result)
		},
		"isWebSearchTool": func(toolName string) bool {
			return toolName == bag.WebSearch.String()
		},
	}

	tmplContent, err := templateFiles.ReadFile("templates/investment_research.md")
	if err != nil {
		return nil, fmt.Errorf("failed to read investment research template: %w", err)
	}

	return template.Must(template.New("investment_research").Funcs(funcMap).Parse(string(tmplContent))), nil
}

// getWebSearchTemplate returns the web search preview template
func (g *Generator) getWebSearchTemplate() (*template.Template, error) {
	funcMap := template.FuncMap{
		"formatFloat": func(f float64, decimals int) string {
			format := fmt.Sprintf("%%.%df", decimals)
			return fmt.Sprintf(format, f)
		},
		"formatDuration": func(v any) string {
			if durStr, ok := v.(string); ok {
				if dur, err := time.ParseDuration(durStr); err == nil {
					return dur.String()
				}
				return durStr
			}
			if durFloat, ok := v.(float64); ok {
				dur := time.Duration(durFloat * float64(time.Nanosecond))
				return dur.String()
			}
			return fmt.Sprintf("%v", v)
		},
		"getStatus": func(result map[string]any) string {
			if status, ok := result["status"].(string); ok {
				return status
			}
			return "unknown"
		},
		"getDescription": func(result map[string]any) string {
			if desc, ok := result["description"].(string); ok {
				return desc
			}
			return ""
		},
		"getQuery": func(args map[string]any) string {
			if query, ok := args["query"].(string); ok {
				return query
			}
			// Also check for nested query structures
			if actionMap, ok := args["action"].(map[string]any); ok {
				if query, ok := actionMap["query"].(string); ok {
					return query
				}
			}
			return ""
		},
		"getAction": func(args map[string]any) string {
			if action, ok := args["action"].(string); ok {
				return action
			}
			return ""
		},
		"join": strings.Join,
	}

	tmplContent, err := templateFiles.ReadFile("templates/web_search_preview.md")
	if err != nil {
		return nil, fmt.Errorf("failed to read web search template: %w", err)
	}

	return template.Must(template.New(bag.WebSearch.String()).Funcs(funcMap).Parse(string(tmplContent))), nil
}
