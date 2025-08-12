package openai

import (
	"testing"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCitationProcessor_ProcessCitations(t *testing.T) {
	sharedBag := bag.NewSharedBag()
	processor := NewCitationProcessor(sharedBag)

	cases := []struct {
		name              string
		query             string
		response          string
		expectedSuccess   bool
		expectedCitations int
		expectError       bool
	}{
		{
			name:              "JSON structured response",
			query:             "crypto market trends",
			response:          `{"content":"Market analysis shows...","sources":[{"url":"https://example.com/crypto","title":"Crypto News","snippet":"Latest trends"}]}`,
			expectedSuccess:   true,
			expectedCitations: 1,
		},
		{
			name:              "numbered citations",
			query:             "market analysis",
			response:          "Analysis shows volatility [1]. More data [2].\n[1] https://marketwatch.com/crypto\n[2] Reuters - https://reuters.com/market",
			expectedSuccess:   true,
			expectedCitations: 2,
		},
		{
			name:              "markdown links",
			query:             "investment trends",
			response:          "See [Market Analysis](https://bloomberg.com/markets) and [Crypto Report](https://coindesk.com/report)",
			expectedSuccess:   true,
			expectedCitations: 2,
		},
		{
			name:              "URLs only fallback",
			query:             "financial data",
			response:          "Check https://finance.yahoo.com and https://marketwatch.com for updates",
			expectedSuccess:   true,
			expectedCitations: 2,
		},
		{
			name:              "no citations found",
			query:             "general query",
			response:          "This is just text without any URLs or citations",
			expectedSuccess:   true,
			expectedCitations: 0,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result, err := processor.ProcessCitations(c.query, c.response)

			if c.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, c.expectedSuccess, result.Success)
			assert.Equal(t, c.expectedCitations, len(result.Citations))
			assert.Equal(t, c.query, result.Query)
			assert.Equal(t, c.expectedCitations, result.TotalCites)

			// Verify citations have required fields
			for _, citation := range result.Citations {
				assert.NotEmpty(t, citation.URL)
				assert.NotEmpty(t, citation.CitationID)
				assert.Equal(t, c.query, citation.Query)
				assert.False(t, citation.Timestamp.IsZero())
			}
		})
	}
}

func TestGetCitationResults(t *testing.T) {
	t.Run("with nil bag", func(t *testing.T) {
		results, ok := GetCitationResults(nil)
		assert.False(t, ok)
		assert.Nil(t, results)
	})

	t.Run("with empty bag", func(t *testing.T) {
		sharedBag := bag.NewSharedBag()
		results, ok := GetCitationResults(sharedBag)
		assert.False(t, ok)
		assert.Nil(t, results)
	})

	t.Run("with results", func(t *testing.T) {
		sharedBag := bag.NewSharedBag()
		processor := NewCitationProcessor(sharedBag)

		result := &CitationResult{
			Query:      "test query",
			TotalCites: 1,
			Citations: []Citation{
				{
					URL:       "https://example.com",
					Title:     "Example",
					Snippet:   "This is an example snippet",
					Timestamp: time.Now(),
				},
			},
		}

		processor.storeCitationResult(result)

		results, ok := GetCitationResults(sharedBag)
		require.True(t, ok)
		require.Len(t, results, 1)
		assert.Equal(t, "test query", results[0].Query)
	})
}

func TestProcessor_parseJSONResponse(t *testing.T) {
	sharedBag := bag.NewSharedBag()
	processor := NewCitationProcessor(sharedBag)

	cases := []struct {
		name              string
		response          string
		expectedCitations int
		expectError       bool
	}{
		{
			name:              "valid JSON response",
			response:          `{"content":"Test content","sources":[{"url":"https://example.com","title":"Test","snippet":"Test snippet"}]}`,
			expectedCitations: 1,
		},
		{
			name:              "multiple sources",
			response:          `{"content":"Test","sources":[{"url":"https://example1.com","title":"Test1"},{"url":"https://example2.com","title":"Test2"}]}`,
			expectedCitations: 2,
		},
		{
			name:        "invalid JSON",
			response:    `{"content":"Test"`,
			expectError: true,
		},
		{
			name:              "no sources",
			response:          `{"content":"Test content","sources":[]}`,
			expectedCitations: 0,
		},
		{
			name:              "invalid URLs filtered",
			response:          `{"content":"Test","sources":[{"url":"invalid-url","title":"Bad"},{"url":"https://good.com","title":"Good"}]}`,
			expectedCitations: 1,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			citations, content, err := processor.parseJSONResponse(c.response)

			if c.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, c.expectedCitations, len(citations))
			assert.NotEmpty(t, content)
		})
	}
}

func TestProcessor_parseNumberedCitations(t *testing.T) {
	sharedBag := bag.NewSharedBag()
	processor := NewCitationProcessor(sharedBag)

	cases := []struct {
		name              string
		response          string
		expectedCitations int
		expectError       bool
	}{
		{
			name:              "basic numbered citations",
			response:          "Text with [1] and [2].\n[1] https://example1.com\n[2] https://example2.com",
			expectedCitations: 2,
		},
		{
			name:              "citations with titles",
			response:          "Analysis [1] shows trends [2].\n[1] Market Analysis - https://market.com\n[2] Crypto Report - https://crypto.com",
			expectedCitations: 2,
		},
		{
			name:        "no citations",
			response:    "Just plain text without citations",
			expectError: true,
		},
		{
			name:              "duplicate citations",
			response:          "Text [1] more [1] text.\n[1] https://example.com",
			expectedCitations: 1,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			citations, content, err := processor.parseNumberedCitations(c.response)

			if c.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, c.expectedCitations, len(citations))
			assert.Equal(t, c.response, content)
		})
	}
}

func TestProcessor_parseMarkdownLinks(t *testing.T) {
	sharedBag := bag.NewSharedBag()
	processor := NewCitationProcessor(sharedBag)

	cases := []struct {
		name              string
		response          string
		expectedCitations int
		expectError       bool
	}{
		{
			name:              "basic markdown links",
			response:          "Check [Example](https://example.com) and [Test](https://test.com)",
			expectedCitations: 2,
		},
		{
			name:              "duplicate URLs",
			response:          "See [Link1](https://example.com) and [Link2](https://example.com)",
			expectedCitations: 1,
		},
		{
			name:        "no markdown links",
			response:    "Just plain text",
			expectError: true,
		},
		{
			name:              "invalid URLs filtered",
			response:          "Links: [Good](https://example.com) [Bad](not-a-url)",
			expectedCitations: 1,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			citations, content, err := processor.parseMarkdownLinks(c.response)

			if c.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, c.expectedCitations, len(citations))
			assert.Equal(t, c.response, content)
		})
	}
}

func TestProcessor_isValidURL(t *testing.T) {
	sharedBag := bag.NewSharedBag()
	processor := NewCitationProcessor(sharedBag)

	cases := []struct {
		name  string
		url   string
		valid bool
	}{
		{"valid HTTP", "http://example.com", true},
		{"valid HTTPS", "https://example.com", true},
		{"with path", "https://example.com/path", true},
		{"with query", "https://example.com/path?q=test", true},
		{"empty string", "", false},
		{"no scheme", "example.com", false},
		{"invalid scheme", "ftp://example.com", false},
		{"malformed", "https://", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result := processor.isValidURL(c.url)
			assert.Equal(t, c.valid, result)
		})
	}
}

func TestProcessor_storeBagResult(t *testing.T) {
	sharedBag := bag.NewSharedBag()
	processor := NewCitationProcessor(sharedBag)

	result := &CitationResult{
		Query:       "test query",
		Content:     "test content",
		Citations:   []Citation{{URL: "https://example.com", CitationID: "[1]"}},
		TotalCites:  1,
		ProcessedAt: time.Now(),
		Success:     true,
	}

	processor.storeCitationResult(result)

	// Verify result is stored
	results, ok := GetCitationResults(sharedBag)
	require.True(t, ok)
	require.Len(t, results, 1)
	assert.Equal(t, result.Query, results[0].Query)

	// Verify citations are stored separately
	citations, ok := GetWebSearchCitations(sharedBag)
	require.True(t, ok)
	require.Len(t, citations, 1)
	assert.Equal(t, "https://example.com", citations[0].URL)
}

func TestGetWebSearchCitations(t *testing.T) {
	t.Run("with nil bag", func(t *testing.T) {
		citations, ok := GetWebSearchCitations(nil)
		assert.False(t, ok)
		assert.Nil(t, citations)
	})

	t.Run("with empty bag", func(t *testing.T) {
		sharedBag := bag.NewSharedBag()
		citations, ok := GetWebSearchCitations(sharedBag)
		assert.False(t, ok)
		assert.Nil(t, citations)
	})

	t.Run("with citations", func(t *testing.T) {
		sharedBag := bag.NewSharedBag()
		processor := NewCitationProcessor(sharedBag)

		result := &CitationResult{
			Query:     "test",
			Citations: []Citation{{URL: "https://example.com", CitationID: "[1]"}},
			Success:   true,
		}

		processor.storeCitationResult(result)

		citations, ok := GetWebSearchCitations(sharedBag)
		require.True(t, ok)
		require.Len(t, citations, 1)
		assert.Equal(t, "https://example.com", citations[0].URL)
	})
}
