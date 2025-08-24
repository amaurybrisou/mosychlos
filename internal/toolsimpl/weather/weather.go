package weather

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v2"
)

type WeatherTool struct{}

func (WeatherTool) Name() string { return "get_weather" }
func (WeatherTool) Definition() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        "get_weather",
		Description: "Get the current weather in a given city",
		Parameters: openai.FunctionParameters{
			"location": map[string]any{
				"type":        "string",
				"description": "The location to get the weather for",
				"required":    true,
			},
		},
	}
}
func (WeatherTool) Run(ctx context.Context, args string) (string, error) {
	var p struct {
		Location string `json:"location"`
	}
	if err := json.Unmarshal([]byte(args), &p); err != nil {
		return "", err
	}
	return fmt.Sprintf(`{"location":%q,"summary":"sunny","temp_c":25}`, p.Location), nil
}
