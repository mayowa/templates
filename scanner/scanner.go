package scanner

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

type Scanner struct {
	r        *bufio.Reader
	position int
	line     int
	lastCh   rune
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

func (s *Scanner) read() rune {
	var err error
	s.lastCh, _, err = s.r.ReadRune()
	if err != nil {
		return eof
	}
	if s.lastCh == '\n' {
		s.line++
	}

	s.position++
	return s.lastCh
}

func (s *Scanner) unread() {
	_ = s.r.UnreadRune()
	s.position--
	if s.lastCh == '\n' {
		s.line--
	}
}
func (s *Scanner) newTokenItem(token Token, literal string) *TokenItem {
	ti := &TokenItem{
		Token:         token,
		Literal:       literal,
		StartPosition: s.position - len(literal),
		EndPosition:   s.position,
		Line:          s.line,
	}

	return ti
}

//
// func (s *Scanner) ParseTagHead() (*templates.Tag, error) {
// 	var (
// 		tkItem *TokenItem
// 		err    error
// 	)
// 	// look for tag start
// 	for {
// 		tkItem = s.Scan()
// 		if tkItem.Token == EOF {
// 			break
// 		} else if tkItem.Token == TagStart {
// 			s.unread()
// 			break
// 		}
// 	}
//
// 	tag := new(templates.Tag)
// 	if tkItem.Token == TagStart {
// 		return nil, nil
// 	}
//
// 	// next token must be an Identifier
// 	tkItem = s.Scan()
// 	if tkItem.Token != Identifier {
// 		return nil, nil
// 	}
// 	tag.Name = tkItem.Literal
//
// 	// parse args if any
// 	tag.Args, err = s.parseArgs(RightAngleBracket)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return tag, nil
// }

func (s *Scanner) ParseTagArgs() (map[string]string, error) {

	args, err := s.parseArgs(EOF)
	if err != nil {
		return nil, err
	}

	return args, nil
}

func (s *Scanner) parseArgs(until Token) (map[string]string, error) {
	// Identifier [WS] = "|'[WS, Word, Other] "|'
	// name="mayo" class="{{ if eq .Name "mayo" }}123{{end}}"
	// Identifier, Assign, DoubleQuote, Identifier, DoubleQuote, WhiteSpace

	var (
		wrkItems []*TokenItem
		item     *TokenItem
		args     = map[string]string{}
		err      error
	)
	wrkItems, lastTokenItem := s.ScanUntil(until, false)
	if lastTokenItem.Token != until {
		return nil, fmt.Errorf("closing token %s not found", until)
	}

	for len(wrkItems) > 0 {
		item, wrkItems = trimWhiteSpace(wrkItems)
		if item.Token == until || item.Token == EOF {
			break
		}

		// Identifier
		if item.Token != Identifier {
			continue
		}
		name := item.Literal

		// Assign
		item, wrkItems = trimWhiteSpace(wrkItems)
		if item.Token != Assign {
			return nil, fmt.Errorf("expected '=', found %q", item.Literal)
		}

		// String
		var strItems []*TokenItem
		strItems, wrkItems, err = extractArgVal(wrkItems)
		if err != nil {
			return nil, err
		}

		args[name] = concatItems(strItems)
	}

	return args, nil
}

func countToken(token Token, tokens []*TokenItem) (count int) {
	for _, ti := range tokens {
		if ti.Token == token {
			count++
		}
	}

	return count
}

func concatItems(items []*TokenItem) string {
	str := ""
	for _, i := range items {
		str += i.Literal
	}

	return str
}

func nextArg(tokens []*TokenItem) int {
	tCount := len(tokens)
	for pos, tkItem := range tokens {
		if pos+1 < tCount && tkItem.Token == Identifier && tokens[pos+1].Token == Assign {
			return pos
		}
	}

	return -1
}

func extractArgVal(tokens []*TokenItem) ([]*TokenItem, []*TokenItem, error) {
	// Single or Double Quote
	item, wrkItems := trimWhiteSpace(tokens)
	if item.Token != SingleQuote && item.Token != DoubleQuote {
		return nil, nil, fmt.Errorf("expected %s or %s, found %s", SingleQuote, DoubleQuote, item.Token)
	}
	stringStart := item.Token

	naPos := nextArg(wrkItems)
	stringEnd := -1
	if naPos == -1 {
		// no arg found set to last token
		naPos = len(wrkItems) - 1
	}

	// find position of the last quote string
	for i := naPos - 1; i > 0; i-- {
		if wrkItems[i].Token == stringStart {
			stringEnd = i
			break
		}
	}

	// if the quote found is the one at the beginning, we have an unterminated string
	if stringEnd <= 0 || (countToken(stringStart, wrkItems[:stringEnd])%2) != 0 {
		return nil, nil, fmt.Errorf("unterminated string on line:%d", item.Line)
	}

	if stringEnd+1 >= len(wrkItems) {
		return wrkItems[:stringEnd], wrkItems[:], nil
	}

	return wrkItems[:stringEnd], wrkItems[stringEnd+1:], nil
}

func trimWhiteSpace(tokens []*TokenItem) (*TokenItem, []*TokenItem) {
	itm, tokens := pop(tokens)
	if itm.Token == WhiteSpace {
		itm, tokens = pop(tokens)
	}

	return itm, tokens
}

func pop[T any](stack []T) (T, []T) {
	top := stack[0]
	stack = stack[1:]
	return top, stack
}

func (s *Scanner) ScanUntil(token Token, withPeek bool) (items []*TokenItem, lastItem *TokenItem) {
	var (
		oldPosition int
	)

	if withPeek {
		oldPosition = s.position
	}

	for {
		lastItem = s.Scan()
		if lastItem.Token == EOF {
			break
		} else if lastItem.Token == token {
			break
		}
		items = append(items, lastItem)
	}

	items = append(items, lastItem)

	if withPeek {
		for i := s.position; i > oldPosition; i-- {
			s.unread()
		}
	}

	return items, lastItem
}

func (s *Scanner) Scan() *TokenItem {
	// Read the next rune.
	ch := s.read()

	// If we see whitespace then consume all contiguous whitespace.
	// If we see a letter then consume as an ident or reserved word.
	if isWhitespace(ch) {
		s.unread()
		return s.scanWhitespace()
	} else if isLetter(ch) {
		s.unread()
		return s.scanIdentifier()
	}

	// Otherwise read the individual character.
	switch ch {
	case eof:
		return s.newTokenItem(EOF, "")
	case '<':
		nextCh := s.read()
		if isUpperCaseLetter(nextCh) {
			s.unread()
			// look for TagStart "<X"
			return s.newTokenItem(TagStart, string(ch))
		} else if nextCh == '/' {
			// look for start ClosingTag "</X"
			nextCh = s.read()
			s.unread()
			if isUpperCaseLetter(nextCh) {
				return s.newTokenItem(ClosingTagStart, "</")
			}
		}

		if nextCh != eof {
			s.unread()
		}
		return s.newTokenItem(LeftAngleBracket, string(ch))

	case '/':
		nextCh := s.read()
		if nextCh == '>' {
			return s.newTokenItem(TagSelfClosing, "/>")
		}
		if nextCh != eof {
			s.unread()
		}
		return s.newTokenItem(BackSlash, "/")
	case '>':
		return s.newTokenItem(RightAngleBracket, string(ch))
	case '=':
		return s.newTokenItem(Assign, string(ch))
	case '\'':
		return s.newTokenItem(SingleQuote, string(ch))
	case '"':
		return s.newTokenItem(DoubleQuote, string(ch))
	case '\\':
		nextCh := s.read()
		if nextCh == '\'' {
			return s.newTokenItem(EscSingleQuote, string(`\'`))
		} else if nextCh == '"' {
			return s.newTokenItem(EscDoubleQuote, string(`\"`))
		}
		return s.newTokenItem(Slash, string(ch))
	case '`':
		return s.newTokenItem(TripleQuote, string(ch))
	}

	return s.newTokenItem(Other, string(ch))
}

func (s *Scanner) scanWhitespace() *TokenItem {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent whitespace character into the buffer.
	// Non-whitespace characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return s.newTokenItem(WhiteSpace, buf.String())
}

func (s *Scanner) scanIdentifier() *TokenItem {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !isLetter(ch) && !isDigit(ch) && ch != '-' {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	return s.newTokenItem(Identifier, buf.String())
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '-'
}

func isUpperCaseLetter(ch rune) bool {
	return 'A' <= ch && ch <= 'Z'
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n'
}
