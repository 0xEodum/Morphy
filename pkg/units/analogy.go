package units

import (
	"fmt"
	"sort"
	"strings"

	"morphy/pkg/analysis"
	"morphy/pkg/dawg"
	"morphy/pkg/dict"
	"morphy/pkg/tagset"
	"morphy/pkg/utils"
)

// KnownPrefixAnalyzer parses words with known prefixes.
type KnownPrefixAnalyzer struct {
	BaseAnalyzerUnit
	KnownPrefixes   []string
	ScoreMultiplier float64
	MinRemainder    int
	matcher         *dawg.PrefixMatcher
}

func NewKnownPrefixAnalyzer(prefixes []string) *KnownPrefixAnalyzer {
	return &KnownPrefixAnalyzer{KnownPrefixes: prefixes, ScoreMultiplier: 0.75, MinRemainder: 3}
}

func (k *KnownPrefixAnalyzer) Init(morph Analyzer) {
	k.BaseAnalyzerUnit.Init(morph)
	k.matcher = dawg.NewPrefixMatcher(k.KnownPrefixes)
}

func (k *KnownPrefixAnalyzer) possible(word string) []utils.Split {
	prefixes := k.matcher.Prefixes(word)
	sort.Slice(prefixes, func(i, j int) bool { return len(prefixes[i]) > len(prefixes[j]) })
	res := []utils.Split{}
	for _, p := range prefixes {
		if len(word)-len(p) < k.MinRemainder {
			continue
		}
		res = append(res, utils.Split{Prefix: p, Suffix: word[len(p):]})
	}
	return res
}

func (k *KnownPrefixAnalyzer) Parse(word, wordLower string, seen map[string]struct{}) []analysis.Parse {
	res := []analysis.Parse{}
	for _, sp := range k.possible(wordLower) {
		parses := k.Morph.Parse(sp.Suffix)
		for _, p := range parses {
			method := struct {
				Analyzer *KnownPrefixAnalyzer
				Prefix   string
			}{k, sp.Prefix}
			stack := append(append([]interface{}{}, p.MethodsStack...), method)
			np := analysis.NewParse(sp.Prefix+p.Word, p.Tag, sp.Prefix+p.NormalForm, p.Score*k.ScoreMultiplier, stack)
			AddParseIfNotSeen(np, &res, seen)
		}
	}
	return res
}

func (k *KnownPrefixAnalyzer) Tag(word, wordLower string, seen map[string]struct{}) []tagset.Tag {
	res := []tagset.Tag{}
	for _, sp := range k.possible(wordLower) {
		tags := k.Morph.Tag(sp.Suffix)
		for _, t := range tags {
			AddTagIfNotSeen(t, &res, seen)
		}
	}
	return res
}

func (k *KnownPrefixAnalyzer) GetLexeme(p analysis.Parse) []analysis.Parse {
	return []analysis.Parse{p}
}
func (k *KnownPrefixAnalyzer) Normalized(p analysis.Parse) analysis.Parse { return p }

// UnknownPrefixAnalyzer parses words by stripping any prefix and analyzing remainder via dictionary.
type UnknownPrefixAnalyzer struct {
	BaseAnalyzerUnit
	ScoreMultiplier float64
	dictAnalyzer    *DictionaryAnalyzer
}

func NewUnknownPrefixAnalyzer() *UnknownPrefixAnalyzer {
	return &UnknownPrefixAnalyzer{ScoreMultiplier: 0.5}
}

func (u *UnknownPrefixAnalyzer) Init(morph Analyzer) {
	u.BaseAnalyzerUnit.Init(morph)
	da := &DictionaryAnalyzer{}
	da.Init(morph)
	u.dictAnalyzer = da
}

func (u *UnknownPrefixAnalyzer) Parse(word, wordLower string, seen map[string]struct{}) []analysis.Parse {
	res := []analysis.Parse{}
	splits := utils.WordSplits(wordLower, 3, len(wordLower)-1)
	for _, sp := range splits {
		parses := u.dictAnalyzer.Parse(sp.Suffix, sp.Suffix, seen)
		for _, p := range parses {
			method := struct {
				Analyzer *UnknownPrefixAnalyzer
				Prefix   string
			}{u, sp.Prefix}
			stack := append(append([]interface{}{}, p.MethodsStack...), method)
			np := analysis.NewParse(sp.Prefix+p.Word, p.Tag, sp.Prefix+p.NormalForm, p.Score*u.ScoreMultiplier, stack)
			AddParseIfNotSeen(np, &res, seen)
		}
	}
	return res
}

func (u *UnknownPrefixAnalyzer) Tag(word, wordLower string, seen map[string]struct{}) []tagset.Tag {
	res := []tagset.Tag{}
	splits := utils.WordSplits(wordLower, 3, len(wordLower)-1)
	for _, sp := range splits {
		tags := u.dictAnalyzer.Tag(sp.Suffix, sp.Suffix, seen)
		for _, t := range tags {
			AddTagIfNotSeen(t, &res, seen)
		}
	}
	return res
}

func (u *UnknownPrefixAnalyzer) GetLexeme(p analysis.Parse) []analysis.Parse {
	return []analysis.Parse{p}
}
func (u *UnknownPrefixAnalyzer) Normalized(p analysis.Parse) analysis.Parse { return p }

// KnownSuffixAnalyzer predicts tags based on suffix analogies.
type KnownSuffixAnalyzer struct {
	BaseAnalyzerUnit
	ScoreMultiplier  float64
	MinWordLength    int
	paradigmPrefixes []struct {
		ID     int
		Prefix string
	}
	predictionSplits []int
	fakeDict         *DictionaryAnalyzer
}

func NewKnownSuffixAnalyzer() *KnownSuffixAnalyzer {
	return &KnownSuffixAnalyzer{ScoreMultiplier: 0.5, MinWordLength: 4}
}

func (k *KnownSuffixAnalyzer) Init(morph Analyzer) {
	k.BaseAnalyzerUnit.Init(morph)
	dict, _ := k.Dict.(*dict.Dictionary)
	pp := dict.ParadigmPrefixes()
	k.paradigmPrefixes = make([]struct {
		ID     int
		Prefix string
	}, len(pp))
	for i, p := range pp {
		k.paradigmPrefixes[i] = struct {
			ID     int
			Prefix string
		}{ID: i, Prefix: p}
	}
	for i, j := 0, len(k.paradigmPrefixes)-1; i < j; i, j = i+1, j-1 {
		k.paradigmPrefixes[i], k.paradigmPrefixes[j] = k.paradigmPrefixes[j], k.paradigmPrefixes[i]
	}
	maxLen := dict.MaxSuffixLength()
	if maxLen <= 0 {
		maxLen = 5
	}
	k.predictionSplits = make([]int, maxLen)
	for i := 1; i <= maxLen; i++ {
		k.predictionSplits[i-1] = i
	}
	for i, j := 0, len(k.predictionSplits)-1; i < j; i, j = i+1, j-1 {
		k.predictionSplits[i], k.predictionSplits[j] = k.predictionSplits[j], k.predictionSplits[i]
	}
	fd := &DictionaryAnalyzer{}
	fd.Init(morph)
	k.fakeDict = fd
}

func (k *KnownSuffixAnalyzer) Parse(word, wordLower string, seen map[string]struct{}) []analysis.Parse {
	if len(word) < k.MinWordLength {
		return nil
	}
	dict, _ := k.Dict.(*dict.Dictionary)
	subs := k.Morph.CharSubstitutes()
	totalCounts := make([]int, len(k.paradigmPrefixes))
	for i := range totalCounts {
		totalCounts[i] = 1
	}
	type tmp struct {
		cnt      int
		word     string
		tag      tagset.Tag
		normal   string
		prefixID int
		methods  []interface{}
	}
	tmpRes := []tmp{}
	seenPar := map[string]struct{}{}
	for _, pref := range k.paradigmPrefixes {
		if !strings.HasPrefix(wordLower, pref.Prefix) {
			continue
		}
		suffixDawg := dict.PredictionSuffixes()[pref.ID]
		for _, split := range k.predictionSplits {
			if split > len(wordLower) {
				continue
			}
			wordStart := wordLower[:len(wordLower)-split]
			wordEnd := wordLower[len(wordLower)-split:]
			items := suffixDawg.SimilarItems(wordEnd, subs)
			for suffix, parses := range items {
				fixedWord := wordStart + suffix
				for _, p := range parses {
					tag := dict.BuildTagInfo(int(p.ParadigmID), int(p.FormIndex))
					if !tag.IsProductive() {
						continue
					}
					totalCounts[pref.ID] += int(p.Count)
					key := fixedWord + "|" + tag.String() + "|" + fmt.Sprint(p.ParadigmID)
					if _, ok := seenPar[key]; ok {
						continue
					}
					seenPar[key] = struct{}{}
					normal := dict.BuildNormalForm(int(p.ParadigmID), int(p.FormIndex), fixedWord)
					methods := []interface{}{
						dictMethod{Analyzer: k.fakeDict, Word: fixedWord, ParaID: int(p.ParadigmID), Index: int(p.FormIndex)},
						struct {
							Analyzer *KnownSuffixAnalyzer
							Suffix   string
						}{k, suffix},
					}
					tmpRes = append(tmpRes, tmp{cnt: int(p.Count), word: fixedWord, tag: tag, normal: normal, prefixID: pref.ID, methods: methods})
				}
			}
			if totalCounts[pref.ID] > 1 {
				break
			}
		}
	}
	parses := []analysis.Parse{}
	for _, r := range tmpRes {
		score := float64(r.cnt) / float64(totalCounts[r.prefixID]) * k.ScoreMultiplier
		p := analysis.NewParse(r.word, &r.tag, r.normal, score, r.methods)
		AddParseIfNotSeen(p, &parses, seen)
	}
	sort.Slice(parses, func(i, j int) bool { return parses[i].Score > parses[j].Score })
	return parses
}

func (k *KnownSuffixAnalyzer) Tag(word, wordLower string, seen map[string]struct{}) []tagset.Tag {
	if len(word) < k.MinWordLength {
		return nil
	}
	dict, _ := k.Dict.(*dict.Dictionary)
	subs := k.Morph.CharSubstitutes()
	type tmp struct {
		cnt int
		tag tagset.Tag
	}
	tmpTags := []tmp{}
	for _, pref := range k.paradigmPrefixes {
		if !strings.HasPrefix(wordLower, pref.Prefix) {
			continue
		}
		suffixDawg := dict.PredictionSuffixes()[pref.ID]
		for _, split := range k.predictionSplits {
			if split > len(wordLower) {
				continue
			}
			end := wordLower[len(wordLower)-split:]
			items := suffixDawg.SimilarItems(end, subs)
			found := false
			for _, parses := range items {
				for _, p := range parses {
					tag := dict.BuildTagInfo(int(p.ParadigmID), int(p.FormIndex))
					if !tag.IsProductive() {
						continue
					}
					found = true
					if _, ok := seen[tag.String()]; ok {
						continue
					}
					seen[tag.String()] = struct{}{}
					tmpTags = append(tmpTags, tmp{cnt: int(p.Count), tag: tag})
				}
			}
			if found {
				break
			}
		}
	}
	sort.Slice(tmpTags, func(i, j int) bool { return tmpTags[i].cnt > tmpTags[j].cnt })
	res := make([]tagset.Tag, 0, len(tmpTags))
	for _, t := range tmpTags {
		res = append(res, t.tag)
	}
	return res
}
func (k *KnownSuffixAnalyzer) GetLexeme(p analysis.Parse) []analysis.Parse {
	return []analysis.Parse{p}
}
func (k *KnownSuffixAnalyzer) Normalized(p analysis.Parse) analysis.Parse { return p }
