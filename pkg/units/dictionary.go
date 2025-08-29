package units

import (
	"morphy/pkg/analysis"
	"morphy/pkg/dict"
	"morphy/pkg/tagset"
)

// DictionaryAnalyzer analyzes words using a dictionary lookup.
type DictionaryAnalyzer struct {
	BaseAnalyzerUnit
}

// Parse a word using the dictionary.
func (d *DictionaryAnalyzer) Parse(word, wordLower string, seenParses map[string]struct{}) []analysis.Parse {
	dictionary, ok := d.Dict.(*dict.Dictionary)
	if !ok {
		return nil
	}
	subs := map[rune]rune{}
	if d.Morph != nil {
		subs = d.Morph.CharSubstitutes()
	}
	res := []analysis.Parse{}
	items := dictionary.Words().SimilarItems(wordLower, subs)
	for _, it := range items {
		for _, wf := range it.Forms {
			tag := dictionary.BuildTagInfo(int(wf.ParadigmID), int(wf.FormIndex))
			normal := dictionary.BuildNormalForm(int(wf.ParadigmID), int(wf.FormIndex), it.Word)
			method := dictMethod{Analyzer: d, Word: it.Word, ParaID: int(wf.ParadigmID), Index: int(wf.FormIndex)}
			parse := analysis.NewParse(it.Word, &tag, normal, 1.0, []interface{}{method})
			AddParseIfNotSeen(parse, &res, seenParses)
		}
	}
	return res
}

// Tag a word using the dictionary.
func (d *DictionaryAnalyzer) Tag(word, wordLower string, seenTags map[string]struct{}) []tagset.Tag {
	dictionary, ok := d.Dict.(*dict.Dictionary)
	if !ok {
		return nil
	}
	subs := map[rune]rune{}
	if d.Morph != nil {
		subs = d.Morph.CharSubstitutes()
	}
	res := []tagset.Tag{}
	values := dictionary.Words().SimilarItemValues(wordLower, subs)
	for _, forms := range values {
		for _, wf := range forms {
			tag := dictionary.BuildTagInfo(int(wf.ParadigmID), int(wf.FormIndex))
			AddTagIfNotSeen(tag, &res, seenTags)
		}
	}
	return res
}

// GetLexeme returns the lexeme for a parsed word.
func (d *DictionaryAnalyzer) GetLexeme(p analysis.Parse) []analysis.Parse {
	dictionary, ok := d.Dict.(*dict.Dictionary)
	if !ok {
		return nil
	}
	fixedWord, paraID, idx := d.extractParaInfo(p.MethodsStack)
	paradigm := dictionary.BuildParadigmInfo(paraID)
	current := paradigm[idx]
	stem := fixedWord
	if len(current.Prefix) > 0 {
		stem = stem[len(current.Prefix):]
	}
	if len(current.Suffix) > 0 {
		stem = stem[:len(stem)-len(current.Suffix)]
	}
	res := make([]analysis.Parse, 0, len(paradigm))
	for i, form := range paradigm {
		word := form.Prefix + stem + form.Suffix
		newStack := d.fixStack(p.MethodsStack, word, paraID, i)
		parse := analysis.NewParse(word, &form.Tag, p.NormalForm, 1.0, newStack)
		res = append(res, parse)
	}
	return res
}

// Normalized returns the normal form of a parsed word.
func (d *DictionaryAnalyzer) Normalized(p analysis.Parse) analysis.Parse {
	dictionary, ok := d.Dict.(*dict.Dictionary)
	if !ok {
		return p
	}
	_, paraID, idx := d.extractParaInfo(p.MethodsStack)
	if idx == 0 {
		return p
	}
	normal := p.NormalForm
	tag := dictionary.BuildTagInfo(paraID, 0)
	newStack := d.fixStack(p.MethodsStack, normal, paraID, 0)
	return analysis.NewParse(normal, &tag, normal, 1.0, newStack)
}

type dictMethod struct {
	Analyzer *DictionaryAnalyzer
	Word     string
	ParaID   int
	Index    int
}

func (m dictMethod) Unit() AnalyzerUnit { return m.Analyzer }

func (d *DictionaryAnalyzer) extractParaInfo(stack []interface{}) (string, int, int) {
	method := stack[0].(dictMethod)
	return method.Word, method.ParaID, method.Index
}

func (d *DictionaryAnalyzer) fixStack(stack []interface{}, word string, paraID, idx int) []interface{} {
	method0 := dictMethod{Analyzer: d, Word: word, ParaID: paraID, Index: idx}
	newStack := make([]interface{}, len(stack))
	newStack[0] = method0
	copy(newStack[1:], stack[1:])
	return newStack
}

// Clone returns a copy of analyzer.
func (d *DictionaryAnalyzer) Clone() AnalyzerUnit {
	cloned := *d
	return &cloned
}
