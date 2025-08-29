package units

import (
	"morphy/pkg/analysis"
	"morphy/pkg/tagset"
)

// Analyzer represents a morphological analyzer able to provide dictionary access.
type Analyzer interface {
	Dictionary() Dictionary
	Parse(word string) []analysis.Parse
	Tag(word string) []tagset.Tag
	CharSubstitutes() map[rune]rune
}

// Dictionary is a placeholder for dictionary implementation.
type Dictionary interface{}

// AnalyzerUnit is a basic interface for analyzer units.
type AnalyzerUnit interface {
	Init(morph Analyzer)
	Parse(word, wordLower string, seenParses map[string]struct{}) []analysis.Parse
	Tag(word, wordLower string, seenTags map[string]struct{}) []tagset.Tag
	Normalized(p analysis.Parse) analysis.Parse
	GetLexeme(p analysis.Parse) []analysis.Parse
	Clone() AnalyzerUnit
}

// Method represents an entry in methods stack.
type Method interface {
	Unit() AnalyzerUnit
}

// BaseAnalyzerUnit contains common fields for analyzer units.
type BaseAnalyzerUnit struct {
	Morph Analyzer
	Dict  Dictionary
}

// Init saves references to the morph analyzer and dictionary.
func (u *BaseAnalyzerUnit) Init(morph Analyzer) {
	u.Morph = morph
	if morph != nil {
		u.Dict = morph.Dictionary()
	}
}

// Clone returns a shallow copy of the unit.
func (u *BaseAnalyzerUnit) Clone() AnalyzerUnit {
	cloned := *u
	return &cloned
}

// Parse must be implemented by subclasses.
func (u *BaseAnalyzerUnit) Parse(word, wordLower string, seenParses map[string]struct{}) []analysis.Parse {
	panic("Parse not implemented")
}

// Tag returns all unique tags for the word using Parse.
func (u *BaseAnalyzerUnit) Tag(word, wordLower string, seenTags map[string]struct{}) []tagset.Tag {
	parses := u.Parse(word, wordLower, map[string]struct{}{})
	res := make([]tagset.Tag, 0, len(parses))
	for _, p := range parses {
		tagStr := p.Tag.String()
		if _, ok := seenTags[tagStr]; ok {
			continue
		}
		seenTags[tagStr] = struct{}{}
		res = append(res, *p.Tag)
	}
	return res
}

// Normalized must be implemented by subclasses.
func (u *BaseAnalyzerUnit) Normalized(p analysis.Parse) analysis.Parse {
	panic("Normalized not implemented")
}

// GetLexeme must be implemented by subclasses.
func (u *BaseAnalyzerUnit) GetLexeme(p analysis.Parse) []analysis.Parse {
	panic("GetLexeme not implemented")
}
