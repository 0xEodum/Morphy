package dawg

// WordForm stores paradigm information for a word form.
type WordForm struct {
	ParadigmID uint16
	FormIndex  uint16
}

// WordsDawg stores word forms and allows lookup by word.
type WordsDawg struct {
	*DAWG[WordForm]
}

// NewWordsDawg creates a WordsDawg instance from the provided data map.
func NewWordsDawg(data map[string][]WordForm) *WordsDawg {
	return &WordsDawg{New[WordForm](data)}
}

// Lookup returns paradigm records for a word.
func (w *WordsDawg) Lookup(word string) []WordForm {
	return w.Items(word)
}

// WordItem couples a dictionary word with its paradigm records.
type WordItem struct {
	Word  string
	Forms []WordForm
}

// SimilarItems returns all dictionary entries reachable from the given word by
// applying character substitutions.
func (w *WordsDawg) SimilarItems(word string, subs map[rune]rune) []WordItem {
	items := w.DAWG.SimilarItems(word, subs)
	res := make([]WordItem, 0, len(items))
	for k, v := range items {
		cp := make([]WordForm, len(v))
		copy(cp, v)
		res = append(res, WordItem{Word: k, Forms: cp})
	}
	return res
}

// SimilarItemValues returns paradigm records for all similar words.
func (w *WordsDawg) SimilarItemValues(word string, subs map[rune]rune) [][]WordForm {
	values := w.DAWG.SimilarItemValues(word, subs)
	res := make([][]WordForm, len(values))
	for i, v := range values {
		cp := make([]WordForm, len(v))
		copy(cp, v)
		res[i] = cp
	}
	return res
}
