package analyzer

import (
	"testing"

	"morphy/pkg/analysis"
	"morphy/pkg/tagset"
	"morphy/pkg/units"
)

type dummyUnit struct {
	units.BaseAnalyzerUnit
	lexeme []analysis.Parse
}

func (u *dummyUnit) Parse(word, wordLower string, seen map[string]struct{}) []analysis.Parse {
	return nil
}
func (u *dummyUnit) Normalized(p analysis.Parse) analysis.Parse  { return p }
func (u *dummyUnit) GetLexeme(p analysis.Parse) []analysis.Parse { return u.lexeme }
func (u *dummyUnit) Clone() units.AnalyzerUnit                   { return u }

type dummyMethod struct{ u *dummyUnit }

func (m dummyMethod) Unit() units.AnalyzerUnit { return m.u }

func TestInflectAndAgree(t *testing.T) {
	tagSingNomn, _ := tagset.New("NOUN,anim,femn sing,nomn")
	tagSingGent, _ := tagset.New("NOUN,anim,femn sing,gent")
	tagPlurNomn, _ := tagset.New("NOUN,anim,femn plur,nomn")
	tagPlurGent, _ := tagset.New("NOUN,anim,femn plur,gent")
	du := &dummyUnit{}
	method := dummyMethod{u: du}
	lexeme := []analysis.Parse{
		analysis.NewParse("мама", tagSingNomn, "мама", 1.0, []interface{}{method}),
		analysis.NewParse("мамы", tagSingGent, "мама", 1.0, []interface{}{method}),
		analysis.NewParse("мамы", tagPlurNomn, "мама", 1.0, []interface{}{method}),
		analysis.NewParse("мам", tagPlurGent, "мама", 1.0, []interface{}{method}),
	}
	du.lexeme = lexeme
	base := lexeme[0]
	m := &MorphAnalyzer{}
	res, ok := m.Inflect(base, []string{"plur", "gent"})
	if !ok || res.Word != "мам" {
		t.Fatalf("expected мам, got %v ok=%v", res.Word, ok)
	}
	res2, ok := m.MakeAgreeWithNumber(base, 3)
	if !ok || res2.Word != "мамы" {
		t.Fatalf("expected мамы, got %v ok=%v", res2.Word, ok)
	}
}
