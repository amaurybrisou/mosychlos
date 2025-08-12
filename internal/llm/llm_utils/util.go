// internal/llm/util.go
package llmutils

import (
	"strings"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// DetectModelClass determines the model class based on model name
// Invariant: reasoning models must NOT set unsupported params like temperature/tool_choice.
func DetectModelClass(model string) models.ModelClass {
	if strings.HasPrefix(model, "gpt-5") || strings.HasPrefix(model, "o1-") {
		return models.ModelClassReasoning
	}
	return models.ModelClassStandard
}

// IsReasoningModel is a convenience function to check if a model is reasoning class
func IsReasoningModel(model string) bool {
	return DetectModelClass(model) == models.ModelClassReasoning
}
