package symbols

// Symbols ASCII special symbols (Byte)
const (
	QuoteByte              byte = '"'
	TickByte               byte = '`'
	TildeByte              byte = '~'
	ExclamationByte        byte = '!'
	AtByte                 byte = '@'
	HashByte               byte = '#'
	DollarByte             byte = '$'
	PercentByte            byte = '%'
	CircumflexByte         byte = '^'
	AndByte                byte = '&'
	AsteriskByte           byte = '*'
	BracketOpenByte        byte = '('
	BracketCloseByte       byte = ')'
	DashByte               byte = '-'
	UnderscoreByte         byte = '_'
	PlusByte               byte = '+'
	EqualByte              byte = '='
	SquareBracketOpenByte  byte = '['
	SquareBracketCloseByte byte = ']'
	CurlyBracketOpenByte   byte = '{'
	CurlyBracketCloseByte  byte = '}'
	OrByte                 byte = '|'
	SlashByte              byte = '/'
	BackslashByte          byte = '\\'
	ColonByte              byte = ':'
	SemicolonByte          byte = ';'
	QuestionByte           byte = '?'
	ApostropheByte         byte = '\''
	LesserThanByte         byte = '<'
	GreaterThanByte        byte = '>'
	CommaByte              byte = ','
	DotByte                byte = '.'
	SpaceByte              byte = ' '
	TabByte                byte = '\t'
	NewLineByte            byte = '\n'
)

// NullBytes byte string representing JSON null value
var NullBytes = []byte("null")
