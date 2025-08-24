package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

type Config struct {
	// CacheDir is the base directory where the app persists data/cache artifacts
	CacheDir string `mapstructure:"cache_dir" yaml:"cache_dir"`
	// DataDir is the base directory where the app stores portfolio data
	DataDir string `mapstructure:"data_dir" yaml:"data_dir"`

	// ConfigDir is the base directory where the app stores configuration files
	ConfigDir string `mapstructure:"config_dir" yaml:"config_dir"`

	// Centralized localization configuration
	Localization models.LocalizationConfig `mapstructure:"localization" yaml:"localization"`

	LLM          LLMConfig          `mapstructure:"llm" yaml:"llm"`
	Tools        ToolsConfig        `mapstructure:"tools" yaml:"tools"`
	Binance      BinanceConfig      `mapstructure:"binance" yaml:"binance"`
	Jurisdiction JurisdictionConfig `mapstructure:"jurisdiction" yaml:"jurisdiction"`
	Report       ReportConfig       `mapstructure:"report" yaml:"report"`

	Logging LoggingConfig `mapstructure:"logging" yaml:"logging"`

	// internal validation state
	validated bool
}

// End of Config struct

// LoggingConfig holds the configuration for logging
type LoggingConfig struct {
	Level     string `mapstructure:"level" yaml:"level"`
	Format    string `mapstructure:"format" yaml:"format"`
	AddSource bool   `mapstructure:"add_source" yaml:"add_source"`
	Output    string `mapstructure:"output" yaml:"output"`
	// End of LoggingConfig struct
}

// ReportConfig holds the configuration for report generation
type ReportConfig struct {
	// OutputDir is the directory where reports are saved (relative to DataDir)
	OutputDir string `mapstructure:"output_dir" yaml:"output_dir"`
	// DefaultFormat is the default output format for reports (markdown, pdf, json)
	DefaultFormat string `mapstructure:"default_format" yaml:"default_format"`
	// IncludeTimestamp whether to include timestamp in report filenames
	IncludeTimestamp bool `mapstructure:"include_timestamp" yaml:"include_timestamp"`
	// DefaultCustomerName is the default customer name for reports
	DefaultCustomerName string `mapstructure:"default_customer_name" yaml:"default_customer_name"`
	// PDFEngine specifies the LaTeX engine for PDF generation (xelatex, lualatex)
	PDFEngine string `mapstructure:"pdf_engine" yaml:"pdf_engine"`
	// EnablePDFUnicodeSanitization enables unicode sanitization fallback for PDF
	EnablePDFUnicodeSanitization bool `mapstructure:"enable_pdf_unicode_sanitization" yaml:"enable_pdf_unicode_sanitization"`
	// ArchiveReports whether to keep generated reports in an archive
	ArchiveReports bool `mapstructure:"archive_reports" yaml:"archive_reports"`
	// MaxArchivedReports maximum number of reports to keep in archive (0 = unlimited)
	MaxArchivedReports int `mapstructure:"max_archived_reports" yaml:"max_archived_reports"`
	// End of ReportConfig struct
}

// Validate validates the Report configuration
func (rc *ReportConfig) Validate(dataDir string) error {
	// set defaults if not provided
	if strings.TrimSpace(rc.OutputDir) == "" {
		rc.OutputDir = "reports"
	}

	if strings.TrimSpace(rc.DefaultFormat) == "" {
		rc.DefaultFormat = "markdown"
	}

	if strings.TrimSpace(rc.PDFEngine) == "" {
		rc.PDFEngine = "xelatex"
	}

	// validate default format
	validFormats := []string{"markdown", "pdf", "json"}
	formatValid := false
	for _, format := range validFormats {
		if rc.DefaultFormat == format {
			formatValid = true
			break
		}
	}
	if !formatValid {
		return fmt.Errorf("DefaultFormat must be one of %v, got: %s", validFormats, rc.DefaultFormat)
	}

	// validate PDF engine
	validEngines := []string{"xelatex", "lualatex"}
	engineValid := false
	for _, engine := range validEngines {
		if rc.PDFEngine == engine {
			engineValid = true
			break
		}
	}
	if !engineValid {
		return fmt.Errorf("PDFEngine must be one of %v, got: %s", validEngines, rc.PDFEngine)
	}

	// validate MaxArchivedReports is non-negative
	if rc.MaxArchivedReports < 0 {
		return fmt.Errorf("MaxArchivedReports must be non-negative, got: %d", rc.MaxArchivedReports)
	}

	// construct absolute output directory path
	var outputPath string
	if filepath.IsAbs(rc.OutputDir) {
		outputPath = rc.OutputDir
	} else {
		outputPath = filepath.Join(dataDir, rc.OutputDir)
	}

	// validate output directory can be created and is writable
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("failed to create report output directory %s: %w", outputPath, err)
	}

	// test writability
	testFile := filepath.Join(outputPath, ".write_test_report")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("report output directory %s is not writable: %w", outputPath, err)
	}

	// clean up test file
	_ = os.Remove(testFile)

	return nil
}

// GetReportOutputDir returns the absolute path to the report output directory
func (rc *ReportConfig) GetReportOutputDir(dataDir string) string {
	if filepath.IsAbs(rc.OutputDir) {
		return rc.OutputDir
	}
	return filepath.Join(dataDir, rc.OutputDir)
}

// Validate performs strict validation on the configuration
func (c *Config) Validate() error {
	if c.validated {
		return nil // already validated
	}

	// validate required directories
	if strings.TrimSpace(c.DataDir) == "" {
		return fmt.Errorf("DataDir cannot be empty")
	}
	if strings.TrimSpace(c.CacheDir) == "" {
		return fmt.Errorf("CacheDir cannot be empty")
	}

	// validate directory paths are absolute
	if !filepath.IsAbs(c.DataDir) {
		return fmt.Errorf("DataDir must be an absolute path, got: %s", c.DataDir)
	}
	if !filepath.IsAbs(c.CacheDir) {
		return fmt.Errorf("CacheDir must be an absolute path, got: %s", c.CacheDir)
	}

	// validate directory permissions (create if needed)
	if err := c.validateDirectory(c.DataDir); err != nil {
		return fmt.Errorf("DataDir validation failed: %w", err)
	}
	if err := c.validateDirectory(c.CacheDir); err != nil {
		return fmt.Errorf("CacheDir validation failed: %w", err)
	}

	// validate localization config
	if err := c.Localization.Validate(); err != nil {
		return fmt.Errorf("localization config validation failed: %w", err)
	}

	// populate computed fields from centralized localization
	c.populateComputedFields()

	// validate jurisdiction config
	if err := c.Jurisdiction.Validate(); err != nil {
		return fmt.Errorf("jurisdiction config validation failed: %w", err)
	}

	// validate binance config if provided
	if err := c.Binance.Validate(); err != nil {
		return fmt.Errorf("binance config validation failed: %w", err)
	}

	// validate LLM config
	if err := c.LLM.Validate(); err != nil {
		return fmt.Errorf("LLM config validation failed: %w", err)
	}

	// validate report config
	if err := c.Report.Validate(c.DataDir); err != nil {
		return fmt.Errorf("report config validation failed: %w", err)
	}

	// mark as validated
	c.validated = true
	return nil
}

// populateComputedFields sets computed fields from centralized localization
func (c *Config) populateComputedFields() {
	// Populate Jurisdiction.Country from centralized localization
	c.Jurisdiction.Country = models.CountryCode(c.Localization.Country)

	// Populate LLM.Locale from centralized localization.language
	c.LLM.Locale = c.Localization.Language

	// Populate OpenAI.WebSearchUserLocation from centralized localization
	c.LLM.OpenAI.WebSearchUserLocation = &WebSearchUserLocationConfig{
		Country:  &c.Localization.Country,
		City:     &c.Localization.City,
		Region:   &c.Localization.Region,
		Timezone: &c.Localization.Timezone,
	}

	// Populate tools configs from centralized localization
	if c.Tools.NewsAPI != nil {
		c.Tools.NewsAPI.Locale = c.Localization.Language
	}
	if c.Tools.FRED != nil {
		c.Tools.FRED.Country = c.Localization.Country
	}
}

// IsValid returns whether the config has been successfully validated
func (c *Config) IsValid() bool {
	return c.validated
}

// MustValidate validates the config and panics on error
func (c *Config) MustValidate() {
	if err := c.Validate(); err != nil {
		panic(fmt.Sprintf("Configuration validation failed: %v", err))
	}
}

// validateDirectory ensures directory exists and is writable
func (c *Config) validateDirectory(dir string) error {
	// create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// check if directory is writable
	testFile := filepath.Join(dir, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("directory %s is not writable: %w", dir, err)
	}

	// clean up test file
	if err := os.Remove(testFile); err != nil {
		// log warning but don't fail
		return nil
	}

	return nil
}

type ToolsConfig struct {
	EnabledTools        []string                   `mapstructure:"enabled_tools" yaml:"enabled_tools"`
	NewsAPI             *NewsAPIConfig             `mapstructure:"newsapi" yaml:"newsapi"`
	FRED                *FREDConfig                `mapstructure:"fred" yaml:"fred"`
	FMP                 *FMPConfig                 `mapstructure:"fmp" yaml:"fmp"`
	FMPAnalystEstimates *FMPAnalystEstimatesConfig `mapstructure:"fmp_analyst_estimates" yaml:"fmp_analyst_estimates"`
	YFinance            *YFinanceConfig            `mapstructure:"yfinance" yaml:"yfinance"`
	SECEdgar            *SECEdgarConfig            `mapstructure:"sec_edgar" yaml:"sec_edgar"`
}

func (c *Config) GetToolConfig(name string) any {
	switch name {
	case bag.NewsAPI.String():
		return c.Tools.NewsAPI
	case bag.Fred.String():
		return c.Tools.FRED
	case bag.FMP.String():
		return c.Tools.FMP
	case bag.FMPAnalystEstimates.String():
		return c.Tools.FMPAnalystEstimates
	case bag.YFinanceDividends.String(),
		bag.YFinanceFinancials.String(),
		bag.YFinanceMarketData.String(),
		bag.YFinanceStockData.String(),
		bag.YFinanceStockInfo.String():
		return c.Tools.YFinance
	case bag.SECFilings.String():
		return c.Tools.SECEdgar
	case bag.WebSearch.String():
		return &c.LLM.OpenAI
	default:
		return nil
	}
}

type NewsAPIConfig struct {
	APIKey   string `mapstructure:"api_key"`
	BaseURL  string `mapstructure:"base_url"`
	Provider string `mapstructure:"provider"`
	// Locale is computed at runtime from centralized localization.language
	Locale      string
	CacheEnable bool `mapstructure:"cache_enable"`
}

type FREDConfig struct {
	APIKey string `mapstructure:"api_key"`
	// Country is computed at runtime from centralized localization.country
	Country     string
	Series      FREDSeriesConfig `mapstructure:"series"`
	CacheEnable bool             `mapstructure:"cache_enable"`
}

type FREDSeriesConfig struct {
	GDP            string `mapstructure:"gdp"`
	Inflation      string `mapstructure:"inflation"`
	InterestRate   string `mapstructure:"interest_rate"`
	Unemployment   string `mapstructure:"unemployment"`
	InflationUnits string `mapstructure:"inflation_units"`
}

type FMPConfig struct {
	APIKey      string `mapstructure:"api_key"`
	Provider    string `mapstructure:"provider"`
	MaxDaily    int    `mapstructure:"max_daily"`
	CacheEnable bool   `mapstructure:"cache_enable"`
}

// FMPAnalystEstimatesConfig holds the configuration for FMP Analyst Estimates tool
type FMPAnalystEstimatesConfig struct {
	APIKey      string `mapstructure:"api_key"`
	Provider    string `mapstructure:"provider"`
	CacheDir    string `mapstructure:"cache_dir"`
	CacheEnable bool   `mapstructure:"cache_enable"`
	MaxDaily    int    `mapstructure:"max_daily"`
}

type YFinanceConfig struct {
	BaseURL     string `mapstructure:"base_url" yaml:"base_url"`
	CacheEnable bool   `mapstructure:"cache_enable" yaml:"cache_enable"`
	MaxDaily    int    `mapstructure:"max_daily" yaml:"max_daily"`
	Timeout     int    `mapstructure:"timeout" yaml:"timeout"`
	MaxRequests int    `mapstructure:"max_requests" yaml:"max_requests"`
}

// BinanceConfig holds the configuration for Binance API
type BinanceConfig struct {
	APIKey      string `mapstructure:"api_key" yaml:"api_key"`
	APISecret   string `mapstructure:"api_secret" yaml:"api_secret"`
	BaseURL     string `mapstructure:"base_url" yaml:"base_url"`
	CacheEnable bool   `mapstructure:"cache_enable" yaml:"cache_enable"`
}

// SECEdgarConfig holds SEC EDGAR tool configuration
type SECEdgarConfig struct {
	UserAgent   string `mapstructure:"user_agent" yaml:"user_agent"`
	BaseURL     string `mapstructure:"base_url" yaml:"base_url"`
	CacheEnable bool   `mapstructure:"cache_enable" yaml:"cache_enable"`
	MaxDaily    int    `mapstructure:"max_daily" yaml:"max_daily"`
}

// Validate validates the Binance configuration
func (bc *BinanceConfig) Validate() error {
	// if all fields are empty, consider it as not configured (optional)
	if bc.APIKey == "" && bc.APISecret == "" && bc.BaseURL == "" {
		return nil // not configured, skip validation
	}

	if strings.TrimSpace(bc.APIKey) == "" {
		return fmt.Errorf("APIKey cannot be empty when Binance is configured")
	}
	if strings.TrimSpace(bc.APISecret) == "" {
		return fmt.Errorf("APISecret cannot be empty when Binance is configured")
	}
	if strings.TrimSpace(bc.BaseURL) == "" {
		return fmt.Errorf("BaseURL cannot be empty when Binance is configured")
	}

	// validate URL format
	if !strings.HasPrefix(bc.BaseURL, "http://") && !strings.HasPrefix(bc.BaseURL, "https://") {
		return fmt.Errorf("BaseURL must be a valid HTTP(S) URL, got: %s", bc.BaseURL)
	}

	return nil
}

// JurisdictionConfig holds the configuration for jurisdiction validation
type JurisdictionConfig struct {
	// Country is computed at runtime from centralized localization.country
	Country models.CountryCode
	// CustomSchemaPath allows overriding the default embedded policy schema
	CustomSchemaPath string `mapstructure:"custom_schema_path" yaml:"custom_schema_path"`
	// Rules contains the compliance rules for this jurisdiction
	Rules models.ComplianceRules `mapstructure:"rules" yaml:"rules"`
}

// Validate validates the Jurisdiction configuration
func (jc *JurisdictionConfig) Validate() error {
	// country code is required (should be populated from centralized localization)
	if jc.Country == "" {
		return fmt.Errorf("country code cannot be empty")
	}

	// validate country code format (2-letter ISO code)
	countryStr := string(jc.Country)
	if len(countryStr) != 2 {
		return fmt.Errorf("country code must be 2 characters (ISO 3166-1 alpha-2), got: %s", countryStr)
	}
	if strings.ToUpper(countryStr) != countryStr {
		return fmt.Errorf("country code must be uppercase, got: %s", countryStr)
	}

	// validate custom schema path if provided
	if jc.CustomSchemaPath != "" {
		if !filepath.IsAbs(jc.CustomSchemaPath) {
			return fmt.Errorf("CustomSchemaPath must be absolute, got: %s", jc.CustomSchemaPath)
		}
		if _, err := os.Stat(jc.CustomSchemaPath); os.IsNotExist(err) {
			return fmt.Errorf("CustomSchemaPath file does not exist: %s", jc.CustomSchemaPath)
		}
	}

	// validate compliance rules
	if err := jc.Rules.Validate(); err != nil {
		return fmt.Errorf("compliance rules validation failed: %w", err)
	}

	return nil
}

type LLMModel string

func (m LLMModel) String() string {
	return string(m)
}

const (
	// Model names
	LLMModelGPT5      LLMModel = "gpt-5"
	LLMModelGPT5Mini  LLMModel = "gpt-5-mini"
	LLMModelGPT5Nano  LLMModel = "gpt-5-nano"
	LLMModelGPT4o     LLMModel = "gpt-4o"
	LLMModelGPT4oMini LLMModel = "gpt-4o-mini"
	LLMModelGPT4oNano LLMModel = "gpt-4o-nano"
	LLMModelClaude    LLMModel = "claude"
)

// WebSearchUserLocationConfig represents user location for web search (computed at runtime)
type WebSearchUserLocationConfig struct {
	Country  *string
	City     *string
	Region   *string
	Timezone *string
}

// LLMConfig holds the configuration for LLM/AI providers
type LLMConfig struct {
	// Provider specifies the AI provider (openai, claude, etc.)
	Provider string `mapstructure:"provider" yaml:"provider"`
	// Model specifies the model to use for completions
	Model LLMModel `mapstructure:"model" yaml:"model"`
	// APIKey for the provider API
	APIKey string `mapstructure:"api_key" yaml:"api_key"`
	// BaseURL for custom API endpoints or proxies
	BaseURL string `mapstructure:"base_url" yaml:"base_url"`
	// Locale is computed at runtime from centralized localization.language
	Locale string
	// OpenAI contains OpenAI-specific configuration
	OpenAI OpenAIConfig `mapstructure:"openai" yaml:"openai"`
}

// Validate validates the LLM configuration
func (lc *LLMConfig) Validate() error {
	if strings.TrimSpace(lc.Provider) == "" {
		return fmt.Errorf("provider cannot be empty")
	}

	if strings.TrimSpace(string(lc.Model)) == "" {
		return fmt.Errorf("model cannot be empty")
	}

	// validate lc.Model is in our defined models
	validModels := []LLMModel{
		LLMModelGPT5,
		LLMModelGPT5Mini,
		LLMModelGPT5Nano,
		LLMModelGPT4o,
		LLMModelGPT4oMini,
		LLMModelGPT4oNano,
		LLMModelClaude,
	}
	valid := slices.Contains(validModels, lc.Model)
	if !valid {
		return fmt.Errorf("model must be one of %v, got: %s", validModels, lc.Model)
	}

	if strings.TrimSpace(lc.APIKey) == "" {
		return fmt.Errorf("APIKey cannot be empty")
	}

	// validate base URL if provided
	if lc.BaseURL != "" && !strings.HasPrefix(lc.BaseURL, "http://") && !strings.HasPrefix(lc.BaseURL, "https://") {
		return fmt.Errorf("BaseURL must be a valid HTTP(S) URL, got: %s", lc.BaseURL)
	}

	// validate OpenAI config if provider is openai
	if strings.ToLower(lc.Provider) == "openai" {
		if err := lc.OpenAI.Validate(); err != nil {
			return fmt.Errorf("OpenAI config validation failed: %w", err)
		}
	}

	return nil
}

// OpenAIConfig holds OpenAI-specific configuration parameters
type OpenAIConfig struct {
	// OrganizationID for API requests
	OrganizationID string `mapstructure:"organization_id" yaml:"organization_id"`
	// ProjectID for API requests
	ProjectID string `mapstructure:"project_id" yaml:"project_id"`
	// MaxCompletionTokens sets the token limit for responses
	MaxCompletionTokens int64 `mapstructure:"max_completion_tokens" yaml:"max_completion_tokens"`
	// Temperature controls randomness (0-2)
	Temperature *float64 `mapstructure:"temperature" yaml:"temperature"`
	// TopP controls nucleus sampling (0-1), alternative to temperature
	TopP *float64 `mapstructure:"top_p" yaml:"top_p"`
	// ReasoningEffort for reasoning models: minimal, low, medium, high
	ReasoningEffort *string `mapstructure:"reasoning_effort" yaml:"reasoning_effort"`
	// Verbosity controls response length: low, medium, high
	Verbosity *string `mapstructure:"verbosity" yaml:"verbosity"`
	// ParallelToolCalls enables parallel function calling
	ParallelToolCalls bool `mapstructure:"parallel_tool_calls" yaml:"parallel_tool_calls"`
	// ServiceTier for request processing: auto, default, flex, priority
	ServiceTier *string `mapstructure:"service_tier" yaml:"service_tier"`
	// PresencePenalty controls topic diversity (-2.0 to 2.0)
	PresencePenalty *float64 `mapstructure:"presence_penalty" yaml:"presence_penalty"`
	// FrequencyPenalty controls repetition (-2.0 to 2.0)
	FrequencyPenalty *float64 `mapstructure:"frequency_penalty" yaml:"frequency_penalty"`
	// Seed for reproducible outputs
	Seed *int64 `mapstructure:"seed" yaml:"seed"`
	// PromptCacheKey for response caching optimization
	PromptCacheKey *string `mapstructure:"prompt_cache_key" yaml:"prompt_cache_key"`
	// WebSearch enables OpenAI's web search capability (uses Responses API)
	WebSearch bool `mapstructure:"web_search" yaml:"web_search"`
	// WebSearchContextSize controls web search context: low, medium, high
	WebSearchContextSize string `mapstructure:"web_search_context_size" yaml:"web_search_context_size"`
	// WebSearchUserLocation is computed at runtime from centralized localization
	WebSearchUserLocation *WebSearchUserLocationConfig
	// RateLimit holds rate limiting configuration
	RateLimit models.RateLimitConfig `yaml:"rate_limit"`
	// Retry holds retry configuration
	Retry models.RetryConfig `yaml:"retry"`
}

// Validate validates the OpenAI configuration
func (oc *OpenAIConfig) Validate() error {
	// validate temperature range
	if oc.Temperature != nil && (*oc.Temperature < 0 || *oc.Temperature > 2) {
		return fmt.Errorf("temperature must be between 0 and 2, got: %f", *oc.Temperature)
	}

	// validate top_p range
	if oc.TopP != nil && (*oc.TopP < 0 || *oc.TopP > 1) {
		return fmt.Errorf("TopP must be between 0 and 1, got: %f", *oc.TopP)
	}

	// validate reasoning effort values
	if oc.ReasoningEffort != nil {
		validReasoningEfforts := []string{"minimal", "low", "medium", "high"}
		valid := false
		for _, v := range validReasoningEfforts {
			if *oc.ReasoningEffort == v {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("ReasoningEffort must be one of %v, got: %s", validReasoningEfforts, *oc.ReasoningEffort)
		}
	}

	// validate verbosity values
	if oc.Verbosity != nil {
		validVerbosities := []string{"low", "medium", "high"}
		valid := false
		for _, v := range validVerbosities {
			if *oc.Verbosity == v {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("verbosity must be one of %v, got: %s", validVerbosities, *oc.Verbosity)
		}
	}

	// validate service tier values
	if oc.ServiceTier != nil {
		validServiceTiers := []string{"auto", "default", "flex", "scale", "priority"}
		valid := false
		for _, v := range validServiceTiers {
			if *oc.ServiceTier == v {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("ServiceTier must be one of %v, got: %s", validServiceTiers, *oc.ServiceTier)
		}
	}

	// validate penalty ranges
	if oc.PresencePenalty != nil && (*oc.PresencePenalty < -2.0 || *oc.PresencePenalty > 2.0) {
		return fmt.Errorf("PresencePenalty must be between -2.0 and 2.0, got: %f", *oc.PresencePenalty)
	}

	if oc.FrequencyPenalty != nil && (*oc.FrequencyPenalty < -2.0 || *oc.FrequencyPenalty > 2.0) {
		return fmt.Errorf("FrequencyPenalty must be between -2.0 and 2.0, got: %f", *oc.FrequencyPenalty)
	}

	// validate max completion tokens (0 means not configured, which is valid)
	if oc.MaxCompletionTokens < 0 {
		return fmt.Errorf("MaxCompletionTokens must be non-negative, got: %d", oc.MaxCompletionTokens)
	}

	// validate web search context size
	if oc.WebSearchContextSize != "" {
		validContextSizes := []string{"low", "medium", "high"}
		valid := false
		for _, v := range validContextSizes {
			if oc.WebSearchContextSize == v {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("WebSearchContextSize must be one of %v, got: %s", validContextSizes, oc.WebSearchContextSize)
		}
	}

	return nil
}
