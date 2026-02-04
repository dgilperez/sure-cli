package rules

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dgilperez/sure-cli/internal/models"
)

type RuleProposal struct {
	Type            string   `json:"type"`    // "category" | "tag" | "merchant"
	Pattern         string   `json:"pattern"` // merchant name or pattern
	Action          string   `json:"action"`  // e.g. "set_category", "add_tag"
	Value           string   `json:"value"`   // category name or tag
	Confidence      float64  `json:"confidence"`
	Reason          string   `json:"reason"`
	AffectedCount   int      `json:"affected_count"`
	SampleTxIDs     []string `json:"sample_tx_ids"`
	SuggestedAction string   `json:"suggested_action"`
}

type ProposeResult struct {
	Proposals     []RuleProposal `json:"proposals"`
	TotalTx       int            `json:"total_transactions"`
	Uncategorized int            `json:"uncategorized_count"`
}

// ProposeRules analyzes transactions and suggests categorization rules.
// Heuristics:
// - Group by merchant name
// - If a merchant always has the same category, suggest a rule
// - If a merchant is uncategorized but similar names have categories, suggest
func ProposeRules(txs []models.Transaction) ProposeResult {
	// Group by merchant/name
	byName := make(map[string][]models.Transaction)
	for _, tx := range txs {
		name := strings.TrimSpace(tx.Name)
		if name == "" {
			continue
		}
		byName[name] = append(byName[name], tx)
	}

	var proposals []RuleProposal
	var uncategorized int

	// Analyze each merchant group
	for name, txList := range byName {
		if len(txList) < 2 {
			continue // need at least 2 occurrences
		}

		// Count categories
		catCounts := make(map[string]int)
		var hasCategory bool
		for _, tx := range txList {
			cat := tx.CategoryName
			if cat == "" {
				uncategorized++
			} else {
				hasCategory = true
				catCounts[cat]++
			}
		}

		if !hasCategory {
			// All uncategorized - can't infer
			continue
		}

		// Find dominant category
		var dominantCat string
		var dominantCount int
		for cat, count := range catCounts {
			if count > dominantCount {
				dominantCat = cat
				dominantCount = count
			}
		}

		// Calculate confidence based on consistency
		consistency := float64(dominantCount) / float64(len(txList))
		if consistency < 0.7 {
			continue // not consistent enough
		}

		// Count how many would be affected (currently not matching)
		affected := 0
		var sampleIDs []string
		for _, tx := range txList {
			if tx.CategoryName != dominantCat {
				affected++
				if len(sampleIDs) < 3 {
					sampleIDs = append(sampleIDs, tx.ID)
				}
			}
		}

		if affected == 0 {
			continue // already consistent
		}

		conf := 0.6 + (consistency * 0.3)
		if len(txList) >= 5 {
			conf += 0.1
		}
		if conf > 1 {
			conf = 1
		}

		proposals = append(proposals, RuleProposal{
			Type:            "category",
			Pattern:         name,
			Action:          "set_category",
			Value:           dominantCat,
			Confidence:      conf,
			Reason:          "consistent_categorization",
			AffectedCount:   affected,
			SampleTxIDs:     sampleIDs,
			SuggestedAction: fmt.Sprintf("Review and apply: would categorize %d transactions as %s", affected, dominantCat),
		})
	}

	// Sort by confidence then affected count
	sort.Slice(proposals, func(i, j int) bool {
		if proposals[i].Confidence == proposals[j].Confidence {
			return proposals[i].AffectedCount > proposals[j].AffectedCount
		}
		return proposals[i].Confidence > proposals[j].Confidence
	})

	// Limit to top 20
	if len(proposals) > 20 {
		proposals = proposals[:20]
	}

	return ProposeResult{
		Proposals:     proposals,
		TotalTx:       len(txs),
		Uncategorized: uncategorized,
	}
}
