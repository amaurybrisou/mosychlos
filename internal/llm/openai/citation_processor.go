package openai

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
)

// Citation represents a parsed web search citation
type Citation struct {
	URL        string    `json:"url"`
	Title      string    `json:"title,omitempty"`
	Snippet    string    `json:"snippet,omitempty"`
	Source     string    `json:"source,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	Query      string    `json:"query"`
	Relevance  float64   `json:"relevance,omitempty"`
	CitationID string    `json:"citation_id"` // For referencing in text
}

// CitationResult represents the complete result from web search citation processing
type CitationResult struct {
	Query       string     `json:"query"`
	Content     string     `json:"content"`
	Citations   []Citation `json:"citations"`
	TotalCites  int        `json:"total_cites"`
	ProcessedAt time.Time  `json:"processed_at"`
	Success     bool       `json:"success"`
	Error       string     `json:"error,omitempty"`
}

// CitationProcessor handles web search citation extraction and storage
type CitationProcessor struct {
	sharedBag bag.SharedBag
}

// NewCitationProcessor creates a new citation processor
func NewCitationProcessor(sharedBag bag.SharedBag) *CitationProcessor {
	return &CitationProcessor{
		sharedBag: sharedBag,
	}
}

// ProcessCitations processes a web search response and extracts citations
func (p *CitationProcessor) ProcessCitations(query, response string) (*CitationResult, error) {
	result := &CitationResult{
		Query:       query,
		Content:     response,
		Citations:   []Citation{},
		ProcessedAt: time.Now(),
		Success:     false,
	}

	// Try different parsing strategies
	citations, content, err := p.parseResponse(response)
	if err != nil {
		result.Error = err.Error()
		p.storeCitationResult(result)
		return result, err
	}

	// Enhance citations with query context
	for i := range citations {
		citations[i].Query = query
		citations[i].Timestamp = result.ProcessedAt
		if citations[i].CitationID == "" {
			citations[i].CitationID = fmt.Sprintf("[%d]", i+1)
		}
	}

	result.Citations = citations
	result.Content = content
	result.TotalCites = len(citations)
	result.Success = true

	p.storeCitationResult(result)

	slog.Info("Citations processed successfully",
		"query", query,
		"citations_found", len(citations),
		"content_length", len(content),
	)

	return result, nil
}

// parseResponse attempts to parse citations from different response formats
func (p *CitationProcessor) parseResponse(response string) ([]Citation, string, error) {
	// Try JSON structured response first
	if citations, content, err := p.parseJSONResponse(response); err == nil {
		return citations, content, nil
	}

	// Try numbered citations format [1], [2], etc.
	if citations, content, err := p.parseNumberedCitations(response); err == nil && len(citations) > 0 {
		return citations, content, nil
	}

	// Try markdown links format
	if citations, content, err := p.parseMarkdownLinks(response); err == nil && len(citations) > 0 {
		return citations, content, nil
	}

	// Try URL extraction as fallback
	if citations, content, err := p.parseURLsOnly(response); err == nil && len(citations) > 0 {
		return citations, content, nil
	}

	// If all parsing fails, return empty result but no error
	return []Citation{}, response, nil
}

// parseJSONResponse attempts to parse structured JSON response
func (p *CitationProcessor) parseJSONResponse(response string) ([]Citation, string, error) {
	var structured struct {
		Content string `json:"content"`
		Sources []struct {
			URL     string  `json:"url"`
			Title   string  `json:"title"`
			Snippet string  `json:"snippet"`
			Source  string  `json:"source"`
			Score   float64 `json:"relevance"`
		} `json:"sources"`
	}

	if err := json.Unmarshal([]byte(response), &structured); err != nil {
		return nil, "", err
	}

	var citations []Citation
	for i, source := range structured.Sources {
		if !p.isValidURL(source.URL) {
			continue
		}

		citations = append(citations, Citation{
			URL:        source.URL,
			Title:      source.Title,
			Snippet:    source.Snippet,
			Source:     source.Source,
			Relevance:  source.Score,
			CitationID: fmt.Sprintf("[%d]", i+1),
		})
	}

	return citations, structured.Content, nil
}

// parseNumberedCitations parses citations in format: text [1] more text
// with sources listed as [1] https://example.com
func (p *CitationProcessor) parseNumberedCitations(response string) ([]Citation, string, error) {
	// Regex to find numbered citations in text
	citationRegex := regexp.MustCompile(`\[(\d+)\]`)
	matches := citationRegex.FindAllStringSubmatch(response, -1)

	if len(matches) == 0 {
		return nil, "", fmt.Errorf("no numbered citations found")
	}

	var citations []Citation
	citationMap := make(map[string]Citation)

	// Look for citation definitions: [1] URL or [1] Title - URL
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[") {
			if citation := p.parseNumberedCitationLine(line); citation != nil {
				citationMap[citation.CitationID] = *citation
			}
		}
	}

	// Collect unique citations in order
	seen := make(map[string]bool)
	for _, match := range matches {
		citationID := fmt.Sprintf("[%s]", match[1])
		if !seen[citationID] {
			if citation, exists := citationMap[citationID]; exists {
				citations = append(citations, citation)
			}
			seen[citationID] = true
		}
	}

	return citations, response, nil
}

// parseNumberedCitationLine parses a single citation line like "[1] Title - https://example.com"
func (p *CitationProcessor) parseNumberedCitationLine(line string) *Citation {
	// Match patterns like: [1] https://example.com or [1] Title - https://example.com
	urlRegex := regexp.MustCompile(`\[(\d+)\]\s*(?:(.+?)\s*-\s*)?(.+)`)
	matches := urlRegex.FindStringSubmatch(line)

	if len(matches) < 4 {
		return nil
	}

	url := strings.TrimSpace(matches[3])
	if !p.isValidURL(url) {
		return nil
	}

	return &Citation{
		URL:        url,
		Title:      strings.TrimSpace(matches[2]),
		CitationID: fmt.Sprintf("[%s]", matches[1]),
	}
}

// parseMarkdownLinks extracts markdown-style links [text](url)
func (p *CitationProcessor) parseMarkdownLinks(response string) ([]Citation, string, error) {
	linkRegex := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	matches := linkRegex.FindAllStringSubmatch(response, -1)

	if len(matches) == 0 {
		return nil, "", fmt.Errorf("no markdown links found")
	}

	var citations []Citation
	seen := make(map[string]bool)

	for i, match := range matches {
		url := strings.TrimSpace(match[2])
		if !p.isValidURL(url) || seen[url] {
			continue
		}

		citations = append(citations, Citation{
			URL:        url,
			Title:      strings.TrimSpace(match[1]),
			CitationID: fmt.Sprintf("[%d]", i+1),
		})
		seen[url] = true
	}

	return citations, response, nil
}

// parseURLsOnly extracts any valid URLs as fallback
func (p *CitationProcessor) parseURLsOnly(response string) ([]Citation, string, error) {
	urlRegex := regexp.MustCompile(`https?://[^\s<>"{}|\\^` + "`" + `\[\]]+`)
	matches := urlRegex.FindAllString(response, -1)

	if len(matches) == 0 {
		return nil, "", fmt.Errorf("no URLs found")
	}

	var citations []Citation
	seen := make(map[string]bool)

	for i, match := range matches {
		urlStr := strings.TrimSpace(match)
		if !p.isValidURL(urlStr) || seen[urlStr] {
			continue
		}

		// Extract domain as title
		if parsed, err := url.Parse(urlStr); err == nil {
			citations = append(citations, Citation{
				URL:        urlStr,
				Source:     parsed.Host,
				CitationID: fmt.Sprintf("[%d]", i+1),
			})
			seen[urlStr] = true
		}
	}

	return citations, response, nil
}

// isValidURL validates if a string is a valid HTTP/HTTPS URL
func (p *CitationProcessor) isValidURL(s string) bool {
	if s == "" {
		return false
	}

	parsed, err := url.Parse(s)
	if err != nil {
		return false
	}

	return (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}

// storeCitationResult stores the citation result in the shared bag
func (p *CitationProcessor) storeCitationResult(result *CitationResult) {
	if p.sharedBag == nil {
		slog.Warn("SharedBag is nil, cannot store citation result")
		return
	}

	// Store individual result
	resultKey := fmt.Sprintf("%s.result.%d", keys.WebSearch, time.Now().Unix())
	p.sharedBag.Set(keys.Key(resultKey), result)

	// Store or update aggregated results
	var allResults []*CitationResult
	if existing, ok := p.sharedBag.Get(keys.WebSearch); ok {
		if existingResults, ok := existing.([]*CitationResult); ok {
			allResults = existingResults
		}
	}

	allResults = append(allResults, result)
	p.sharedBag.Set(keys.WebSearch, allResults)

	// Store citations separately for easy access
	if result.Success && len(result.Citations) > 0 {
		citationsKey := keys.Key(fmt.Sprintf("%s.citations", keys.WebSearch))
		var allCitations []Citation
		if existing, ok := p.sharedBag.Get(citationsKey); ok {
			if existingCitations, ok := existing.([]Citation); ok {
				allCitations = existingCitations
			}
		}
		allCitations = append(allCitations, result.Citations...)
		p.sharedBag.Set(citationsKey, allCitations)
	}

	slog.Debug("Citation result stored in SharedBag",
		"key", resultKey,
		"citations_count", len(result.Citations),
		"success", result.Success,
	)
}

// GetCitationResults retrieves all citation results from the bag
func GetCitationResults(sharedBag bag.SharedBag) ([]*CitationResult, bool) {
	if sharedBag == nil {
		return nil, false
	}

	if results, ok := sharedBag.Get(keys.WebSearch); ok {
		if citationResults, ok := results.([]*CitationResult); ok {
			return citationResults, true
		}
	}

	return nil, false
}

// GetWebSearchCitations retrieves all citations from the bag
func GetWebSearchCitations(sharedBag bag.SharedBag) ([]Citation, bool) {
	if sharedBag == nil {
		return nil, false
	}

	citationsKey := keys.Key(fmt.Sprintf("%s.citations", keys.WebSearch))
	if citations, ok := sharedBag.Get(citationsKey); ok {
		if webCitations, ok := citations.([]Citation); ok {
			return webCitations, true
		}
	}

	return nil, false
}
