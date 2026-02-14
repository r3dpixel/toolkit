package stringsx

import (
	"testing"

	"github.com/r3dpixel/toolkit/symbols"
	"github.com/stretchr/testify/assert"
)

func TestFromBytes(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expected string
	}{
		{"Empty bytes", []byte{}, ""},
		{"Simple string", []byte("hello"), "hello"},
		{"Unicode string", []byte("你好世界"), "你好世界"},
		{"Binary data", []byte{0, 1, 2, 255}, string([]byte{0, 1, 2, 255})},
		{"Whitespace", []byte(" \t\n\r"), " \t\n\r"},
		{"Numbers", []byte("12345"), "12345"},
		{"Special chars", []byte("!@#$%^&*()"), "!@#$%^&*()"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FromBytes(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestToBytes(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []byte
	}{
		{"Empty string", "", []byte{}},
		{"Simple string", "hello", []byte("hello")},
		{"Unicode string", "你好世界", []byte("你好世界")},
		{"Whitespace", " \t\n\r", []byte(" \t\n\r")},
		{"Numbers", "12345", []byte("12345")},
		{"Special chars", "!@#$%^&*()", []byte("!@#$%^&*()")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ToBytes(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestFromBytesToBytesRoundTrip(t *testing.T) {
	testCases := []string{
		"",
		"hello world",
		"你好世界",
		"test with 123 numbers",
		"special !@#$% chars",
		" \t\n\r whitespace ",
	}

	for _, original := range testCases {
		t.Run("RoundTrip: "+original, func(t *testing.T) {
			// String -> Bytes -> String
			bytes := ToBytes(original)
			result := FromBytes(bytes)
			assert.Equal(t, original, result)
		})
	}
}

func TestContainsAny(t *testing.T) {
	testCases := []struct {
		name          string
		str           string
		matches       []string
		expectedMatch string
		expectedBool  bool
	}{
		{name: "Successful Match", str: "hello beautiful world", matches: []string{"world", "foo"}, expectedMatch: "world", expectedBool: true},
		{name: "No Match", str: "hello world", matches: []string{"foo", "bar"}, expectedMatch: "", expectedBool: false},
		{name: "Overlapping Match", str: "hello beautiful world", matches: []string{"beautiful", "world"}, expectedMatch: "beautiful", expectedBool: true},
		{name: "Substring Match", str: "this is a test", matches: []string{"is a", "is"}, expectedMatch: "is a", expectedBool: true},
		{name: "Empty String Input", str: "", matches: []string{"world"}, expectedMatch: "", expectedBool: false},
		{name: "Empty Matches Slice", str: "hello world", matches: []string{}, expectedMatch: "", expectedBool: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			match, ok := ContainsAny(tc.str, tc.matches...)
			assert.Equal(t, tc.expectedBool, ok)
			assert.Equal(t, tc.expectedMatch, match)
		})
	}
}

func TestIsBlankAndIsNotBlank(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		isBlank bool
	}{
		{name: "Empty String", input: "", isBlank: true},
		{name: "Single Space", input: " ", isBlank: true},
		{name: "Tab", input: "\t", isBlank: true},
		{name: "Newline", input: "\n", isBlank: true},
		{name: "Carriage Return", input: "\r", isBlank: true},
		{name: "Mixed Whitespace", input: " \t \n \r ", isBlank: true},
		{name: "Unicode Whitespace", input: "\u2002\u2003", isBlank: true},
		{name: "String with Content", input: "hello", isBlank: false},
		{name: "Whitespace with Content", input: " hello ", isBlank: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.isBlank, IsBlank(tc.input))
			assert.Equal(t, !tc.isBlank, IsNotBlank(tc.input))
		})
	}
}

func TestIsBlankAndIsNotBlankPtr(t *testing.T) {
	strWithContent := "hello"
	strWithSpace := " "
	emptyStr := ""

	testCases := []struct {
		name     string
		input    *string
		expected bool
	}{
		{name: "Nil Pointer", input: nil, expected: true},
		{name: "Pointer to Empty String", input: &emptyStr, expected: true},
		{name: "Pointer to Whitespace String", input: &strWithSpace, expected: true},
		{name: "Pointer to String with Content", input: &strWithContent, expected: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, IsBlankPtr(tc.input))
			assert.Equal(t, !tc.expected, IsNotBlankPtr(tc.input))
		})
	}
}

func TestUpdateIfExists(t *testing.T) {
	t.Run("Update with non-blank value", func(t *testing.T) {
		initial := "initial"
		UpdateIfExists(&initial, "updated")
		assert.Equal(t, "updated", initial)
	})

	t.Run("Do not update with blank value", func(t *testing.T) {
		initial := "initial"
		UpdateIfExists(&initial, "  ")
		assert.Equal(t, "initial", initial)
	})

	t.Run("Do not panic on nil pointer", func(t *testing.T) {
		// This test has no assertion, it just ensures the function doesn't panic.
		var ptr *string
		UpdateIfExists(ptr, "some value")
	})
}

func TestJoin(t *testing.T) {
	t.Run("JoinNonBlank", func(t *testing.T) {
		result := JoinNonBlank(", ", "one", "  ", "two", "", " three ")
		assert.Equal(t, "one, two, three", result)
	})

	t.Run("JoinNonBlank with all blank", func(t *testing.T) {
		result := JoinNonBlank(", ", " ", "\t", "")
		assert.Equal(t, "", result)
	})

	t.Run("Join with custom filter", func(t *testing.T) {
		filter := func(str string) bool { return len(str) > 3 }
		result := Join(" | ", filter, "apple", "cat", "apricot", "dog", "cherry")
		assert.Equal(t, "apple | apricot | cherry", result)
	})
}

func TestWhiteSpaceRegExp(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Mixed whitespace", "  hello \t  world\n\r", "helloworld"},
		{"Only whitespace", " \t\n\r", ""},
		{"Single space", "a b", "ab"},
		{"Unicode whitespace", "hello\u2002world\u2003test", "hello\u2002world\u2003test"},
		{"No whitespace", "helloworld", "helloworld"},
		{"Leading/trailing whitespace", "  text  ", "text"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Remove(tc.input, symbols.WhiteSpaceRegExp)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestNonAsciiRegExp(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Mixed Unicode", "HelloÂ World你好 éçà", "Hello World "},
		{"Japanese/Chinese", "Hello世界こんにちは", "Hello"},
		{"Emojis", "Hello 🌍 World 🚀", "Hello  World "},
		{"Arabic/Cyrillic", "Hello مرحبا Привет", "Hello  "},
		{"Only ASCII", "Hello World 123", "Hello World 123"},
		{"Only non-ASCII", "你好世界", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Remove(tc.input, symbols.NonAsciiRegExp)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestNonAsciiWhiteSpaceRegExp(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Non-ASCII and whitespace", "Hello 你好 World", "HelloWorld"},
		{"Mixed with spaces", "Test 测试 Space", "TestSpace"},
		{"Only whitespace", "   \t\n", ""},
		{"Only non-ASCII", "你好世界", ""},
		{"ASCII only", "Hello World", "HelloWorld"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Remove(tc.input, symbols.NonAsciiWhiteSpaceRegExp)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestNonAlphaNumericRegExp(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Mixed symbols", "Hello!@# World$%^ 123", "Hello World 123"},
		{"Only symbols", "!@#$%^&*()", ""},
		{"Alphanumeric only", "Hello123World", "Hello123World"},
		{"With whitespace preserved", "Test 123 End", "Test 123 End"},
		{"Unicode letters preserved", "Café naïve", "Caf nave"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Remove(tc.input, symbols.NonAlphaNumericRegExp)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestNonAlphaNumericWhiteSpaceRegExp(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Remove symbols and whitespace", "Hello!@# World$%^ 123", "HelloWorld123"},
		{"Only symbols and spaces", "!@# $%^ ", ""},
		{"Letters and numbers only", "Hello123", "Hello123"},
		{"Mixed Unicode", "Test测试123", "Test123"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Remove(tc.input, symbols.NonAlphaNumericWhiteSpaceRegExp)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSymbolsRegExp(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"All symbols", "`Hello~!@#$World%^&*()-_=+[]{}|\\:;'<,>.?/123", "HelloWorld123"},
		{"Unicode symbols", "Hello™© World® 123", "Hello World 123"},
		{"Mixed symbols and emojis", "Test🚀!@# World", "Test World"},
		{"Letters/numbers/whitespace only", "Hello World 123", "Hello World 123"},
		{"Only symbols", "!@#$%^&*()", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Remove(tc.input, symbols.SymbolsRegExp)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSymbolsWhiteSpaceRegExp(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Symbols and whitespace", "Hello!@# World$%^ 123", "HelloWorld123"},
		{"Only symbols and spaces", "!@# $%^ ", ""},
		{"Unicode symbols and spaces", "Test™© World® ", "TestWorld"},
		{"Letters and numbers only", "Hello123", "Hello123"},
		{"Letters and numbers only", "Café naïve123", "Cafénaïve123"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Remove(tc.input, symbols.SymbolsWhiteSpaceRegExp)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestNonAsciiSymbolsRegExp(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Keep ASCII symbols, remove non-ASCII symbols", "Hello!@# 你好™© World$%^", "Hello!@# 你好 World$%^"},
		{"Only non-ASCII symbols", "™©®", ""},
		{"Only ASCII symbols", "!@#$%", "!@#$%"},
		{"Mixed with letters", "Test™Hello©World®", "TestHelloWorld"},
		{"Emojis removed", "Hello🚀World💫", "HelloWorld"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Remove(tc.input, symbols.NonAsciiSymbolsRegExp)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestNonAsciiSymbolsWhiteSpaceRegExp(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"Remove non-ASCII symbols and whitespace", "Hello™© World® Test", "HelloWorldTest"},
		{"Keep ASCII symbols", "Test!@# World$%^", "Test!@#World$%^"},
		{"Only non-ASCII symbols and spaces", "™© ®", ""},
		{"Mixed content", "Hello🚀 World!@# 123", "HelloWorld!@#123"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Remove(tc.input, symbols.NonAsciiSymbolsWhiteSpaceRegExp)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestToTitle(t *testing.T) {
	assert.Equal(t, "This Is A Title", ToTitle("this is a title"))
	assert.Equal(t, "Already Titled", ToTitle("Already Titled"))
	assert.Equal(t, "Mixedcase With Spaces", ToTitle("mixedCase with spaces"))
	assert.Equal(t, "123 Numbers", ToTitle("123 numbers"))
}

func TestStripQuotes(t *testing.T) {
	// Assuming symbols.QuoteByte is '"'
	assert.Equal(t, `content`, Unquote(`"content"`), "Should strip matching double quotes")
	assert.Equal(t, ``, Unquote(`""`), "Should handle empty quoted string")
	assert.Equal(t, `no quotes`, Unquote(`no quotes`), "Should not change unquoted string")
	assert.Equal(t, `"unmatched`, Unquote(`"unmatched`), "Should not strip unmatched leading quote")
	assert.Equal(t, `unmatched"`, Unquote(`unmatched"`), "Should not strip unmatched trailing quote")
	assert.Equal(t, `"`, Unquote(`"`), "Should not change single quote")
}

func TestNormalizeSymbols(t *testing.T) {
	input := `〝opposites〞 ‹angular› 〈gills〉 ‘smart’ “curly” «guillemets» „low“ 「corner」 《brackets》 `
	expected := `"opposites" "angular" "gills" 'smart' "curly" "guillemets" "low" "corner" "brackets"`
	assert.Equal(t, expected, NormalizeSymbols(input))
}

func TestNormalizeSymbolsPtr(t *testing.T) {
	t.Run("With valid pointer", func(t *testing.T) {
		inputStr := ` 'test' `
		NormalizeSymbolsPtr(&inputStr)
		assert.Equal(t, `'test'`, inputStr)
	})

	t.Run("With nil pointer", func(t *testing.T) {
		var nilPtr *string
		NormalizeSymbolsPtr(nilPtr) // Should not panic
		assert.Nil(t, nilPtr)
	})
}

func TestIsAsciiSpace(t *testing.T) {
	testCases := []struct {
		name     string
		input    byte
		expected bool
	}{
		{"Space character", ' ', true},
		{"Tab character", '\t', true},
		{"Newline character", '\n', true},
		{"Carriage return", '\r', true},
		{"Vertical tab", '\v', true},
		{"Form feed", '\f', true},
		{"Regular letter", 'a', false},
		{"Number", '5', false},
		{"Symbol", '!', false},
		{"Null byte", 0, false},
		{"High ASCII", 255, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsAsciiSpace(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
