---
applyTo: '**/*.go'
---

# Run Instructions

```bash
docker compose --env-file .env up

go run main.go
```

# Usage

```An interactive command-line interface for managing and analyzing your portfolio.

Usage:
  mosychlos [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  portfolio   Display portfolio information

Flags:
  -h, --help   help for mosychlos

Use "mosychlos [command] --help" for more information about a command.
```

```Interactively display your portfolio with various view options.

Usage:
  mosychlos portfolio [flags]
  mosychlos portfolio [command]

Available Commands:
  analyze     Analyze portfolio with AI insights

Flags:
  -h, --help   help for portfolio

Use "mosychlos portfolio [command] --help" for more information about a command.
```

```Generate AI-powered portfolio analysis. Run without arguments for interactive mode,
or specify analysis type directly: risk, allocation, performance, compliance.

Examples:
  mosychlos portfolio analyze          # Interactive mode - select analysis type
  mosychlos portfolio analyze risk     # Direct risk analysis
  mosychlos portfolio analyze allocation # Direct allocation analysis

Usage:
  mosychlos portfolio analyze [analysis-type] [flags]

Flags:
      --all-formats   Generate reports in all formats (markdown, PDF, JSON)
  -h, --help          help for analyze
      --json          Generate JSON reports
      --markdown      Generate markdown reports
      --pdf           Generate PDF reports
      --reports       Generate reports after analysis
  -v, --verbose       Show detailed analysis process including prompts and AI conversation
```
