package llmutils

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/amaurybrisou/mosychlos/pkg/models"
)

func TestDetectModelClass(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		expected models.ModelClass
	}{
		{
			name:     "gpt-5 model",
			model:    "gpt-5",
			expected: models.ModelClassReasoning,
		},
		{
			name:     "gpt-5-mini model",
			model:    "gpt-5-mini",
			expected: models.ModelClassReasoning,
		},
		{
			name:     "o1-preview model",
			model:    "o1-preview",
			expected: models.ModelClassReasoning,
		},
		{
			name:     "o1-mini model",
			model:    "o1-mini",
			expected: models.ModelClassReasoning,
		},
		{
			name:     "gpt-4o model",
			model:    "gpt-4o",
			expected: models.ModelClassStandard,
		},
		{
			name:     "gpt-4o-mini model",
			model:    "gpt-4o-mini",
			expected: models.ModelClassStandard,
		},
		{
			name:     "claude model",
			model:    "claude-3-5-sonnet",
			expected: models.ModelClassStandard,
		},
		{
			name:     "empty model",
			model:    "",
			expected: models.ModelClassStandard,
		},
		{
			name:     "random model",
			model:    "random-model-123",
			expected: models.ModelClassStandard,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectModelClass(tt.model)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsReasoningModel(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		expected bool
	}{
		{
			name:     "gpt-5 is reasoning",
			model:    "gpt-5",
			expected: true,
		},
		{
			name:     "gpt-5-turbo is reasoning",
			model:    "gpt-5-turbo",
			expected: true,
		},
		{
			name:     "o1-preview is reasoning",
			model:    "o1-preview",
			expected: true,
		},
		{
			name:     "o1-mini is reasoning",
			model:    "o1-mini",
			expected: true,
		},
		{
			name:     "gpt-4o is not reasoning",
			model:    "gpt-4o",
			expected: false,
		},
		{
			name:     "gpt-4o-mini is not reasoning",
			model:    "gpt-4o-mini",
			expected: false,
		},
		{
			name:     "claude is not reasoning",
			model:    "claude-3-5-sonnet",
			expected: false,
		},
		{
			name:     "empty model is not reasoning",
			model:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsReasoningModel(tt.model)
			assert.Equal(t, tt.expected, result)
		})
	}
}
