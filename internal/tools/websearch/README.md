# Web Search Tool Integration

This document explains the integration of OpenAI's internal `web_search_preview` tool into the Mosychlos batch processing system.

## Overview

The `web_search_preview` tool is an **internal OpenAI tool** that OpenAI executes automatically when included in tool definitions. This tool allows the AI to search the web for current market information, news, and analysis to complement portfolio data.

## Architecture

### Virtual Tool Implementation

- **Location**: `internal/tools/websearch/`
- **Type**: Virtual tool (no local execution)
- **Purpose**: Provides tool definition for OpenAI's internal web search

### Key Components

1. **websearch.go**: Virtual tool implementation that provides the tool definition
2. **tools_config.go**: Configuration provider that conditionally registers the tool
3. **Tool Registry Integration**: Automatic registration when web search is enabled in config

## Configuration

Web search is controlled by the OpenAI configuration:

```yaml
llm:
  openai:
    web_search: true # Enable web search
    web_search_context_size: 'medium' # Context size: low, medium, high
    # web_search_user_location is computed automatically from localization
```

## Integration Points

### 1. Tool Registration

- **File**: `internal/tools/tool_registry.go`
- **Logic**: Web search tool is registered only when `cfg.LLM.OpenAI.WebSearch` is `true`

### 2. Engine Constraints

- **File**: `internal/engine/wiring.go`
- **Logic**: Risk engine includes web search in preferred tools when enabled
- **Constraints**:
  - Minimum calls: 1 (for market context)
  - Maximum calls: 3 (for comprehensive analysis)

### 3. Batch Processing

- **File**: `internal/engine/risk/risk_batch.go`
- **Behavior**: Web search tool calls are included in batch requests
- **Execution**: OpenAI executes web searches internally and integrates results

## Usage in Risk Analysis

When web search is enabled, the risk analysis engine will:

1. **Include web search in tool constraints**
2. **Request 1-3 web searches** for current market conditions
3. **Receive integrated results** from OpenAI's internal execution
4. **Generate enhanced risk analysis** with real-time market context

## Web Search Processing Pipeline

### Citation Parsing Architecture

The web search integration includes a dedicated processing pipeline:

- **Location**: `internal/tools/websearch/processor.go`
- **Purpose**: Parse citations from OpenAI web search responses
- **Storage**: Citations stored in SharedBag under `keys.WebSearch`

### Processor Features

1. **Multiple Citation Formats Support**:

   - **JSON Structured**: `{"content":"...", "sources":[{"url":"...", "title":"..."}]}`
   - **Numbered Citations**: Text with `[1]`, `[2]` and reference list
   - **Markdown Links**: `[title](url)` format
   - **URL Extraction**: Fallback to any HTTP/HTTPS URLs found

2. **Citation Data Model**:

   ```go
   type Citation struct {
       URL         string    `json:"url"`
       Title       string    `json:"title,omitempty"`
       Snippet     string    `json:"snippet,omitempty"`
       Source      string    `json:"source,omitempty"`
       Timestamp   time.Time `json:"timestamp"`
       Query       string    `json:"query"`
       Relevance   float64   `json:"relevance,omitempty"`
       CitationID  string    `json:"citation_id"`
   }
   ```

3. **Bag Storage Strategy**:
   - Individual results: `web_search_preview.result.{timestamp}`
   - Aggregated results: `web_search_preview`
   - Citations only: `web_search_preview.citations`

### Usage Examples

```go
// Create processor
processor := websearch.NewProcessor(sharedBag)

// Process OpenAI web search response
result, err := processor.ProcessWebSearchResponse(query, response)

// Retrieve all web search results from bag
results, ok := websearch.GetWebSearchResults(sharedBag)

// Get just the citations
citations, ok := websearch.GetWebSearchCitations(sharedBag)
```

## Example Tool Call

The AI can call web search like any other tool:

```json
{
  "name": keys.WebSearch,
  "arguments": {
    "query": "current cryptocurrency market conditions PEPE BNB price volatility"
  }
}
```

OpenAI executes this internally and provides the results directly in the response content.

## Benefits

1. **Real-time Market Data**: Access to current market conditions and news
2. **Enhanced Analysis**: Complement portfolio data with web-sourced insights
3. **Zero Infrastructure**: No local web search implementation required
4. **Batch Compatible**: Works seamlessly with existing batch processing
5. **Config Driven**: Easy to enable/disable based on needs

## Monitoring

Web search usage is tracked through:

- **Tool Metrics**: Tracked via the existing metrics system
- **Computation Logs**: Recorded in shared bag for analysis
- **Health Monitoring**: Included in external data health checks

## Troubleshooting

If web search is not working:

1. **Check Configuration**: Ensure `web_search: true` in OpenAI config
2. **Verify Tool Registration**: Look for "Web search tool registered" in logs
3. **Check Engine Constraints**: Verify web search is in preferred tools
4. **Review Batch Results**: Check if web search tool calls are present

## Future Enhancements

Potential improvements:

- Query optimization based on portfolio composition
- Regional search context customization
- Web search result caching (if beneficial)
- Enhanced prompt engineering for better search queries
