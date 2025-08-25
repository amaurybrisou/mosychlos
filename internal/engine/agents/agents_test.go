package agents

import (
	"testing"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/nlpodyssey/openai-agents-go/agents"
)

func TestBaseAgentEngine_Creation(t *testing.T) {
	config := BaseAgentEngineConfig{
		Name:         "test-agent",
		ResultKey:    bag.Key("test-result"),
		Instructions: "You are a test agent.",
		Model:        "gpt-4o",
		Tools:        []agents.Tool{},
	}

	engine := NewBaseAgentEngine(config)

	if engine.Name() != "test-agent" {
		t.Errorf("Expected name 'test-agent', got %s", engine.Name())
	}

	if engine.ResultKey() != bag.Key("test-result") {
		t.Errorf("Expected result key 'test-result', got %s", engine.ResultKey())
	}

	if engine.GetAgent() == nil {
		t.Error("Expected agent to be created")
	}
}

func TestToolConverter(t *testing.T) {
	converter := NewToolConverter()
	
	if converter == nil {
		t.Error("Expected converter to be created")
	}

	// Test with empty slice
	agentTools := converter.ConvertTools([]models.Tool{})
	if len(agentTools) != 0 {
		t.Errorf("Expected empty slice, got %d tools", len(agentTools))
	}
}