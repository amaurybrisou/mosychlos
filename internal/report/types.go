package report

import (
	"context"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/fs"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// ReportGenerator interface for generating different types of reports
type ReportGenerator interface {
	GenerateCustomerReport(ctx context.Context, format models.ReportFormat) (*models.ReportOutput, error)
	GenerateSystemReport(ctx context.Context, format models.ReportFormat) (*models.ReportOutput, error)
	GenerateFullReport(ctx context.Context, format models.ReportFormat) (*models.ReportOutput, error)
}

// Dependencies contains the dependencies needed for report generation
type Dependencies struct {
	Config     *config.Config
	DataBag    bag.Bag
	FileSystem fs.FS
}
