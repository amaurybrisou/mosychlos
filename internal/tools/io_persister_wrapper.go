package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	pkgfs "github.com/amaurybrisou/mosychlos/pkg/fs"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

type IOPersister struct {
	tool models.Tool
	fs   pkgfs.FS
}

func NewIOPersistingTool(tool models.Tool, dataDir string) models.Tool {
	dataDir = filepath.Join(dataDir, "tools_input_output")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create cache directory: dir=%s err=%v", dataDir, err))
	}

	return &IOPersister{
		tool: tool,
		fs:   pkgfs.New(dataDir),
	}
}

func (p *IOPersister) Name() string {
	return p.tool.Name()
}

func (p *IOPersister) Key() bag.Key {
	return p.tool.Key()
}

func (p *IOPersister) Description() string {
	return p.tool.Description()
}

func (p *IOPersister) Definition() models.ToolDef {
	return p.tool.Definition()
}

func (p *IOPersister) Tags() []string {
	return p.tool.Tags()
}

func (p *IOPersister) IsExternal() bool {
	return p.tool.IsExternal()
}

func (p *IOPersister) Run(ctx context.Context, args any) (any, error) {
	result, err := p.tool.Run(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("tool %s invocation error: %w", p.tool.Name(), err)
	}

	if err := p.persistInput(args); err != nil {
		return nil, fmt.Errorf("tool %s failed to persist input: %w", p.tool.Name(), err)
	}

	if err := p.persistOutput(result); err != nil {
		return nil, fmt.Errorf("tool %s failed to persist output: %w", p.tool.Name(), err)
	}

	return result, nil
}

func (p *IOPersister) persistInput(input any) error {
	fileName := fmt.Sprintf("%s_input.json", p.tool.Name())

	err := p.fs.MkdirAll(filepath.Dir(fileName), 0755)
	if err != nil {
		return fmt.Errorf("failed to create input directory: %w", err)
	}

	return p.fs.WriteFile(fileName, []byte(fmt.Sprintf("%v", input)), 0644)
}

func (p *IOPersister) persistOutput(output any) error {
	filePath := fmt.Sprintf("%s_output.json", p.tool.Name())

	err := p.fs.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	var data []byte

	switch v := output.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		var err error
		data, err = json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal output: %w", err)
		}
	}

	return p.fs.WriteFile(filePath, data, 0644)
}
