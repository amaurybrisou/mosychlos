package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/nlpodyssey/openai-agents-go/agents"
)

// Usage: OPENAI_API_KEY=<your-key> go run main.go

// Tool params type
type GetWeatherParams struct {
	City string `json:"city"`
}

// Tool implementation
func getWeather(_ context.Context, params GetWeatherParams) (string, error) {
	return fmt.Sprintf("The weather in %s is sunny.", params.City), nil
}

// Tool registration (using SDK's NewFunctionTool)
var getWeatherTool = agents.NewFunctionTool("GetWeather", "", getWeather)

func main() {

	type output struct {
		Name string `json:"name"`
		Temp string `json:"temp"`
	}
	r := runner[output]{
		ag: agents.New("hello").
			WithInstructions("What's the weather in Tokyo?").
			WithModel("gpt-5-mini").
			WithTools(getWeatherTool).
			WithOutputType(agents.OutputType[output]()),
	}

	result, err := r.Run(context.Background(), "What's the weather in Tokyo?")

	// result, err := agents.Run(context.Background(), agent, "What's the weather in Tokyo?")
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
	// The weather in Tokyo is sunny.
}

type runner[OutputType any] struct {
	ag *agents.Agent
}

func (r *runner[OutputType]) Run(ctx context.Context, input string) (*OutputType, error) {
	result, err := agents.Run(ctx, r.ag, input)
	if err != nil {
		slog.Error("agent run error", "error", err)
		return nil, err
	}

	bb, err := json.Marshal(result.FinalOutput)
	if err != nil {
		slog.Error("failed to marshal final output", "error", err)
		return nil, err
	}

	var output *OutputType
	if err == nil {
		slog.Debug("marshalled final output", "output", string(bb))
		err = json.Unmarshal(bb, &output)
	}
	return output, err
}
