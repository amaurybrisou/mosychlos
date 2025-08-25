package mosychlos

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/internal/report"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/fs"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func NewReportCommand(cfg *config.Config) *cobra.Command {
	inputPath := ""
	inputDir := filepath.Join(cfg.DataDir, "bag")
	outputDir := filepath.Join(cfg.DataDir, "reports")
	var formats []string

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Generate a report from a saved bag file",
		Long:  "Generate a portfolio or system report from a previously saved bag JSON file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// list available input files
			files, err := os.ReadDir(inputDir)
			if err != nil {
				return fmt.Errorf("failed to read input directory: %w", err)
			}

			// use the promptui package to browse available files
			prompt := promptui.Select{
				Label: "Select a bag file",
				Items: files,
			}

			index, _, err := prompt.Run()
			if err != nil {
				return fmt.Errorf("failed to select bag file: %w", err)
			}

			if index < 0 || index >= len(files) {
				return fmt.Errorf("invalid selection")
			}

			inputPath = filepath.Join(inputDir, files[index].Name())

			if inputPath == "" {
				return fmt.Errorf("input bag file required (--input)")
			}

			file, err := os.Open(inputPath)
			if err != nil {
				return fmt.Errorf("failed to open bag file: %w", err)
			}
			defer file.Close()

			if inputPath == "" {
				return fmt.Errorf("input bag file required (--input)")
			}

			file, err = os.Open(inputPath)
			if err != nil {
				return fmt.Errorf("failed to open bag file: %w", err)
			}
			defer file.Close()

			sharedBag, err := bag.LoadSharedBagFromJSON(file)
			if err != nil {
				return fmt.Errorf("failed to load bag: %w", err)
			}

			if outputDir == "" {
				outputDir = filepath.Join(cfg.DataDir, "reports")
			}

			fsys := fs.New(outputDir)
			loader := report.NewBagLoader(fsys)
			ctx := context.Background()

			fullData, err := loader.LoadFullData(ctx, sharedBag)
			if err != nil {
				return fmt.Errorf("failed to extract report data: %w", err)
			}

			if len(formats) == 0 {
				formats = []string{"markdown"}
			}

			for _, format := range formats {
				if err := report.GenerateReport(fullData, outputDir, format); err != nil {
					return fmt.Errorf("failed to generate %s report: %w", format, err)
				}
			}

			fmt.Printf("Report generated in %s\n", outputDir)
			return nil
		},
	}

	cmd.Flags().StringVar(&inputPath, "input", "", "Path to saved bag JSON file")
	cmd.Flags().StringVar(&outputDir, "output", "mosychlos-data/reports", "Output directory for reports")
	cmd.Flags().StringSliceVar(&formats, "format", []string{"markdown"}, "Report formats (markdown, pdf, json)")

	return cmd
}
