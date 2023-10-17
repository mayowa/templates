package scanner

var eof = rune(0)

type Token int

const (
	OTHER Token = iota - 1
	EOF
	WHITESPACE
	IDENTIFIER
	ASSIGN
	SLASH
	LT
	GT
	SingleQuote
	DoubleQuote
	TripleQuote
)

var TokenLiterals = map[Token]string{
	ASSIGN:      "=",
	SLASH:       "/",
	LT:          "<",
	GT:          ">",
	SingleQuote: "'",
	DoubleQuote: "\"",
	TripleQuote: "`",
}

func getTokenLiteral(token Token) string {
	lt, found := TokenLiterals[token]
	if !found {
		return ""
	}

	return lt
}
