package dawg

// DAWG is a minimal map-backed replacement for a Directed Acyclic Word Graph.
// It stores a mapping from string keys to a slice of values of generic type T.
// It also provides prefix-based queries used by the morphological analyzer.
type DAWG[T any] struct {
	data map[string][]T
}

// New creates a DAWG instance from the provided data map. The map is used as-is;
// callers should not modify it after passing to New unless such modifications
// are intentional.
func New[T any](data map[string][]T) *DAWG[T] {
	if data == nil {
		data = make(map[string][]T)
	}
	return &DAWG[T]{data: data}
}

// Items returns a copy of values associated with the key.
func (d *DAWG[T]) Items(key string) []T {
	vals, ok := d.data[key]
	if !ok {
		return nil
	}
	res := make([]T, len(vals))
	copy(res, vals)
	return res
}

// Prefixes returns all prefixes of word that exist in the DAWG.
func (d *DAWG[T]) Prefixes(word string) []string {
	res := make([]string, 0)
	for i := 1; i <= len(word); i++ {
		if _, ok := d.data[word[:i]]; ok {
			res = append(res, word[:i])
		}
	}
	return res
}

// IsPrefixed reports whether word has at least one prefix stored in the DAWG.
func (d *DAWG[T]) IsPrefixed(word string) bool {
	for i := 1; i <= len(word); i++ {
		if _, ok := d.data[word[:i]]; ok {
			return true
		}
	}
	return false
}

// Data returns the underlying map. It is intended for serialization helpers
// and callers should treat the returned map as read-only.
func (d *DAWG[T]) Data() map[string][]T {
	return d.data
}

// SimilarItems returns all stored keys obtainable from the given word by
// applying character substitutions from the provided map. The result maps
// substituted words to their associated values.
func (d *DAWG[T]) SimilarItems(word string, subs map[rune]rune) map[string][]T {
	res := map[string][]T{}
	for _, variant := range generateVariants(word, subs) {
		if vals, ok := d.data[variant]; ok {
			res[variant] = vals
		}
	}
	return res
}

// SimilarItemValues returns values for all words similar to the given one
// according to the substitution map.
func (d *DAWG[T]) SimilarItemValues(word string, subs map[rune]rune) [][]T {
	items := d.SimilarItems(word, subs)
	res := make([][]T, 0, len(items))
	for _, vals := range items {
		cp := make([]T, len(vals))
		copy(cp, vals)
		res = append(res, cp)
	}
	return res
}

// SimilarKeys returns words similar to the given one using the substitution map.
func (d *DAWG[T]) SimilarKeys(word string, subs map[rune]rune) []string {
	items := d.SimilarItems(word, subs)
	res := make([]string, 0, len(items))
	for k := range items {
		res = append(res, k)
	}
	return res
}

// generateVariants returns all variants of word by replacing characters using
// substitutions from the map. The original word is always included.
func generateVariants(word string, subs map[rune]rune) []string {
	variants := map[string]struct{}{word: {}}
	runes := []rune(word)
	for i, r := range runes {
		if sub, ok := subs[r]; ok {
			keys := make([]string, 0, len(variants))
			for k := range variants {
				keys = append(keys, k)
			}
			for _, v := range keys {
				rr := []rune(v)
				rr[i] = sub
				variants[string(rr)] = struct{}{}
			}
		}
	}
	res := make([]string, 0, len(variants))
	for v := range variants {
		res = append(res, v)
	}
	return res
}
