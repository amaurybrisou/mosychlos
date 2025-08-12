package cli

import (
	"fmt"
	"strings"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// SelectAnalysisType prompts the user to select an analysis type
func SelectAnalysisType() (models.AnalysisType, error) {
	options := []string{
		"risk - Portfolio risk assessment and concentration analysis",
		"investment_research - In-depth analysis of investment opportunities",
	}

	fmt.Println("\n📊 Select Analysis Type:")
	for i, option := range options {
		fmt.Printf("  %d. %s\n", i+1, option)
	}

	var choice int
	fmt.Print("\nEnter your choice (1-5): ")
	if _, err := fmt.Scanf("%d", &choice); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	if choice < 1 || choice > len(options) {
		return "", fmt.Errorf("invalid choice: must be between 1 and %d", len(options))
	}

	// Extract the analysis type from the option (before the " - " separator)
	analysisType := strings.Split(options[choice-1], " - ")[0]
	return models.AnalysisType(analysisType), nil
}
