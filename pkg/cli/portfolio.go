package cli

import (
	"fmt"

	"github.com/manifoldco/promptui"
)

// DisplayMode represents the different ways to display portfolio data
type DisplayMode string

const (
	SummaryView      DisplayMode = "Summary View"
	DetailedHoldings DisplayMode = "Detailed Holdings"
	ByAccount        DisplayMode = "By Account"
	ComplianceCheck  DisplayMode = "Compliance Check"
	AIAnalysis       DisplayMode = "AI Analysis"
)

// SelectDisplayMode prompts the user to select how they want to view portfolio data
func SelectDisplayMode() (string, error) {
	options := []struct {
		Display string
		Value   string
	}{
		{"Summary View", "summary"},
		{"Detailed Holdings", "detailed"},
		{"By Account", "by-account"},
		{"Compliance Check", "compliance"},
		{"AI Analysis", "ai-analysis"},
	}

	// Convert to display strings for prompt
	items := make([]string, len(options))
	for i, option := range options {
		items[i] = option.Display
	}

	prompt := promptui.Select{
		Label: "Display mode",
		Items: items,
		Size:  len(items),
	}

	index, _, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("display mode selection failed: %w", err)
	}

	return options[index].Value, nil
}
