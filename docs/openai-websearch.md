# OpenAI Web Search Implementation

## Overview

This document describes the implementation of OpenAI's web search functionality in the mosychlos v2 project. The implementation uses the OpenAI Responses API with the `web_search_preview` tool to provide real-time access to web information.

## Implementation Status: IN PROGRESS 🚧

The OpenAI web search functionality is partially implemented with configuration support but requires additional integration work:

### ✅ Completed Components:

- Configuration structure in `config.go`
- Provider web search detection (`HasWebSearch()`)
- Response API infrastructure
- Test framework setup

### 🚧 Pending Implementation:

- Full Responses API integration
- Web search tool registration
- Citation handling and rendering
- UI integration for search results

## Current Implementation

## Configuration

Web search is configured via the OpenAI configuration section:

```yaml
llm:
  openai:
    # Enable web search capability (requires Responses API)
    web_search: true
    # Control search context depth: low, medium, high
    web_search_context_size: medium

# Localization used for geographic targeting
localization:
  country: 'US'
  city: 'San Francisco'
  region: 'California'
  timezone: 'America/Los_Angeles'
```

The `WebSearchUserLocation` is automatically populated from the centralized localization configuration.

## API Integration

### Current Provider Implementation

The OpenAI provider includes basic web search detection:

```go
// HasWebSearch returns true if web search is enabled in the configuration
func (p *Provider) HasWebSearch() bool {
    return p.config.OpenAI.WebSearch != nil && *p.config.OpenAI.WebSearch
}
```

### Configuration Structure

```go
type OpenAIConfig struct {
    // WebSearch enables OpenAI's web search capability (uses Responses API)
    WebSearch *bool `mapstructure:"web_search" yaml:"web_search"`
    // WebSearchContextSize controls web search context: low, medium, high
    WebSearchContextSize *string `mapstructure:"web_search_context_size" yaml:"web_search_context_size"`
    // WebSearchUserLocation is computed at runtime from centralized localization
    WebSearchUserLocation *WebSearchUserLocationConfig
}

type WebSearchUserLocationConfig struct {
    Country  *string
    City     *string
    Region   *string
    Timezone *string
}
```

## Responses API vs Chat Completions

**Important**: Web search requires the **Responses API**, not the traditional Chat Completions API.

### Key Differences:

| Chat Completions API     | Responses API                     |
| ------------------------ | --------------------------------- |
| Traditional OpenAI API   | New API for tools like web search |
| JSON response format     | Structured response items         |
| No built-in citations    | Automatic URL citations           |
| Limited tool integration | Native web search support         |

### Tool Definition Format

```go
// For Responses API
tools := []models.ToolDef{
    {
        Type: "web_search_preview",
        SearchContextSize: "medium", // low, medium, high
        UserLocation: &ai.UserLocation{
            Type:     "approximate",
            Country:  "US",
            City:     "San Francisco",
            Region:   "California",
            Timezone: "America/Los_Angeles",
        },
    },
}
```

## Implementation Roadmap

### Phase 1: Core Integration

- [ ] Implement full Responses API client
- [ ] Add web search tool registration
- [ ] Handle web search response format
- [ ] Basic citation extraction

### Phase 2: Advanced Features

- [ ] Citation rendering in UI
- [ ] Geographic targeting integration
- [ ] Context size optimization
- [ ] Error handling and fallbacks

### Phase 3: Production Features

- [ ] Caching layer for search results
- [ ] Rate limiting and quota management
- [ ] Search result filtering
- [ ] Performance monitoring

## Integration Notes

### Automatic Location Population

The system automatically populates `WebSearchUserLocation` from the centralized localization config:

```go
// In config.go - populateComputedFields()
c.LLM.OpenAI.WebSearchUserLocation = &WebSearchUserLocationConfig{
    Country:  &c.Localization.Country,
    City:     &c.Localization.City,
    Region:   &c.Localization.Region,
    Timezone: &c.Localization.Timezone,
}
```

### Provider Detection

The system can detect if web search is available:

```go
if provider.HasWebSearch() {
    // Use Responses API with web search tools
} else {
    // Fall back to Chat Completions API
}
```

## Reference: OpenAI Web Search API

Based on OpenAI's official documentation: https://platform.openai.com/docs/guides/tools-web-search

### Tool Versions

- **Current Default**: `web_search_preview`
- **Specific Version**: `web_search_preview_2025_03_11`

Future dated versions will be documented in the API reference.

### Response Structure

Web search responses include two main components:

#### 1. Web Search Call Item

```json
{
  "type": "web_search_call",
  "id": "ws_67c9fa0502748190b7dd390736892e100be649c1a5ff9609",
  "status": "completed",
  "action": "search",
  "query": "positive news today",
  "domains": ["example.com", "news.org"]
}
```

#### 2. Message Item with Citations

```json
{
  "type": "message",
  "content": [
    {
      "type": "output_text",
      "text": "Based on today's news...",
      "annotations": [
        {
          "type": "url_citation",
          "start_index": 25,
          "end_index": 45,
          "url": "https://example.com/news",
          "title": "Positive News Story"
        }
      ]
    }
  ]
}
```

## Search Actions

The tool can perform different types of actions:

1. **`search`**: Standard web search with query and domains
2. **`open_page`**: Opens specific pages (Deep Research models)
3. **`find_in_page`**: Searches within a page (Deep Research models)

## Context Size Options

| Size     | Description                     | Use Case                           |
| -------- | ------------------------------- | ---------------------------------- |
| `low`    | Fastest response, least context | Quick facts, simple queries        |
| `medium` | Balanced (default)              | General purpose searches           |
| `high`   | Most comprehensive, slower      | Complex research, detailed answers |

## User Location Parameters

- **`country`**: Two-letter ISO country code (e.g., "US", "GB", "CA")
- **`city`**: Free text city name (e.g., "New York", "London")
- **`region`**: Free text region/state (e.g., "California", "Ontario")
- **`timezone`**: IANA timezone (e.g., "America/New_York")

## Implementation in Mosychlos

### Current Status

The basic infrastructure is in place but requires completion:

```go
// Provider detection (✅ Implemented)
if provider.HasWebSearch() {
    // Web search is configured and available
}

// Configuration (✅ Implemented)
type OpenAIConfig struct {
    WebSearch *bool
    WebSearchContextSize *string
    WebSearchUserLocation *WebSearchUserLocationConfig
}

// Response handling (🚧 Needs Implementation)
// Citation processing (🚧 Needs Implementation)
// Tool registration (🚧 Needs Implementation)
```

### Configuration Example

```yaml
llm:
  provider: 'openai'
  model: 'gpt-4o'
  api_key: '${OPENAI_API_KEY}'
  openai:
    web_search: true
    web_search_context_size: 'medium'

localization:
  country: 'US'
  city: 'San Francisco'
  region: 'California'
  timezone: 'America/Los_Angeles'
```

### Next Steps

1. **Complete Responses API client implementation**
2. **Add web search tool to tool registry**
3. **Implement citation handling and rendering**
4. **Add UI support for displaying search results with citations**
5. **Test geographic targeting with different locales**

## Supported Models

### Full Support

- `gpt-4o`
- `gpt-4o-mini`
- `gpt-4.1`
- `gpt-4.1-mini`
- `o1`
- `o1-mini`
- `o3`
- `o3-mini`
- `o4`
- `o4-mini`

### Limited Support

- `gpt-4o-search-preview` (Chat Completions only)
- `gpt-4o-mini-search-preview` (Chat Completions only)

### Not Supported

- `gpt-4.1-nano`

## Rate Limits and Pricing

- **Rate Limits**: Same tiered limits as the underlying model
- **Search Tokens**: May be free or billed at model rates (see pricing)
- **Tool Calls**: Incur additional tool call costs
- **Context Window**: Limited to 128,000 tokens for search content

## Best Practices

1. **Citation Display**: Make inline citations clearly visible and clickable
2. **Location Targeting**: Use user location for geographically relevant queries
3. **Context Size**: Choose appropriate size based on query complexity
4. **Error Handling**: Handle web search failures gracefully
5. **Caching**: Consider caching frequent search results

## Limitations

- Not available in `gpt-4.1-nano`
- Context window limited to 128K tokens
- Location not supported for Deep Research models
- Context size configuration not supported for o3/o4/Deep Research models
- Search tokens don't carry over between turns

## Migration Notes

This functionality requires using the **Responses API** rather than the traditional Chat Completions API. The mosychlos implementation will need to:

1. Support both APIs in the OpenAI provider
2. Automatically use Responses API when web search tools are detected
3. Handle the different response format with citations
4. Provide proper citation rendering in the UI

## Security Considerations

- Web search results are processed by OpenAI's systems
- User location data is sent to OpenAI for geographic targeting
- Search queries and results follow OpenAI's data retention policies
- Consider privacy implications when using location-based search
