package models

//go:generate mockgen -source=prompt.go -destination=mocks/mock_prompt.go -package=mocks

import (
	"context"
)

// PromptBuilder provides prompt building functionality with dependency injection
type PromptBuilder interface {
	// BuildPrompt creates a prompt for the specified analysis type using
	// normalized data from the shared bag and user configuration
	BuildPrompt(ctx context.Context, analysisType AnalysisType) (string, error)
}
