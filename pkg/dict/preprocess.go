package dict

import (
	"sort"
	"strings"
)

// SimplifyTags normalizes tag strings and removes duplicates.
func SimplifyTags(pd *ParsedDictionary, skipSpaceAmbiguity bool) {
	spellings := getTagSpellings(pd)
	replaces := getDuplicateTagReplaces(spellings, skipSpaceAmbiguity)
	for lexID, forms := range pd.Lexemes {
		nf := make([]WordForm, len(forms))
		for i, wf := range forms {
			tag := replaceRedundantGrammemes(wf.Tag)
			if r, ok := replaces[tag]; ok {
				tag = r
			}
			nf[i] = WordForm{Word: wf.Word, Tag: tag}
		}
		pd.Lexemes[lexID] = nf
	}
}

// DropUnsupportedParses removes lexemes with unsupported tags.
func DropUnsupportedParses(pd *ParsedDictionary) {
	for lexID, forms := range pd.Lexemes {
		nf := make([]WordForm, 0, len(forms))
		for _, wf := range forms {
			if !strings.Contains(wf.Tag, "Init") {
				nf = append(nf, wf)
			}
		}
		pd.Lexemes[lexID] = nf
	}
}

func tag2grammemes(tag string) []string {
	tag = replaceRedundantGrammemes(tag)
	tag = strings.Replace(tag, " ", ",", 1)
	if tag == "" {
		return nil
	}
	parts := strings.Split(tag, ",")
	sort.Strings(parts)
	return parts
}

func replaceRedundantGrammemes(tag string) string {
	tag = strings.ReplaceAll(tag, "loc1", "loct")
	tag = strings.ReplaceAll(tag, "gen1", "gent")
	tag = strings.ReplaceAll(tag, "acc1", "accs")
	return tag
}

func getTagSpellings(pd *ParsedDictionary) map[string]map[string]int {
	res := map[string]map[string]int{}
	for _, forms := range pd.Lexemes {
		for _, wf := range forms {
			grams := strings.Join(tag2grammemes(wf.Tag), ",")
			if _, ok := res[grams]; !ok {
				res[grams] = map[string]int{}
			}
			res[grams][wf.Tag]++
		}
	}
	return res
}

func getDuplicateTagReplaces(spellings map[string]map[string]int, skipSpaceAmbiguity bool) map[string]string {
	replaces := map[string]string{}
	for _, tags := range spellings {
		if isAmbiguous(mapKeys(tags), skipSpaceAmbiguity) {
			type kv struct {
				tag string
				cnt int
			}
			items := make([]kv, 0, len(tags))
			for t, c := range tags {
				items = append(items, kv{t, c})
			}
			sort.Slice(items, func(i, j int) bool { return items[i].cnt > items[j].cnt })
			top := items[0].tag
			for _, it := range items[1:] {
				replaces[it.tag] = top
			}
		}
	}
	return replaces
}

func mapKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func isAmbiguous(tags []string, skipSpace bool) bool {
	if len(tags) < 2 {
		return false
	}
	if skipSpace {
		pos := map[int]struct{}{}
		for _, t := range tags {
			p := strings.Index(t, " ")
			pos[p] = struct{}{}
		}
		if len(pos) == len(tags) {
			return false
		}
	}
	return true
}
