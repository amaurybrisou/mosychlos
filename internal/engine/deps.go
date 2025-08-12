// internal/engine/deps.go
package engine

import (
	"context"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/internal/llm"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/fs"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// Deps is the dependency bundle passed to engine builders/constructors.
type Deps struct {
	Ctx       context.Context
	Config    *config.Config
	SharedBag bag.SharedBag
	FS        fs.FS

	// Core services built by the orchestrator:
	AI      *llm.Client
	Prompts models.PromptBuilder // if some engines need prompt building
	Tools   []models.Tool
	// Add other shared services here if needed (portfolio svc, news svc, etc.)
}
