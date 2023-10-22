package scanner

var eof = rune(0)

type Token int

const (
	None Token = iota - 1
	EOF
	Other
	NewLine
	WhiteSpace
	// Identifier matches [a-ZA-Z0-9-]
	Identifier
	// String matches "|'[\w\W\s]"|'
	String
	Assign
	Slash
	BackSlash
	LeftAngleBracket
	RightAngleBracket
	SingleQuote
	DoubleQuote
	TripleQuote
	TagStart
	TagSelfClosing
	ClosingTagStart
	EscSingleQuote
	EscDoubleQuote
)

func (t Token) String() string {
	var retv string
	switch t {
	case None:
		retv = "None"
	case EOF:
		retv = "EOF"
	case Other:
		retv = "Other"
	case NewLine:
		retv = "NewLine"
	case WhiteSpace:
		retv = "WhiteSpace"
	case Identifier:
		retv = "Identifier"
	case String:
		retv = "String"
	case Assign:
		retv = "Assign"
	case Slash:
		retv = "Slash"
	case BackSlash:
		retv = "BackSlash"
	case LeftAngleBracket:
		retv = "LeftAngleBracket"
	case RightAngleBracket:
		retv = "RightAngleBracket"
	case SingleQuote:
		retv = "SingleQuote"
	case DoubleQuote:
		retv = "DoubleQuote"
	case TripleQuote:
		retv = "TripleQuote"
	case TagStart:
		retv = "TagStart"
	case TagSelfClosing:
		retv = "TagSelfClosing"
	case ClosingTagStart:
		retv = "ClosingTagStart"
	case EscSingleQuote:
		retv = "EscSingleQuote"
	case EscDoubleQuote:
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
