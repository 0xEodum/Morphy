package dawg

// Prediction stores data used for suffix prediction.
type Prediction struct {
	Count      uint16
	ParadigmID uint16
	FormIndex  uint16
}

// PredictionSuffixesDAWG stores suffix prediction information.
type PredictionSuffixesDAWG struct {
	*DAWG[Prediction]
}

// NewPredictionSuffixesDAWG creates a PredictionSuffixesDAWG from the data map.
func NewPredictionSuffixesDAWG(data map[string][]Prediction) *PredictionSuffixesDAWG {
	return &PredictionSuffixesDAWG{New[Prediction](data)}
}

// Lookup returns prediction records for a suffix.
func (p *PredictionSuffixesDAWG) Lookup(suffix string) []Prediction {
	return p.Items(suffix)
}
