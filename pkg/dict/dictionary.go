package dict

import (
	"strings"

	"morphy/pkg/dawg"
	"morphy/pkg/tagset"
)

// Dictionary is a wrapper around loaded dictionary data.
type Dictionary struct {
	paradigms        [][]uint16
	gramtab          []tagset.Tag
	paradigmPrefixes []string
	suffixes         []string
	words            *dawg.WordsDawg
	predictionDAWGs  []*dawg.PredictionSuffixesDAWG
	meta             map[string]any
	path             string
}

// NewDictionary loads dictionary from path.
func NewDictionary(path string) (*Dictionary, error) {
	ld, err := LoadDict(path)
	if err != nil {
		return nil, err
	}
	return &Dictionary{
		paradigms:        ld.Paradigms,
		gramtab:          ld.Gramtab,
		paradigmPrefixes: ld.ParadigmPrefixes,
		suffixes:         ld.Suffixes,
		words:            ld.Words,
		predictionDAWGs:  ld.PredictionSuffixes,
		meta:             ld.Meta,
		path:             path,
	}, nil
}

// BuildTagInfo returns tag for given paradigm and form index.
func (d *Dictionary) BuildTagInfo(paraID int, idx int) tagset.Tag {
	paradigm := d.paradigms[paraID]
	n := len(paradigm) / 3
	tagID := paradigm[n+idx]
	return d.gramtab[tagID]
}

// ParadigmForm represents single form info.
type ParadigmForm struct {
	Prefix string
	Tag    tagset.Tag
	Suffix string
}

// BuildParadigmInfo returns paradigm description for paraID.
func (d *Dictionary) BuildParadigmInfo(paraID int) []ParadigmForm {
	paradigm := d.paradigms[paraID]
	n := len(paradigm) / 3
	res := make([]ParadigmForm, n)
	for i := 0; i < n; i++ {
		pref := d.paradigmPrefixes[paradigm[2*n+i]]
		suff := d.suffixes[paradigm[i]]
		tag := d.gramtab[paradigm[n+i]]
		res[i] = ParadigmForm{Prefix: pref, Tag: tag, Suffix: suff}
	}
	return res
}

// BuildNormalForm constructs a normal form for word in paradigm.
func (d *Dictionary) BuildNormalForm(paraID, idx int, fixedWord string) string {
	if idx == 0 {
		return fixedWord
	}
	paradigm := d.paradigms[paraID]
	n := len(paradigm) / 3
	stem := d.BuildStem(paradigm, idx, fixedWord)
	npref := d.paradigmPrefixes[paradigm[2*n+0]]
	nsuff := d.suffixes[paradigm[0]]
	return npref + stem + nsuff
}

// BuildStem returns stem of word according to paradigm.
func (d *Dictionary) BuildStem(paradigm []uint16, idx int, fixedWord string) string {
	n := len(paradigm) / 3
	pref := d.paradigmPrefixes[paradigm[2*n+idx]]
	suff := d.suffixes[paradigm[idx]]
	if len(suff) > 0 {
		return fixedWord[len(pref) : len(fixedWord)-len(suff)]
	}
	return fixedWord[len(pref):]
}

// Words returns underlying words DAWG.
func (d *Dictionary) Words() *dawg.WordsDawg { return d.words }

// PredictionSuffixes returns DAWGs used for suffix prediction.
func (d *Dictionary) PredictionSuffixes() []*dawg.PredictionSuffixesDAWG {
	return d.predictionDAWGs
}

// ParadigmPrefixes returns paradigm prefixes list.
func (d *Dictionary) ParadigmPrefixes() []string { return d.paradigmPrefixes }

// MaxSuffixLength retrieves maximum suffix length used for predictions from meta.
func (d *Dictionary) MaxSuffixLength() int {
	if opts, ok := d.meta["compile_options"].(map[string]any); ok {
		if v, ok := opts["max_suffix_length"].(float64); ok {
			return int(v)
		}
	}
	if opts, ok := d.meta["prediction_options"].(map[string]any); ok {
		if v, ok := opts["max_suffix_length"].(float64); ok {
			return int(v)
		}
	}
	return 0
}

// WordIsKnown reports whether word is present in dictionary, accounting for
// character substitutes.
func (d *Dictionary) WordIsKnown(word string, subs map[rune]rune) bool {
	if len(d.words.Lookup(word)) > 0 {
		return true
	}
	if len(subs) > 0 {
		vals := d.words.SimilarItemValues(word, subs)
		return len(vals) > 0
	}
	return false
}

// KnownWord holds information returned by IterKnownWords.
type KnownWord struct {
	Word       string
	Tag        tagset.Tag
	NormalForm string
	ParadigmID uint16
	Index      uint16
}

// IterKnownWords returns slice of known words with prefix.
func (d *Dictionary) IterKnownWords(prefix string) []KnownWord {
	res := []KnownWord{}
	for word, forms := range d.words.Data() {
		if !strings.HasPrefix(word, prefix) {
			continue
		}
		for _, wf := range forms {
			tag := d.BuildTagInfo(int(wf.ParadigmID), int(wf.FormIndex))
			normal := d.BuildNormalForm(int(wf.ParadigmID), int(wf.FormIndex), word)
			res = append(res, KnownWord{Word: word, Tag: tag, NormalForm: normal, ParadigmID: wf.ParadigmID, Index: wf.FormIndex})
		}
	}
	return res
}
