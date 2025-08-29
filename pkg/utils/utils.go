package utils

import (
	"fmt"
	"sort"
	"strings"
)

// CombinationsOfAllLengths returns all combinations of the input slice with all possible lengths >=1.
func CombinationsOfAllLengths(items []string) [][]string {
	var res [][]string
	n := len(items)
	for r := 1; r <= n; r++ {
		comb := make([]int, r)
		var gen func(int, int)
		gen = func(start, idx int) {
			if idx == r {
				tmp := make([]string, r)
				for i, c := range comb {
					tmp[i] = items[c]
				}
				res = append(res, tmp)
				return
			}
			for i := start; i < n; i++ {
				comb[idx] = i
				gen(i+1, idx+1)
			}
		}
		gen(0, 0)
	}
	return res
}

// LongestCommonSubstring returns the longest common substring for all strings in data.
func LongestCommonSubstring(data []string) string {
	if len(data) == 0 {
		return ""
	}
	if len(data) == 1 {
		return data[0]
	}
	first := data[0]
	substr := ""
	for i := 0; i < len(first); i++ {
		for j := i + 1; j <= len(first); j++ {
			candidate := first[i:j]
			if len(candidate) <= len(substr) {
				continue
			}
			ok := true
			for _, s := range data[1:] {
				if !strings.Contains(s, candidate) {
					ok = false
					break
				}
			}
			if ok {
				substr = candidate
			}
		}
	}
	return substr
}

// LargestElements returns elements that have one of the top-n key values.
func LargestElements[T any](iter []T, key func(T) float64, n int) []T {
	if n <= 0 {
		return nil
	}
	uniq := make(map[float64]struct{})
	keys := make([]float64, 0)
	for _, item := range iter {
		k := key(item)
		if _, ok := uniq[k]; !ok {
			uniq[k] = struct{}{}
			keys = append(keys, k)
		}
	}
	sort.Float64s(keys)
	if n > len(keys) {
		n = len(keys)
	}
	topKeys := make(map[float64]struct{})
	for _, k := range keys[len(keys)-n:] {
		topKeys[k] = struct{}{}
	}
	res := make([]T, 0)
	for _, item := range iter {
		if _, ok := topKeys[key(item)]; ok {
			res = append(res, item)
		}
	}
	return res
}

// Split represents a word split into prefix and suffix.
type Split struct {
	Prefix string
	Suffix string
}

// WordSplits returns all possible splits for a word given minReminder and maxPrefixLength.
func WordSplits(word string, minReminder, maxPrefixLength int) []Split {
	maxSplit := maxPrefixLength
	if l := len(word) - minReminder; maxSplit > l {
		maxSplit = l
	}
	res := make([]Split, 0, maxSplit)
	for i := 1; i <= maxSplit; i++ {
		res = append(res, Split{Prefix: word[:i], Suffix: word[i:]})
	}
	return res
}

// KwargsRepr returns a string representation of keyword arguments map.
func KwargsRepr(kwargs map[string]interface{}, dontShow []string) string {
	if len(kwargs) == 0 {
		return ""
	}
	hide := make(map[string]struct{}, len(dontShow))
	for _, k := range dontShow {
		hide[k] = struct{}{}
	}
	keys := make([]string, 0, len(kwargs))
	for k := range kwargs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		if _, ok := hide[k]; ok {
			parts = append(parts, fmt.Sprintf("%s=<...>", k))
		} else {
			parts = append(parts, fmt.Sprintf("%s=%v", k, kwargs[k]))
		}
	}
	return strings.Join(parts, ", ")
}
