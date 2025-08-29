package units

import (
	"morphy/pkg/analysis"
	"morphy/pkg/tagset"
)

// UnknAnalyzer adds UNKN parse when other analyzers return nothing.
type UnknAnalyzer struct {
	BaseAnalyzerUnit
	tag   *tagset.Tag
	score float64
}

type unknMethod struct{ Analyzer *UnknAnalyzer }

func (m unknMethod) Unit() AnalyzerUnit { return m.Analyzer }

// NewUnknAnalyzer creates a new unknown analyzer with default score.
func NewUnknAnalyzer() *UnknAnalyzer {
	return &UnknAnalyzer{score: 1.0}
}

// Init prepares analyzer and registers grammeme.
func (u *UnknAnalyzer) Init(morph Analyzer) {
	u.BaseAnalyzerUnit.Init(morph)
	tagset.AddGrammemeToKnown("UNKN", "НЕИЗВ", false)
	t, _ := tagset.New("UNKN")
	u.tag = t
}

// Parse returns UNKN parse if no other parses were found.
func (u *UnknAnalyzer) Parse(word, wordLower string, seenParses map[string]struct{}) []analysis.Parse {
	if len(seenParses) > 0 {
		return nil
	}
	m := unknMethod{Analyzer: u}
	p := analysis.NewParse(wordLower, u.tag, wordLower, u.score, []interface{}{m})
	return []analysis.Parse{p}
}

// Tag returns UNKN tag if no tags were found.
func (u *UnknAnalyzer) Tag(word, wordLower string, seenTags map[string]struct{}) []tagset.Tag {
	if len(seenTags) > 0 {
		return nil
	}
	return []tagset.Tag{*u.tag}
}

// GetLexeme returns the form itself as its only lexeme.
func (u *UnknAnalyzer) GetLexeme(form analysis.Parse) []analysis.Parse { return []analysis.Parse{form} }

// Normalized returns the form unchanged.
func (u *UnknAnalyzer) Normalized(form analysis.Parse) analysis.Parse { return form }
