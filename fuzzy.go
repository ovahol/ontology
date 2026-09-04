package ontology

import "strings"

// This file adds typo-tolerant (fuzzy) matching for identity fields such as
// device names and EMDN terms. Exact matching stays the primary, highest
// confidence path; fuzzy matching is a secondary fallback that lets a
// misspelled/malformed name still reconcile to the correct dictionary row.
//
// The algorithm is token-aware: two names are "close" when their normalized
// word sequences align, allowing a bounded number of insertions/substitutions
// per token (Levenshtein distance). It is deliberately conservative so it does
// not pull unrelated entries in, and it is fully system-agnostic — it compares
// whatever name/term strings the catalog holds.

// editThreshold returns the maximum Levenshtein distance allowed between two
// tokens for them to be considered a match. Longer tokens tolerate more typos.
func editThreshold(tt string) int {
	return 1 + len(tt)/4
}

// editDistance returns the Levenshtein distance between a and b (case
// matters; callers should pass already-normalized lowercase tokens).
func editDistance(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost
			m := del
			if ins < m {
				m = ins
			}
			if sub < m {
				m = sub
			}
			curr[j] = m
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

// fuzzyScore reports how well the query token sequence aligns with the
// candidate token sequence, as a fraction in [0,1]. A score of 1 means every
// candidate token matched a distinct query token within the typo budget.
func fuzzyScore(query, candidate []string) float64 {
	if len(query) == 0 || len(candidate) == 0 {
		return 0
	}
	used := make([]bool, len(query))
	var total, matched float64
	for _, ct := range candidate {
		w := float64(len(ct))
		total += w
		bestDist := -1
		best := -1
		for i, qt := range query {
			if used[i] {
				continue
			}
			d := editDistance(qt, ct)
			if d <= editThreshold(ct) && (bestDist == -1 || d < bestDist) {
				bestDist = d
				best = i
			}
		}
		if best >= 0 {
			used[best] = true
			matched += w
		}
	}
	if total == 0 {
		return 0
	}
	return matched / total
}

// fuzzyNameMatch reports whether query (a possibly-misspelled device name or
// term) is close enough to candidate to be treated as the same entry. It
// requires a solid token-coverage score and a name long enough to be
// meaningful, so short inputs (e.g. "bed") do not spuriously match.
func fuzzyNameMatch(query, candidate string) bool {
	q := strings.Fields(Normalized(query))
	c := strings.Fields(Normalized(candidate))
	if len(q) == 0 || len(c) == 0 {
		return false
	}
	score := fuzzyScore(q, c)
	// Require meaningful coverage (>=70%) and enough query substance.
	return score >= 0.70 && len(Normalized(query)) >= 4
}
