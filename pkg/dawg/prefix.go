package dawg

// PrefixMatcher provides utilities to test word prefixes.
type PrefixMatcher struct {
	prefixes []string
}

// NewPrefixMatcher creates a new matcher from given prefixes.
func NewPrefixMatcher(prefixes []string) *PrefixMatcher {
	return &PrefixMatcher{prefixes: prefixes}
}

// IsPrefixed returns true if word starts with any prefix.
func (pm *PrefixMatcher) IsPrefixed(word string) bool {
	for _, p := range pm.prefixes {
		if len(word) >= len(p) && word[:len(p)] == p {
			return true
		}
	}
	return false
}

// Prefixes returns all prefixes that match start of word.
func (pm *PrefixMatcher) Prefixes(word string) []string {
	res := []string{}
	for _, p := range pm.prefixes {
		if len(word) >= len(p) && word[:len(p)] == p {
			res = append(res, p)
		}
	}
	return res
}
