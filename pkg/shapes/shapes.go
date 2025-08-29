package shapes

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var romanNumbersRe = regexp.MustCompile(`(?i)^M{0,4}(CM|CD|D?C{0,3})(XC|XL|L?X{0,3})(IX|IV|V?I{0,3})$`)

// IsLatinChar reports whether rune belongs to Latin script.
func IsLatinChar(r rune) bool {
	return unicode.In(r, unicode.Latin)
}

// IsLatin returns true if token contains only latin letters and at least one letter.
func IsLatin(token string) bool {
	hasAlpha := false
	for _, r := range token {
		if unicode.IsLetter(r) {
			if !IsLatinChar(r) {
				return false
			}
			hasAlpha = true
		}
	}
	return hasAlpha
}

// IsPunctuation returns true if token consists only of punctuation and spaces
// and contains at least one punctuation mark.
func IsPunctuation(token string) bool {
	if token == "" {
		return false
	}
	hasPunct := false
	for _, r := range token {
		if unicode.IsSpace(r) {
			continue
		}
		if unicode.IsPunct(r) {
			hasPunct = true
		} else {
			return false
		}
	}
	return hasPunct
}

// IsRomanNumber checks if token is a valid Roman numeral.
func IsRomanNumber(token string) bool {
	if token == "" {
		return false
	}
	return romanNumbersRe.MatchString(token)
}

// RestoreCapitalization makes the capitalization of word the same as example.
func RestoreCapitalization(word, example string) string {
	if strings.ContainsRune(example, '-') {
		wordParts := strings.Split(word, "-")
		exampleParts := strings.Split(example, "-")
		res := make([]string, len(wordParts))
		for i, part := range wordParts {
			if i < len(exampleParts) {
				res[i] = makeTheSameCase(part, exampleParts[i])
			} else {
				res[i] = strings.ToLower(part)
			}
		}
		return strings.Join(res, "-")
	}
	return makeTheSameCase(word, example)
}

func makeTheSameCase(word, example string) string {
	if example == strings.ToLower(example) {
		return strings.ToLower(word)
	}
	if example == strings.ToUpper(example) {
		return strings.ToUpper(word)
	}
	if isTitle(example) {
		return toTitle(word)
	}
	return strings.ToLower(word)
}

func isTitle(s string) bool {
	hasLetter := false
	first := true
	for _, r := range s {
		if !unicode.IsLetter(r) {
			continue
		}
		if first {
			if !unicode.IsUpper(r) {
				return false
			}
			first = false
		} else {
			if unicode.IsUpper(r) {
				return false
			}
		}
		hasLetter = true
	}
	return hasLetter && !first
}

func toTitle(s string) string {
	if s == "" {
		return ""
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToTitle(r)) + strings.ToLower(s[size:])
}
