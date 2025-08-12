# LLM Batch Processing System

## Overview

The LLM Batch Processing system enables asynchronous AI analysis of portfolios and investment data using OpenAI's Batch API. This system provides significant cost savings (up to 50%) compared to synchronous API calls, making it ideal for non-time-critical batch analyses.

## Architecture

### Core Components

- **pkg/models/ai_batch.go**: Core contracts and interfaces for batch processing
- **internal/llm/util.go**: Model classification and utility functions
- **internal/llm/openai/**: OpenAI-specific batch processing implementation
- **internal/llm/batch/**: Batch management, result aggregation, and CLI integration
- **internal/llm/factory.go**: Factory for creating batch services
- **cmd/mosychlos/batch.go**: CLI commands for batch operations

### Key Features

1. **Model Class Detection**: Automatic detection of standard vs reasoning models (GPT-5 class models)
2. **Cost Optimization**: Batch processing with 50% cost savings compared to sync calls
3. **JSONL Format**: Proper formatting for OpenAI Batch API compliance
4. **Result Aggregation**: Processing of batch results with error handling
5. **CLI Integration**: Full command-line interface for batch job management

## Usage

### CLI Commands

#### Submit a Batch Job

```bash
# Submit portfolio analysis batch job
mosychlos batch submit risk portfolio.json

# Submit multiple portfolios with multiple analysis types
mosychlos batch submit --types=risk,allocation *.json

# Submit and wait for completion
mosychlos batch submit --wait --timeout=30m performance portfolio.json
```

#### Check Job Status

```bash
# Check status of a batch job
mosychlos batch status batch_1234567890

# Wait for job completion
mosychlos batch wait batch_1234567890
```

#### Retrieve Results

```bash
# Get results from completed job
mosychlos batch results batch_1234567890
```

#### Cancel Job

```bash
# Cancel running job
mosychlos batch cancel batch_1234567890
```

### Programmatic Usage

```go
// Create batch service
factory := llm.NewBatchServiceFactory(cfg)
manager, err := factory.CreateManager()
if err != nil {
    return err
}

// Submit batch job
requests := []models.BatchRequest{
    {
        CustomID: "analysis_1",
        Method:   "POST",
        URL:      "/v1/chat/completions",
        Body: map[string]any{
            "model": "gpt-4o-mini",
            "messages": []map[string]any{
                {"role": "system", "content": "You are a financial analyst"},
                {"role": "user", "content": "Analyze this portfolio..."},
            },
        },
    },
}

opts := models.BatchOptions{
    CompletionWindow: "24h",
    CostOptimize:     true,
}

job, err := manager.ProcessBatch(ctx, requests, opts, false)
if err != nil {
    return err
}

fmt.Printf("Batch job submitted: %s\n", job.ID)
```

## Analysis Types

The system supports the following portfolio analysis types:

- **risk**: Risk assessment and analysis
- **allocation**: Asset allocation optimization
- **performance**: Performance evaluation
- **compliance**: Regulatory compliance checking
- **reallocation**: Portfolio rebalancing recommendations
- **investment_research**: In-depth investment research

## Configuration

The batch processing system uses the existing LLM configuration in your `config.yaml`:

```yaml
llm:
  provider: 'openai'
  api_key: '${OPENAI_API_KEY}'
  model: 'gpt-4o-mini'
  openai:
    max_completion_tokens: 2000
    temperature: 0.3
```

## Cost Optimization

### Model Class Detection

The system automatically detects model classes to optimize API parameters:

- **Standard Models**: GPT-4o, GPT-4o-mini (supports temperature, tool_choice)
- **Reasoning Models**: GPT-5 class models (limited parameter support)

### Batch Processing Benefits

- **50% Cost Savings**: Compared to synchronous API calls
- **Efficient Processing**: JSONL format with streaming support
- **Bulk Operations**: Process multiple portfolios simultaneously
- **Background Processing**: Non-blocking job submission

## Data Flow

1. **Job Submission**: Convert requests to JSONL format and submit to OpenAI
2. **Status Monitoring**: Poll job status until completion
3. **Result Retrieval**: Stream results and errors from completed jobs
4. **Aggregation**: Process results into unified data structure
5. **Local Storage**: Save job info and results for future reference

## File Structure

```
internal/llm/
├── factory.go                    # Service factory
├── util.go                      # Utility functions
├── batch/
│   ├── manager.go               # Batch job orchestration
│   ├── result_aggregator.go     # Result processing
│   └── cli_integration.go       # CLI service layer
└── openai/
    ├── batch_client.go          # OpenAI batch API client
    └── batch_formatter.go       # JSONL formatting

pkg/models/
└── ai_batch.go                  # Core contracts and types

cmd/mosychlos/
└── batch.go                     # CLI command definitions
```

## Data Storage

### Job Information

- Location: `{DataDir}/batch/`
- Files: `job_{batch_id}.json`
- Content: Job metadata, status, cost estimates

### Results

- Location: `{DataDir}/batch/`
- Files: `results_{batch_id}.json`
- Content: Aggregated success/failure results

### Working Directory Structure

```
mosychlos-data/batch/
├── job_batch_1234567890.json     # Job metadata
├── results_batch_1234567890.json # Aggregated results
└── ...
```

## Error Handling

The system provides comprehensive error handling:

- **Submission Errors**: Invalid requests, configuration issues
- **Processing Errors**: Individual request failures within batch
- **Network Errors**: Connection issues, API rate limits
- **Timeout Errors**: Jobs exceeding completion windows

## Monitoring

### Job Status States

- `validating`: Job is being validated by OpenAI
- `in_progress`: Job is actively being processed
- `finalizing`: Job processing complete, preparing results
- `completed`: Job finished successfully
- `failed`: Job failed during processing
- `expired`: Job exceeded completion window
- `cancelled`: Job was manually cancelled

### Progress Tracking

Jobs provide detailed progress information:

- Total requests submitted
- Completed requests count
- Failed requests count
- Cost estimates and savings

## Integration Notes

### OpenAI API Integration

Current implementation includes placeholder methods for:

- Actual OpenAI batch API calls
- File upload/download operations
- Real-time status polling

To complete integration:

1. Implement actual OpenAI Batch API methods
2. Add file handling for JSONL uploads
3. Implement result streaming from OpenAI files

### Extensibility

The system is designed for easy extension to support:

- Additional AI providers (Claude, etc.)
- Custom analysis types
- Advanced result processing
- Integration with existing portfolio analysis engines

## Development Status

### Completed

- ✅ Core batch processing contracts
- ✅ Model class detection utilities
- ✅ JSONL formatter for OpenAI
- ✅ Batch manager with job orchestration
- ✅ Result aggregation system
- ✅ CLI integration and commands
- ✅ Factory pattern for service creation
- ✅ Configuration integration

### Pending

- 🔄 Complete OpenAI Batch API integration
- 🔄 File upload/download implementation
- 🔄 Real batch API method calls
- 🔄 Integration with existing prompt system
- 🔄 Error recovery and retry logic

### Future Enhancements

- 📋 Support for additional AI providers
- 📋 Advanced job scheduling and prioritization
- 📋 Result caching and persistence
- 📋 Webhook notifications for job completion
- 📋 Batch job templates and presets
