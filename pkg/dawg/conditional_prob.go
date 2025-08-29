package dawg

import "fmt"

// ConditionalProbDistDAWG stores probabilities for word-tag pairs.
// Probabilities are kept as integers scaled by MULTIPLIER to avoid
// floating point issues during serialization.
type ConditionalProbDistDAWG struct {
	data map[string]int
}

const MULTIPLIER = 1000000

// ProbEntry represents a probability for (word, tag).
type ProbEntry struct {
	Word string
	Tag  string
	Prob float64
}

// NewConditionalProbDist creates a ConditionalProbDistDAWG from the provided entries.
func NewConditionalProbDist(entries []ProbEntry) *ConditionalProbDistDAWG {
	data := make(map[string]int, len(entries))
	for _, e := range entries {
		key := fmt.Sprintf("%s:%s", e.Word, e.Tag)
		data[key] = int(e.Prob * MULTIPLIER)
	}
	return &ConditionalProbDistDAWG{data: data}
}

// Prob returns the stored probability for the given word and tag pair.
func (d *ConditionalProbDistDAWG) Prob(word, tag string) float64 {
	key := fmt.Sprintf("%s:%s", word, tag)
	if v, ok := d.data[key]; ok {
		return float64(v) / MULTIPLIER
	}
	return 0
}
