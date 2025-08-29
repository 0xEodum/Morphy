package dict

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"morphy/pkg/dawg"
	"morphy/pkg/tagset"
)

// CurrentFormatVersion describes format of saved dictionaries.
const CurrentFormatVersion = "0.1"

// LoadedDictionary holds dictionary data loaded from disk.
type LoadedDictionary struct {
	Meta               map[string]any
	Gramtab            []tagset.Tag
	Suffixes           []string
	Paradigms          [][]uint16
	Words              *dawg.WordsDawg
	PredictionSuffixes []*dawg.PredictionSuffixesDAWG
	ParadigmPrefixes   []string
}

// LoadDict reads dictionary data from path.
func LoadDict(path string) (*LoadedDictionary, error) {
	f := func(name string) string { return filepath.Join(path, name) }
	meta := map[string]any{}
	if err := jsonRead(f("meta.json"), &meta); err != nil {
		return nil, err
	}

	// load grammemes to register in tagset
	var grammemes []string
	_ = jsonRead(f("grammemes.json"), &grammemes)
	for _, g := range grammemes {
		tagset.AddGrammemeToKnown(g, g, true)
	}

	// load gramtab
	var gramtabStr []string
	_ = jsonRead(f("gramtab.json"), &gramtabStr)
	gramtab := make([]tagset.Tag, 0, len(gramtabStr))
	for _, t := range gramtabStr {
		tg, err := tagset.New(t)
		if err != nil {
			return nil, err
		}
		gramtab = append(gramtab, *tg)
	}

	// load suffixes
	var suffixes []string
	_ = jsonRead(f("suffixes.json"), &suffixes)

	// load paradigms
	var paradigms [][]uint16
	_ = jsonRead(f("paradigms.json"), &paradigms)

	// load words
	wordsMap := map[string][]dawg.WordForm{}
	_ = jsonRead(f("words.json"), &wordsMap)
	words := dawg.NewWordsDawg(wordsMap)

	// load paradigm prefixes
	var paradigmPrefixes []string
	_ = jsonRead(f("paradigm-prefixes.json"), &paradigmPrefixes)

	// load prediction suffix dawgs if present
	prediction := []*dawg.PredictionSuffixesDAWG{}
	for i := 0; ; i++ {
		name := f(fmt.Sprintf("prediction-suffixes-%d.json", i))
		if _, err := os.Stat(name); err != nil {
			break
		}
		data := map[string][]dawg.Prediction{}
		if err := jsonRead(name, &data); err != nil {
			return nil, err
		}
		prediction = append(prediction, dawg.NewPredictionSuffixesDAWG(data))
	}

	return &LoadedDictionary{
		Meta:               meta,
		Gramtab:            gramtab,
		Suffixes:           suffixes,
		Paradigms:          paradigms,
		Words:              words,
		PredictionSuffixes: prediction,
		ParadigmPrefixes:   paradigmPrefixes,
	}, nil
}

// SaveCompiledDict saves compiled dictionary to outPath.
func SaveCompiledDict(cd *CompiledDictionary, outPath, sourceName, languageCode string) error {
	if err := os.MkdirAll(outPath, 0o755); err != nil {
		return err
	}
	f := func(name string) string { return filepath.Join(outPath, name) }

	if err := jsonWrite(f("grammemes.json"), cd.ParsedDict.Grammemes); err != nil {
		return err
	}
	if err := jsonWrite(f("gramtab.json"), cd.Gramtab); err != nil {
		return err
	}
	if err := jsonWrite(f("suffixes.json"), cd.Suffixes); err != nil {
		return err
	}
	if err := jsonWrite(f("paradigms.json"), cd.Paradigms); err != nil {
		return err
	}
	if err := jsonWrite(f("words.json"), cd.WordsDawg.Data()); err != nil {
		return err
	}
	if err := jsonWrite(f("paradigm-prefixes.json"), cd.ParadigmPrefixes); err != nil {
		return err
	}
	for i, pd := range cd.PredictionSuffixesDawgs {
		name := f(fmt.Sprintf("prediction-suffixes-%d.json", i))
		if err := jsonWrite(name, pd.Data()); err != nil {
			return err
		}
	}

	meta := map[string]any{
		"language_code":  languageCode,
		"format_version": CurrentFormatVersion,
		"compiled_at":    time.Now().UTC().Format(time.RFC3339),
		"source":         sourceName,
	}
	return jsonWrite(f("meta.json"), meta)
}

func jsonRead(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func jsonWrite(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
