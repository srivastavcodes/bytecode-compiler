package parser

import (
	"fmt"
	"strconv"

	"comp/ast"
	"comp/lexer"
	"comp/token"
)

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -x or !x
	CALL        // myFunc(x)
	INDEX       // array[index]
)

var precedences = map[token.TokenType]int{
	token.EQ:        EQUALS,
	token.NOT_EQ:    EQUALS,
	token.LT:        LESSGREATER,
	token.GT:        LESSGREATER,
	token.PLUS:      SUM,
	token.MINUS:     SUM,
	token.SLASH:     PRODUCT,
	token.ASTERISK:  PRODUCT,
	token.L_PAREN:   CALL,
	token.L_BRACKET: INDEX,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	lxr    *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func NewParser(lxr *lexer.Lexer) *Parser {
	psr := &Parser{lxr: lxr, errors: []string{}}

	// Read two tokens, so that curToken and peekToken are set
	psr.nextToken()
	psr.nextToken()

	registerParseFunctions(psr)

	return psr
}

func (psr *Parser) ParseRootStatement() *ast.RootStatement {
	root := &ast.RootStatement{}
	root.Statements = []ast.Statement{}

	for !psr.currentTokenIs(token.EOF) {
		stmt := psr.parseStatement()
		if stmt != nil {
			root.Statements = append(root.Statements, stmt)
		}
		psr.nextToken()
	}
	return root
}

func (psr *Parser) parseStatement() ast.Statement {
	switch psr.curToken.Type {
	case token.LET:
		return psr.parseLetStatement()
	case token.RETURN:
		return psr.parseReturnStatement()
	default:
		return psr.parseExpressionStatement()
	}
}

func (psr *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: psr.curToken}
	if !psr.expectPeek(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: psr.curToken, Value: psr.curToken.Literal}

	if !psr.expectPeek(token.ASSIGN) {
		return nil
	}
	psr.nextToken()
	stmt.Value = psr.parseExpression(LOWEST)

	if psr.peekTokenIs(token.SEMICOLON) {
		psr.nextToken()
	}
	return stmt
}

func (psr *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: psr.curToken}
	psr.nextToken()
	stmt.ReturnValue = psr.parseExpression(LOWEST)

	if psr.peekTokenIs(token.SEMICOLON) {
		psr.nextToken()
	}
	return stmt
}

func (psr *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: psr.curToken}
	stmt.Expression = psr.parseExpression(LOWEST)

	if psr.peekTokenIs(token.SEMICOLON) {
		psr.nextToken()
	}
	return stmt
}

func (psr *Parser) parseExpression(precedence int) ast.Expression {
	prefix := psr.prefixParseFns[psr.curToken.Type]
	if nil == prefix {
		psr.noPrefixParseFnError(psr.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !psr.peekTokenIs(token.SEMICOLON) && precedence < psr.peekPrecedence() {
		infix := psr.infixParseFns[psr.peekToken.Type]
		if nil == infix {
			return leftExp
		}
		psr.nextToken()
		leftExp = infix(leftExp)
	}
	return leftExp
}

func (psr *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: psr.curToken, Value: psr.curToken.Literal}
}

func (psr *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: psr.curToken, Value: psr.currentTokenIs(token.TRUE)}
}

func (psr *Parser) parseGroupedExpression() ast.Expression {
	psr.nextToken()
	expr := psr.parseExpression(LOWEST)

	if !psr.expectPeek(token.R_PAREN) {
		return nil
	}
	return expr
}

func (psr *Parser) parseArrayLiteral() ast.Expression {
	al := &ast.ArrayLiteral{Token: psr.curToken}
	al.Elements = psr.parseExpressionList(token.R_BRACKET)
	return al
}

func (psr *Parser) parseExpressionList(rb token.TokenType) []ast.Expression {
	var list []ast.Expression

	if psr.peekTokenIs(rb) {
		psr.nextToken()
		return list
	}
	psr.nextToken()
	list = append(list, psr.parseExpression(LOWEST))

	for psr.peekTokenIs(token.COMMA) {
		psr.nextToken()
		psr.nextToken()
		list = append(list, psr.parseExpression(LOWEST))
	}
	if !psr.expectPeek(rb) {
		return nil
	}
	return list
}

func (psr *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: psr.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	for !psr.peekTokenIs(token.R_BRACE) {
		psr.nextToken()

		key := psr.parseExpression(LOWEST)
		if !psr.expectPeek(token.COLON) {
			return nil
		}

		psr.nextToken()
		value := psr.parseExpression(LOWEST)

		hash.Pairs[key] = value
		if !psr.peekTokenIs(token.R_BRACE) && !psr.expectPeek(token.COMMA) {
			return nil
		}
	}
	if !psr.expectPeek(token.R_BRACE) {
		return nil
	}
	return hash
}

func (psr *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: psr.curToken}

	value, err := strconv.ParseInt(psr.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", psr.curToken.Literal)
		psr.errors = append(psr.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

func (psr *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: psr.curToken, Value: psr.curToken.Literal}
}

func (psr *Parser) parsePrefixExpression() ast.Expression {
	expr := &ast.PrefixExpression{
		Token:    psr.curToken,
		Operator: psr.curToken.Literal,
	}
	psr.nextToken()
	expr.Right = psr.parseExpression(PREFIX)
	return expr
}

func (psr *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expr := &ast.InfixExpression{
		Token:    psr.curToken,
		Operator: psr.curToken.Literal,
		Left:     left,
	}
	precedence := psr.curPrecedence()
	psr.nextToken()
	expr.Right = psr.parseExpression(precedence)
	return expr
}

func (psr *Parser) parseIfExpression() ast.Expression {
	expr := &ast.IfExpression{Token: psr.curToken}
	if !psr.expectPeek(token.L_PAREN) {
		return nil
	}
	psr.nextToken()
	expr.Condition = psr.parseExpression(LOWEST)
	if !psr.expectPeek(token.R_PAREN) {
		return nil
	}
	if !psr.expectPeek(token.L_BRACE) {
		return nil
	}
	expr.Consequence = psr.parseBlockStatement()

	if psr.peekTokenIs(token.ELSE) {
		psr.nextToken()

		if !psr.expectPeek(token.L_BRACE) {
			return nil
		}
		expr.Alternative = psr.parseBlockStatement()
	}
	return expr
}

func (psr *Parser) parseFunctionLiteral() ast.Expression {
	fnLit := &ast.FunctionLiteral{Token: psr.curToken}

	if !psr.expectPeek(token.L_PAREN) {
		return nil
	}
	fnLit.Parameters = psr.parseFunctionParameters()
	if !psr.expectPeek(token.L_BRACE) {
		return nil
	}
	fnLit.Body = psr.parseBlockStatement()
	return fnLit
}

func (psr *Parser) parseFunctionParameters() []*ast.Identifier {
	var identifiers []*ast.Identifier

	if psr.peekTokenIs(token.R_PAREN) {
		psr.nextToken()
		return identifiers
	}
	psr.nextToken()
	ident := &ast.Identifier{Token: psr.curToken, Value: psr.curToken.Literal}
	identifiers = append(identifiers, ident)

	for psr.peekTokenIs(token.COMMA) {
		psr.nextToken()
		psr.nextToken()
		ident := &ast.Identifier{Token: psr.curToken, Value: psr.curToken.Literal}
		identifiers = append(identifiers, ident)
	}
	if !psr.expectPeek(token.R_PAREN) {
		return nil
	}
	return identifiers
}

func (psr *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: psr.curToken}
	block.Statements = []ast.Statement{}

	psr.nextToken()

	for !psr.currentTokenIs(token.R_BRACE) && !psr.currentTokenIs(token.EOF) {
		stmt := psr.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		psr.nextToken()
	}
	return block
}

func (psr *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	expr := &ast.CallExpression{Token: psr.curToken, Function: function}
	expr.Arguments = psr.parseExpressionList(token.R_PAREN)
	return expr
}

func (psr *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	expr := &ast.IndexExpression{Token: psr.curToken, Left: left}

	psr.nextToken()
	expr.Index = psr.parseExpression(LOWEST)

	if !psr.expectPeek(token.R_BRACKET) {
		return nil
	}
	return expr
}

func (psr *Parser) Errors() []string {
	return psr.errors
}

func (psr *Parser) peekError(tokn token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		tokn, psr.peekToken.Type)
	psr.errors = append(psr.errors, msg)
}

func (psr *Parser) noPrefixParseFnError(tokn token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", tokn)
	psr.errors = append(psr.errors, msg)
}

func (psr *Parser) nextToken() {
	psr.curToken = psr.peekToken
	psr.peekToken = psr.lxr.NextToken()
}

func (psr *Parser) currentTokenIs(tokn token.TokenType) bool {
	return psr.curToken.Type == tokn
}

func (psr *Parser) peekTokenIs(tokn token.TokenType) bool {
	return psr.peekToken.Type == tokn
}

func (psr *Parser) expectPeek(tokn token.TokenType) bool {
	if psr.peekTokenIs(tokn) {
		psr.nextToken()
		return true
	} else {
		psr.peekError(tokn)
		return false
	}
}

func (psr *Parser) peekPrecedence() int {
	if pdc, ok := precedences[psr.peekToken.Type]; ok {
		return pdc
	}
	return LOWEST
}

func (psr *Parser) curPrecedence() int {
	if pdc, ok := precedences[psr.curToken.Type]; ok {
		return pdc
	}
	return LOWEST
}

func (psr *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	psr.prefixParseFns[tokenType] = fn
}

func (psr *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	psr.infixParseFns[tokenType] = fn
}

// make this a method
func registerParseFunctions(psr *Parser) {
	registerPrefixParseFunctions(psr)
	registerInfixParseFunctions(psr)
}

func registerPrefixParseFunctions(psr *Parser) {
	psr.prefixParseFns = make(map[token.TokenType]prefixParseFn)

	psr.registerPrefix(token.IDENT, psr.parseIdentifier)

	psr.registerPrefix(token.STRING, psr.parseStringLiteral)
	psr.registerPrefix(token.INT, psr.parseIntegerLiteral)

	psr.registerPrefix(token.BANG, psr.parsePrefixExpression)
	psr.registerPrefix(token.MINUS, psr.parsePrefixExpression)

	psr.registerPrefix(token.TRUE, psr.parseBoolean)
	psr.registerPrefix(token.FALSE, psr.parseBoolean)

	psr.registerPrefix(token.L_PAREN, psr.parseGroupedExpression)
	psr.registerPrefix(token.L_BRACE, psr.parseHashLiteral)
	psr.registerPrefix(token.L_BRACKET, psr.parseArrayLiteral)

	psr.registerPrefix(token.IF, psr.parseIfExpression)
	psr.registerPrefix(token.FUNCTION, psr.parseFunctionLiteral)
}

func registerInfixParseFunctions(psr *Parser) {
	psr.infixParseFns = make(map[token.TokenType]infixParseFn)

	psr.registerInfix(token.PLUS, psr.parseInfixExpression)
	psr.registerInfix(token.MINUS, psr.parseInfixExpression)
	psr.registerInfix(token.SLASH, psr.parseInfixExpression)
	psr.registerInfix(token.ASTERISK, psr.parseInfixExpression)

	psr.registerInfix(token.EQ, psr.parseInfixExpression)
	psr.registerInfix(token.NOT_EQ, psr.parseInfixExpression)

	psr.registerInfix(token.LT, psr.parseInfixExpression)
	psr.registerInfix(token.GT, psr.parseInfixExpression)

	psr.registerInfix(token.L_PAREN, psr.parseCallExpression)
	psr.registerInfix(token.L_BRACKET, psr.parseIndexExpression)
}
