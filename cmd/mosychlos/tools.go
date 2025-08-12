package mosychlos

import (
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/amaurybrisou/mosychlos/internal/tools"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/config"
)

// toolsCommand handles the tools command
func toolsCommand(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize shared bag for metrics tracking
	sharedBag := bag.NewSharedBag()
	tools.SetSharedBag(sharedBag)

	// Initialize tools
	if err := tools.NewTools(cfg); err != nil {
		slog.Error("Failed to initialize tools", "error", err)
		os.Exit(1)
	}

	// Get verbose flag
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Display tools information
	displayTools(verbose)
}

// displayTools shows all available tools
func displayTools(verbose bool) {
	fmt.Println("📊 Mosychlos Financial Data Tools")
	fmt.Println("=================================")
	fmt.Println()

	// Get all registered tools
	allTools := tools.GetTools()
	if len(allTools) == 0 {
		fmt.Println("❌ No tools are currently registered.")
		fmt.Println("   Check your configuration file and ensure tool API keys are set.")
		return
	}

	fmt.Printf("✅ %d tools are currently available:\n\n", len(allTools))

	// Sort tools by name for consistent display
	sort.Slice(allTools, func(i, j int) bool {
		return allTools[i].Name() < allTools[j].Name()
	})

	if verbose {
		// Detailed view with descriptions
		for _, tool := range allTools {
			fmt.Printf("🔧 %s\n", tool.Name())
			fmt.Printf("   Description: %s\n", tool.Description())
			fmt.Printf("   Key: %s\n", tool.Key())
			fmt.Printf("   Tags: %s\n", strings.Join(tool.Tags(), ", "))
			fmt.Println()
		}
	} else {
		// Compact table view
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "TOOL NAME\tKEY\tTAGS\tDESCRIPTION\n")
		fmt.Fprintf(w, "---------\t---\t----\t-----------\n")

		for _, tool := range allTools {
			tags := strings.Join(tool.Tags(), ", ")
			if len(tags) > 30 {
				tags = tags[:27] + "..."
			}
			description := tool.Description()
			if len(description) > 150 {
				description = description[:147] + "..."
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				tool.Name(),
				tool.Key(),
				tags,
				description,
			)
		}
		w.Flush()
	}

	fmt.Println()
	fmt.Println("💡 Use --verbose flag for detailed information about each tool")
	fmt.Println("📖 See individual tool documentation for usage examples")
}

// CreateToolsCommand creates the tools command
func CreateToolsCommand() *cobra.Command {
	var toolsCmd = &cobra.Command{
		Use:   "tools",
		Short: "Display available financial data tools",
		Long: `Display all available financial data tools in the Mosychlos system.

Tools provide access to various financial data sources including:
- Market data (prices, volumes, indices)
- Company information (fundamentals, news, analysis)
- Economic data (indicators, rates, statistics)
- Alternative data (weather, sentiment, events)

Examples:
  mosychlos tools              # Show compact list of all tools
  mosychlos tools --verbose    # Show detailed information for each tool`,
		Run: toolsCommand,
	}

	// Add flags
	toolsCmd.Flags().BoolP("verbose", "v", false, "Show detailed information for each tool including full descriptions and definitions")

	return toolsCmd
}
