package mosychlos

import (
	"fmt"
	"strings"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/pkg/pdf"
	"github.com/spf13/cobra"
)

func NewMarkdownToPDFCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "markdown-to-pdf",
		Short:   "Convert Markdown to PDF",
		Aliases: []string{"md-to-pdf", "mdpdf"},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.MarkFlagRequired("input"); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			pdfConverter := pdf.New(pdf.WithSanitize(true))

			input, err := cmd.Flags().GetString("input")
			if err != nil {
				return fmt.Errorf("failed to get input flag: %w", err)
			}
			// output, err := cmd.Flags().GetString("output")

			pdfPath, err := pdfConverter.Convert(input)
			if err != nil {
				return fmt.Errorf("failed to convert markdown to PDF: %w", err)
			}

			fmt.Println(strings.Repeat("=", 20))
			fmt.Printf("PDF generated at: %s\n", pdfPath)
			fmt.Println(strings.Repeat("=", 20))

			return nil
		},
	}

	cmd.Flags().StringP("input", "i", "input.md", "Input Markdown file path")
	// cmd.Flags().StringP("output", "o", "output.pdf", "Output PDF file path")

	return cmd
}
