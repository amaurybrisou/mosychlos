// internal/engine/wiring.go
package engine

import (
	"context"
	"fmt"
	"sort"

	"github.com/amaurybrisou/mosychlos/internal/budget"
	"github.com/amaurybrisou/mosychlos/internal/engine/risk"
	"github.com/amaurybrisou/mosychlos/internal/tools"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// Builder builds engines from a dependency bundle.
type Builder interface {
	Build(ctx context.Context, deps Deps) ([]models.Engine, error)
}

// EngineFactory creates a single engine using the provided Deps.
type EngineFactory func(Deps) (models.Engine, error)

// RegistryBuilder is a simple DI container for engines.
// - Register factories under a name.
// - Optionally define an explicit order (otherwise names are sorted).
// - Build constructs all engines using the shared Deps.
type RegistryBuilder struct {
	order     []string
	factories map[string]EngineFactory
}

// NewRegistryBuilder returns an empty registry.
func NewRegistryBuilder() *RegistryBuilder {
	return &RegistryBuilder{
		factories: make(map[string]EngineFactory),
	}
}

// Register adds (or replaces) a named engine factory.
func (r *RegistryBuilder) Register(name string, f EngineFactory) *RegistryBuilder {
	r.factories[name] = f
	return r
}

// WithOrder sets the exact construction order by names previously registered.
// Any registered engines not listed here will be appended after, in alpha order.
func (r *RegistryBuilder) WithOrder(names ...string) *RegistryBuilder {
	r.order = append([]string(nil), names...)
	return r
}

// Build constructs all engines using Deps.
// - Enforces explicit order if provided, then appends the rest alphabetically.
// - Fails fast on the first construction error.
func (r *RegistryBuilder) Build(ctx context.Context, deps Deps) ([]models.Engine, error) {
	if deps.Config == nil || deps.SharedBag == nil || deps.AI == nil {
		return nil, fmt.Errorf("engine deps incomplete: Config/SharedBag/AI are required")
	}

	// Carry the context into deps for factories that care.
	deps.Ctx = ctx

	seen := make(map[string]bool, len(r.factories))
	var names []string

	// 1) explicit order (keep only registered names)
	for _, n := range r.order {
		if _, ok := r.factories[n]; ok && !seen[n] {
			seen[n] = true
			names = append(names, n)
		}
	}
	// 2) remaining, sorted alphabetically
	var rest []string
	for n := range r.factories {
		if !seen[n] {
			rest = append(rest, n)
		}
	}
	sort.Strings(rest)
	names = append(names, rest...)

	// 3) construct
	out := make([]models.Engine, 0, len(names))
	for _, n := range names {
		f := r.factories[n]
		eng, err := f(deps)
		if err != nil {
			return nil, fmt.Errorf("build engine %q: %w", n, err)
		}
		if eng == nil {
			return nil, fmt.Errorf("build engine %q: factory returned nil", n)
		}
		out = append(out, eng)
	}
	return out, nil
}

// Ensure RegistryBuilder implements Builder.
var _ Builder = (*RegistryBuilder)(nil)

// ----- Optional: default registration helpers --------------------------------

// DefaultRegistry returns a pre-populated registry for your common engines.
// Use this as-is or copy/modify in your app wiring.
func DefaultRegistry() *RegistryBuilder {
	return NewRegistryBuilder().
		// Register("risk", func(d Deps) (models.Engine, error) {
		// 	if d.Prompts == nil {
		// 		return nil, fmt.Errorf("risk engine requires Deps.Prompts")
		// 	}
		// 	// Example: per-engine tool constraints (tweak to your needs)
		// 	constraints := models.ToolConstraints{
		// 		Tools:          tools.ToolsToToolDefs(d.Tools),
		// 		PreferredTools: []keys.Key{keys.FMP, keys.NewsApi},
		// 		MinCallsPerTool: map[keys.Key]int{
		// 			keys.NewsApi: 2,
		// 			keys.FMP:     1,
		// 		},
		// 		MaxCallsPerTool: map[keys.Key]int{
		// 			keys.NewsApi: 2,
		// 			keys.FMP:     2,
		// 		},
		// 	}
		// 	// Option 1: set consumer globally here
		// 	d.AI.SetToolConsumer(budget.NewToolConsumer(&constraints))
		// 	// Option 2: or let the engine set one inside Execute()
		// 	d.AI.RegisterTool(d.Tools...)

		// 	return risk.New("risk-engine", d.Prompts, constraints), nil
		// }).
		Register("batch-risk-engine", func(d Deps) (models.Engine, error) {
			if d.Prompts == nil {
				return nil, fmt.Errorf("batch-risk-engine requires Deps.Prompts")
			}

			// Base tools for risk analysis
			preferredTools := []keys.Key{keys.FMP, keys.NewsApi}
			minCalls := map[keys.Key]int{
				keys.NewsApi: 2,
				keys.FMP:     1,
			}
			maxCalls := map[keys.Key]int{
				keys.NewsApi: 2,
				keys.FMP:     2,
			}

			// Add web search if enabled in OpenAI config
			if d.Config.LLM.OpenAI.WebSearch {
				preferredTools = append(preferredTools, keys.WebSearch)
				minCalls[keys.WebSearch] = 1 // At least 1 web search for market context
				maxCalls[keys.WebSearch] = 3 // Up to 3 web searches for comprehensive analysis
			}

			// Example: per-engine tool constraints (tweak to your needs)
			constraints := models.BatchToolConstraints{
				Tools:           tools.ToolsToToolDefs(d.Tools),
				PreferredTools:  preferredTools,
				MinCallsPerTool: minCalls,
				MaxCallsPerTool: maxCalls,
			}
			// Option 1: set consumer globally here
			d.AI.SetToolConsumer(budget.NewToolConsumer(&constraints))
			// Option 2: or let the engine set one inside Execute()
			d.AI.RegisterTool(d.Tools...)

			return risk.NewRiskBatchEngine("batch-risk-engine", d.Config.LLM, d.Prompts, constraints), nil
		})

	// .Register("news", func(d Deps) (models.Engine, error) { ... })
	// .Register("screener", func(d Deps) (models.Engine, error) { ... })
}

// DefaultRegistryWithOrder sets an explicit construction order.
// Engines not listed will be appended afterward in alphabetical order.
func DefaultRegistryWithOrder() *RegistryBuilder {
	return DefaultRegistry().
		WithOrder(
			// "risk",
			"batch-risk-engine",
			// "news",
			// "screener",
		)
}
