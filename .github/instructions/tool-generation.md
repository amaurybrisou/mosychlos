---
applyTo: 'internal/tools/**/*.go'
---

# Tool Generation Instructions

This guide provides step-by-step instructions for creating new tools in the Mosychlos tools ecosystem.

## **Quick Reference**

New tools should be created in `internal/tools/<toolname>/` following established patterns for consistency, caching, and monitoring.

## **Tool Architecture Overview**

```
┌─────────────────────────────────────────────────────────┐
│                    Tool Ecosystem                       │
│                                                         │
│ ┌─────────────┐   ┌─────────────┐   ┌─────────────────┐ │
│ │   Raw Tool  │──►│ Cached Tool │──►│ Metrics Wrapper │ │
│ │             │   │             │   │                 │ │
│ │ • API calls │   │ • File cache│   │ • Performance   │ │
│ │ • Business  │   │ • TTL mgmt  │   │ • Error rates   │ │
│ │   logic     │   │ • Key hash  │   │ • Usage stats   │ │
│ └─────────────┘   └─────────────┘   └─────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

## **Step 1: Create Tool Directory Structure**

```bash
# Create tool directory
mkdir -p internal/tools/<toolname>

# Create required files
touch internal/tools/<toolname>/<toolname>.go
touch internal/tools/<toolname>/<toolname>_test.go
touch internal/tools/<toolname>/README.md
```

## **Step 2: Implement Core Tool Structure**

Create `internal/tools/<toolname>/<toolname>.go`:

```go
package toolname

import (
    "context"
    "encoding/json"
    "fmt"
    "log/slog"
    "net/http"
    "time"

    "github.com/amaurybrisou/mosychlos/internal/config"
    "github.com/amaurybrisou/mosychlos/pkg/keys"
    "github.com/amaurybrisou/mosychlos/pkg/models"
)

// Provider implements the Tool interface for <ToolName>
// the name should follow this logic:
// <ToolNameFileName>Tool in camel case.
// for example stock_data.go in yfinance folder would result in: YFinanceStockDataTool
type Provider struct {
    apiKey  string
    baseURL string
    http    *http.Client
    // Add tool-specific fields
}

var _ models.Tool = &Provider{}

// NewFromConfig creates a new Provider from config
func NewFromConfig(cfg *config.ToolNameConfig) (*Provider, error) {
    return New(cfg.APIKey, cfg.BaseURL)
}

// New constructs a provider with the given configuration
func New(apiKey, baseURL string) (*Provider, error) {
    if apiKey == "" {
        return nil, fmt.Errorf("toolname: missing API key")
    }

    if baseURL == "" {
        baseURL = "https://api.example.com"
    }

    return &Provider{
        apiKey:  apiKey,
        baseURL: baseURL,
        http: &http.Client{
            Timeout: 30 * time.Second,
        },
    }, nil
}

// Name returns the tool name for AI function calling
func (p *Provider) Name() string {
    return "tool_name_function"
}

// Key returns the unique tool key
func (p *Provider) Key() bag.Key {
    return bag.ToolName // Add to pkg/keys/bag.go
}

// Description returns the tool description for AI
func (p *Provider) Description() string {
    return "Description of what this tool does and when to use it"
}

// Definition returns the OpenAI tool definition
func (p *Provider) Definition() models.ToolDef {
    return models.ToolDef{
        Type: "function",
        Function: models.FunctionDef{
            Name:        p.Name(),
            Description: p.Description(),
            Parameters: map[string]any{
                "type": "object",
                "properties": map[string]any{
                    "parameter1": map[string]any{
                        "type":        "string",
                        "description": "Description of parameter1",
                    },
                    "parameter2": map[string]any{
                        "type":        "array",
                        "items":       map[string]any{"type": "string"},
                        "description": "Array of items for parameter2",
                    },
                },
                "required": []string{"parameter1"},
            },
        },
    }
}

// Tags returns tool tags for categorization
func (p *Provider) Tags() []string {
    return []string{"financial", "data", "external-api"} // Customize tags
}

// Run executes the tool with the given arguments
func (p *Provider) Run(ctx context.Context, args string) (string, error) {
    slog.Debug("Running tool",
        "tool", p.Name(),
        "args", args,
    )

    // Parse arguments
    var params struct {
        Parameter1 string   `json:"parameter1"`
        Parameter2 []string `json:"parameter2,omitempty"`
    }

    if err := json.Unmarshal([]byte(args), &params); err != nil {
        return "", fmt.Errorf("failed to parse arguments: %w", err)
    }

    // Validate required parameters
    if params.Parameter1 == "" {
        return "", fmt.Errorf("parameter1 is required")
    }

    // Execute tool logic
    result, err := p.executeToolLogic(ctx, params.Parameter1, params.Parameter2)
    if err != nil {
        slog.Error("Tool execution failed",
            "tool", p.Name(),
            "error", err,
            "parameter1", params.Parameter1,
        )
        return "", fmt.Errorf("tool execution failed: %w", err)
    }

    // Return JSON response
    response, err := json.Marshal(result)
    if err != nil {
        return "", fmt.Errorf("failed to marshal response: %w", err)
    }

    slog.Info("Tool executed successfully",
        "tool", p.Name(),
        "parameter1", params.Parameter1,
        "result_size", len(response),
    )

    return string(response), nil
}

// executeToolLogic implements the core business logic
func (p *Provider) executeToolLogic(ctx context.Context, param1 string, param2 []string) (any, error) {
    // Implement your tool's core functionality here
    // This is where you make API calls, process data, etc.

    // Example HTTP request
    url := fmt.Sprintf("%s/endpoint?param1=%s", p.baseURL, param1)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("Authorization", "Bearer "+p.apiKey)
    req.Header.Set("Content-Type", "application/json")

    resp, err := p.http.Do(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
    }

    var result map[string]any
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }

    return result, nil
}
```

## **Step 3: Add Configuration Support**

Add configuration struct to `internal/config/config.go`:

```go
type ToolNameConfig struct {
    APIKey      string `mapstructure:"api_key" yaml:"api_key"`
    BaseURL     string `mapstructure:"base_url" yaml:"base_url"`
    CacheEnable bool   `mapstructure:"cache_enable" yaml:"cache_enable"`
    MaxDaily    int    `mapstructure:"max_daily" yaml:"max_daily"`
}
```

Add to `ToolsConfig` struct:

```go
type ToolsConfig struct {
    // ... existing tools
    ToolName *ToolNameConfig `mapstructure:"toolname" yaml:"toolname"`
}
```

## **Step 4: Add Tool Key**

Add to `pkg/keys/bag.go`:

```go
const (
    // ... existing keys
    ToolName Key = "toolname"
)
```

## **Step 5: Register Tool in Tools Registry**

Add to `internal/tools/tools.go`:

```go
import (
    // ... existing imports
    "github.com/amaurybrisou/mosychlos/internal/tools/toolname"
)

// Add to NewTools function
func NewTools(cfg *config.Config) error {
    // ... existing tool registrations

    if cfg.Tools.ToolName != nil {
        if err := NewToolNameTool(cfg); err != nil {
            return err
        }
    }

    return nil
}

// Add tool initialization function
func NewToolNameTool(cfg *config.Config) error {
    if cfg.Tools.ToolName == nil {
        return nil
    }

    tool, err := toolname.NewFromConfig(cfg.Tools.ToolName)
    if err != nil {
        return fmt.Errorf("failed to create toolname tool: %w", err)
    }

    var wrappedTool models.Tool = tool

    // Wrap with cache if enabled
    if cfg.Tools.ToolName.CacheEnable {
        wrappedTool = NewCachedToolWithMonitoring(wrappedTool, cfg.CacheDir, 24*time.Hour, sharedBag)
    }

    // Wrap with metrics tracking if shared bag is available
    if sharedBag != nil {
        wrappedTool = NewMetricsWrapper(wrappedTool, sharedBag)
    }

    tools[bag.ToolName] = wrappedTool

    slog.Info("ToolName tool initialized",
        "cache_enabled", cfg.Tools.ToolName.CacheEnable,
        "base_url", cfg.Tools.ToolName.BaseURL,
    )

    return nil
}
```

## **Step 6: Create Comprehensive Tests**

Create `internal/tools/<toolname>/<toolname>_test.go`:

```go
package toolname

import (
    "context"
    "testing"
    "time"

    "github.com/amaurybrisou/mosychlos/internal/config"
)

func TestProvider_New(t *testing.T) {
    cases := []struct {
        name      string
        apiKey    string
        baseURL   string
        wantError bool
    }{
        {
            name:    "valid config",
            apiKey:  "test-key",
            baseURL: "https://api.test.com",
        },
        {
            name:      "missing api key",
            apiKey:    "",
            baseURL:   "https://api.test.com",
            wantError: true,
        },
        {
            name:    "default base url",
            apiKey:  "test-key",
            baseURL: "",
        },
    }

    for _, c := range cases {
        t.Run(c.name, func(t *testing.T) {
            provider, err := New(c.apiKey, c.baseURL)

            if c.wantError {
                if err == nil {
                    t.Error("expected error but got nil")
                }
                return
            }

            if err != nil {
                t.Errorf("unexpected error: %v", err)
                return
            }

            if provider == nil {
                t.Error("expected provider but got nil")
            }
        })
    }
}

func TestProvider_ToolInterface(t *testing.T) {
    provider, err := New("test-key", "https://api.test.com")
    if err != nil {
        t.Fatalf("failed to create provider: %v", err)
    }

    // Test tool interface methods
    if provider.Name() == "" {
        t.Error("Name() returned empty string")
    }

    if provider.Description() == "" {
        t.Error("Description() returned empty string")
    }

    if len(provider.Tags()) == 0 {
        t.Error("Tags() returned empty slice")
    }

    def := provider.Definition()
    if def.Type != "function" {
        t.Errorf("Definition() Type = %v, want function", def.Type)
    }

    if def.Function.Name != provider.Name() {
        t.Errorf("Definition() Function.Name = %v, want %v", def.Function.Name, provider.Name())
    }
}

// Add integration test if needed
func TestProvider_Run_Integration(t *testing.T) {
    // Skip if no API key available
    apiKey := os.Getenv("TOOLNAME_API_KEY")
    if apiKey == "" {
        t.Skip("TOOLNAME_API_KEY not set, skipping integration test")
    }

    provider, err := New(apiKey, "")
    if err != nil {
        t.Fatalf("failed to create provider: %v", err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    args := `{"parameter1": "test-value"}`
    result, err := provider.Run(ctx, args)
    if err != nil {
        t.Fatalf("Run() failed: %v", err)
    }

    if result == "" {
        t.Error("Run() returned empty result")
    }

    t.Logf("Result: %s", result)
}
```

Add integration test to main tools package in `internal/tools/<toolname>_tool_test.go`:

```go
//go:build integration

package tools

import (
    "context"
    "os"
    "testing"
    "time"

    "github.com/amaurybrisou/mosychlos/internal/config"
)

func TestToolNameTool_Integration(t *testing.T) {
    // Test configuration
    apiKey := os.Getenv("TOOLNAME_API_KEY")
    if apiKey == "" {
        t.Skip("TOOLNAME_API_KEY not set")
    }

    cfg := &config.Config{
        CacheDir: t.TempDir(),
        Tools: config.ToolsConfig{
            ToolName: &config.ToolNameConfig{
                APIKey:      apiKey,
                CacheEnable: true,
                MaxDaily:    100,
            },
        },
    }

    err := NewToolNameTool(cfg)
    if err != nil {
        t.Fatalf("Failed to create ToolName tool: %v", err)
    }

    tool := GetTool(bag.ToolName)
    if tool == nil {
        t.Fatal("ToolName tool not registered")
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    args := `{"parameter1": "test"}`
    result, err := tool.Run(ctx, args)
    if err != nil {
        t.Fatalf("Tool run failed: %v", err)
    }

    if result == "" {
        t.Error("Empty result returned")
    }

    t.Logf("Result: %s", result)
}
```

## **Step 7: Create Documentation**

Create `internal/tools/<toolname>/README.md`:

````markdown
# ToolName Tool

Description of what the tool does and its purpose in the Mosychlos ecosystem.

## Configuration

```yaml
tools:
  toolname:
    api_key: 'your-api-key'
    base_url: 'https://api.example.com' # Optional
    cache_enable: true
    max_daily: 1000
```
````

## Environment Variables

- `TOOLNAME_API_KEY`: API key for the service

## Usage Examples

The tool is automatically registered and available through the AI system when configured.

### Function Call Format

```json
{
  "name": "tool_name_function",
  "arguments": {
    "parameter1": "required-value",
    "parameter2": ["optional", "array"]
  }
}
```

### Response Format

```json
{
  "status": "success",
  "data": {
    "result": "processed data"
  }
}
```

## API Integration

- **Provider**: ExampleAPI
- **Rate Limits**: 1000 requests/day
- **Caching**: 24 hour TTL when enabled
- **Timeout**: 30 seconds

## Error Handling

The tool handles the following error scenarios:

- Missing required parameters
- API authentication failures
- Network timeouts
- Rate limiting
- Invalid responses

## Testing

```bash
# Unit tests
go test ./internal/tools/toolname/ -v

# Integration tests (requires API key)
export TOOLNAME_API_KEY=your-key
go test ./internal/tools/ -run TestToolNameTool_Integration -v
```

## **Step 8: Configuration Examples**

Add example configuration to `config/config.default.yaml`:

```yaml
tools:
  # ... existing tools

  toolname:
    api_key: '${TOOLNAME_API_KEY}'
    base_url: 'https://api.example.com'
    cache_enable: true
    max_daily: 1000
```

## **Best Practices**

### **Error Handling**

- Always validate input parameters
- Provide descriptive error messages
- Log errors with appropriate context
- Handle API failures gracefully

### **Performance**

- Use context for cancellation
- Set reasonable timeouts (30s default)
- Enable caching for expensive operations
- Implement rate limiting awareness

### **Security**

- Never log API keys or sensitive data
- Use environment variables for secrets
- Validate all external data
- Sanitize user inputs

### **Monitoring**

- Use structured logging with slog
- Include tool execution metrics
- Track error rates and performance
- Monitor API usage quotas

### **Testing**

- Write comprehensive unit tests
- Include integration tests for API calls
- Test error scenarios thoroughly
- Use table-driven test patterns

## **Integration Checklist**

Before submitting your tool:

- [ ] Tool implements `models.Tool` interface correctly
- [ ] Configuration added to `config.go`
- [ ] Tool key added to `bag.go`
- [ ] Tool registered in `tools.go`
- [ ] Comprehensive tests written
- [ ] Documentation created
- [ ] Example configuration provided
- [ ] Caching and monitoring enabled
- [ ] Error handling implemented
- [ ] Security considerations addressed

## **Common Patterns**

### **HTTP Client Configuration**

```go
http: &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        10,
        MaxIdleConnsPerHost: 5,
        IdleConnTimeout:     30 * time.Second,
    },
},
```

### **Parameter Validation**

```go
if params.RequiredField == "" {
    return "", fmt.Errorf("required_field is required")
}

if len(params.ArrayField) > maxItems {
    return "", fmt.Errorf("array_field cannot exceed %d items", maxItems)
}
```

### **Response Formatting**

```go
response := map[string]any{
    "status": "success",
    "data": result,
    "metadata": map[string]any{
        "timestamp": time.Now().UTC(),
        "source": "toolname",
    },
}
```

Follow these patterns to ensure your tool integrates seamlessly with the Mosychlos ecosystem and provides reliable, monitored, and cached functionality.

---
