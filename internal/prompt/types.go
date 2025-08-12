package prompt

import (
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// Config holds the configuration needed for prompt building
type Config struct {
	UserLocalization models.LocalizationConfig
	UserProfile      *models.InvestmentProfile // Optional user preferences
}

// Dependencies holds the injected dependencies needed for prompt building
type Dependencies struct {
	Bag    bag.Bag
	Config Config
}

// PromptData holds all the data needed for prompt template execution
type PromptData struct {
	EngineVersion string `json:"engine_version"`
	// User Context
	Localization models.LocalizationConfig `json:"localization"`
	// TODO: UserProfile should be different than InvestmentProfile
	UserProfile *models.InvestmentProfile `json:"user_profile,omitempty"`
	Timestamp   time.Time                 `json:"timestamp"`

	// Portfolio Data
	Portfolio  *models.NormalizedPortfolio  `json:"portfolio,omitempty"`
	MarketData *models.NormalizedMarketData `json:"market_data,omitempty"`
	MacroData  *models.NormalizedMacroData  `json:"macro_data,omitempty"`

	// Analysis-specific context
	AnalysisType models.AnalysisType `json:"analysis_type"`
	Context      map[string]any      `json:"context,omitempty"`

	// Regional extensions (optional fields) - NEW
	InvestmentProfile *models.InvestmentProfile `json:"investment_profile,omitempty"`
	RegionalOverlay   *models.RegionalOverlay   `json:"regional_overlay,omitempty"`
	RegionalConfig    *models.RegionalConfig    `json:"regional_config,omitempty"`
}
