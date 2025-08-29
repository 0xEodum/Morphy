package units

import (
	"fmt"

	"morphy/pkg/analysis"
	"morphy/pkg/tagset"
)

// AddParseIfNotSeen appends parse to resultList if it wasn't seen before.
func AddParseIfNotSeen(parse analysis.Parse, resultList *[]analysis.Parse, seenParses map[string]struct{}) {
	paraID := -1
	if len(parse.MethodsStack) > 0 {
		if info, ok := parse.MethodsStack[0].(dictMethod); ok {
			paraID = info.ParaID
		}
	}
	key := fmt.Sprintf("%s|%s|%d", parse.Word, parse.Tag.String(), paraID)
	if _, ok := seenParses[key]; ok {
		return
	}
	seenParses[key] = struct{}{}
	*resultList = append(*resultList, parse)
}

// AddTagIfNotSeen appends tag to resultList if it wasn't seen before.
func AddTagIfNotSeen(tag tagset.Tag, resultList *[]tagset.Tag, seenTags map[string]struct{}) {
	key := tag.String()
	if _, ok := seenTags[key]; ok {
		return
	}
	seenTags[key] = struct{}{}
	*resultList = append(*resultList, tag)
}

// WithSuffix returns a new parse with suffix appended.
func WithSuffix(p analysis.Parse, suffix string) analysis.Parse {
	return analysis.NewParse(p.Word+suffix, p.Tag, p.NormalForm+suffix, p.Score, p.MethodsStack)
}

// WithoutFixedSuffix returns a new parse with suffixLength characters removed from end.
func WithoutFixedSuffix(p analysis.Parse, suffixLength int) analysis.Parse {
	return analysis.NewParse(p.Word[:len(p.Word)-suffixLength], p.Tag, p.NormalForm[:len(p.NormalForm)-suffixLength], p.Score, p.MethodsStack)
}

// WithoutFixedPrefix returns a new parse with prefixLength characters removed from start.
func WithoutFixedPrefix(p analysis.Parse, prefixLength int) analysis.Parse {
	return analysis.NewParse(p.Word[prefixLength:], p.Tag, p.NormalForm[prefixLength:], p.Score, p.MethodsStack)
}

// WithPrefix returns a new parse with prefix added.
func WithPrefix(p analysis.Parse, prefix string) analysis.Parse {
	return analysis.NewParse(prefix+p.Word, p.Tag, prefix+p.NormalForm, p.Score, p.MethodsStack)
}

// ReplaceMethodsStack returns a new parse with provided methods stack.
func ReplaceMethodsStack(p analysis.Parse, newStack []interface{}) analysis.Parse {
	return analysis.NewParse(p.Word, p.Tag, p.NormalForm, p.Score, newStack)
}

// WithoutLastMethod returns a new parse without last method in stack.
func WithoutLastMethod(p analysis.Parse) analysis.Parse {
	stack := p.MethodsStack[:len(p.MethodsStack)-1]
	return analysis.NewParse(p.Word, p.Tag, p.NormalForm, p.Score, stack)
}

// AppendMethod returns a new parse with method appended to stack.
func AppendMethod(p analysis.Parse, method interface{}) analysis.Parse {
	stack := append(append([]interface{}{}, p.MethodsStack...), method)
	return analysis.NewParse(p.Word, p.Tag, p.NormalForm, p.Score, stack)
}
