package units

import (
	"strings"

	"morphy/pkg/analysis"
	"morphy/pkg/tagset"
)

// InitialsAnalyzer handles single-letter abbreviations.
type InitialsAnalyzer struct {
	BaseAnalyzerUnit
	letters    string
	tagPattern string
	score      float64
	letterSet  map[string]struct{}
	tags       []tagset.Tag
}

type abbrMethod struct{ Analyzer AnalyzerUnit }

func (m abbrMethod) Unit() AnalyzerUnit { return m.Analyzer }

// NewInitialsAnalyzer creates analyzer for given letters and tag pattern.
func NewInitialsAnalyzer(letters, pattern string, score float64) *InitialsAnalyzer {
	return &InitialsAnalyzer{letters: letters, tagPattern: pattern, score: score}
}

func (a *InitialsAnalyzer) Init(morph Analyzer) {
	a.BaseAnalyzerUnit.Init(morph)
	if a.tagPattern == "" {
		a.tagPattern = "NOUN,anim,%[gender]s,Sgtm,Fixd,Abbr,Init sing,%[case]s"
	}
	tagset.AddGrammemeToKnown("Init", "иниц", false)
	a.letterSet = make(map[string]struct{})
	for _, r := range a.letters {
		a.letterSet[string(r)] = struct{}{}
	}
	genders := []string{"masc", "femn"}
	cases := []string{"nomn", "gent", "datv", "accs", "ablt", "loct"}
	for _, g := range genders {
		for _, c := range cases {
			t, _ := tagset.New(strings.ReplaceAll(strings.ReplaceAll(a.tagPattern, "%[gender]s", g), "%[case]s", c))
			a.tags = append(a.tags, *t)
		}
	}
}

func (a *InitialsAnalyzer) Parse(word, wordLower string, seen map[string]struct{}) []analysis.Parse {
	if _, ok := a.letterSet[word]; !ok {
		return nil
	}
	res := make([]analysis.Parse, 0, len(a.tags))
	method := abbrMethod{Analyzer: a}
	for _, t := range a.tags {
		p := analysis.NewParse(wordLower, &t, wordLower, a.score, []interface{}{method})
		res = append(res, p)
	}
	return res
}

func (a *InitialsAnalyzer) Tag(word, wordLower string, seen map[string]struct{}) []tagset.Tag {
	if _, ok := a.letterSet[word]; !ok {
		return nil
	}
	return append([]tagset.Tag(nil), a.tags...)
}

func (a *InitialsAnalyzer) GetLexeme(form analysis.Parse) []analysis.Parse {
	return []analysis.Parse{form}
}

func (a *InitialsAnalyzer) Normalized(form analysis.Parse) analysis.Parse { return form }

// AbbreviatedFirstNameAnalyzer handles first name initials.
type AbbreviatedFirstNameAnalyzer struct {
	InitialsAnalyzer
	tagsMasc []tagset.Tag
	tagsFemn []tagset.Tag
}

func NewAbbreviatedFirstNameAnalyzer(letters string) *AbbreviatedFirstNameAnalyzer {
	return &AbbreviatedFirstNameAnalyzer{InitialsAnalyzer: *NewInitialsAnalyzer(letters, "NOUN,anim,%[gender]s,Sgtm,Name,Fixd,Abbr,Init sing,%[case]s", 0.1)}
}

func (a *AbbreviatedFirstNameAnalyzer) Init(morph Analyzer) {
	a.InitialsAnalyzer.Init(morph)
	for _, t := range a.tags {
		if ok, _ := t.Contains("masc"); ok {
			a.tagsMasc = append(a.tagsMasc, t)
		} else {
			a.tagsFemn = append(a.tagsFemn, t)
		}
	}
}

func (a *AbbreviatedFirstNameAnalyzer) GetLexeme(form analysis.Parse) []analysis.Parse {
	var tags []tagset.Tag
	if ok, _ := form.Tag.Contains("masc"); ok {
		tags = a.tagsMasc
	} else {
		tags = a.tagsFemn
	}
	res := make([]analysis.Parse, 0, len(tags))
	for _, t := range tags {
		res = append(res, analysis.NewParse(form.Word, &t, form.NormalForm, form.Score, form.MethodsStack))
	}
	return res
}

func (a *AbbreviatedFirstNameAnalyzer) Normalized(form analysis.Parse) analysis.Parse {
	tags := a.tagsMasc
	if ok, _ := form.Tag.Contains("masc"); !ok {
		tags = a.tagsFemn
	}
	return analysis.NewParse(form.Word, &tags[0], form.NormalForm, form.Score, form.MethodsStack)
}

// AbbreviatedPatronymicAnalyzer handles patronymic initials.
type AbbreviatedPatronymicAnalyzer struct{ InitialsAnalyzer }

func NewAbbreviatedPatronymicAnalyzer(letters string) *AbbreviatedPatronymicAnalyzer {
	return &AbbreviatedPatronymicAnalyzer{InitialsAnalyzer: *NewInitialsAnalyzer(letters, "NOUN,anim,%[gender]s,Sgtm,Patr,Fixd,Abbr,Init sing,%[case]s", 0.1)}
}

func (a *AbbreviatedPatronymicAnalyzer) Init(morph Analyzer) {
	a.InitialsAnalyzer.Init(morph)
	tagset.AddGrammemeToKnown("Patr", "отч", false)
}

func (a *AbbreviatedPatronymicAnalyzer) GetLexeme(form analysis.Parse) []analysis.Parse {
	res := make([]analysis.Parse, 0, len(a.tags))
	for _, t := range a.tags {
		res = append(res, analysis.NewParse(form.Word, &t, form.NormalForm, form.Score, form.MethodsStack))
	}
	return res
}

func (a *AbbreviatedPatronymicAnalyzer) Normalized(form analysis.Parse) analysis.Parse {
	return analysis.NewParse(form.Word, &a.tags[0], form.NormalForm, form.Score, form.MethodsStack)
}
