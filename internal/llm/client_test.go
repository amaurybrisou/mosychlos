package llm

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/internal/llm/mocks"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	bagmocks "github.com/amaurybrisou/mosychlos/pkg/bag/mocks"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	modelsmocks "github.com/amaurybrisou/mosychlos/pkg/models/mocks"
)

func TestClient_RegisterTool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTool := modelsmocks.NewMockTool(ctrl)
	toolKey := bag.Key("test-tool")
	mockTool.EXPECT().Key().Return(toolKey).AnyTimes()

	// Create a mock responses strategy
	mockResponsesStrat := mocks.NewMockResponsesStrategyInterface(ctrl)
	mockResponsesStrat.EXPECT().RegisterTool(mockTool).Times(1)

	client := &Client{
		toolRegistry:   make(map[bag.Key]models.Tool),
		responsesStrat: mockResponsesStrat,
	}

	client.RegisterTool(mockTool)

	assert.Equal(t, mockTool, client.toolRegistry[toolKey])
}

func TestClient_SetToolConsumer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConsumer := modelsmocks.NewMockToolConsumer(ctrl)
	mockResponsesStrat := mocks.NewMockResponsesStrategyInterface(ctrl)
	mockResponsesStrat.EXPECT().SetToolConsumer(mockConsumer).Times(1)

	client := &Client{
		responsesStrat: mockResponsesStrat,
	}
	client.SetToolConsumer(mockConsumer)

	assert.Equal(t, mockConsumer, client.consumer)
}

func TestClient_DoSync(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("nil responses strategy", func(t *testing.T) {
		client := &Client{
			responsesStrat: nil,
		}

		response, err := client.DoSync(context.Background(), models.PromptRequest{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider not set")
		assert.Nil(t, response)
	})

	t.Run("successful sync call", func(t *testing.T) {
		mockResponsesStrat := mocks.NewMockResponsesStrategyInterface(ctrl)
		expectedReq := models.PromptRequest{
			Messages: []map[string]any{
				{"role": "user", "content": "Hello"},
			},
		}
		expectedResp := &models.LLMResponse{
			Content: "Hi there",
		}

		mockResponsesStrat.EXPECT().Ask(gomock.Any(), expectedReq).Return(expectedResp, nil).Times(1)

		client := &Client{
			responsesStrat: mockResponsesStrat,
		}

		response, err := client.DoSync(context.Background(), expectedReq)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "Hi there", response.Content)
	})
}

// Skip NewLLMClient test for now as it requires complex mocking
func TestNewLLMClient_Simple(t *testing.T) {
	mockBag := bagmocks.NewMockSharedBag(gomock.NewController(t))
	mockBag.EXPECT().Snapshot().Return(nil).AnyTimes()

	// Test missing provider error
	cfg := &config.Config{
		LLM: config.LLMConfig{
			Provider: "",
		},
	}

	client, err := NewLLMClient(cfg, mockBag)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LLM provider not configured")
	assert.Nil(t, client)
}
