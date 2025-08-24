package budget

import (
	"testing"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

func TestNewToolConsumer(t *testing.T) {
	cases := []struct {
		name        string
		constraints *models.BaseToolConstraints
		expectNil   bool
	}{
		{
			name: "valid constraints",
			constraints: &models.BaseToolConstraints{
				MaxCallsPerTool: map[bag.Key]int{
					bag.Fred: 2,
					bag.FMP:  3,
				},
			},
			expectNil: false,
		},
		{
			name:        "nil constraints",
			constraints: nil,
			expectNil:   false, // should create with empty constraints
		},
		{
			name:        "empty constraints",
			constraints: &models.BaseToolConstraints{},
			expectNil:   false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			consumer := NewToolConsumer(c.constraints)

			if c.expectNil && consumer != nil {
				t.Error("expected nil consumer")
			}
			if !c.expectNil && consumer == nil {
				t.Error("expected non-nil consumer")
			}

			// consumer already implements models.ToolConsumer interface
		})
	}
}

func TestHasCreditsFor(t *testing.T) {
	constraints := &models.BaseToolConstraints{
		MaxCallsPerTool: map[bag.Key]int{
			bag.Fred: 2,
			bag.FMP:  1,
		},
	}

	consumer := NewToolConsumer(constraints).(*defaultToolConsumer)

	cases := []struct {
		name         string
		toolKey      bag.Key
		expectCredit bool
	}{
		{
			name:         "tool with limit and no usage",
			toolKey:      bag.Fred,
			expectCredit: true,
		},
		{
			name:         "tool with no limit",
			toolKey:      bag.NewsAPI, // not in constraints
			expectCredit: true,
		},
		{
			name:         "tool with limit at boundary",
			toolKey:      bag.FMP,
			expectCredit: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			hasCredit := consumer.HasCreditsFor(c.toolKey)
			if hasCredit != c.expectCredit {
				t.Errorf("expected %v, got %v", c.expectCredit, hasCredit)
			}
		})
	}
}

func TestIncrementCallCount(t *testing.T) {
	constraints := &models.BaseToolConstraints{
		MaxCallsPerTool: map[bag.Key]int{
			bag.Fred: 2,
		},
	}

	consumer := NewToolConsumer(constraints).(*defaultToolConsumer)
	toolKey := bag.Fred

	// initially should have credit
	if !consumer.HasCreditsFor(toolKey) {
		t.Error("should have credit initially")
	}

	// after first call
	consumer.IncrementCallCount(toolKey)
	if !consumer.HasCreditsFor(toolKey) {
		t.Error("should still have credit after first call")
	}
	if consumer.GetCallCount(toolKey) != 1 {
		t.Errorf("expected call count 1, got %d", consumer.GetCallCount(toolKey))
	}

	// after second call (at limit)
	consumer.IncrementCallCount(toolKey)
	if consumer.HasCreditsFor(toolKey) {
		t.Error("should not have credit after reaching limit")
	}
	if consumer.GetCallCount(toolKey) != 2 {
		t.Errorf("expected call count 2, got %d", consumer.GetCallCount(toolKey))
	}

	// after third call (over limit)
	consumer.IncrementCallCount(toolKey)
	if consumer.HasCreditsFor(toolKey) {
		t.Error("should not have credit after exceeding limit")
	}
	if consumer.GetCallCount(toolKey) != 3 {
		t.Errorf("expected call count 3, got %d", consumer.GetCallCount(toolKey))
	}
}

func TestGetRemainingCredits(t *testing.T) {
	constraints := &models.BaseToolConstraints{
		MaxCallsPerTool: map[bag.Key]int{
			bag.Fred: 3,
			bag.FMP:  2,
		},
	}

	consumer := NewToolConsumer(constraints).(*defaultToolConsumer)

	// initial state
	remaining := consumer.GetRemainingCredits()
	if remaining[bag.Fred] != 3 {
		t.Errorf("expected 3 remaining for risk assessment, got %d", remaining[bag.Fred])
	}
	if remaining[bag.FMP] != 2 {
		t.Errorf("expected 2 remaining for FMP, got %d", remaining[bag.FMP])
	}

	// after using some credits
	consumer.IncrementCallCount(bag.Fred)
	consumer.IncrementCallCount(bag.FMP)
	consumer.IncrementCallCount(bag.FMP)

	remaining = consumer.GetRemainingCredits()
	if remaining[bag.Fred] != 2 {
		t.Errorf("expected 2 remaining for risk assessment, got %d", remaining[bag.Fred])
	}
	if remaining[bag.FMP] != 0 {
		t.Errorf("expected 0 remaining for FMP, got %d", remaining[bag.FMP])
	}

	// after exceeding limits
	consumer.IncrementCallCount(bag.FMP) // over limit
	remaining = consumer.GetRemainingCredits()
	if remaining[bag.FMP] != 0 {
		t.Errorf("expected 0 remaining for FMP (not negative), got %d", remaining[bag.FMP])
	}
}

func TestReset(t *testing.T) {
	constraints := &models.BaseToolConstraints{
		MaxCallsPerTool: map[bag.Key]int{
			bag.Fred: 2,
		},
	}

	consumer := NewToolConsumer(constraints).(*defaultToolConsumer)
	toolKey := bag.Fred

	// use up credits
	consumer.IncrementCallCount(toolKey)
	consumer.IncrementCallCount(toolKey)
	if consumer.HasCreditsFor(toolKey) {
		t.Error("should not have credit after using all")
	}

	// reset
	consumer.Reset()
	if !consumer.HasCreditsFor(toolKey) {
		t.Error("should have credit after reset")
	}
	if consumer.GetCallCount(toolKey) != 0 {
		t.Errorf("expected call count 0 after reset, got %d", consumer.GetCallCount(toolKey))
	}
}

func TestGetConstraints(t *testing.T) {
	constraints := &models.BaseToolConstraints{
		MaxCallsPerTool: map[bag.Key]int{
			bag.Fred: 2,
		},
	}

	consumer := NewToolConsumer(constraints).(*defaultToolConsumer)
	returnedConstraints := consumer.GetConstraints()

	if returnedConstraints != constraints {
		t.Error("expected same constraints object")
	}
}
