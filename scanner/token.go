package scanner

var eof = rune(0)

type Token int

const (
	Other Token = iota - 1
	EOF
	NewLine
	Whitespace
	Identifier
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
