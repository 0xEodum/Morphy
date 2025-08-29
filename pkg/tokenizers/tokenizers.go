package tokenizers

import (
	"regexp"
	"unicode"
)

var groupingSpaceRegex = regexp.MustCompile(`([^\p{L}\p{M}\p{N}_-]|[+])`)

func SimpleWordTokenize(text string) []string {
	if text == "" {
		return nil
	}
	locs := groupingSpaceRegex.FindAllStringIndex(text, -1)
	tokens := make([]string, 0, len(locs)+1)
	last := 0
	for _, loc := range locs {
		if loc[0] > last {
			seg := text[last:loc[0]]
			if seg != "" && !isSpace(seg) {
				tokens = append(tokens, seg)
			}
		}
		sep := text[loc[0]:loc[1]]
		if sep != "" && !isSpace(sep) {
			tokens = append(tokens, sep)
		}
		last = loc[1]
	}
	if last < len(text) {
		seg := text[last:]
		if seg != "" && !isSpace(seg) {
			tokens = append(tokens, seg)
		}
	}
	return tokens
}

func isSpace(s string) bool {
	for _, r := range s {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}
