package scanner

var eof = rune(0)

type Token int

const (
	TokenNone Token = iota - 1
	TokenEOF
	TokenOther
	TokenNewLine
	TokenWhiteSpace
	// TokenIdentifier matches [a-ZA-Z0-9-]
	TokenIdentifier
	// TokenString matches "|'[\w\W\s]"|'
	TokenString
	TokenAssign
	TokenSlash
	TokenBackSlash
	TokenLeftAngleBracket
	TokenRightAngleBracket
	TokenSingleQuote
	TokenDoubleQuote
	TokenTripleQuote
	TokenTagStart
	TokenTagSelfClosing
	TokenClosingTagStart
	TokenEscSingleQuote
	TokenEscDoubleQuote
)

func (t Token) String() string {
	var retv string
	switch t {
	case TokenNone:
		retv = "None"
	case TokenEOF:
		retv = "EOF"
	case TokenOther:
		retv = "Other"
	case TokenNewLine:
		retv = "NewLine"
	case TokenWhiteSpace:
		retv = "WhiteSpace"
	case TokenIdentifier:
		retv = "Identifier"
	case TokenString:
		retv = "String"
	case TokenAssign:
		retv = "Assign"
	case TokenSlash:
		retv = "Slash"
	case TokenBackSlash:
		retv = "BackSlash"
	case TokenLeftAngleBracket:
		retv = "LeftAngleBracket"
	case TokenRightAngleBracket:
		retv = "RightAngleBracket"
	case TokenSingleQuote:
		retv = "SingleQuote"
	case TokenDoubleQuote:
		retv = "DoubleQuote"
	case TokenTripleQuote:
		retv = "TripleQuote"
	case TokenTagStart:
		retv = "TagStart"
	case TokenTagSelfClosing:
		retv = "TagSelfClosing"
	case TokenClosingTagStart:
		retv = "ClosingTagStart"
	case TokenEscSingleQuote:
		retv = "EscSingleQuote"
	case TokenEscDoubleQuote:
		retv = "EscDoubleQuote"
	}

	return retv
}

type TokenItem struct {
	Token         Token
	Literal       string
	StartPosition int
	EndPosition   int
	Line          int
}
