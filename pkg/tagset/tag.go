package tagset

import (
	"fmt"
	"strings"
)

// Tag represents an OpenCorpora tag.
type Tag struct {
	text      string
	grammemes []string
	gramSet   map[string]struct{}
}

// New creates a Tag from the string representation.
func New(tag string) (*Tag, error) {
	grams := parseTag(tag)
	for _, g := range grams {
		if !GrammemeIsKnown(g) {
			return nil, fmt.Errorf("unknown grammeme: %s", g)
		}
	}
	set := make(map[string]struct{}, len(grams))
	for _, g := range grams {
		set[g] = struct{}{}
	}
	return &Tag{text: tag, grammemes: grams, gramSet: set}, nil
}

func parseTag(tag string) []string {
	tag = strings.ReplaceAll(tag, " ", ",")
	parts := strings.Split(tag, ",")
	res := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			res = append(res, p)
		}
	}
	return res
}

func (t *Tag) String() string { return t.text }

// Grammemes returns a copy of grammemes slice.
func (t *Tag) Grammemes() []string {
	return append([]string(nil), t.grammemes...)
}

func (t *Tag) contains(g string) bool {
	_, ok := t.gramSet[g]
	return ok
}

// Contains checks if grammeme g is in the tag. It returns an error for unknown grammemes.
func (t *Tag) Contains(g string) (bool, error) {
	if t.contains(g) {
		return true, nil
	}
	if !GrammemeIsKnown(g) {
		return false, fmt.Errorf("grammeme is unknown: %s", g)
	}
	return false, nil
}

func selectFrom(set map[string]struct{}, grammemes []string) string {
	for _, g := range grammemes {
		if _, ok := set[g]; ok {
			return g
		}
	}
	return ""
}

func (t *Tag) POS() string          { return selectFrom(PARTS_OF_SPEECH, t.grammemes) }
func (t *Tag) Animacy() string      { return selectFrom(ANIMACY, t.grammemes) }
func (t *Tag) Aspect() string       { return selectFrom(ASPECTS, t.grammemes) }
func (t *Tag) Case() string         { return selectFrom(CASES, t.grammemes) }
func (t *Tag) Gender() string       { return selectFrom(GENDERS, t.grammemes) }
func (t *Tag) Involvement() string  { return selectFrom(INVOLVEMENT, t.grammemes) }
func (t *Tag) Mood() string         { return selectFrom(MOODS, t.grammemes) }
func (t *Tag) Number() string       { return selectFrom(NUMBERS, t.grammemes) }
func (t *Tag) Person() string       { return selectFrom(PERSONS, t.grammemes) }
func (t *Tag) Tense() string        { return selectFrom(TENSES, t.grammemes) }
func (t *Tag) Transitivity() string { return selectFrom(TRANSITIVITY, t.grammemes) }
func (t *Tag) Voice() string        { return selectFrom(VOICES, t.grammemes) }

// IsProductive reports whether tag belongs to a productive part of speech.
func (t *Tag) IsProductive() bool {
	for g := range NON_PRODUCTIVE_GRAMMEMES {
		if _, ok := t.gramSet[g]; ok {
			return false
		}
	}
	return true
}

func newSet(items ...string) map[string]struct{} {
	m := make(map[string]struct{}, len(items))
	for _, it := range items {
		m[it] = struct{}{}
	}
	return m
}

var (
	PARTS_OF_SPEECH = newSet(
		"NOUN", "ADJF", "ADJS", "COMP", "VERB", "INFN",
		"PRTF", "PRTS", "GRND", "NUMR", "ADVB", "NPRO",
		"PRED", "PREP", "CONJ", "PRCL", "INTJ",
	)
	ANIMACY = newSet("anim", "inan")
	GENDERS = newSet("masc", "femn", "neut")
	NUMBERS = newSet("sing", "plur")
	CASES   = newSet(
		"nomn", "gent", "datv", "accs", "ablt", "loct",
		"voct", "gen1", "gen2", "acc2", "loc1", "loc2",
	)
	ASPECTS                  = newSet("perf", "impf")
	TRANSITIVITY             = newSet("tran", "intr")
	PERSONS                  = newSet("1per", "2per", "3per")
	TENSES                   = newSet("pres", "past", "futr")
	MOODS                    = newSet("indc", "impr")
	VOICES                   = newSet("actv", "pssv")
	INVOLVEMENT              = newSet("incl", "excl")
	NON_PRODUCTIVE_GRAMMEMES = newSet("NUMR", "NPRO", "PRED", "PREP", "CONJ", "PRCL", "INTJ", "Apro")
)

var (
	KnownGrammemes = map[string]struct{}{}
	LatToCyr       = map[string]string{}
	CyrToLat       = map[string]string{}
)

func AddGrammemeToKnown(lat, cyr string, overwrite bool) {
	if _, ok := KnownGrammemes[lat]; ok && !overwrite {
		return
	}
	KnownGrammemes[lat] = struct{}{}
	LatToCyr[lat] = cyr
	CyrToLat[cyr] = lat
}

func GrammemeIsKnown(g string) bool {
	_, ok := KnownGrammemes[g]
	return ok
}

func TranslateTag(tag string, mapping map[string]string) string {
	parts := strings.Fields(tag)
	for i, part := range parts {
		grams := strings.Split(part, ",")
		for j, g := range grams {
			if val, ok := mapping[g]; ok {
				grams[j] = val
			}
		}
		parts[i] = strings.Join(grams, ",")
	}
	return strings.Join(parts, " ")
}

func Cyr2Lat(tag string) string { return TranslateTag(tag, CyrToLat) }
func Lat2Cyr(tag string) string { return TranslateTag(tag, LatToCyr) }

func init() {
	categories := []map[string]struct{}{
		PARTS_OF_SPEECH, ANIMACY, GENDERS, NUMBERS, CASES,
		ASPECTS, TRANSITIVITY, PERSONS, TENSES, MOODS, VOICES, INVOLVEMENT,
	}
	for _, cat := range categories {
		for g := range cat {
			AddGrammemeToKnown(g, g, true)
		}
	}
}
