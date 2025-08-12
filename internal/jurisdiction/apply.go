package jurisdiction

import (
	"strings"

	models "github.com/amaurybrisou/mosychlos/pkg/models"
)

type EligibilityNote struct {
	Ticker string
	Note   string
}

type Result struct {
	Portfolio models.Portfolio
	Notes     []EligibilityNote
}

// Apply filters holdings by rules and applies ticker substitutions. It does not delete holdings; it marks ineligible ones.
func Apply(pf models.Portfolio, r models.ComplianceRules) (Result, map[string]bool) {
	blocked := make(map[string]bool)
	notes := []EligibilityNote{}

	allowed := setFrom(r.AllowedAssetTypes)
	disallowed := setFrom(r.DisallowedAssetTypes)
	blockTick := setFrom(r.TickerBlocklist)
	subs := r.TickerSubstitutes

	out := pf // shallow copy; we don't mutate nested fields besides ticker replacement

	for ai := range out.Accounts {
		for hi := range out.Accounts[ai].Holdings {
			h := &out.Accounts[ai].Holdings[hi]
			// asset type check (case-insensitive; treat crypto + crypto_core compatibly)
			tkey := strings.ToUpper(string(h.Type))
			if len(allowed) > 0 && !allowed[tkey] {
				if (tkey == strings.ToUpper(string(models.Crypto)) && allowed[strings.ToUpper(string(models.CryptoCore))]) ||
					(tkey == strings.ToUpper(string(models.CryptoCore)) && allowed[strings.ToUpper(string(models.Crypto))]) {
					// allow cross-compat crypto types
				} else {
					blocked[h.Ticker] = true
					notes = append(notes, EligibilityNote{Ticker: h.Ticker, Note: "asset type not allowed"})
					continue
				}
			}
			if disallowed[tkey] {
				blocked[h.Ticker] = true
				notes = append(notes, EligibilityNote{Ticker: h.Ticker, Note: "asset type disallowed"})
				continue
			}
			// ticker blocklist
			if blockTick[h.Ticker] {
				blocked[h.Ticker] = true
				note := "ticker blocked"
				if alt, ok := subs[h.Ticker]; ok {
					note += ", consider substitute: " + alt
					// optional: perform substitution inline for analysis/suggest
					h.Ticker = alt
				}
				notes = append(notes, EligibilityNote{Ticker: h.Ticker, Note: note})
			}
		}
	}

	return Result{Portfolio: out, Notes: notes}, blocked
}

func setFrom(list []string) map[string]bool {
	m := map[string]bool{}
	for _, v := range list {
		m[strings.ToUpper(v)] = true
	}
	return m
}
