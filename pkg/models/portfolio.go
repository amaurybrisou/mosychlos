package models

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"time"
)

type AssetType string

type AccountType string

type CurrencyCode string

const (
	// Equities & funds
	Stock      AssetType = "stock"
	ETF        AssetType = "etf"
	MutualFund AssetType = "mutual_fund"

	// Bonds
	BondIG   AssetType = "bond_ig"
	BondHY   AssetType = "bond_hy"
	BondIL   AssetType = "bond_il"
	BondGov  AssetType = "bond_gov"
	BondCorp AssetType = "bond_corp"
	BondEM   AssetType = "bond_em"
	BondMuni AssetType = "bond_muni"

	// Real assets & commodities
	REIT            AssetType = "reit"
	CommodityBroad  AssetType = "commodity_broad"
	CommodityEnergy AssetType = "commodity_energy"
	CommodityAgri   AssetType = "commodity_agri"
	Metal           AssetType = "metal"

	// Crypto & FX
	Crypto     AssetType = "crypto"
	CryptoCore AssetType = "crypto_core"
	Stablecoin AssetType = "stablecoin"
	FX         AssetType = "fx"

	// Cash & liquidity
	Cash   AssetType = "cash"
	CashEQ AssetType = "cash_eq"
	MoneyM AssetType = "money_market"

	// Derivatives
	DerivativeOption AssetType = "derivative_option"
	DerivativeFuture AssetType = "derivative_future"
)

const (
	AccountBrokerage AccountType = "brokerage"
	AccountSavings   AccountType = "savings"
	AccountExchange  AccountType = "exchange"
	AccountVault     AccountType = "vault"
)

type Holding struct {
	Ticker    string    `yaml:"ticker"`
	Quantity  float64   `yaml:"quantity"`
	CostBasis float64   `yaml:"cost_basis"`
	Currency  string    `yaml:"currency"`
	Type      AssetType `yaml:"type"`
	ISIN      string    `yaml:"isin,omitempty"`
	Name      string    `yaml:"name,omitempty"`
	Sector    string    `yaml:"sector,omitempty"`
	Region    string    `yaml:"region,omitempty"`
}

type Account struct {
	Name     string      `yaml:"name"`
	Type     AccountType `yaml:"type"`
	Currency string      `yaml:"currency"`
	Balance  float64     `yaml:"balance,omitempty"`
	Holdings []Holding   `yaml:"holdings,omitempty"`
	// optional metadata
	ID       string   `yaml:"id,omitempty"`
	Provider string   `yaml:"provider,omitempty"`
	Tags     []string `yaml:"tags,omitempty"`
}

type Portfolio struct {
	AsOf         string    `yaml:"as_of"`
	BaseCurrency string    `yaml:"base_currency,omitempty"`
	Accounts     []Account `yaml:"accounts"`
	Validated    bool      `yaml:"validated,omitempty"`
}

// AsOfTime parses AsOf using common formats (YYYY-MM-DD, RFC3339).
func (p Portfolio) AsOfTime() (time.Time, error) {
	if p.AsOf == "" {
		return time.Time{}, nil
	}
	if t, err := time.Parse("2006-01-02", p.AsOf); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, p.AsOf)
}

// Tickers returns the unique list of tickers in the portfolio.
func (p Portfolio) Tickers() []string {
	m := make(map[string]struct{})
	for _, a := range p.Accounts {
		for _, h := range a.Holdings {
			if h.Ticker == "" {
				continue
			}
			m[h.Ticker] = struct{}{}
		}
	}
	out := make([]string, 0, len(m))
	for t := range m {
		out = append(out, t)
	}
	return out
}

// AccountsByType filters accounts by type.
func (p Portfolio) AccountsByType(t AccountType) []Account {
	var out []Account
	for _, a := range p.Accounts {
		if a.Type == t {
			out = append(out, a)
		}
	}
	return out
}

// UserID generates a unique identifier for this portfolio based on its structure.
// This is used for OpenAI's user tracking parameter to enable proper caching and abuse detection.
// The ID is deterministic and based on portfolio characteristics rather than sensitive data.
func (p Portfolio) UserID() string {
	// Create a deterministic hash based on portfolio structure
	h := sha256.New()

	// Include base currency and as_of date
	h.Write([]byte(p.BaseCurrency))
	h.Write([]byte(p.AsOf))

	// Include account structure (types and names, but not sensitive values)
	accountSigs := make([]string, 0, len(p.Accounts))
	for _, account := range p.Accounts {
		accountSig := fmt.Sprintf("%s:%s:%d", account.Type, account.Name, len(account.Holdings))
		accountSigs = append(accountSigs, accountSig)
	}
	sort.Strings(accountSigs)

	for _, sig := range accountSigs {
		h.Write([]byte(sig))
	}

	// Include ticker list (for portfolio composition fingerprint)
	tickers := p.Tickers()
	sort.Strings(tickers)
	for _, ticker := range tickers {
		h.Write([]byte(ticker))
	}

	// Return first 16 characters of hex hash as user ID
	return fmt.Sprintf("portfolio_%x", h.Sum(nil))[:24]
}

// HoldingsByType filters holdings in an account by asset type.
func (a Account) HoldingsByType(t AssetType) []Holding {
	var out []Holding
	for _, h := range a.Holdings {
		if h.Type == t {
			out = append(out, h)
		}
	}
	return out
}

// CashBalance sums cash-type holdings quantities (assumes cost_basis=1 for cash amounts).
func (a Account) CashBalance() float64 {
	total := 0.0
	for _, h := range a.Holdings {
		if h.Type == Cash || h.Type == CashEQ || h.Type == MoneyM {
			total += h.Quantity
		}
	}
	return total
}

// Value computes a holding value given a price, falling back to cost basis when price <= 0.
func (h Holding) Value(price float64) float64 {
	if price > 0 {
		return price * h.Quantity
	}
	if h.CostBasis > 0 {
		return h.CostBasis * h.Quantity
	}
	return 0
}
