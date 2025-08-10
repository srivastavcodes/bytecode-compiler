package lexer

import (
	"Flint-v2/token"
)

type Lexer struct {
	input        string
	position     int // current position in input (points to current char)
	readPosition int // current reading position in input (after reading char)
	char         byte
}

func NewLexer(input string) *Lexer {
	lex := &Lexer{input: input}
	lex.readChar()
	return lex
}

func (lex *Lexer) readChar() {
	if lex.readPosition >= len(lex.input) {
		lex.char = 0
	} else {
		lex.char = lex.input[lex.readPosition]
	}
	lex.position = lex.readPosition
	lex.readPosition += 1
}

func (lex *Lexer) peekChar() byte {
	if lex.readPosition >= len(lex.input) {
		return 0
	} else {
		return lex.input[lex.readPosition]
	}
}

func (lex *Lexer) NextToken() token.Token {
	var tokn token.Token
	lex.skipWhiteSpace()

	switch lex.char {
	case '=':
		tokn = lex.readTwoCharToken('=', token.EQ, token.ASSIGN)
	case '+':
		tokn = newToken(token.PLUS, lex.char)
	case '-':
		tokn = newToken(token.MINUS, lex.char)
	case '!':
		tokn = lex.readTwoCharToken('=', token.NOT_EQ, token.BANG)
	case '/':
		tokn = newToken(token.SLASH, lex.char)
	case '*':
		tokn = newToken(token.ASTERISK, lex.char)
	case '<':
		tokn = newToken(token.LT, lex.char)
	case '>':
		tokn = newToken(token.GT, lex.char)
	case ';':
		tokn = newToken(token.SEMICOLON, lex.char)
	case ',':
		tokn = newToken(token.COMMA, lex.char)
	case ':':
		tokn = newToken(token.COLON, lex.char)
	case '(':
		tokn = newToken(token.L_PAREN, lex.char)
	case ')':
		tokn = newToken(token.R_PAREN, lex.char)
	case '{':
		tokn = newToken(token.L_BRACE, lex.char)
	case '}':
		tokn = newToken(token.R_BRACE, lex.char)
	case '"':
		tokn.Type = token.STRING
		tokn.Literal = lex.readString()
	case '[':
		tokn = newToken(token.L_BRACKET, lex.char)
	case ']':
		tokn = newToken(token.R_BRACKET, lex.char)
	case 0:
		tokn.Literal = ""
		tokn.Type = token.EOF
	default:
		return lex.readDefaultToken()
	}
	lex.readChar()
	return tokn
}

func (lex *Lexer) skipWhiteSpace() {
	for lex.char == ' ' || lex.char == '\t' || lex.char == '\n' || lex.char == '\r' {
		lex.readChar()
	}
}

func (lex *Lexer) readTwoCharToken(expectedChar byte, twoCharType,
	singleCharType token.TokenType) token.Token {

	if lex.peekChar() == expectedChar {
		char := lex.char
		lex.readChar()

		return token.Token{Type: twoCharType, Literal: string(char) + string(lex.char)}
	}
	return newToken(singleCharType, lex.char)
}

func (lex *Lexer) readString() string {
	position := lex.position + 1
	for {
		lex.readChar()
		if lex.char == '"' || lex.char == 0 {
			break
		}
	}
	return lex.input[position:lex.position]
}

func (lex *Lexer) readDefaultToken() token.Token {
	var tokn token.Token

	if isLetter(lex.char) {
		tokn.Literal = lex.readIdentifier()
		tokn.Type = token.LookupIdent(tokn.Literal)
		return tokn
	}
	if isDigit(lex.char) {
		tokn.Type = token.INT
		tokn.Literal = lex.readNumber()
		return tokn
	}
	tokn = newToken(token.ILLEGAL, lex.char)
	lex.readChar()
	return tokn
}

func (lex *Lexer) readIdentifier() string {
	position := lex.position
	for isLetter(lex.char) {
		lex.readChar()
	}
	return lex.input[position:lex.position]
}

func (lex *Lexer) readNumber() string {
	position := lex.position
	for isDigit(lex.char) {
		lex.readChar()
	}
	return lex.input[position:lex.position]
}

func isLetter(char byte) bool {
	return 'a' <= char && char <= 'z' || 'A' <= char && char <= 'Z' || char == '_'
}

func isDigit(char byte) bool {
	return '0' <= char && char <= '9'
}

func newToken(tokenType token.TokenType, char byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(char)}
}
