package main

import (
	"fmt"
	"strconv"
)

type scanner struct {
	source  string
	start   int
	current int
	line    int
	tokens  []token
}

var keywords = map[string]tokenType{
	"and":    And,
	"class":  Class,
	"else":   Else,
	"false":  False,
	"for":    For,
	"fun":    Fun,
	"if":     If,
	"nil":    Nil,
	"or":     Or,
	"print":  Print,
	"return": Return,
	"super":  Super,
	"this":   This,
	"true":   True,
	"var":    Var,
	"while":  While,
}

func (s *scanner) scanToken() {
	c := s.advance()

	switch c {
	case '(':
		s.addToken(LeftParen)
	case ')':
		s.addToken(RightParen)
	case '{':
		s.addToken(LeftBrace)
	case '}':
		s.addToken(RightBrace)
	case ',':
		s.addToken(Comma)
	case '.':
		s.addToken(Dot)
	case '-':
		s.addToken(Minus)
	case '+':
		s.addToken(Plus)
	case ';':
		s.addToken(Semicolon)
	case '*':
		s.addToken(Star)
	case '!':
		if s.match('=') {
			s.addToken(BangEqual)
		} else {
			s.addToken(Bang)
		}
	case '=':
		if s.match('=') {
			s.addToken(EqualEqual)
		} else {
			s.addToken(Equal)
		}
	case '<':
		if s.match('=') {
			s.addToken(LessEqual)
		} else {
			s.addToken(Less)
		}
	case '>':
		if s.match('=') {
			s.addToken(GreaterEqual)
		} else {
			s.addToken(Greater)
		}
	case '/':
		if s.match('/') {
			// a comment goes until the end of the line
			for s.peek() != '\n' && !s.isAtEnd() {
				s.advance()
			}
		} else {
			s.addToken(Slash)
		}
	case ' ', '\r', '\t':
		// ignore whitespace
	case '\n':
		s.line++
	case '"':
		s.scanString()

	default:
		if s.isDigit(c) {
			s.scanNumber()
		} else if s.isAlpha(c) {
			s.scanIdentifier()

		} else {
			reportLineError(s.line, fmt.Sprintf("unexpected character %q", c))
		}

	}
}

func (s *scanner) scanTokens() {
	for !s.isAtEnd() {
		s.start = s.current
		s.scanToken()
	}

	s.tokens = append(s.tokens, token{
		Eof, "", nil, s.line})

}

func (s *scanner) addToken(typ tokenType) {
	s.addTokenLiteral(typ, nil)
}

func (s *scanner) addTokenLiteral(typ tokenType, literal any) {
	text := s.source[s.start:s.current]
	s.tokens = append(s.tokens, token{
		typ:     typ,
		lexeme:  text,
		literal: literal,
		line:    s.line,
	})

}

func (s *scanner) advance() byte { // consumes the current character and advances
	res := s.source[s.current]
	s.current++
	return res
}

func (s *scanner) match(expected byte) bool { // checks for and consumes second character in cases of two-character lexemes
	if s.isAtEnd() {
		return false
	}
	if s.source[s.current] != expected {
		return false
	}

	s.current++
	return true
}
func (s *scanner) peek() byte { // looks ahead at the unconsumed character
	if s.isAtEnd() {
		return 0
	}
	return s.source[s.current]
}

func (s *scanner) peekNext() byte { // looking past decimal point when parsing decimals require 2nd char of lookahead
	if s.current+1 >= len(s.source) {
		return 0
	}
	return s.source[s.current+1]
}

func (s *scanner) scanString() {
	for s.peek() != '"' && !s.isAtEnd() {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}

	if s.isAtEnd() {
		reportLineError(s.line, "unterminated string")
		return
	}

	s.advance() // the closing "

	// trimming surrounding quotes
	value := s.source[s.start+1 : s.current-1]
	s.addTokenLiteral(String, value)
}

func (s *scanner) scanNumber() {
	for s.isDigit(s.peek()) {
		s.advance()
	}

	// look for fractional part
	if s.peek() == '.' && s.isDigit(s.peekNext()) {
		// consume the .
		s.advance()
	}

	for s.isDigit(s.peek()) {
		s.advance()
	}

	num, err := strconv.ParseFloat(s.source[s.start:s.current], 64)

	if err != nil {
		reportLineError(s.line, "number failed to parse")
		return
	}

	s.addTokenLiteral(Number, num)
}

func (s *scanner) scanIdentifier() {
	for s.isAlphaNumeric(s.peek()) {
		s.advance()
	}

	text := s.source[s.start:s.current]

	typ, ok := keywords[text]

	if !ok {
		typ = Identifier
	}
	s.addToken(typ)
}

func (s *scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func (s *scanner) isAlphaNumeric(c byte) bool {
	return s.isAlpha(c) || s.isDigit(c)
}

func (s *scanner) isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func (s *scanner) isAlpha(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_'
}
