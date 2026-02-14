package iterx

import (
	"iter"
	"unicode/utf8"
)

// Runes returns an iterator over the runes of a string (no allocation).
func Runes(s string) iter.Seq[rune] {
	return func(yield func(rune) bool) {
		for _, r := range s {
			if !yield(r) {
				return
			}
		}
	}
}

// RunesReverse returns an iterator over the runes of a string in reverse (no allocation).
func RunesReverse(s string) iter.Seq[rune] {
	return func(yield func(rune) bool) {
		for len(s) > 0 {
			r, size := utf8.DecodeLastRuneInString(s)
			s = s[:len(s)-size]
			if !yield(r) {
				return
			}
		}
	}
}
