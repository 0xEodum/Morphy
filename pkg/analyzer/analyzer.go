package analyzer

import (
	"strings"

	"morphy/pkg/analysis"
	"morphy/pkg/dict"
	ru "morphy/pkg/lang/ru"
	"morphy/pkg/tagset"
	"morphy/pkg/units"
)

// unitItem couples analyzer unit with a terminal flag.
type unitItem struct {
	unit     units.AnalyzerUnit
	terminal bool
}

// MorphAnalyzer provides morphological parsing using a set of units.
type MorphAnalyzer struct {
	dict     *dict.Dictionary
	units    []unitItem
	prob     *ProbabilityEstimator
	charSubs map[rune]rune
}

// New creates MorphAnalyzer for dictionary at path with provided units configuration.
// unitsCfg can contain units.AnalyzerUnit or []units.AnalyzerUnit to denote groups.
func New(path string, unitsCfg []interface{}) (*MorphAnalyzer, error) {
	d, err := dict.NewDictionary(path)
	if err != nil {
		return nil, err
	}
	m := &MorphAnalyzer{dict: d, charSubs: ru.CharSubstitutes}
	m.initUnits(unitsCfg)
	if pe, err := NewProbabilityEstimator(path); err == nil {
		m.prob = pe
	}
	return m, nil
}

func (m *MorphAnalyzer) initUnits(cfg []interface{}) {
	if cfg == nil {
		cfg = []interface{}{[]units.AnalyzerUnit{&units.DictionaryAnalyzer{}}, &units.UnknAnalyzer{}}
	}
	for _, item := range cfg {
		switch v := item.(type) {
		case []units.AnalyzerUnit:
			for i, u := range v {
				b := u.Clone()
				b.Init(m)
				term := i == len(v)-1
				m.units = append(m.units, unitItem{unit: b, terminal: term})
			}
		case units.AnalyzerUnit:
			b := v.Clone()
			b.Init(m)
			m.units = append(m.units, unitItem{unit: b, terminal: true})
		}
	}
}

// Dictionary returns underlying dictionary.
func (m *MorphAnalyzer) Dictionary() units.Dictionary { return m.dict }

// CharSubstitutes returns compiled character substitute table.
func (m *MorphAnalyzer) CharSubstitutes() map[rune]rune { return m.charSubs }

// Parse analyzes a word and returns parses.
func (m *MorphAnalyzer) Parse(word string) []analysis.Parse {
	res := []analysis.Parse{}
	seen := map[string]struct{}{}
	wl := strings.ToLower(word)
	for _, it := range m.units {
		res = append(res, it.unit.Parse(word, wl, seen)...)
		if it.terminal && len(res) > 0 {
			break
		}
	}
	if m.prob != nil {
		res = m.prob.ApplyToParses(word, wl, res)
	}
	return res
}

// Tag returns tags for a word.
func (m *MorphAnalyzer) Tag(word string) []tagset.Tag {
	res := []tagset.Tag{}
	seen := map[string]struct{}{}
	wl := strings.ToLower(word)
	for _, it := range m.units {
		res = append(res, it.unit.Tag(word, wl, seen)...)
		if it.terminal && len(res) > 0 {
			break
		}
	}
	if m.prob != nil {
		res = m.prob.ApplyToTags(word, wl, res)
	}
	return res
}

// NormalForms returns list of normal forms for word.
func (m *MorphAnalyzer) NormalForms(word string) []string {
	seen := map[string]struct{}{}
	res := []string{}
	for _, p := range m.Parse(word) {
		if _, ok := seen[p.NormalForm]; !ok {
			seen[p.NormalForm] = struct{}{}
			res = append(res, p.NormalForm)
		}
	}
	return res
}

// GetLexeme returns lexeme for parse.
func (m *MorphAnalyzer) GetLexeme(p analysis.Parse) []analysis.Parse {
	if len(p.MethodsStack) == 0 {
		return []analysis.Parse{p}
	}
	if method, ok := p.MethodsStack[len(p.MethodsStack)-1].(units.Method); ok {
		return method.Unit().GetLexeme(p)
	}
	return []analysis.Parse{p}
}

// Normalized returns normalized form parse.
func (m *MorphAnalyzer) Normalized(p analysis.Parse) analysis.Parse {
	if len(p.MethodsStack) == 0 {
		return p
	}
	if method, ok := p.MethodsStack[len(p.MethodsStack)-1].(units.Method); ok {
		return method.Unit().Normalized(p)
	}
	return p
}

// WordIsKnown checks if word is in dictionary.
func (m *MorphAnalyzer) WordIsKnown(word string) bool {
	return m.dict.WordIsKnown(strings.ToLower(word), m.charSubs)
}

// TagClass parses tag string to Tag.
func (m *MorphAnalyzer) TagClass(tag string) tagset.Tag {
	t, _ := tagset.New(tag)
	return *t
}

// Cyr2Lat transliterates grammemes.
func (m *MorphAnalyzer) Cyr2Lat(s string) string { return tagset.Cyr2Lat(s) }

// Lat2Cyr transliterates grammemes.
func (m *MorphAnalyzer) Lat2Cyr(s string) string { return tagset.Lat2Cyr(s) }
