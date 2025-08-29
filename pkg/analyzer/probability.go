package analyzer

import (
	"encoding/json"
	"os"
	"path/filepath"

	"morphy/pkg/analysis"
	"morphy/pkg/dawg"
	"morphy/pkg/tagset"
)

// ProbabilityEstimator adjusts parse scores using P(t|w) data.
type ProbabilityEstimator struct {
	probs *dawg.ConditionalProbDistDAWG
}

// NewProbabilityEstimator loads probabilities from dictionary path.
func NewProbabilityEstimator(dictPath string) (*ProbabilityEstimator, error) {
	file := filepath.Join(dictPath, "p_t_given_w.json")
	data := map[string]int{}
	if b, err := os.ReadFile(file); err == nil {
		if err := json.Unmarshal(b, &data); err != nil {
			return nil, err
		}
		entries := make([]dawg.ProbEntry, 0, len(data))
		for k, v := range data {
			// key format "word:tag"
			for i := 0; i < len(k); i++ {
				if k[i] == ':' {
					word := k[:i]
					tag := k[i+1:]
					entries = append(entries, dawg.ProbEntry{Word: word, Tag: tag, Prob: float64(v) / dawg.MULTIPLIER})
					break
				}
			}
		}
		return &ProbabilityEstimator{probs: dawg.NewConditionalProbDist(entries)}, nil
	}
	return nil, os.ErrNotExist
}

// ApplyToParses replaces scores with conditional probabilities.
func (pe *ProbabilityEstimator) ApplyToParses(word, wordLower string, parses []analysis.Parse) []analysis.Parse {
	if pe == nil || len(parses) == 0 {
		return parses
	}
	probs := make([]float64, len(parses))
	sum := 0.0
	for i, p := range parses {
		prob := pe.probs.Prob(wordLower, p.Tag.String())
		probs[i] = prob
		sum += prob
	}
	if sum == 0 {
		total := 0.0
		for _, p := range parses {
			total += p.Score
		}
		k := 1.0 / total
		for i, p := range parses {
			parses[i].Score = p.Score * k
		}
		return parses
	}
	for i := range parses {
		parses[i].Score = probs[i]
	}
	// simple sort descending by score
	for i := 0; i < len(parses)-1; i++ {
		for j := i + 1; j < len(parses); j++ {
			if parses[j].Score > parses[i].Score {
				parses[i], parses[j] = parses[j], parses[i]
			}
		}
	}
	return parses
}

// ApplyToTags sorts tags according to P(t|w).
func (pe *ProbabilityEstimator) ApplyToTags(word, wordLower string, tags []tagset.Tag) []tagset.Tag {
	if pe == nil || len(tags) == 0 {
		return tags
	}
	for i := 0; i < len(tags)-1; i++ {
		for j := i + 1; j < len(tags); j++ {
			if pe.probs.Prob(wordLower, tags[j].String()) > pe.probs.Prob(wordLower, tags[i].String()) {
				tags[i], tags[j] = tags[j], tags[i]
			}
		}
	}
	return tags
}
