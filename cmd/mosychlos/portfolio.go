package mosychlos

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/amaurybrisou/mosychlos/internal/config"
	"github.com/amaurybrisou/mosychlos/internal/engine"
	"github.com/amaurybrisou/mosychlos/pkg/bag"
	"github.com/amaurybrisou/mosychlos/pkg/models"
	"github.com/spf13/cobra"
)

const (
	modeSummary    = "summary"
	modeDetailed   = "detailed"
	modeAccounts   = "accounts"
	modeCompliance = "compliance"
)

func NewPortfolioCommand(cfg *config.Config) *cobra.Command {
	var (
		mode           string
		nonInteractive bool
	)

	cmd := &cobra.Command{
		Use:   "portfolio",
		Short: "Display portfolio information",
		Long:  "Interactively display your portfolio with various view options (via the Engine Orchestrator).",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			return runPortfolioUI(ctx, cfg, mode, nonInteractive)
		},
	}

	cmd.Flags().StringVarP(&mode, "mode", "m", "", "Display mode: summary | detailed | accounts | compliance")
	cmd.Flags().BoolVar(&nonInteractive, "no-input", false, "Non-interactive mode (requires --mode)")

	return cmd
}

func runPortfolioUI(ctx context.Context, cfg *config.Config, mode string, nonInteractive bool) error {
	// 1) Build orchestrator (single source of truth)
	// If you have options/registry, pass them here (e.g., engine.WithFS(...), engine.WithRegistry(...))
	orch := engine.New(cfg)

	// 2) Initialize once (sets up tools, portfolio, profile, LLM, etc.)
	if err := orch.Init(ctx); err != nil {
		return fmt.Errorf("orchestrator init: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		// Resolve mode (flag or prompt)
		var err error
		if mode == "" {
			if nonInteractive {
				return errors.New("non-interactive mode requires --mode")
			}
			mode, err = promptDisplayMode(reader)
			if err != nil {
				return err
			}
		}

		// Pull portfolio from the orchestrator's shared state
		// NOTE: Adjust if your orchestrator exposes the bag differently (e.g., orch.Bag(), orch.SharedBag(), orch.Snapshot()).
		shareBag := orch.Bag() // <- common pattern; replace with your actual accessor if different
		portfolioData, ok := shareBag.Get(bag.KPortfolio)
		if !ok || portfolioData == nil {
			return errors.New("portfolio not available in orchestrator state (keys.KPortfolio)")
		}

		// If you have a concrete type, assert it here (example):
		portfolio, ok := portfolioData.(*models.Portfolio)
		if !ok {
			return errors.New("unexpected portfolio type in bag")
		}

		// Dispatch rendering by mode — plug your real renderers in place of the fmt.Println
		switch mode {
		case modeSummary:
			// display portfolio
			fmt.Println(strings.Repeat("=", 20))
			for _, account := range portfolio.Accounts {
				fmt.Printf("%s %s %s:\n", account.ID, account.Name, account.Provider)
				fmt.Printf("Currency: %s, Value: %f\n", account.Currency, account.CashBalance())
				for _, h := range account.Holdings {
					fmt.Printf("%s: Quantity: %f, Value: %f, Total Value: %f\n", h.Name, h.Quantity, h.CostBasis, h.CostBasis*h.Quantity)
				}
			}
			fmt.Println(strings.Repeat("=", 20))
		case modeDetailed:
			// display detailed portfolio
		case modeAccounts:
			// cli.DisplayByAccount(portfolioData)
			fmt.Println("[accounts] TODO: render account view via orchestrator outputs")
		case modeCompliance:
			// jurisdiction := cfg.Localization.Country
			// cli.DisplayComplianceCheck(portfolioData, jurisdiction)
			fmt.Printf("[compliance] TODO: render compliance (jurisdiction=%s)\n", cfg.Localization.Country)
		default:
			slog.Warn("unknown mode", "mode", mode)
			if nonInteractive {
				return fmt.Errorf("unknown mode: %s", mode)
			}
		}

		// In non-interactive or CLI-provided mode: exit after one render
		if nonInteractive || isKnownMode(mode) {
			return nil
		}

		ok2, err := confirm(reader, "Would you like to see another view? [y/N]: ")
		if err != nil {
			return fmt.Errorf("confirm: %w", err)
		}
		if !ok2 {
			return nil
		}
		mode = "" // loop back to prompt
	}
}

func promptDisplayMode(r *bufio.Reader) (string, error) {
	options := []string{
		"summary - Portfolio overview and summary statistics",
		"detailed - Detailed holdings breakdown",
		"accounts - View by account structure",
		"compliance - Compliance and regulatory check",
	}

	fmt.Println("\n📊 Select Display Mode:")
	for i, option := range options {
		fmt.Printf("  %d. %s\n", i+1, option)
	}

	fmt.Print("\nEnter your choice (1-4 or name): ")
	line, err := readLine(r, 6*time.Minute)
	if err != nil {
		return "", fmt.Errorf("input: %w", err)
	}
	choice := strings.TrimSpace(line)
	switch strings.ToLower(choice) {
	case "1", modeSummary:
		return modeSummary, nil
	case "2", modeDetailed:
		return modeDetailed, nil
	case "3", modeAccounts:
		return modeAccounts, nil
	case "4", modeCompliance:
		return modeCompliance, nil
	default:
		return "", fmt.Errorf("invalid choice: %q", choice)
	}
}

func confirm(r *bufio.Reader, prompt string) (bool, error) {
	fmt.Print(prompt)
	line, err := readLine(r, 3*time.Minute)
	if err != nil {
		return false, err
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

func readLine(r *bufio.Reader, timeout time.Duration) (string, error) {
	type res struct {
		line string
		err  error
	}
	ch := make(chan res, 1)
	go func() {
		s, err := r.ReadString('\n')
		ch <- res{line: s, err: err}
	}()
	select {
	case out := <-ch:
		return strings.TrimRight(out.line, "\r\n"), out.err
	case <-time.After(timeout):
		return "", errors.New("input timed out")
	}
}

func isKnownMode(mode string) bool {
	switch mode {
	case modeSummary, modeDetailed, modeAccounts, modeCompliance:
		return true
	default:
		return false
	}
}
