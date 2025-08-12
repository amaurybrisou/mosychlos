package jurisdiction

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

//go:embed policy.json
var defaultPolicySchema []byte

// service provides jurisdiction compliance functionality
type service struct {
	schema *jsonschema.Schema
	rules  models.ComplianceRules
}

// New creates a new jurisdiction service with configuration
func New(cfg config.JurisdictionConfig) (*service, error) {
	var schemaData []byte

	// load custom schema if specified, otherwise use default
	if cfg.CustomSchemaPath != "" {
		customSchema, err := os.ReadFile(cfg.CustomSchemaPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read custom schema from %s: %w", cfg.CustomSchemaPath, err)
		}
		schemaData = customSchema
	} else {
		schemaData = defaultPolicySchema
	}

	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2019

	// parse and compile the schema
	if err := compiler.AddResource("embedded://policy.json", strings.NewReader(string(schemaData))); err != nil {
		return nil, fmt.Errorf("failed to add policy schema: %w", err)
	}

	schema, err := compiler.Compile("embedded://policy.json")
	if err != nil {
		return nil, fmt.Errorf("failed to compile policy schema: %w", err)
	}

	// use rules from config - they have already been validated by config.Validate()
	rules := cfg.Rules

	// apply default rules if config rules are empty
	if len(rules.AllowedAssetTypes) == 0 && len(rules.DisallowedAssetTypes) == 0 {
		rules = models.ComplianceRules{
			AllowedAssetTypes: []string{"stock", "etf", "bond_gov", "cash"},
			MaxLeverage:       1, // no leverage
		}
	}

	return &service{
		schema: schema,
		rules:  rules,
	}, nil
}

// Validate validates a portfolio against jurisdiction rules
func (s *service) Validate(ctx context.Context, portfolio *models.Portfolio) error {
	if portfolio == nil {
		return fmt.Errorf("portfolio cannot be nil")
	}

	// apply compliance rules to portfolio
	result, violations := Apply(*portfolio, s.rules)

	// check if there are any violations
	if len(violations) > 0 {
		var violationList []string
		for asset, isViolation := range violations {
			if isViolation {
				violationList = append(violationList, asset)
			}
		}
		return fmt.Errorf("jurisdiction compliance violations for assets: %v", violationList)
	}

	// check if there are compliance notes (warnings/issues)
	if len(result.Notes) > 0 {
		var noteMessages []string
		for _, note := range result.Notes {
			noteMessages = append(noteMessages, fmt.Sprintf("%s: %s", note.Ticker, note.Note))
		}
		return fmt.Errorf("jurisdiction compliance issues: %v", noteMessages)
	}

	return nil
}
