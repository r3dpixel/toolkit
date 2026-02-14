package symbols

import (
	"regexp"
	"strings"
)

// Symbols all symbols in slice format
// Dash must be last because regexp.QuoteMeta doesn't escape dash
var Symbols = []string{
	Tick, Tilde, Exclamation, At, Hash, Dollar, Percent, Circumflex, And, Asterisk, BracketOpen, BracketClose,
	Underscore, Plus, Equal, SquareBracketOpen, SquareBracketClose, CurlyBracketOpen, CurlyBracketClose,
	Or, Slash, Backslash, Colon, Semicolon, Question, Apostrophe, LesserThan, GreaterThan, Comma, Dot, Dash,
}

// CapitalizeAfter symbols that [ToTitle] needs to capitalize after
// Dash must be last because regexp.QuoteMeta doesn't escape dash
var CapitalizeAfter = []string{
	Exclamation, At, Hash, And, BracketOpen, BracketClose, Underscore, Plus, Equal, SquareBracketOpen, SquareBracketClose,
	CurlyBracketOpen, CurlyBracketClose, Or, Slash, Backslash, Colon, Semicolon, Question, LesserThan, GreaterThan,
	Comma, Dot, Dash,
}

// CapitalizeAfterString symbols that [ToTitle] needs to capitalize after (in string format)
var CapitalizeAfterString = strings.Join(CapitalizeAfter, "")

// CapitalizeAfterRegExp regex that matches all the symbols after capitalization is necessary
var CapitalizeAfterRegExp = regexp.MustCompile(`[` + regexp.QuoteMeta(CapitalizeAfterString) + `]`)

// WhiteSpaceRegExp regex matching any whitespace
var WhiteSpaceRegExp = regexp.MustCompile(`\s+`)

// NonAsciiRegExp regex matching: non-ASCII characters, non-ASCII whitespace
var NonAsciiRegExp = regexp.MustCompile(`[^\x00-\x7F]+`)

// NonAsciiWhiteSpaceRegExp regex matching: non-ASCII characters, ALL whitespace
var NonAsciiWhiteSpaceRegExp = regexp.MustCompile(`[^\x00-\x7F]+|\s+`)

// NonAlphaNumericRegExp regex matching: non-alphanumeric ASCII characters, non-ASCII whitespace
var NonAlphaNumericRegExp = regexp.MustCompile(`[^a-zA-Z0-9\s]+`)

// NonAlphaNumericWhiteSpaceRegExp regex matching: non-alphanumeric ASCII characters, ALL whitespace
var NonAlphaNumericWhiteSpaceRegExp = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// SymbolsRegExp regex matching: non-alphanumeric characters
var SymbolsRegExp = regexp.MustCompile(`[^\p{L}\p{N}\s]+`)

// SymbolsWhiteSpaceRegExp regex matching: non-alphanumeric characters, ALL whitespace
var SymbolsWhiteSpaceRegExp = regexp.MustCompile(`[^\p{L}\p{N}]+`)

// NonAsciiSymbolsRegExp regex matching: non-ASCII symbols (ie. emojis, ligatures, misc symbols), non-ASCII whitespace
var NonAsciiSymbolsRegExp = regexp.MustCompile(`[^\x00-\x7F\p{L}\p{N}]+`)

// NonAsciiSymbolsWhiteSpaceRegExp regex matching: non-ASCII symbols (ie. emojis, ligatures, misc symbols), ALL whitespace
var NonAsciiSymbolsWhiteSpaceRegExp = regexp.MustCompile(`[^\x00-\x7F\p{L}\p{N}]+|\s+`)

// InvalidPathRegExp regex for path string sanitization (matches everything except path allowed characters)
var InvalidPathRegExp = regexp.MustCompile(`[^a-zA-Z0-9-_\s.]`)

// RFC3339NanoRegExp regex matching a timestamp in RFC3339Nano format
var RFC3339NanoRegExp = regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(\.\d{1,9})?([+-](\d{2}):(\d{2})|Z)$`)

// AbnormalQuotes string containing various abnormal quote characters                                                                                                           │ │
var AbnormalQuotes = `‹›〈〉“”«»「」｢｣《》„‟〝〞`

// AbnormalApostrophes string containing various abnormal apostrophes characters
var AbnormalApostrophes = `‘‛’`

// AbnormalCommas string containing various abnormal commas characters
var AbnormalCommas = `‚`
