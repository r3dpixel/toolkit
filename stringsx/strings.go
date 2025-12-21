package stringsx

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"github.com/r3dpixel/toolkit/symbols"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// asciiSpace characters that represent whitespace in the ASCII code
var asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}

// FromBytes Converts a byte array to a string (without copying the data)
func FromBytes(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// ToBytes Converts a string to a byte array (without copying the data)
func ToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}

// ContainsAny returns the first string from matches that is contained in str, and true (empty string and false otherwise)
func ContainsAny(str string, matches ...string) (string, bool) {
	for _, match := range matches {
		if strings.Contains(str, match) {
			return match, true
		}
	}
	return "", false
}

// IsNotBlankPtr Checks if the string pointer is NOT nil, empty, or whitespace
func IsNotBlankPtr(ptr *string) bool {
	return !IsBlankPtr(ptr)
}

// IsBlankPtr Checks if the string pointer IS nil, empty, or whitespace
func IsBlankPtr(ptr *string) bool {
	return ptr == nil || IsBlank(*ptr)
}

// IsNotBlank Checks if the string is NOT empty or whitespace
func IsNotBlank(value string) bool {
	return !IsBlank(value)
}

// IsBlank Checks if the string is empty or whitespace
func IsBlank(value string) bool {
	// If the string is empty, return true
	if len(value) == 0 {
		return true
	}

	// Check whitespace
	for i := 0; i < len(value); i++ {
		// Get the character
		c := value[i]
		// If not ASCII, use the full Unicode check
		if c >= utf8.RuneSelf {
			return isBlankUnicode(value[i:])
		}
		// ASCII fast path check
		if asciiSpace[c] == 0 {
			// Non-whitespace character found, return false
			return false
		}
	}
	// If no non-whitespace characters were found, return true
	return true
}

// isBlankUnicode slow path for full Unicode whitespace check
func isBlankUnicode(value string) bool {
	// Check for non-whitespace
	for _, r := range value {
		// If a non-whitespace character is found, return false
		if !unicode.IsSpace(r) {
			return false
		}
	}
	// If no non-whitespace characters were found, return true
	return true
}

// UpdateIfExists Updates the string pointer if the value is not blank
func UpdateIfExists(ptr *string, value string) {
	if ptr != nil && IsNotBlank(value) {
		*ptr = value
	}
}

// Join filters and joins strings
func Join(separator string, filter func(str string) bool, strs ...string) string {
	// Create slice for storing non-blank strings
	filtered := make([]string, 0, len(strs))
	// Filter values and eliminate blank strings
	for index := range strs {
		if filter(strs[index]) {
			filtered = append(filtered, strings.TrimSpace(strs[index]))
		}
	}
	// Join the filtered strings using the separator
	return strings.Join(filtered, separator)
}

// JoinNonBlank joins non-blank strings
func JoinNonBlank(separator string, strs ...string) string {
	return Join(separator, IsNotBlank, strs...)
}

// Remove removes from the string all matches of the given regex
func Remove(value string, rxp *regexp.Regexp) string {
	return rxp.ReplaceAllString(value, "")
}

// ToTitle transforms the given string into the title counterpart ("this is a title -> This Is A Title")
func ToTitle(value string) string {
	// Caser for transforming strings into their title counterparts
	return cases.Title(language.English).String(value)
}

// Unquote removes quotes from the string if it is a correctly quoted string (`"abc"` -> `abc`)
func Unquote(s string) string {
	if unquoted, err := strconv.Unquote(s); err == nil {
		return unquoted
	}
	return s
}

// NormalizeSymbols - Replaces all abnormal quotes characters with the ASCII version '"' (input is a string)
// NOTE: Also removes all trailing whitespace
var quoteReplacer = initQuoteReplacer()

func NormalizeSymbols(value string) string {
	return quoteReplacer.Replace(strings.TrimSpace(value))
}

// NormalizeSymbolsPtr - Replaces all abnormal quotes characters with the ASCII version '"' (input is a pointer to string)
func NormalizeSymbolsPtr(value *string) {
	if value == nil {
		return
	}

	*value = NormalizeSymbols(*value)
}

// IsAsciiSpace returns true if the given byte represents and ASCII code for whitespace, false otherwise
func IsAsciiSpace(c byte) bool {
	return asciiSpace[c] == 1
}

// initQuoteReplacer creates a string replacer that will correct all abnormal quotes, apostrophes, or commas symbols
func initQuoteReplacer() *strings.Replacer {
	// Create a slice of pairs for replacing all abnormal quotes, apostrophes, or commas
	noSymbols := len(symbols.AbnormalQuotes) + len(symbols.AbnormalApostrophes) + len(symbols.AbnormalCommas)
	// Create the replacer (the length of the pairs is twice the number of abnormal symbols)
	pairs := make([]string, 0, 2*noSymbols)

	// InsertIter pairs for replacing all abnormal quotes
	for _, char := range symbols.AbnormalQuotes {
		pairs = append(pairs, string(char), symbols.Quote)
	}

	// InsertIter pairs for replacing all abnormal apostrophes
	for _, char := range symbols.AbnormalApostrophes {
		pairs = append(pairs, string(char), symbols.Apostrophe)
	}

	// InsertIter pairs for replacing all abnormal commas
	for _, char := range symbols.AbnormalCommas {
		pairs = append(pairs, string(char), symbols.Comma)
	}

	// Create the replacer
	return strings.NewReplacer(pairs...)
}
