package analysis

import "morphy/pkg/tagset"

// Parse represents a morphological analysis result for a single word.
type Parse struct {
	Word         string
	Tag          *tagset.Tag
	NormalForm   string
	Score        float64
	MethodsStack []interface{}
}

// NewParse creates a new Parse instance.
func NewParse(word string, tag *tagset.Tag, normalForm string, score float64, stack []interface{}) Parse {
	return Parse{
		Word:         word,
		Tag:          tag,
		NormalForm:   normalForm,
		Score:        score,
		MethodsStack: stack,
	}
}
