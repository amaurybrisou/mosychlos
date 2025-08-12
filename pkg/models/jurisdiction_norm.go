package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

// MarshalJSON provides compact JSON representation optimized for AI agents
func (cr ComplianceRules) MarshalJSON() ([]byte, error) {
	compact := map[string]any{
		"max_leverage": cr.MaxLeverage,
	}

	// asset type summaries instead of full lists
	if len(cr.AllowedAssetTypes) > 0 {
		compact["allowed_count"] = len(cr.AllowedAssetTypes)
		compact["allowed_types"] = cr.AllowedAssetTypes
	}

	if len(cr.DisallowedAssetTypes) > 0 {
		compact["disallowed_count"] = len(cr.DisallowedAssetTypes)
		compact["disallowed_types"] = cr.DisallowedAssetTypes
	}

	// ETF domicile rules
	if len(cr.ETFDomicileAllow) > 0 {
		compact["etf_allowed_domiciles"] = cr.ETFDomicileAllow
	}

	if len(cr.ETFDomicileBlock) > 0 {
		compact["etf_blocked_domiciles"] = cr.ETFDomicileBlock
	}

	// ticker restrictions
	if len(cr.TickerBlocklist) > 0 {
		compact["blocked_tickers_count"] = len(cr.TickerBlocklist)
		// only show first few for space efficiency
		if len(cr.TickerBlocklist) <= 5 {
			compact["blocked_tickers"] = cr.TickerBlocklist
		} else {
			compact["blocked_tickers_sample"] = cr.TickerBlocklist[:5]
		}
	}

	if len(cr.TickerSubstitutes) > 0 {
		compact["substitute_count"] = len(cr.TickerSubstitutes)
		compact["substitutes"] = cr.TickerSubstitutes
	}

	// include notes if meaningful
	if cr.Notes != "" && len(cr.Notes) < 200 {
		compact["notes"] = cr.Notes
	} else if cr.Notes != "" {
		compact["notes"] = cr.Notes[:200] + "..."
	}

	return json.Marshal(compact)
}

// String provides human-readable summary optimized for AI understanding
func (cr ComplianceRules) String() string {
	var parts []string

	// leverage limitation
	if cr.MaxLeverage > 0 {
		parts = append(parts, fmt.Sprintf("max leverage %dx", cr.MaxLeverage))
	}

	// asset type restrictions
	if len(cr.AllowedAssetTypes) > 0 {
		parts = append(parts, fmt.Sprintf("%d allowed types", len(cr.AllowedAssetTypes)))
	}

	if len(cr.DisallowedAssetTypes) > 0 {
		parts = append(parts, fmt.Sprintf("%d blocked types", len(cr.DisallowedAssetTypes)))
	}

	// ticker restrictions
	if len(cr.TickerBlocklist) > 0 {
		parts = append(parts, fmt.Sprintf("%d blocked tickers", len(cr.TickerBlocklist)))
	}

	if len(cr.TickerSubstitutes) > 0 {
		parts = append(parts, fmt.Sprintf("%d substitutes", len(cr.TickerSubstitutes)))
	}

	// ETF domicile rules
	if len(cr.ETFDomicileAllow) > 0 {
		parts = append(parts, fmt.Sprintf("ETF domiciles: %s allowed", strings.Join(cr.ETFDomicileAllow, ",")))
	}

	if len(cr.ETFDomicileBlock) > 0 {
		parts = append(parts, fmt.Sprintf("ETF domiciles: %s blocked", strings.Join(cr.ETFDomicileBlock, ",")))
	}

	if len(parts) == 0 {
		return "No compliance restrictions"
	}

	return strings.Join(parts, " | ")
}

// MarshalJSON provides compact JSON representation for AI agents
func (cp CountryPolicy) MarshalJSON() ([]byte, error) {
	compact := map[string]any{
		"allowed_count":    len(cp.Allowed),
		"optional_count":   len(cp.Optional),
		"restricted_count": len(cp.Restricted),
	}

	// include actual lists if reasonable size
	if len(cp.Allowed) <= 10 {
		compact["allowed"] = cp.Allowed
	} else {
		compact["allowed_sample"] = cp.Allowed[:10]
	}

	if len(cp.Optional) <= 10 {
		compact["optional"] = cp.Optional
	} else {
		compact["optional_sample"] = cp.Optional[:10]
	}

	if len(cp.Restricted) <= 10 {
		compact["restricted"] = cp.Restricted
	} else {
		compact["restricted_sample"] = cp.Restricted[:10]
	}

	return json.Marshal(compact)
}

// String provides human-readable policy summary
func (cp CountryPolicy) String() string {
	return fmt.Sprintf("Policy: %d allowed, %d optional, %d restricted",
		len(cp.Allowed), len(cp.Optional), len(cp.Restricted))
}

// MarshalJSON provides compact JSON representation for AI agents
func (cc CountryConfig) MarshalJSON() ([]byte, error) {
	policyData, err := json.Marshal(cc.Policy)
	if err != nil {
		return nil, err
	}

	var policy map[string]any
	if err := json.Unmarshal(policyData, &policy); err != nil {
		return nil, err
	}

	compact := map[string]any{
		"country": cc.Country,
		"policy":  policy,
	}

	return json.Marshal(compact)
}

// String provides human-readable country config summary
func (cc CountryConfig) String() string {
	return fmt.Sprintf("%s: %s", cc.Country, cc.Policy.String())
}

// MarshalJSON provides compact JSON representation for AI agents
func (ar AssetRestriction) MarshalJSON() ([]byte, error) {
	compact := map[string]any{
		"asset_type":    ar.AssetType,
		"is_allowed":    ar.IsAllowed,
		"is_restricted": ar.IsRestricted,
	}

	// include notes if meaningful and not too long
	if len(ar.Notes) > 0 {
		if len(ar.Notes) == 1 {
			compact["note"] = ar.Notes[0]
		} else if len(ar.Notes) <= 3 {
			compact["notes"] = ar.Notes
		} else {
			compact["notes"] = ar.Notes[:3]
			compact["note_count"] = len(ar.Notes)
		}
	}

	return json.Marshal(compact)
}

// String provides human-readable asset restriction summary
func (ar AssetRestriction) String() string {
	var status string
	if ar.IsAllowed && !ar.IsRestricted {
		status = "allowed"
	} else if !ar.IsAllowed && ar.IsRestricted {
		status = "restricted"
	} else if ar.IsAllowed && ar.IsRestricted {
		status = "conditional"
	} else {
		status = "not specified"
	}

	noteCount := ""
	if len(ar.Notes) > 0 {
		noteCount = fmt.Sprintf(" (%d notes)", len(ar.Notes))
	}

	return fmt.Sprintf("%s: %s%s", ar.AssetType, status, noteCount)
}
