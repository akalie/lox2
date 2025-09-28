package inner

type TokenType int

const (
	TOKEN_NOTHING     TokenType = iota // then go creates empty token, it looks like this
	TOKEN_LEFT_PAREN  TokenType = iota
	TOKEN_RIGHT_PAREN TokenType = iota

	TOKEN_LEFT_BRACE  TokenType = iota
	TOKEN_RIGHT_BRACE TokenType = iota

	TOKEN_COMMA TokenType = iota
	TOKEN_DOT   TokenType = iota
	TOKEN_MINUS TokenType = iota
	TOKEN_PLUS  TokenType = iota

	TOKEN_SEMICOLON TokenType = iota
	TOKEN_SLASH     TokenType = iota
	TOKEN_STAR      TokenType = iota

	// One or two character tokens.
	TOKEN_BANG       TokenType = iota
	TOKEN_BANG_EQUAL TokenType = iota

	TOKEN_EQUAL       TokenType = iota
	TOKEN_EQUAL_EQUAL TokenType = iota

	TOKEN_GREATER       TokenType = iota
	TOKEN_GREATER_EQUAL TokenType = iota

	TOKEN_LESS       TokenType = iota
	TOKEN_LESS_EQUAL TokenType = iota

	// Literals.
	TOKEN_IDENTIFIER TokenType = iota
	TOKEN_STRING     TokenType = iota
	TOKEN_NUMBER     TokenType = iota

	// Keywords.
	TOKEN_AND   TokenType = iota
	TOKEN_CLASS TokenType = iota
	TOKEN_ELSE  TokenType = iota
	TOKEN_FALSE TokenType = iota

	TOKEN_FOR TokenType = iota
	TOKEN_FUN TokenType = iota
	TOKEN_IF  TokenType = iota
	TOKEN_NIL TokenType = iota
	TOKEN_OR  TokenType = iota

	TOKEN_PRINT  TokenType = iota
	TOKEN_RETURN TokenType = iota
	TOKEN_SUPER  TokenType = iota
	TOKEN_THIS   TokenType = iota

	TOKEN_TRUE  TokenType = iota
	TOKEN_VAR   TokenType = iota
	TOKEN_WHILE TokenType = iota

	TOKEN_ERROR TokenType = iota
	TOKEN_EOF   TokenType = iota
)

type Scanner struct {
	Start   int
	Current int
	Line    int
	Source  []byte
}

type Token struct {
	Type        TokenType
	Start       int
	Line        int
	Source      []byte
	CurrentChar int
}

func (t *Token) GetSource() []byte {
	return t.Source
}

func NewScanner(source string) *Scanner {
	charSource := []byte(source)

	return &Scanner{
		Start:   0,
		Current: 0,
		Line:    1,
		Source:  charSource,
	}
}

func (s *Scanner) scanToken() Token {
	s.skipWhiteSpaces()
	s.Start = s.Current
	if s.atEnd() {
		return s.makeToken(TOKEN_EOF)
	}
	c := s.advance()

	if s.isDigit(c) {
		return s.number()
	}

	if s.isAlpha(c) {
		return s.identifier()
	}

	switch c {
	case '(':
		return s.makeToken(TOKEN_LEFT_PAREN)
	case ')':
		return s.makeToken(TOKEN_RIGHT_PAREN)
	case '{':
		return s.makeToken(TOKEN_LEFT_BRACE)
	case '}':
		return s.makeToken(TOKEN_RIGHT_BRACE)
	case ';':
		return s.makeToken(TOKEN_SEMICOLON)
	case ',':
		return s.makeToken(TOKEN_COMMA)
	case '.':
		return s.makeToken(TOKEN_DOT)
	case '-':
		return s.makeToken(TOKEN_MINUS)
	case '+':
		return s.makeToken(TOKEN_PLUS)
	case '/':
		return s.makeToken(TOKEN_SLASH)
	case '*':
		return s.makeToken(TOKEN_STAR)
	case '!':
		return s.makeToken(
			to(s.match('='), TOKEN_BANG_EQUAL, TOKEN_BANG),
		)
	case '=':
		return s.makeToken(
			to(s.match('='), TOKEN_EQUAL_EQUAL, TOKEN_EQUAL),
		)
	case '<':
		return s.makeToken(
			to(s.match('='), TOKEN_LESS_EQUAL, TOKEN_LESS),
		)
	case '>':
		return s.makeToken(
			to(s.match('='), TOKEN_GREATER_EQUAL, TOKEN_GREATER),
		)
	case '"':
		return s.string()

	}

	return s.errorToken("Unexpected character.")
}

func (s *Scanner) errorToken(msg string) Token {
	return Token{
		Type:        TOKEN_ERROR,
		Start:       0,
		Line:        s.Line,
		Source:      []byte(msg),
		CurrentChar: s.Current,
	}
}

func (s *Scanner) atEnd() bool {
	return s.Current == len(s.Source)
}

func (s *Scanner) makeToken(tokenType TokenType) Token {
	return Token{
		Type:        tokenType,
		Start:       0,
		Line:        s.Line,
		Source:      s.Source[s.Start:s.Current],
		CurrentChar: s.Current,
	}
}

func (s *Scanner) advance() rune {
	s.Current++
	return rune(s.Source[s.Current-1])
}

func (s *Scanner) match(expected rune) bool {
	if s.atEnd() {
		return false
	}
	if rune(s.Source[s.Current]) != expected {
		return false
	}
	s.Current++

	return true
}

func (s *Scanner) skipWhiteSpaces() {
	for s.Current < len(s.Source) {
		c := s.peek()
		switch c {
		case '\n':
			s.Line++
			s.advance()
		case ' ', '\r', '\t':
			s.advance()
		case '/':
			if s.peekNext() == '/' {
				for s.peek() != '\n' && !s.atEnd() {
					s.advance()
				}
			} else {
				return
			}
		default:
			return
		}
	}
}

func (s *Scanner) peek() rune {
	return rune(s.Source[s.Current])
}

func (s *Scanner) peekNext() rune {
	if s.atEnd() {
		return ' '
	}
	return rune(s.Source[s.Current+1])
}

func (s *Scanner) string() Token {
	for s.peek() != '"' && !s.atEnd() {
		if s.peek() == '\n' {
			s.Line++
		}
		s.advance()
	}

	if s.atEnd() {
		return s.errorToken("Unterminated string.")
	}

	// The closing quote.
	s.advance()
	return s.makeToken(TOKEN_STRING)
}

func (s *Scanner) isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

func (s *Scanner) number() Token {
	for s.isDigit(s.peek()) {
		s.advance()
	}

	// Look for a fractional part.
	if s.peek() == '.' && s.isDigit(s.peekNext()) {
		// Consume the ".".
		s.advance()
		for s.isDigit(s.peek()) {
			s.advance()
		}
	}

	return s.makeToken(TOKEN_NUMBER)
}

func (s *Scanner) isAlpha(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		c == '_'
}

func (s *Scanner) identifier() Token {
	for s.isAlpha(s.peek()) || s.isDigit(s.peek()) {
		s.advance()
	}
	return s.makeToken(s.identifierType())
}

func (s *Scanner) identifierType() TokenType {
	switch s.Source[s.Start] {
	case 'a':
		return s.checkKeyword(1, 2, "nd", TOKEN_AND)
	case 'c':
		return s.checkKeyword(1, 4, "lass", TOKEN_CLASS)
	case 'e':
		return s.checkKeyword(1, 3, "lse", TOKEN_ELSE)
	case 'i':
		return s.checkKeyword(1, 1, "f", TOKEN_IF)
	case 'n':
		return s.checkKeyword(1, 2, "il", TOKEN_NIL)
	case 'o':
		return s.checkKeyword(1, 1, "r", TOKEN_OR)
	case 'p':
		return s.checkKeyword(1, 4, "rint", TOKEN_PRINT)
	case 'r':
		return s.checkKeyword(1, 5, "eturn", TOKEN_RETURN)
	case 's':
		return s.checkKeyword(1, 4, "uper", TOKEN_SUPER)
	case 'v':
		return s.checkKeyword(1, 2, "ar", TOKEN_VAR)
	case 'w':
		return s.checkKeyword(1, 4, "hile", TOKEN_WHILE)
	case 'f':
		if s.Current-s.Start > 1 {
			switch s.Source[s.Start+1] {
			case 'a':
				return s.checkKeyword(2, 3, "lse", TOKEN_FALSE)
			case 'o':
				return s.checkKeyword(2, 1, "r", TOKEN_FOR)
			case 'u':
				return s.checkKeyword(2, 1, "n", TOKEN_FUN)
			}
		}
	case 't':
		if s.Current-s.Start > 1 {
			switch s.Source[s.Start+1] {
			case 'h':
				return s.checkKeyword(2, 2, "is", TOKEN_THIS)
			case 'r':
				return s.checkKeyword(2, 2, "ue", TOKEN_TRUE)
			}
		}
	}

	return TOKEN_IDENTIFIER
}

func (s *Scanner) checkKeyword(start int, length int, rest string, ttype TokenType) TokenType {
	actualRest := string(s.Source[s.Start+start : s.Start+start+length])
	if s.Current-s.Start == start+length && rest == actualRest {
		return ttype
	}

	return TOKEN_IDENTIFIER
}

func to(iif bool, then TokenType, eelse TokenType) TokenType {
	if iif {
		return then
	}

	return eelse
}
