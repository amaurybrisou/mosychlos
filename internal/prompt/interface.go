package prompt

import (
	"context"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// RegionalManager extends template management with regional capabilities
type RegionalManager interface {
	// GenerateRegionalPrompt creates a regionally-aware prompt for the specified analysis type
	GenerateRegionalPrompt(
		ctx context.Context,
		analysisType models.AnalysisType,
		bag bag.SharedBag,
		data PromptData,
	) (string, error)
}
