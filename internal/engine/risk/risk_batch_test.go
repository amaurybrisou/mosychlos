package risk

import (
	"context"
	"fmt"
	"testing"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/amaurybrisou/mosychlos/pkg/models/mocks"
	"github.com/golang/mock/gomock"
)

func TestNewRiskBatchEngine(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPB := mocks.NewMockPromptBuilder(ctrl)

	cases := []struct {
		name        string
		engineName  string
		cfg         config.LLMConfig
		constraints models.BaseToolConstraints
		wantName    string
	}{
		{
			name:       "valid creation with name",
			engineName: "custom-risk-engine",
			cfg: config.LLMConfig{
				Model: config.LLMModelGPT4o,
			},
			constraints: models.BaseToolConstraints{
				Tools: []models.ToolDef{},
			},
			wantName: "custom-risk-engine",
		},
		{
			name:       "valid creation with empty name uses default",
			engineName: "",
			cfg: config.LLMConfig{
				Model: config.LLMModelGPT4oMini,
			},
			constraints: models.BaseToolConstraints{
				Tools: []models.ToolDef{},
			},
			wantName: "risk-batch-engine",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			engine := NewRiskBatchEngine(c.engineName, c.cfg, mockPB, c.constraints)

			if engine == nil {
				t.Fatal("expected engine but got nil")
			}

			if engine.Name() != c.wantName {
				t.Errorf("Name() = %v, want %v", engine.Name(), c.wantName)
			}

			// Test that it implements models.Engine interface
			var _ models.Engine = engine

			// Test that the embedded BatchEngine is accessible
			if engine.BatchEngine == nil {
				t.Error("BatchEngine should not be nil")
			}

			// Test that the result key is correct
			if engine.ResultKey() != bag.KRiskAnalysisResult {
				t.Errorf("ResultKey() = %v, want %v", engine.ResultKey(), bag.KRiskAnalysisResult)
			}
		})
	}
}

func TestRiskBatchEngineHooks_GetInitialPrompt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPB := mocks.NewMockPromptBuilder(ctrl)
	hooks := &RiskBatchEngineHooks{
		promptBuilder: mockPB,
	}

	cases := []struct {
		name           string
		expectedPrompt string
		expectError    bool
		setupMock      func()
	}{
		{
			name:           "successful prompt building",
			expectedPrompt: "comprehensive risk analysis prompt",
			expectError:    false,
			setupMock: func() {
				mockPB.EXPECT().
					BuildPrompt(gomock.Any(), models.AnalysisRisk).
					Return("comprehensive risk analysis prompt", nil).
					Times(1)
			},
		},
		{
			name:        "prompt building fails",
			expectError: true,
			setupMock: func() {
				mockPB.EXPECT().
					BuildPrompt(gomock.Any(), models.AnalysisRisk).
					Return("", fmt.Errorf("prompt building failed")).
					Times(1)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.setupMock()

			prompt, err := hooks.GetInitialPrompt(context.Background())

			if c.expectError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if prompt != c.expectedPrompt {
				t.Errorf("GetInitialPrompt() = %v, want %v", prompt, c.expectedPrompt)
			}
		})
	}
}

func TestRiskBatchEngineHooks_GenerateCustomID(t *testing.T) {
	hooks := &RiskBatchEngineHooks{}

	cases := []struct {
		name      string
		iteration int
		jobIndex  int
		want      string
	}{
		{
			name:      "initial iteration with zero index",
			iteration: 0,
			jobIndex:  0,
			want:      "task0",
		},
		{
			name:      "first iteration with zero index",
			iteration: 1,
			jobIndex:  0,
			want:      "task_1_0",
		},
		{
			name:      "multiple iterations and jobs",
			iteration: 3,
			jobIndex:  5,
			want:      "task_3_5",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := hooks.GenerateCustomID(c.iteration, c.jobIndex)
			if got != c.want {
				t.Errorf("GenerateCustomID(%d, %d) = %v, want %v", c.iteration, c.jobIndex, got, c.want)
			}
		})
	}
}

func TestRiskBatchEngineHooks_IterationHooks(t *testing.T) {
	hooks := &RiskBatchEngineHooks{}

	t.Run("PreIteration", func(t *testing.T) {
		jobs := []models.BatchJob{
			{CustomID: "task0"},
			{CustomID: "task1"},
		}

		err := hooks.PreIteration(1, jobs)
		if err != nil {
			t.Errorf("PreIteration() error = %v, want nil", err)
		}
	})

	t.Run("PostIteration", func(t *testing.T) {
		results := &models.BatchResult{
			JobID:     "test-job",
			Successes: 2,
			Failures:  0,
		}

		err := hooks.PostIteration(1, results)
		if err != nil {
			t.Errorf("PostIteration() error = %v, want nil", err)
		}
	})
}

func TestRiskBatchEngineHooks_ShouldContinueIteration(t *testing.T) {
	hooks := &RiskBatchEngineHooks{}

	cases := []struct {
		name      string
		iteration int
		jobsCount int
		want      bool
	}{
		{
			name:      "continue with jobs under limit",
			iteration: 5,
			jobsCount: 3,
			want:      true,
		},
		{
			name:      "stop at max iterations",
			iteration: 20,
			jobsCount: 3,
			want:      false,
		},
		{
			name:      "stop with no jobs",
			iteration: 5,
			jobsCount: 0,
			want:      false,
		},
		{
			name:      "stop over max iterations",
			iteration: 25,
			jobsCount: 3,
			want:      false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Create dummy jobs
			jobs := make([]models.BatchJob, c.jobsCount)
			for i := range jobs {
				jobs[i] = models.BatchJob{CustomID: fmt.Sprintf("task%d", i)}
			}

			got := hooks.ShouldContinueIteration(c.iteration, jobs)
			if got != c.want {
				t.Errorf("ShouldContinueIteration(%d, %d jobs) = %v, want %v", c.iteration, c.jobsCount, got, c.want)
			}
		})
	}
}

func TestRiskBatchEngineHooks_ProcessResults(t *testing.T) {
	hooks := &RiskBatchEngineHooks{}
	sharedBag := bag.NewSharedBag()

	t.Run("ProcessToolResult", func(t *testing.T) {
		customID := "task0"
		toolName := "fmp_tool"
		result := "financial data result"

		err := hooks.ProcessToolResult(customID, toolName, result, sharedBag)
		if err != nil {
			t.Errorf("ProcessToolResult() error = %v, want nil", err)
		}

		// Verify result was stored in shared bag
		storedResults := sharedBag.MustGet(bag.KRiskAnalysisResult)
		resultMap, ok := storedResults.(map[string]any)
		if !ok {
			t.Fatal("expected map[string]any in shared bag")
		}

		expectedKey := fmt.Sprintf("%s_tool_%s", customID, toolName)
		if resultMap[expectedKey] != result {
			t.Errorf("expected %s to be stored, got %v", result, resultMap[expectedKey])
		}
	})

	t.Run("ProcessFinalResult", func(t *testing.T) {
		customID := "task1"
		content := "final risk analysis result"

		err := hooks.ProcessFinalResult(customID, content, sharedBag)
		if err != nil {
			t.Errorf("ProcessFinalResult() error = %v, want nil", err)
		}

		// Verify result was stored in shared bag
		storedResults := sharedBag.MustGet(bag.KRiskAnalysisResult)
		resultMap, ok := storedResults.(map[string]any)
		if !ok {
			t.Fatal("expected map[string]any in shared bag")
		}

		if resultMap["result"] != content {
			t.Errorf("expected %s to be stored, got %v", content, resultMap["result"])
		}
	})
}

func TestRiskBatchEngineHooks_ResultKey(t *testing.T) {
	hooks := &RiskBatchEngineHooks{}

	key := hooks.ResultKey()
	if key != bag.KRiskAnalysisResult {
		t.Errorf("ResultKey() = %v, want %v", key, bag.KRiskAnalysisResult)
	}
}

func TestRiskBatchEngineHooks_InterfaceCompliance(t *testing.T) {
	hooks := &RiskBatchEngineHooks{}

	// Test interface compliance
	var _ models.BatchEngineHooks = hooks
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
