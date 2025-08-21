package report

import (
	"context"
	"time"

	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/fs"
	"github.com/amaurybrisou/mosychlos/pkg/keys"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

//go:generate mockgen -source=bag_loader.go -destination=mocks/bag_loader_mock.go -package=mocks

// BagLoader provides data loading capabilities from SharedBag using fs.FS
// This is a reusable component that can be used across different modules
type BagLoader interface {
	LoadData(ctx context.Context, bag bag.SharedBag, reportType string) (interface{}, error)

	LoadCustomerData(ctx context.Context, bag bag.SharedBag) (*models.CustomerReportData, error)
	LoadSystemData(ctx context.Context, bag bag.SharedBag) (*models.SystemReportData, error)
	LoadFullData(ctx context.Context, bag bag.SharedBag) (*models.FullReportData, error)
}

// bagLoader implements the BagLoader interface
type bagLoader struct {
	fs fs.FS
}

// NewBagLoader creates a new bag loader with file system access
func NewBagLoader(filesystem fs.FS) BagLoader {
	return &bagLoader{
		fs: filesystem,
	}
}

func (r *bagLoader) LoadData(ctx context.Context, bag bag.SharedBag, reportType string) (interface{}, error) {
	switch reportType {
	case "customer":
		return r.LoadCustomerData(ctx, bag)
	case "system":
		return r.LoadSystemData(ctx, bag)
	case "full":
		return r.LoadFullData(ctx, bag)
	default:
		return r.LoadFullData(ctx, bag)
	}
}

// LoadCustomerData extracts customer-facing data from the bag
func (l *bagLoader) LoadCustomerData(ctx context.Context, sharedBag bag.SharedBag) (*models.CustomerReportData, error) {
	data := &models.CustomerReportData{
		GeneratedAt: time.Now(),
	}

	// Load portfolio data
	if portfolio, ok := sharedBag.Get(keys.KPortfolio); ok {
		// if p, ok := portfolio.(map[string]any); ok {
		data.Portfolio = portfolio
		// }
	}

	// Load analysis results
	if risk, ok := sharedBag.Get(keys.KRiskMetrics); ok {
		data.RiskMetrics = risk
	}

	if allocation, ok := sharedBag.Get(keys.KPortfolioAllocationData); ok {
		data.AllocationData = allocation
	}

	if performance, ok := sharedBag.Get(keys.KPortfolioPerformanceData); ok {
		data.PerformanceData = performance
	}

	if compliance, ok := sharedBag.Get(keys.KPortfolioComplianceData); ok {
		data.ComplianceData = compliance
	}

	if stockAnalysis, ok := sharedBag.Get(keys.KStockAnalysis); ok {
		data.StockAnalysis = stockAnalysis
	}

	if insights, ok := sharedBag.Get(keys.KInsights); ok {
		data.Insights = insights
	}

	if newsAnalyzed, ok := sharedBag.Get(keys.KNewsAnalyzed); ok {
		data.NewsAnalyzed = newsAnalyzed
	}

	if fundamentals, ok := sharedBag.Get(keys.KFundamentals); ok {
		data.Fundamentals = fundamentals
	}

	return data, nil
}

// LoadSystemData extracts system diagnostic data from the bag
func (l *bagLoader) LoadSystemData(ctx context.Context, sharedBag bag.SharedBag) (*models.SystemReportData, error) {
	data := &models.SystemReportData{
		GeneratedAt: time.Now(),
	}

	// Track batch mode status
	value, ok := sharedBag.Get(keys.KBatchMode)
	if ok && value.(bool) {
		data.BatchMode = true
	}

	// Load system health data
	if appHealth, ok := sharedBag.Get(keys.KApplicationHealth); ok {
		if health, ok := appHealth.(models.ApplicationHealth); ok {
			data.ApplicationHealth = health
		}
	}

	if toolMetrics, ok := sharedBag.Get(keys.KToolMetrics); ok {
		if metrics, ok := toolMetrics.(*models.ToolMetrics); ok {
			data.ToolMetrics = metrics
		}
	}

	if cacheStats, ok := sharedBag.Get(keys.KCacheHealthStatus); ok {
		if stats, ok := cacheStats.(*models.CacheHealthStatus); ok {
			data.CacheStats = stats
		}
	}

	if extDataHealth, ok := sharedBag.Get(keys.KExternalDataHealth); ok {
		if health, ok := extDataHealth.(*models.ExternalDataHealth); ok {
			data.ExternalDataHealth = health
		}
	}

	if marketDataFreshness, ok := sharedBag.Get(keys.KMarketDataFreshness); ok {
		if freshness, ok := marketDataFreshness.(*models.MarketDataFreshness); ok {
			data.MarketDataFreshness = freshness
		}
	}

	// Load tool computations
	if toolComputations, ok := sharedBag.Get(keys.KToolComputations); ok {
		if computations, ok := toolComputations.([]models.ToolComputation); ok {
			data.ToolComputations = computations
		}
	}

	return data, nil
}

// LoadFullData extracts both customer and system data from the bag
func (l *bagLoader) LoadFullData(ctx context.Context, sharedBag bag.SharedBag) (*models.FullReportData, error) {
	customerData, err := l.LoadCustomerData(ctx, sharedBag)
	if err != nil {
		return nil, err
	}

	systemData, err := l.LoadSystemData(ctx, sharedBag)
	if err != nil {
		return nil, err
	}

	return &models.FullReportData{
		Customer: customerData,
		System:   systemData,
	}, nil
}

// IsBatchMode returns true if the bag is in batch mode
func (l *bagLoader) IsBatchMode(sharedBag bag.SharedBag) bool {
	return sharedBag.MustGet(keys.KBatchMode).(bool)
}
