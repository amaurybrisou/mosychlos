package normalize

// Registry holds available normalizers.
type Registry []Normalizer

// Find returns the first normalizer that can handle the given tool.
func (rs Registry) Find(tool string) (Normalizer, bool) {
	for _, n := range rs {
		if n.Can(tool) {
			return n, true
		}
	}
	return nil, false
}

// DefaultRegistry returns a ready-to-use set of normalizers.
func DefaultRegistry() Registry {
	return Registry{
		YFinanceStockData{},
		YFinanceMarketData{},
		NewsAPINormalizer{},
	}
}
