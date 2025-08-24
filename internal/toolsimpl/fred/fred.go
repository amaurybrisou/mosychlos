// Package fred
package fred

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/fred"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// FredTool fetches macroeconomic indicators from the FRED API.
// Docs: https://fred.stlouisfed.org/docs/api/fred/
type FredTool struct {
	client    *fred.Client
	sharedBag bag.SharedBag
}

// Ensure Provider implements models.Tool interface
var _ models.Tool = &FredTool{}

// new returns a FRED provider using the given API key.
func new(apiKey string, sharedBag bag.SharedBag) (*FredTool, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("fred: missing API key")
	}

	// Create FRED client
	clientCfg := fred.Config{
		APIKey:  apiKey,
		BaseURL: "https://api.stlouisfed.org/fred",
		Timeout: 30 * time.Second,
	}

	client, err := fred.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create FRED client: %w", err)
	}

	return &FredTool{
		client:    client,
		sharedBag: sharedBag,
	}, nil
}

// models.Tool interface implementation

func (p *FredTool) Name() string {
	return bag.Fred.String()
}

func (p *FredTool) Key() bag.Key {
	return bag.Fred
}

func (p *FredTool) Tags() []string {
	return []string{"economic", "macro", "finance", "data", "federal-reserve"}
}

func (p *FredTool) Description() string {
	return "Fetches macroeconomic indicators from the Federal Reserve Economic Data (FRED) API. Available data includes: Per Capita Personal Income (882), GDP (249), Unemployment Rate (158), Federal Funds Rate (115), and other economic series. Use specific series group IDs, dates, and regions to get accurate economic data. For most reliable results, use series_group=882 with frequency=a (annual) and dates between 2013-01-01 and 2024-01-01."
}

func (p *FredTool) IsExternal() bool {
	return false
}

func (p *FredTool) Definition() models.ToolDef {
	return &models.CustomToolDef{
		Type: models.CustomToolDefType,
		FunctionDef: models.FunctionDef{
			Name:        p.Name(),
			Description: p.Description(),
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"series_group": map[string]any{
						"type":        "string",
						"description": "The series group ID for FRED GeoFRED regional data. ONLY use '882' for Per Capita Personal Income - this is the only guaranteed working series group. Do not use any other values.",
						"enum":        []string{"882"},
						"default":     "882",
					},
					"date": map[string]any{
						"type":        "string",
						"description": "The date you want to pull series group data from (YYYY-MM-DD format). Use dates between 2013-01-01 and 2023-01-01 for best results with annual frequency",
						"pattern":     "^\\d{4}-\\d{2}-\\d{2}$",
						"default":     "2022-01-01",
					},
					"region_type": map[string]any{
						"type":        "string",
						"description": "The region type for the data",
						"enum":        []string{"state", "msa", "county"},
						"default":     "state",
					},
					"units": map[string]any{
						"type":        "string",
						"description": "The units for the data",
						"enum":        []string{"Dollars", "Percent"},
						"default":     "Dollars",
					},
					"frequency": map[string]any{
						"type":        "string",
						"description": "The frequency of the data - annual is most reliable for series 882",
						"enum":        []string{"a", "q", "m"},
						"default":     "a",
					},
					"season": map[string]any{
						"type":        "string",
						"description": "Seasonal adjustment",
						"enum":        []string{"SA", "NSA"},
						"default":     "NSA",
					},
				},
				"required": []string{"series_group", "date", "region_type", "units", "frequency", "season"},
			},
		},
	}
}

// Run executes the tool with given arguments
func (p *FredTool) Run(ctx context.Context, args string) (string, error) {
	// Parse input arguments
	var input struct {
		SeriesGroup string `json:"series_group"`
		Date        string `json:"date"`
		RegionType  string `json:"region_type"`
		Units       string `json:"units"`
		Frequency   string `json:"frequency"`
		Season      string `json:"season"`
	}

	if args != "" {
		if err := json.Unmarshal([]byte(args), &input); err != nil {
			return "", fmt.Errorf("invalid JSON arguments: %w", err)
		}
	}

	// Set defaults if not provided
	if input.SeriesGroup == "" {
		input.SeriesGroup = "882"
	}
	if input.Date == "" {
		input.Date = "2022-01-01"
	}
	if input.RegionType == "" {
		input.RegionType = "state"
	}
	if input.Units == "" {
		input.Units = "Dollars"
	}
	if input.Frequency == "" {
		input.Frequency = "a"
	}
	if input.Season == "" {
		input.Season = "NSA"
	}

	// Fetch regional series data using the new client-based approach
	result, err := p.fetchWithClient(ctx, input.SeriesGroup, input.Date, input.RegionType, input.Units, input.Frequency, input.Season)
	if err != nil {
		return "", fmt.Errorf("failed to fetch regional data: %w", err)
	}

	// Convert the entire response to map using JSON marshaling/unmarshaling
	// This automatically handles all fields without manual enumeration
	jsonData, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	var response map[string]any
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	// Add metadata
	response["metadata"] = map[string]any{
		"timestamp":    time.Now().UTC(),
		"source":       "fred_geofred",
		"series_group": input.SeriesGroup,
		"date":         input.Date,
		"region_type":  input.RegionType,
		"units":        input.Units,
		"frequency":    input.Frequency,
		"season":       input.Season,
	}

	// Convert back to JSON string for AI response
	resultJSON, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal final response: %w", err)
	}

	return string(resultJSON), nil
}

// fetchWithClient uses the FRED provider client to fetch regional data
func (p *FredTool) fetchWithClient(ctx context.Context, seriesGroup, date, regionType, units, frequency, season string) (map[string]any, error) {
	// Use the FRED client to get GeoFRED regional data
	geoData, err := p.client.GetGeoFREDRegionalData(ctx, seriesGroup, date, regionType, units, frequency, season)
	if err != nil {
		slog.Error("Failed to get FRED regional data", "error", err, "series_group", seriesGroup, "date", date)
		return nil, fmt.Errorf("failed to get regional data: %w", err)
	}

	// Convert to the expected format: map[string]Series
	result := make(map[string]any)

	// Parse the nested date -> []observation structure
	for dateKey, observations := range geoData.Meta.Data {
		for _, obs := range observations {
			// Convert value to string (handling different types)
			var valueStr string
			if val, ok := obs.Value.(string); ok {
				valueStr = val
			} else if val, ok := obs.Value.(float64); ok {
				valueStr = fmt.Sprintf("%.0f", val)
			} else if val, ok := obs.Value.(int); ok {
				valueStr = fmt.Sprintf("%d", val)
			}

			series := models.FREDGeoRegionalSeries{
				Code:      obs.Code,
				Region:    obs.Region,
				SeriesID:  obs.SeriesID,
				Value:     valueStr,
				Units:     obs.Units,
				Frequency: obs.Frequency,
				Date:      dateKey,
			}

			result[obs.Code] = series
		}
	}

	return result, nil
}
