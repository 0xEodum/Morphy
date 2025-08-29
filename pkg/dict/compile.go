package dict

import (
	"fmt"
	"os"
	"strings"

	"morphy/pkg/dawg"
	"morphy/pkg/utils"
)

// CompiledDictionary stores compacted dictionary data.
type CompiledDictionary struct {
	Gramtab                 []string
	Suffixes                []string
	Paradigms               [][]uint16
	WordsDawg               *dawg.WordsDawg
	PredictionSuffixesDawgs []*dawg.PredictionSuffixesDAWG
	ParsedDict              *ParsedDictionary
	CompileOptions          map[string]any
	ParadigmPrefixes        []string
}

// ConvertToPymorphy2 converts OpenCorpora XML dict to compiled format and saves it.
func ConvertToPymorphy2(xmlPath, outPath, sourceName, languageCode string, overwrite bool, options map[string]any) error {
	if !overwrite {
		if _, err := os.Stat(outPath); err == nil {
			return fmt.Errorf("output path exists")
		}
	}
	parsed, err := ParseOpencorporaXML(xmlPath)
	if err != nil {
		return err
	}
	SimplifyTags(parsed, true)
	DropUnsupportedParses(parsed)
	compiled, err := CompileParsedDict(parsed, options)
	if err != nil {
		return err
	}
	return SaveCompiledDict(compiled, outPath, sourceName, languageCode)
}

// CompileParsedDict builds compact representation from parsed dictionary.
func CompileParsedDict(parsed *ParsedDictionary, compileOptions map[string]any) (*CompiledDictionary, error) {
	suffixes := []string{}
	suffixIDs := map[string]uint16{}
	gramtab := []string{}
	tagIDs := map[string]uint16{}
	prefixes := []string{""}
	prefixIDs := map[string]uint16{"": 0}

	paradigms := [][]uint16{}
	words := map[string][]dawg.WordForm{}

	for _, lexeme := range parsed.Lexemes {
		stem, para := toParadigm(lexeme)
		paraArr := make([]uint16, len(para)*3)
		for i, f := range para {
			sid, ok := suffixIDs[f.Suffix]
			if !ok {
				sid = uint16(len(suffixes))
				suffixes = append(suffixes, f.Suffix)
				suffixIDs[f.Suffix] = sid
			}
			tid, ok := tagIDs[f.Tag]
			if !ok {
				tid = uint16(len(gramtab))
				gramtab = append(gramtab, f.Tag)
				tagIDs[f.Tag] = tid
			}
			pid, ok := prefixIDs[f.Prefix]
			if !ok {
				pid = uint16(len(prefixes))
				prefixes = append(prefixes, f.Prefix)
				prefixIDs[f.Prefix] = pid
			}
			paraArr[i] = sid
			paraArr[len(para)+i] = tid
			paraArr[2*len(para)+i] = pid
			word := f.Prefix + stem + f.Suffix
			words[word] = append(words[word], dawg.WordForm{ParadigmID: uint16(len(paradigms)), FormIndex: uint16(i)})
		}
		paradigms = append(paradigms, paraArr)
	}

	wd := dawg.NewWordsDawg(words)
	return &CompiledDictionary{
		Gramtab:                 gramtab,
		Suffixes:                suffixes,
		Paradigms:               paradigms,
		WordsDawg:               wd,
		PredictionSuffixesDawgs: []*dawg.PredictionSuffixesDAWG{},
		ParsedDict:              parsed,
		CompileOptions:          compileOptions,
		ParadigmPrefixes:        prefixes,
	}, nil
}

// formInfo represents part of paradigm.
type formInfo struct {
	Suffix string
	Tag    string
	Prefix string
}

func toParadigm(lexeme []WordForm) (string, []formInfo) {
	forms := make([]string, len(lexeme))
	tags := make([]string, len(lexeme))
	for i, wf := range lexeme {
		forms[i] = wf.Word
		tags[i] = wf.Tag
	}
	var stem string
	if len(forms) == 1 {
		stem = forms[0]
	} else {
		stem = utils.LongestCommonSubstring(forms)
	}
	prefixes := make([]string, len(forms))
	for i, form := range forms {
		idx := strings.Index(form, stem)
		if idx < 0 {
			idx = 0
		}
		prefixes[i] = form[:idx]
	}
	res := make([]formInfo, len(forms))
	for i, form := range forms {
		suff := form[len(prefixes[i])+len(stem):]
		res[i] = formInfo{Suffix: suff, Tag: tags[i], Prefix: prefixes[i]}
	}
	return stem, res
}
