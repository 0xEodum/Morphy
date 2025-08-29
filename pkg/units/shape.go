package units

import (
	"strconv"
	"strings"

	"morphy/pkg/analysis"
	"morphy/pkg/shapes"
	"morphy/pkg/tagset"
)

// PunctuationAnalyzer tags punctuation marks as PNCT.
type PunctuationAnalyzer struct {
	BaseAnalyzerUnit
	tag   *tagset.Tag
	score float64
}

type simpleMethod struct{ Analyzer AnalyzerUnit }

func (m simpleMethod) Unit() AnalyzerUnit { return m.Analyzer }

// NewPunctuationAnalyzer creates analyzer with default score.
func NewPunctuationAnalyzer() *PunctuationAnalyzer {
	return &PunctuationAnalyzer{score: 0.9}
}

// Init registers grammemes and builds tag.
func (a *PunctuationAnalyzer) Init(morph Analyzer) {
	a.BaseAnalyzerUnit.Init(morph)
	tagset.AddGrammemeToKnown("PNCT", "ЗПР", false)
	t, _ := tagset.New("PNCT")
	a.tag = t
}

// Parse checks if word consists of punctuation characters.
func (a *PunctuationAnalyzer) Parse(word, wordLower string, seenParses map[string]struct{}) []analysis.Parse {
	if !shapes.IsPunctuation(word) {
		return nil
	}
	method := simpleMethod{Analyzer: a}
	p := analysis.NewParse(wordLower, a.tag, wordLower, a.score, []interface{}{method})
	return []analysis.Parse{p}
}

// Tag returns PNCT tag for punctuation tokens.
func (a *PunctuationAnalyzer) Tag(word, wordLower string, seenTags map[string]struct{}) []tagset.Tag {
	if !shapes.IsPunctuation(word) {
		return nil
	}
	return []tagset.Tag{*a.tag}
}

// GetLexeme returns the form itself.
func (a *PunctuationAnalyzer) GetLexeme(form analysis.Parse) []analysis.Parse {
	return []analysis.Parse{form}
}

// Normalized returns the form unchanged.
func (a *PunctuationAnalyzer) Normalized(form analysis.Parse) analysis.Parse { return form }

// LatinAnalyzer marks latin words with LATN tag.
type LatinAnalyzer struct {
	BaseAnalyzerUnit
	tag   *tagset.Tag
	score float64
}

func NewLatinAnalyzer() *LatinAnalyzer { return &LatinAnalyzer{score: 0.9} }

func (a *LatinAnalyzer) Init(morph Analyzer) {
	a.BaseAnalyzerUnit.Init(morph)
	tagset.AddGrammemeToKnown("LATN", "ЛАТ", false)
	t, _ := tagset.New("LATN")
	a.tag = t
}

func (a *LatinAnalyzer) Parse(word, wordLower string, seenParses map[string]struct{}) []analysis.Parse {
	if !shapes.IsLatin(word) {
		return nil
	}
	method := simpleMethod{Analyzer: a}
	p := analysis.NewParse(wordLower, a.tag, wordLower, a.score, []interface{}{method})
	return []analysis.Parse{p}
}

func (a *LatinAnalyzer) Tag(word, wordLower string, seenTags map[string]struct{}) []tagset.Tag {
	if !shapes.IsLatin(word) {
		return nil
	}
	return []tagset.Tag{*a.tag}
}

func (a *LatinAnalyzer) GetLexeme(form analysis.Parse) []analysis.Parse {
	return []analysis.Parse{form}
}
func (a *LatinAnalyzer) Normalized(form analysis.Parse) analysis.Parse { return form }

// NumberAnalyzer marks numbers with NUMB,intg or NUMB,real tags.
type NumberAnalyzer struct {
	BaseAnalyzerUnit
	tags  map[string]*tagset.Tag
	score float64
}

func NewNumberAnalyzer() *NumberAnalyzer { return &NumberAnalyzer{score: 0.9} }

func (a *NumberAnalyzer) Init(morph Analyzer) {
	a.BaseAnalyzerUnit.Init(morph)
	pairs := [][2]string{{"NUMB", "ЧИСЛО"}, {"intg", "цел"}, {"real", "вещ"}}
	for _, p := range pairs {
		tagset.AddGrammemeToKnown(p[0], p[1], false)
	}
	a.tags = make(map[string]*tagset.Tag)
	t1, _ := tagset.New("NUMB,intg")
	t2, _ := tagset.New("NUMB,real")
	a.tags["intg"] = t1
	a.tags["real"] = t2
}

func (a *NumberAnalyzer) checkShape(word string) string {
	if _, err := strconv.Atoi(word); err == nil {
		return "intg"
	}
	if _, err := strconv.ParseFloat(strings.Replace(word, ",", ".", 1), 64); err == nil {
		return "real"
	}
	return ""
}

func (a *NumberAnalyzer) Parse(word, wordLower string, seenParses map[string]struct{}) []analysis.Parse {
	shape := a.checkShape(word)
	if shape == "" {
		return nil
	}
	method := simpleMethod{Analyzer: a}
	p := analysis.NewParse(wordLower, a.tags[shape], wordLower, a.score, []interface{}{method})
	return []analysis.Parse{p}
}

func (a *NumberAnalyzer) Tag(word, wordLower string, seenTags map[string]struct{}) []tagset.Tag {
	shape := a.checkShape(word)
	if shape == "" {
		return nil
	}
	return []tagset.Tag{*a.tags[shape]}
}

func (a *NumberAnalyzer) GetLexeme(form analysis.Parse) []analysis.Parse {
	return []analysis.Parse{form}
}
func (a *NumberAnalyzer) Normalized(form analysis.Parse) analysis.Parse { return form }

// RomanNumberAnalyzer marks Roman numerals with ROMN tag.
type RomanNumberAnalyzer struct {
	BaseAnalyzerUnit
	tag   *tagset.Tag
	score float64
}

func NewRomanNumberAnalyzer() *RomanNumberAnalyzer { return &RomanNumberAnalyzer{score: 0.9} }

func (a *RomanNumberAnalyzer) Init(morph Analyzer) {
	a.BaseAnalyzerUnit.Init(morph)
	tagset.AddGrammemeToKnown("ROMN", "РИМ", false)
	t, _ := tagset.New("ROMN")
	a.tag = t
}

func (a *RomanNumberAnalyzer) Parse(word, wordLower string, seenParses map[string]struct{}) []analysis.Parse {
	if !shapes.IsRomanNumber(word) {
		return nil
	}
	method := simpleMethod{Analyzer: a}
	p := analysis.NewParse(wordLower, a.tag, wordLower, a.score, []interface{}{method})
	return []analysis.Parse{p}
}

func (a *RomanNumberAnalyzer) Tag(word, wordLower string, seenTags map[string]struct{}) []tagset.Tag {
	if !shapes.IsRomanNumber(word) {
		return nil
	}
	return []tagset.Tag{*a.tag}
}

func (a *RomanNumberAnalyzer) GetLexeme(form analysis.Parse) []analysis.Parse {
	return []analysis.Parse{form}
}
func (a *RomanNumberAnalyzer) Normalized(form analysis.Parse) analysis.Parse { return form }
