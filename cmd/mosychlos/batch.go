// cmd/mosychlos/batch.go
package mosychlos

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/internal/llm"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/fs"
	"github.com/amaurybrisou/mosychlos/pkg/models"
)

// CreateBatchCommand creates the batch command for CLI
func CreateBatchCommand(cfg *config.Config) *cobra.Command {
	var batchCmd = &cobra.Command{
		Use:   "batch",
		Short: "Manage AI batch processing jobs",
		Long: `Submit, monitor, and retrieve results from AI batch processing jobs.
Batch processing offers significant cost savings (up to 50%) for non-time-critical workloads.`,
	}

	// Submit command
	var submitCmd = &cobra.Command{
		Use:   "submit [analysis-type] [portfolio-files...]",
		Short: "Submit batch processing job",
		Long: `Submit a batch job for portfolio analysis.

Examples:
  mosychlos batch submit risk
  mosychlos batch submit allocation *.json           # All JSON portfolios
  mosychlos batch submit --wait performance p1.json p2.json  # Wait for completion`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBatchSubmit(cmd, args, cfg)
		},
	}

	// Status command
	var statusCmd = &cobra.Command{
		Use:   "status [job-id]",
		Short: "Check batch job status",
		Long:  `Check the status of a batch processing job.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBatchStatus(cmd, args, cfg)
		},
	}

	// Results command
	var resultsCmd = &cobra.Command{
		Use:   "results [job-id]",
		Short: "Retrieve batch job results",
		Long:  `Retrieve and display results from a completed batch job.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBatchResults(cmd, args, cfg)
		},
	}

	// Errors command
	var errorsCmd = &cobra.Command{
		Use:   "errors [job-id]",
		Short: "Retrieve batch job errors",
		Long: `Retrieve and display errors from a batch job.
Useful when a batch job fails or has no output file.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBatchErrors(cmd, args, cfg)
		},
	}

	// Wait command
	var waitCmd = &cobra.Command{
		Use:   "wait [job-id]",
		Short: "Wait for batch job completion",
		Long:  `Wait for a batch job to complete and retrieve results.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBatchWait(cmd, args, cfg)
		},
	}

	// Cancel command
	var cancelCmd = &cobra.Command{
		Use:   "cancel [job-id]",
		Short: "Cancel a batch job",
		Long:  `Cancel a running or queued batch job.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBatchCancel(cmd, args, cfg)
		},
	}

	// List command
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List batch jobs",
		Long: `List all batch processing jobs with their current status.

Examples:
  mosychlos batch list                    # List all jobs
  mosychlos batch list --limit 10         # Limit to 10 most recent jobs
  mosychlos batch list --status completed # Only show completed jobs`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBatchList(cmd, args, cfg)
		},
	}

	// Add flags for submit command
	submitCmd.Flags().Bool("wait", false, "Wait for job completion")
	submitCmd.Flags().Duration("timeout", 30*time.Minute, "Timeout for waiting")
	submitCmd.Flags().StringSlice("types", []string{"risk"}, "Analysis types to perform")

	// Add timeout flag for wait command
	waitCmd.Flags().Duration("timeout", 30*time.Minute, "Timeout for waiting")

	// Add flags for list command
	listCmd.Flags().Int("limit", 20, "Maximum number of jobs to return")
	listCmd.Flags().String("status", "", "Filter by status (validating, in_progress, completed, failed, etc.)")
	listCmd.Flags().String("after", "", "Show jobs created after this job ID (pagination)")

	batchCmd.AddCommand(submitCmd)
	batchCmd.AddCommand(statusCmd)
	batchCmd.AddCommand(resultsCmd)
	batchCmd.AddCommand(errorsCmd)
	batchCmd.AddCommand(waitCmd)
	batchCmd.AddCommand(cancelCmd)
	batchCmd.AddCommand(listCmd)

	return batchCmd
}

// getBatchManager creates a minimal batch service for status-only operations
func getBatchManager(cfg *config.Config) (models.BatchManager, error) {
	sharedBag := bag.NewSharedBag()
	return llm.NewBatchServiceFactory(cfg, fs.OS{}, sharedBag).CreateManager()
}

func runBatchSubmit(cmd *cobra.Command, args []string, cfg *config.Config) error {
	ctx := context.Background()

	// Parse arguments
	// analysisTypes, _ := cmd.Flags().GetStringSlice("types")
	// if len(args) > 0 {
	// 	// First arg could be analysis type or portfolio file
	// 	if isAnalysisType(args[0]) {
	// 		analysisTypes = []string{args[0]}
	// 		args = args[1:]
	// 	}
	// }

	// Add timeout if waiting
	wait, _ := cmd.Flags().GetBool("wait")
	if wait {
		timeout, _ := cmd.Flags().GetDuration("timeout")
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	bm, err := getBatchManager(cfg)
	if err != nil {
		return fmt.Errorf("failed to get batch manager: %w", err)
	}

	job, err := bm.ProcessBatch(
		ctx,
		[]models.BatchRequest{},
		models.BatchOptions{},
		wait,
	)
	if err != nil {
		return fmt.Errorf("failed to process batch: %w", err)
	}

	fmt.Printf("Submitted batch job: %s\n", job.ID)
	return nil
}

func runBatchStatus(cmd *cobra.Command, args []string, cfg *config.Config) error {
	jobID := args[0]
	ctx := context.Background()

	bm, err := getBatchManager(cfg)
	if err != nil {
		return fmt.Errorf("failed to get batch manager: %w", err)
	}

	resp, err := bm.GetJobStatus(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job status: %w", err)
	}

	fmt.Printf("Job status: %s\n", resp.Status)
	return nil
}

func runBatchResults(cmd *cobra.Command, args []string, cfg *config.Config) error {
	jobID := args[0]
	ctx := context.Background()

	bm, err := getBatchManager(cfg)
	if err != nil {
		return fmt.Errorf("failed to get batch manager: %w", err)
	}

	resp, err := bm.GetResults(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job results: %w", err)
	}

	fmt.Printf("Job results: %v\n", resp)
	return nil
}

func runBatchErrors(cmd *cobra.Command, args []string, cfg *config.Config) error {
	jobID := args[0]
	ctx := context.Background()

	bm, err := getBatchManager(cfg)
	if err != nil {
		return fmt.Errorf("failed to get batch manager: %w", err)
	}

	errors, err := bm.GetError(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job errors: %w", err)
	}

	if len(errors) == 0 {
		fmt.Printf("✅ No errors found for this job.\n")
	} else {
		fmt.Printf("❌ Found %d error(s):\n\n", len(errors))
		for customID, errorMsg := range errors {
			fmt.Printf("Request ID: %s\n", customID)
			fmt.Printf("Error: %s\n\n", errorMsg)
		}
	}

	return nil
}

func runBatchWait(cmd *cobra.Command, args []string, cfg *config.Config) error {
	jobID := args[0]
	timeout, _ := cmd.Flags().GetDuration("timeout")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	bm, err := getBatchManager(cfg)
	if err != nil {
		return fmt.Errorf("failed to get batch manager: %w", err)
	}

	job, err := bm.WaitForCompletion(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to wait for job completion: %w", err)
	}

	fmt.Printf("Job completed: %v\n", job)
	return nil
}

func runBatchCancel(cmd *cobra.Command, args []string, cfg *config.Config) error {
	jobID := args[0]
	ctx := context.Background()

	bm, err := getBatchManager(cfg)
	if err != nil {
		return fmt.Errorf("failed to get batch manager: %w", err)
	}

	err = bm.CancelJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to cancel job: %w", err)
	}

	return nil
}

func isAnalysisType(arg string) bool {
	validTypes := []string{
		"risk",
		"allocation",
		"performance",
		"compliance",
		"reallocation",
		"investment_research",
	}

	arg = strings.ToLower(arg)
	for _, t := range validTypes {
		if arg == t {
			return true
		}
	}
	return false
}

func runBatchList(cmd *cobra.Command, args []string, cfg *config.Config) error {
	ctx := context.Background()

	// Get flags
	limit, _ := cmd.Flags().GetInt("limit")
	status, _ := cmd.Flags().GetString("status")
	after, _ := cmd.Flags().GetString("after")

	// Create batch service
	bm, err := getBatchManager(cfg)
	if err != nil {
		return fmt.Errorf("failed to create batch manager: %w", err)
	}

	// Prepare filters
	filters := make(map[string]string)
	if limit > 0 {
		filters["limit"] = fmt.Sprintf("%d", limit)
	}
	if status != "" {
		filters["status"] = status
	}
	if after != "" {
		filters["after"] = after
	}

	jobs, err := bm.ListBatches(ctx, filters)
	if err != nil {
		return fmt.Errorf("failed to list batch jobs: %w", err)
	}

	// Iterate over jobs and print each one nicely
	for _, job := range jobs {
		fmt.Println("--------------------------------------------------")
		fmt.Printf("Job ID          : %s\n", job.ID)
		fmt.Printf("Status          : %s\n", job.Status)
		fmt.Printf("Input File ID   : %s\n", job.InputFileID)
		if job.OutputFileID != nil {
			fmt.Printf("Output File ID  : %s\n", *job.OutputFileID)
		} else {
			fmt.Printf("Output File ID  : none\n")
		}
		if job.ErrorFileID != nil {
			fmt.Printf("Error File ID   : %s\n", *job.ErrorFileID)
		} else {
			fmt.Printf("Error File ID   : none\n")
		}
		createdAt := time.Unix(job.CreatedAt, 0)
		fmt.Printf("Created At      : %s\n", createdAt.Format(time.RFC3339))
		if job.CompletedAt != nil {
			completedAt := time.Unix(*job.CompletedAt, 0)
			fmt.Printf("Completed At    : %s\n", completedAt.Format(time.RFC3339))
		} else {
			fmt.Printf("Completed At    : not completed\n")
		}
		fmt.Printf("Request Counts  : Total: %d, Completed: %d, Failed: %d\n",
			job.RequestCounts.Total, job.RequestCounts.Completed, job.RequestCounts.Failed)
	}

	return nil
}
