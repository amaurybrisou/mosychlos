// internal/llm/factory.go
package llm

import (
	"fmt"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/internal/llm/batch"
	"github.com/amaurybrisou/mosychlos/internal/llm/openai"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/fs"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// BatchServiceFactory creates batch processing services
type BatchServiceFactory struct {
	cfg        *config.Config
	filesystem fs.FS
	sharedBag  bag.SharedBag
}

// NewBatchServiceFactory creates a new factory
func NewBatchServiceFactory(cfg *config.Config, filesystem fs.FS, sharedBag bag.SharedBag) *BatchServiceFactory {
	return &BatchServiceFactory{
		cfg:        cfg,
		filesystem: filesystem,
		sharedBag:  sharedBag,
	}
}

func (f *BatchServiceFactory) CreateManager() (*batch.Manager, error) {
	client, err := f.createBatchClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create batch client: %w", err)
	}
	return batch.NewManager(client), nil
}

// createBatchClient creates the appropriate batch client based on configuration
func (f *BatchServiceFactory) createBatchClient() (models.AiBatchClient, error) {
	// Check which AI provider is configured
	if f.cfg.LLM.Provider == "" {
		return nil, fmt.Errorf("LLM provider not configured")
	}

	// For now, only OpenAI is supported
	if f.cfg.LLM.Provider != "openai" {
		return nil, fmt.Errorf("unsupported AI provider for batch processing: %s", f.cfg.LLM.Provider)
	}

	return openai.NewBatchClient(f.cfg.LLM, f.sharedBag)
}
