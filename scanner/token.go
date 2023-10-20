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
	LeftAngleBracket
	RightAngleBracket
	SingleQuote
	DoubleQuote
	TripleQuote
	TagStart
	TagSelfClosing
	ClosingTagStart
)

func (t Token) String() (retv string) {
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
	}

	return ""
}

var TokenLiterals = map[Token]string{
	Assign:            "=",
	Slash:             "/",
	LeftAngleBracket:  "<",
	RightAngleBracket: ">",
	SingleQuote:       "'",
	DoubleQuote:       "\"",
	TripleQuote:       "`",
	ClosingTagStart:   "</",
	TagSelfClosing:    "/>",
}

type TokenItem struct {
	Token    Token
	Literal  string
	Position int
	Line     int
}

func getTokenLiteral(token Token) string {
	lt, found := TokenLiterals[token]
	if !found {
		return ""
	}

	return lt
}

func NewTokenItem(token Token, literal string, position int) *TokenItem {
	ti := &TokenItem{
		Token:    token,
		Literal:  literal,
		Position: position,
	}

	return ti
}
