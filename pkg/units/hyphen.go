package units

import (
	"strings"

	"morphy/pkg/analysis"
	"morphy/pkg/dawg"
	"morphy/pkg/tagset"
)

// HyphenSeparatedParticleAnalyzer handles words with particles after hyphen.
type HyphenSeparatedParticleAnalyzer struct {
	BaseAnalyzerUnit
	Particles       []string
	ScoreMultiplier float64
}

func NewHyphenSeparatedParticleAnalyzer(particles []string) *HyphenSeparatedParticleAnalyzer {
	return &HyphenSeparatedParticleAnalyzer{Particles: particles, ScoreMultiplier: 0.9}
}

func (h *HyphenSeparatedParticleAnalyzer) Parse(word, wordLower string, seen map[string]struct{}) []analysis.Parse {
	res := []analysis.Parse{}
	for _, part := range h.Particles {
		if !strings.HasSuffix(wordLower, part) {
			continue
		}
		base := wordLower[:len(wordLower)-len(part)]
		if base == "" {
			continue
		}
		parses := h.Morph.Parse(base)
		for _, p := range parses {
			method := struct {
				Analyzer *HyphenSeparatedParticleAnalyzer
				Particle string
			}{h, part}
			stack := append(append([]interface{}{}, p.MethodsStack...), method)
			np := analysis.NewParse(p.Word+part, p.Tag, p.NormalForm+part, p.Score*h.ScoreMultiplier, stack)
			AddParseIfNotSeen(np, &res, seen)
		}
		break
	}
	return res
}

func (h *HyphenSeparatedParticleAnalyzer) Tag(word, wordLower string, seen map[string]struct{}) []tagset.Tag {
	res := []tagset.Tag{}
	for _, part := range h.Particles {
		if !strings.HasSuffix(wordLower, part) {
			continue
		}
		base := wordLower[:len(wordLower)-len(part)]
		tags := h.Morph.Tag(base)
		for _, t := range tags {
			AddTagIfNotSeen(t, &res, seen)
		}
		break
	}
	return res
}

func (h *HyphenSeparatedParticleAnalyzer) GetLexeme(p analysis.Parse) []analysis.Parse {
	return []analysis.Parse{p}
}
func (h *HyphenSeparatedParticleAnalyzer) Normalized(p analysis.Parse) analysis.Parse { return p }

// HyphenAdverbAnalyzer detects adverbs starting with "по-".
type HyphenAdverbAnalyzer struct {
	BaseAnalyzerUnit
	ScoreMultiplier float64
	tag             *tagset.Tag
}

func NewHyphenAdverbAnalyzer() *HyphenAdverbAnalyzer {
	return &HyphenAdverbAnalyzer{ScoreMultiplier: 0.7}
}

func (h *HyphenAdverbAnalyzer) Init(morph Analyzer) {
	h.BaseAnalyzerUnit.Init(morph)
	t, _ := tagset.New("ADVB")
	h.tag = t
}

func (h *HyphenAdverbAnalyzer) shouldParse(word string) bool {
	if len(word) < 5 || !strings.HasPrefix(word, "по-") {
		return false
	}
	tags := h.Morph.Tag(word[3:])
	for _, t := range tags {
		ok1, _ := t.Contains("ADJF")
		ok2, _ := t.Contains("sing")
		ok3, _ := t.Contains("datv")
		if ok1 && ok2 && ok3 {
			return true
		}
	}
	return false
}

func (h *HyphenAdverbAnalyzer) Parse(word, wordLower string, seen map[string]struct{}) []analysis.Parse {
	if !h.shouldParse(wordLower) {
		return nil
	}
	method := struct{ Analyzer *HyphenAdverbAnalyzer }{h}
	p := analysis.NewParse(wordLower, h.tag, wordLower, h.ScoreMultiplier, []interface{}{method})
	res := []analysis.Parse{}
	AddParseIfNotSeen(p, &res, seen)
	return res
}

func (h *HyphenAdverbAnalyzer) Tag(word, wordLower string, seen map[string]struct{}) []tagset.Tag {
	if !h.shouldParse(wordLower) {
		return nil
	}
	if _, ok := seen[h.tag.String()]; ok {
		return nil
	}
	seen[h.tag.String()] = struct{}{}
	return []tagset.Tag{*h.tag}
}

func (h *HyphenAdverbAnalyzer) GetLexeme(p analysis.Parse) []analysis.Parse {
	return []analysis.Parse{p}
}
func (h *HyphenAdverbAnalyzer) Normalized(p analysis.Parse) analysis.Parse { return p }

// HyphenatedWordsAnalyzer parses words composed with hyphen.
type HyphenatedWordsAnalyzer struct {
	BaseAnalyzerUnit
	SkipPrefixes    []string
	ScoreMultiplier float64
	matcher         *dawg.PrefixMatcher
}

func NewHyphenatedWordsAnalyzer(skip []string) *HyphenatedWordsAnalyzer {
	return &HyphenatedWordsAnalyzer{SkipPrefixes: skip, ScoreMultiplier: 0.75}
}

func (h *HyphenatedWordsAnalyzer) Init(morph Analyzer) {
	h.BaseAnalyzerUnit.Init(morph)
	h.matcher = dawg.NewPrefixMatcher(h.SkipPrefixes)
}

func (h *HyphenatedWordsAnalyzer) shouldParse(word string) bool {
	if strings.Count(word, "-") != 1 {
		return false
	}
	if strings.HasPrefix(word, "-") || strings.HasSuffix(word, "-") {
		return false
	}
	if h.matcher.IsPrefixed(word) {
		return false
	}
	return true
}

func (h *HyphenatedWordsAnalyzer) Parse(word, wordLower string, seen map[string]struct{}) []analysis.Parse {
	if !h.shouldParse(wordLower) {
		return nil
	}
	parts := strings.SplitN(wordLower, "-", 2)
	left, right := parts[0], parts[1]
	leftParses := h.Morph.Parse(left)
	rightParses := h.Morph.Parse(right)
	res := []analysis.Parse{}
	method := struct{ Analyzer *HyphenatedWordsAnalyzer }{h}

	for _, rp := range rightParses {
		stack := append(append([]interface{}{}, rp.MethodsStack...), method)
		p := analysis.NewParse(left+"-"+rp.Word, rp.Tag, left+"-"+rp.NormalForm, rp.Score*h.ScoreMultiplier, stack)
		AddParseIfNotSeen(p, &res, seen)
	}
	for _, lp := range leftParses {
		for _, rp := range rightParses {
			wordCombined := lp.Word + "-" + rp.Word
			nf := lp.NormalForm + "-" + rp.NormalForm
			score := (lp.Score + rp.Score) / 2 * h.ScoreMultiplier
			stack := append(append(lp.MethodsStack, rp.MethodsStack...), method)
			p := analysis.NewParse(wordCombined, lp.Tag, nf, score, stack)
			AddParseIfNotSeen(p, &res, seen)
		}
	}
	return res
}

func (h *HyphenatedWordsAnalyzer) Tag(word, wordLower string, seen map[string]struct{}) []tagset.Tag {
	if !h.shouldParse(wordLower) {
		return nil
	}
	parts := strings.SplitN(wordLower, "-", 2)
	res := []tagset.Tag{}
	tags := h.Morph.Tag(parts[1])
	for _, t := range tags {
		AddTagIfNotSeen(t, &res, seen)
	}
	return res
}

func (h *HyphenatedWordsAnalyzer) GetLexeme(p analysis.Parse) []analysis.Parse {
	return []analysis.Parse{p}
}
func (h *HyphenatedWordsAnalyzer) Normalized(p analysis.Parse) analysis.Parse { return p }
