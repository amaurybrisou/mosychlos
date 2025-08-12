package base

import (
	"testing"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/amaurybrisou/mosychlos/pkg/models/mocks"
	"github.com/golang/mock/gomock"
)

func TestNewBatchEngine(t *testing.T) {
	cases := []struct {
		name        string
		engineName  string
		model       config.LLMModel
		constraints models.BatchToolConstraints
	}{
		{
			name:       "valid creation",
			engineName: "test-engine",
			model:      config.LLMModelGPT4o,
			constraints: models.BatchToolConstraints{
				Tools: []models.ToolDef{},
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			hooks := mocks.NewMockBatchEngineHooks(ctrl)
			hooks.EXPECT().ResultKey().Return(keys.KRiskAnalysisResult).AnyTimes()

			engine := NewBatchEngine(c.engineName, c.model, c.constraints, hooks)

			if engine == nil {
				t.Fatal("expected engine but got nil")
			}

			if engine.Name() != c.engineName {
				t.Errorf("Name() = %v, want %v", engine.Name(), c.engineName)
			}

			if engine.ResultKey() != keys.KRiskAnalysisResult {
				t.Errorf("ResultKey() = %v, want %v", engine.ResultKey(), keys.KRiskAnalysisResult)
			}
		})
	}
}

func TestBatchEngine_Interface(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	hooks := mocks.NewMockBatchEngineHooks(ctrl)

	hooks.EXPECT().ResultKey().Return(keys.KRiskAnalysisResult).AnyTimes()

	engine := NewBatchEngine("test", config.LLMModelGPT4o, models.BatchToolConstraints{}, hooks)

	// Test that it implements models.Engine interface
	var _ models.Engine = engine

	if engine.Name() == "" {
		t.Error("Name() returned empty string")
	}

	if engine.ResultKey() == "" {
		t.Error("ResultKey() returned empty key")
	}
}
