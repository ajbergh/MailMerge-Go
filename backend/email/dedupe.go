package email

// Policy controls how duplicate recipients (by normalized address) are resolved.
type Policy string

const (
	// PolicyKeepFirst keeps the first occurrence of each address (default).
	PolicyKeepFirst Policy = "keep_first"
	// PolicyKeepLast keeps the last occurrence of each address.
	PolicyKeepLast Policy = "keep_last"
	// PolicyExcludeAll drops every recipient that has any duplicate.
	PolicyExcludeAll Policy = "exclude_all"
	// PolicyKeepAll keeps every recipient (send more than once).
	PolicyKeepAll Policy = "keep_all"
	// PolicyManual defers resolution to the user; ResolveDuplicates keeps first
	// and flags the groups so the UI can prompt.
	PolicyManual Policy = "manual"
)

// DefaultPolicy is applied when none is specified.
const DefaultPolicy = PolicyKeepFirst

// Valid reports whether p is a recognized policy.
func (p Policy) Valid() bool {
	switch p {
	case PolicyKeepFirst, PolicyKeepLast, PolicyExcludeAll, PolicyKeepAll, PolicyManual:
		return true
	}
	return false
}

// Group is a set of recipient indices that share a normalized address.
type Group struct {
	Key     string `json:"key"`     // normalized address
	Indices []int  `json:"indices"` // positions in the original recipient list
}

// DedupeResult is the outcome of applying a Policy to a recipient list.
type DedupeResult struct {
	// Keep is the ordered list of original indices to send to.
	Keep []int `json:"keep"`
	// Dropped is the ordered list of original indices excluded by the policy.
	Dropped []int `json:"dropped"`
	// Groups contains only the groups that actually had duplicates.
	Groups []Group `json:"groups"`
	// NeedsManualReview is true when Policy is PolicyManual and duplicates exist.
	NeedsManualReview bool `json:"needsManualReview"`
}

// ResolveDuplicates groups emails by normalized key (preserving first-seen order)
// and applies the policy, returning the indices to keep and drop plus the
// duplicate groups for display. Counts should be computed from Keep, before
// preflight.
func ResolveDuplicates(emails []string, policy Policy) DedupeResult {
	if !policy.Valid() {
		policy = DefaultPolicy
	}

	// Build groups preserving first-seen order of keys.
	order := []string{}
	groups := map[string][]int{}
	for i, e := range emails {
		key := NormalizeKey(e)
		if _, ok := groups[key]; !ok {
			order = append(order, key)
		}
		groups[key] = append(groups[key], i)
	}

	res := DedupeResult{}
	keep := map[int]bool{}

	for _, key := range order {
		idx := groups[key]
		if len(idx) == 1 {
			keep[idx[0]] = true
			continue
		}
		// Duplicate group.
		res.Groups = append(res.Groups, Group{Key: key, Indices: append([]int(nil), idx...)})
		switch policy {
		case PolicyKeepAll:
			for _, i := range idx {
				keep[i] = true
			}
		case PolicyKeepLast:
			keep[idx[len(idx)-1]] = true
		case PolicyExcludeAll:
			// keep none
		case PolicyManual:
			// Interim: keep first, but flag for user review.
			keep[idx[0]] = true
			res.NeedsManualReview = true
		default: // PolicyKeepFirst
			keep[idx[0]] = true
		}
	}

	for i := range emails {
		if keep[i] {
			res.Keep = append(res.Keep, i)
		} else {
			res.Dropped = append(res.Dropped, i)
		}
	}
	return res
}
