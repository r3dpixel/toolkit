package lexer

import (
	"testing"

	"github.com/r3dpixel/toolkit/iterx"
)

func TestMatch(t *testing.T) {
	lex := New[rune, int]()
	lex.InsertIter(iterx.Runes("if"), 1)
	lex.InsertIter(iterx.Runes("else"), 2)
	lex.InsertIter(iterx.Runes("for"), 3)

	tests := []struct {
		input string
		want  int
		ok    bool
	}{
		{"if", 1, true},
		{"else", 2, true},
		{"for", 3, true},
		{"i", 0, false},      // partial match
		{"iff", 0, false},    // too long
		{"els", 0, false},    // partial match
		{"elseif", 0, false}, // too long
		{"while", 0, false},  // no match
		{"", 0, false},       // empty input
	}

	for _, tt := range tests {
		got, ok := lex.Match(iterx.Runes(tt.input))
		if got != tt.want || ok != tt.ok {
			t.Errorf("Match(%q) = (%v, %v), want (%v, %v)", tt.input, got, ok, tt.want, tt.ok)
		}
	}
}

func TestFirstMatch(t *testing.T) {
	lex := New[rune, int]()
	lex.InsertIter(iterx.Runes("if"), 1)
	lex.InsertIter(iterx.Runes("iffy"), 2)
	lex.InsertIter(iterx.Runes("else"), 3)
	lex.InsertIter(iterx.Runes("elseif"), 4)

	tests := []struct {
		input   string
		want    int
		wantLen int
		ok      bool
	}{
		{"if", 1, 2, true},
		{"iffy", 1, 2, true},   // returns "if" (shortest)
		{"iffyyy", 1, 2, true}, // returns "if" (shortest)
		{"else", 3, 4, true},
		{"elseif", 3, 4, true},    // returns "else" (shortest)
		{"elseifxxx", 3, 4, true}, // returns "else" (shortest)
		{"i", 0, 0, false},        // no complete match
		{"e", 0, 0, false},        // no complete match
		{"while", 0, 0, false},    // no match at all
		{"", 0, 0, false},         // empty input
	}

	for _, tt := range tests {
		got, gotLen, ok := lex.FirstMatch(iterx.Runes(tt.input))
		if got != tt.want || gotLen != tt.wantLen || ok != tt.ok {
			t.Errorf("FirstMatch(%q) = (%v, %v, %v), want (%v, %v, %v)", tt.input, got, gotLen, ok, tt.want, tt.wantLen, tt.ok)
		}
	}
}

func TestLongestMatch(t *testing.T) {
	lex := New[rune, int]()
	lex.InsertIter(iterx.Runes("if"), 1)
	lex.InsertIter(iterx.Runes("iffy"), 2)
	lex.InsertIter(iterx.Runes("else"), 3)
	lex.InsertIter(iterx.Runes("elseif"), 4)

	tests := []struct {
		input   string
		want    int
		wantLen int
		ok      bool
	}{
		{"if", 1, 2, true},
		{"iffy", 2, 4, true},   // returns "iffy" (longest)
		{"iffyyy", 2, 4, true}, // returns "iffy" (longest available)
		{"iff", 1, 2, true},    // returns "if" (longest available)
		{"else", 3, 4, true},
		{"elseif", 4, 6, true},    // returns "elseif" (longest)
		{"elseifxxx", 4, 6, true}, // returns "elseif" (longest available)
		{"elsei", 3, 4, true},     // returns "else" (longest available)
		{"i", 0, 0, false},        // no complete match
		{"e", 0, 0, false},        // no complete match
		{"while", 0, 0, false},    // no match at all
		{"", 0, 0, false},         // empty input
	}

	for _, tt := range tests {
		got, gotLen, ok := lex.LongestMatch(iterx.Runes(tt.input))
		if got != tt.want || gotLen != tt.wantLen || ok != tt.ok {
			t.Errorf("LongestMatch(%q) = (%v, %v, %v), want (%v, %v, %v)", tt.input, got, gotLen, ok, tt.want, tt.wantLen, tt.ok)
		}
	}
}

func TestSliceMethods(t *testing.T) {
	lex := New[byte, string]()
	lex.InsertSlice([]byte{0x00, 0x01}, "header")
	lex.InsertSlice([]byte{0x00, 0x01, 0x02}, "extended")
	lex.InsertSlice([]byte{0xFF}, "end")

	// MatchSlice
	if v, ok := lex.MatchSlice([]byte{0x00, 0x01}); !ok || v != "header" {
		t.Errorf("MatchSlice header failed: got (%v, %v)", v, ok)
	}
	if v, ok := lex.MatchSlice([]byte{0xFF}); !ok || v != "end" {
		t.Errorf("MatchSlice end failed: got (%v, %v)", v, ok)
	}
	if _, ok := lex.MatchSlice([]byte{0x00}); ok {
		t.Error("MatchSlice partial should fail")
	}

	// FirstMatchSlice
	if v, n, ok := lex.FirstMatchSlice([]byte{0x00, 0x01, 0x02, 0x03}); !ok || v != "header" || n != 2 {
		t.Errorf("FirstMatchSlice should return header with len 2: got (%v, %v, %v)", v, n, ok)
	}

	// LongestMatchSlice
	if v, n, ok := lex.LongestMatchSlice([]byte{0x00, 0x01, 0x02, 0x03}); !ok || v != "extended" || n != 3 {
		t.Errorf("LongestMatchSlice should return extended with len 3: got (%v, %v, %v)", v, n, ok)
	}
}

func TestSliceMethodsWithInts(t *testing.T) {
	lex := New[int, string]()
	lex.InsertSlice([]int{1, 2, 3}, "a")
	lex.InsertSlice([]int{1, 2}, "b")
	lex.InsertSlice([]int{4, 5}, "c")

	if v, ok := lex.MatchSlice([]int{1, 2, 3}); !ok || v != "a" {
		t.Errorf("MatchSlice [1,2,3] failed: got (%v, %v)", v, ok)
	}
	if v, ok := lex.MatchSlice([]int{1, 2}); !ok || v != "b" {
		t.Errorf("MatchSlice [1,2] failed: got (%v, %v)", v, ok)
	}
	if v, n, ok := lex.FirstMatchSlice([]int{1, 2, 3, 4}); !ok || v != "b" || n != 2 {
		t.Errorf("FirstMatchSlice should return b with len 2: got (%v, %v, %v)", v, n, ok)
	}
	if v, n, ok := lex.LongestMatchSlice([]int{1, 2, 3, 4}); !ok || v != "a" || n != 3 {
		t.Errorf("LongestMatchSlice should return a with len 3: got (%v, %v, %v)", v, n, ok)
	}
}

func TestUnicodeStrings(t *testing.T) {
	lex := New[rune, int]()
	lex.InsertIter(iterx.Runes("日本"), 1)
	lex.InsertIter(iterx.Runes("日本語"), 2)
	lex.InsertIter(iterx.Runes("中文"), 3)

	if v, ok := lex.Match(iterx.Runes("日本")); !ok || v != 1 {
		t.Errorf("Match 日本 failed: got (%v, %v)", v, ok)
	}
	if v, ok := lex.Match(iterx.Runes("日本語")); !ok || v != 2 {
		t.Errorf("Match 日本語 failed: got (%v, %v)", v, ok)
	}
	if v, _, ok := lex.FirstMatch(iterx.Runes("日本語テスト")); !ok || v != 1 {
		t.Errorf("FirstMatch should return 1: got (%v, %v)", v, ok)
	}
	if v, _, ok := lex.LongestMatch(iterx.Runes("日本語テスト")); !ok || v != 2 {
		t.Errorf("LongestMatch should return 2: got (%v, %v)", v, ok)
	}
}

func TestOverwritePattern(t *testing.T) {
	lex := New[rune, int]()
	lex.InsertIter(iterx.Runes("test"), 1)
	lex.InsertIter(iterx.Runes("test"), 2) // overwrite

	if v, ok := lex.Match(iterx.Runes("test")); !ok || v != 2 {
		t.Errorf("Match should return overwritten value 2: got (%v, %v)", v, ok)
	}
}

func TestEmptyPattern(t *testing.T) {
	lex := New[rune, int]()
	lex.InsertIter(iterx.Runes(""), 1) // empty pattern

	// Empty input should match empty pattern
	if v, ok := lex.Match(iterx.Runes("")); !ok || v != 1 {
		t.Errorf("Match empty should return 1: got (%v, %v)", v, ok)
	}
	// Non-empty input should not match empty pattern via Match
	if _, ok := lex.Match(iterx.Runes("x")); ok {
		t.Error("Match non-empty should fail for empty pattern")
	}
}

func TestSingleCharPatterns(t *testing.T) {
	lex := New[rune, int]()
	lex.InsertIter(iterx.Runes("a"), 1)
	lex.InsertIter(iterx.Runes("b"), 2)
	lex.InsertIter(iterx.Runes("ab"), 3)

	if v, ok := lex.Match(iterx.Runes("a")); !ok || v != 1 {
		t.Errorf("Match a failed: got (%v, %v)", v, ok)
	}
	if v, _, ok := lex.FirstMatch(iterx.Runes("ab")); !ok || v != 1 {
		t.Errorf("FirstMatch ab should return 1: got (%v, %v)", v, ok)
	}
	if v, _, ok := lex.LongestMatch(iterx.Runes("ab")); !ok || v != 3 {
		t.Errorf("LongestMatch ab should return 3: got (%v, %v)", v, ok)
	}
}

func TestMixedInsert(t *testing.T) {
	lex := New[rune, int]()
	lex.InsertIter(iterx.Runes("hello"), 1)
	lex.InsertSlice([]rune{'w', 'o', 'r', 'l', 'd'}, 2)
	lex.InsertIter(iterx.Runes("helloworld"), 3)

	if v, ok := lex.Match(iterx.Runes("hello")); !ok || v != 1 {
		t.Errorf("Match hello failed: got (%v, %v)", v, ok)
	}
	if v, ok := lex.Match(iterx.Runes("world")); !ok || v != 2 {
		t.Errorf("Match world failed: got (%v, %v)", v, ok)
	}
	if v, ok := lex.Match(iterx.Runes("helloworld")); !ok || v != 3 {
		t.Errorf("Match helloworld failed: got (%v, %v)", v, ok)
	}
}

func TestRunesReverse(t *testing.T) {
	tests := []struct {
		input string
		want  []rune
	}{
		{"hello", []rune{'o', 'l', 'l', 'e', 'h'}},
		{"abc", []rune{'c', 'b', 'a'}},
		{"", []rune{}},
		{"a", []rune{'a'}},
		{"日本語", []rune{'語', '本', '日'}},
	}

	for _, tt := range tests {
		var got []rune
		for r := range iterx.RunesReverse(tt.input) {
			got = append(got, r)
		}
		if len(got) != len(tt.want) {
			t.Errorf("iterx.RunesReverse(%q) length = %d, want %d", tt.input, len(got), len(tt.want))
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("iterx.RunesReverse(%q)[%d] = %c, want %c", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

func TestRunesReverseWithLexer(t *testing.T) {
	lex := New[rune, int]()
	// Insert reversed patterns
	lex.InsertIter(iterx.RunesReverse("abc"), 1)  // stores as "cba"
	lex.InsertIter(iterx.RunesReverse("abcd"), 2) // stores as "dcba"

	// Match reversed input
	if v, ok := lex.Match(iterx.RunesReverse("abc")); !ok || v != 1 {
		t.Errorf("Match reversed abc failed: got (%v, %v)", v, ok)
	}
	if v, ok := lex.Match(iterx.RunesReverse("abcd")); !ok || v != 2 {
		t.Errorf("Match reversed abcd failed: got (%v, %v)", v, ok)
	}
}

func TestRunes(t *testing.T) {
	tests := []struct {
		input string
		want  []rune
	}{
		{"hello", []rune{'h', 'e', 'l', 'l', 'o'}},
		{"", []rune{}},
		{"日本語", []rune{'日', '本', '語'}},
	}

	for _, tt := range tests {
		var got []rune
		for r := range iterx.Runes(tt.input) {
			got = append(got, r)
		}
		if len(got) != len(tt.want) {
			t.Errorf("iterx.Runes(%q) length = %d, want %d", tt.input, len(got), len(tt.want))
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("iterx.Runes(%q)[%d] = %c, want %c", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}
