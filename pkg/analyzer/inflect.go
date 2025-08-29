package analyzer

import (
	"morphy/pkg/analysis"
	"morphy/pkg/tagset"
)

// Inflect inflects a parsed word to match required grammemes.
func (m *MorphAnalyzer) Inflect(p analysis.Parse, required []string) (analysis.Parse, bool) {
	lexeme := m.GetLexeme(p)
	matches := make([]analysis.Parse, 0, len(lexeme))
	for _, f := range lexeme {
		if containsAll(f.Tag, required) {
			matches = append(matches, f)
		}
	}
	if len(matches) == 0 {
		required = tagset.FixRareCases(required)
		for _, f := range lexeme {
			if containsAll(f.Tag, required) {
				matches = append(matches, f)
			}
		}
	}
	if len(matches) == 0 {
		return analysis.Parse{}, false
	}
	grams, err := p.Tag.UpdatedGrammemes(required)
	if err != nil {
		return analysis.Parse{}, false
	}
	best := matches[0]
	bestScore := similarity(grams, best.Tag.Grammemes())
	for _, cand := range matches[1:] {
		s := similarity(grams, cand.Tag.Grammemes())
		if s > bestScore {
			best = cand
			bestScore = s
		}
	}
	return best, true
}

// MakeAgreeWithNumber inflects the word so it agrees with provided number.
func (m *MorphAnalyzer) MakeAgreeWithNumber(p analysis.Parse, num int) (analysis.Parse, bool) {
	grams := p.Tag.NumeralAgreementGrammemes(num)
	return m.Inflect(p, grams)
}

func containsAll(tag *tagset.Tag, grams []string) bool {
	for _, g := range grams {
		ok, _ := tag.Contains(g)
		if !ok {
			return false
		}
	}
	return true
}

func similarity(a []string, b []string) float64 {
	setA := make(map[string]struct{}, len(a))
	setB := make(map[string]struct{}, len(b))
	for _, g := range a {
		setA[g] = struct{}{}
	}
	for _, g := range b {
		setB[g] = struct{}{}
	}
	inter := 0
	for g := range setA {
		if _, ok := setB[g]; ok {
			inter++
		}
	}
	symdiff := 0
	for g := range setA {
		if _, ok := setB[g]; !ok {
			symdiff++
		}
	}
	for g := range setB {
		if _, ok := setA[g]; !ok {
			symdiff++
		}
	}
	return float64(inter) - 0.1*float64(symdiff)
}
