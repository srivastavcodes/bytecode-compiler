package parser

import (
	"fmt"
	"testing"

	"comp/ast"
	"comp/lexer"
)

func TestLetStatement(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foobar = y;", "foobar", "y"},
	}
	for _, tt := range tests {
		lxr := lexer.NewLexer(tt.input)
		psr := NewParser(lxr)
		root := psr.ParseRootStatement()
		checkParserErrors(t, psr)

		if len(root.Statements) != 1 {
			t.Fatalf("root.Statements does not contain 1 statements. got=%d",
				len(root.Statements))
		}
		stmt := root.Statements[0]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
		val := stmt.(*ast.LetStatement).Value
		if !testLiteralExpression(t, val, tt.expectedValue) {
			return
		}
	}
}

func checkParserErrors(t *testing.T, psr *Parser) {
	t.Helper()
	errors := psr.Errors()
	if len(errors) == 0 {
		return
	}
	t.Errorf("parser has %d errors", len(errors))
	for _, err := range errors {
		t.Errorf("parser error %q", err)
	}
	t.FailNow()
}

func testLetStatement(t *testing.T, stmt ast.Statement, name string) bool {
	t.Helper()
	if stmt.TokenLiteral() != "let" {
		t.Errorf("stmt.TokenLiteral not 'let'. got=%q", stmt.TokenLiteral())
		return false
	}
	letStmt, ok := stmt.(*ast.LetStatement)
	if !ok {
		t.Errorf("stmt not *ast.LetStatement. got=%T", stmt)
		return false
	}
	if letStmt.Name.Value != name {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", name, letStmt.Name.Value)
		return false
	}
	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("stmt.Name not '%s'. got=%s", name, letStmt.Name)
		return false
	}
	return true
}

func TestReturnStatement(tst *testing.T) {
	input := `
return 5;
return 10;
return 992233;
`
	lxr := lexer.NewLexer(input)
	psr := NewParser(lxr)

	root := psr.ParseRootStatement()
	checkParserErrors(tst, psr)

	if len(root.Statements) != 3 {
		tst.Fatalf("root.Statements does not contain 3 statement. got=%d",
			len(root.Statements))
	}
	for _, stmt := range root.Statements {
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			tst.Errorf("stmt not *ast.ReturnStatement. got=%T", stmt)
			continue
		}
		if returnStmt.TokenLiteral() != "return" {
			tst.Errorf("returnStmt.TokenLiteral() not 'return'. got=%q",
				returnStmt.TokenLiteral())
		}
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := `foobar;`

	lxr := lexer.NewLexer(input)
	psr := NewParser(lxr)
	root := psr.ParseRootStatement()
	checkParserErrors(t, psr)

	if len(root.Statements) != 1 {
		t.Fatalf("root does not have 1 statement. got=%d", len(root.Statements))
	}
	stmt, ok := root.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("root.Statements[0] is not ast.ExpressionStatement. got=%T", stmt)
	}
	ident, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("Expression is not *ast.Identifier. got=%T", stmt.Expression)
	}
	if ident.Value != "foobar" {
		t.Errorf("ident.Value not '%s'. got=%s", "foobar", ident.Value)
	}
	if ident.TokenLiteral() != "foobar" {
		t.Errorf("ident.TokenLiteral not '%s'. got=%s", "foobar", ident.TokenLiteral())
	}
}

func TestBooleanExpression(t *testing.T) {
	input := `true;`

	lxr := lexer.NewLexer(input)
	psr := NewParser(lxr)
	root := psr.ParseRootStatement()
	checkParserErrors(t, psr)

	if len(root.Statements) != 1 {
		t.Fatalf("root does not have 1 statement. got=%d", len(root.Statements))
	}
	stmt, ok := root.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("root.Statements[0] is not ast.ExpressionStatement. got=%T", stmt)
	}
	ident, ok := stmt.Expression.(*ast.Boolean)
	if !ok {
		t.Fatalf("Expression is not *ast.Identifier. got=%T", ident)
	}
	if ident.Value != true {
		t.Errorf("ident.Value not '%s'. got=%t", "foobar", ident.Value)
	}
	if ident.TokenLiteral() != "true" {
		t.Errorf("ident.TokenLiteral not '%s'. got=%s", "foobar", ident.TokenLiteral())
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := `5;`

	lxr := lexer.NewLexer(input)
	psr := NewParser(lxr)
	root := psr.ParseRootStatement()
	checkParserErrors(t, psr)

	if len(root.Statements) != 1 {
		t.Fatalf("root does not have 1 length statement. got=%d", len(root.Statements))
	}
	stmt, ok := root.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not *ast.ExpressionStatement. got=%T", stmt)
	}
	literal, ok := stmt.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("Expression is not *ast.IntegerLiteral. got=%T", literal)
	}
	if literal.Value != 5 {
		t.Errorf("literal.Value not %d. got=%d", 5, literal.Value)
	}
	if literal.TokenLiteral() != "5" {
		t.Errorf("literal.TokenLiteral not '%s'. got=%s", "5", literal.TokenLiteral())
	}
}

func TestStringLiteralExpression(t *testing.T) {
	input := `"hello world";`

	lxr := lexer.NewLexer(input)
	psr := NewParser(lxr)
	root := psr.ParseRootStatement()
	checkParserErrors(t, psr)

	stmt := root.Statements[0].(*ast.ExpressionStatement)
	literal, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("exp not %T. got=%T", &ast.StringLiteral{}, stmt.Expression)
	}
	if literal.Value != "hello world" {
		t.Errorf("literal.Value not %q. got=%q", "hello world", literal.Value)
	}
}

func TestParsingPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input        string
		operator     string
		integerValue interface{}
	}{
		{"!5", "!", 5},
		{"-15", "-", 15},
		{"!true", "!", true},
		{"!false", "!", false},
	}
	for _, pt := range prefixTests {
		lxr := lexer.NewLexer(pt.input)
		psr := NewParser(lxr)
		root := psr.ParseRootStatement()
		checkParserErrors(t, psr)

		if len(root.Statements) != 1 {
			t.Fatalf("root.Statements does not contain %d statement. got=%d",
				1, len(root.Statements))
		}
		stmt, ok := root.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("RootStatement.Statements[0] is not ast.ExpressionStatement. got=%T", stmt)
		}
		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("stmt.Expression is not ast.PrefixExpression. got=%T", exp)
		}
		if exp.Operator != pt.operator {
			t.Fatalf("exp.Operator not '%s'. got=%s", pt.operator, exp.Operator)
		}
		if !testLiteralExpression(t, exp.Right, pt.integerValue) {
			return
		}
	}
}

func TestParsingInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"true == true", true, "==", true},
		{"true != false", true, "!=", false},
		{"false == false", false, "==", false},
	}
	for _, it := range infixTests {
		lxr := lexer.NewLexer(it.input)
		psr := NewParser(lxr)
		root := psr.ParseRootStatement()
		checkParserErrors(t, psr)

		if len(root.Statements) != 1 {
			t.Fatalf("root.Statements does not contain %d statement. got=%d",
				1, len(root.Statements))
		}
		stmt, ok := root.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("root.Statements[0] is not ast.ExpressionStatement. got=%T", stmt)
		}
		exp, ok := stmt.Expression.(*ast.InfixExpression)
		if !ok {
			t.Fatalf("stmt.Expression is not ast.InfixExpression. got=%T", exp)
		}
		if !testLiteralExpression(t, exp.Left, it.leftValue) {
			return
		}
		if !testLiteralExpression(t, exp.Right, it.rightValue) {
			return
		}
		if exp.Operator != it.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s", it.operator, exp.Operator)
		}
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a + b - c",
			"((a + b) - c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"true",
			"true",
		},
		{
			"false",
			"false",
		},
		{
			"3 > 5 == false",
			"((3 > 5) == false)",
		},
		{
			"3 < 5 == true",
			"((3 < 5) == true)",
		},
		{
			"1 + (2 + 3) + 4",
			"((1 + (2 + 3)) + 4)",
		},
		{
			"(5 + 5) * 2",
			"((5 + 5) * 2)",
		},
		{
			"2 / (5 + 5)",
			"(2 / (5 + 5))",
		},
		{
			"-(5 + 5)",
			"(-(5 + 5))",
		},
		{
			"!(true == true)",
			"(!(true == true))",
		},
		{
			"3 + 4; -5 * 5",
			"(3 + 4)((-5) * 5)",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))"},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"a + add(b * c) + d",
			"((a + add((b * c))) + d)",
		},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7* 8))",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))",
		},
		{
			"add(a + b + c * d / f + g)",
			"add((((a + b) + ((c * d) / f)) + g))",
		},
		{
			"a * [1, 2, 3, 4][b * c] * d",
			"((a * ([1, 2, 3, 4][(b * c)])) * d)",
		},
		{
			"add(a * b[2], b[1], 2 * [1, 2][1])",
			"add((a * (b[2])), (b[1]), (2 * ([1, 2][1])))",
		},
	}
	for _, tt := range tests {
		lxr := lexer.NewLexer(tt.input)
		psr := NewParser(lxr)
		root := psr.ParseRootStatement()
		checkParserErrors(t, psr)

		actual := root.String()
		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
}

func TestIfExpression(t *testing.T) {
	input := `if (x < y) { x }`

	lxr := lexer.NewLexer(input)
	psr := NewParser(lxr)
	root := psr.ParseRootStatement()
	checkParserErrors(t, psr)

	if len(root.Statements) != 1 {
		t.Fatalf("root.body does not contain %d statements. got=%d\n",
			1, len(root.Statements))
	}
	stmt, ok := root.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("root.Statements[0] is not *ast.ExpressionStatement. got=%T", stmt)
	}
	expr, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.IfExpression. got=%T", stmt.Expression)
	}
	if !testInfixExpression(t, expr.Condition, "x", "<", "y") {
		return
	}
	if len(expr.Consequence.Statements) != 1 {
		t.Errorf("consequence is not 1 statements. got=%d\n", len(expr.Consequence.Statements))
	}
	consequence, ok := expr.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not *ast.ExpressionStatement. got=%T", consequence)
	}
	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}
	if expr.Alternative != nil {
		t.Errorf("expr.Alternative.Statement was not nil. got=%+v", expr.Alternative)
	}
}

func TestIfElseExpression(t *testing.T) {
	input := `if (x < y) { x } else { y }`

	lxr := lexer.NewLexer(input)
	psr := NewParser(lxr)
	root := psr.ParseRootStatement()
	checkParserErrors(t, psr)

	if len(root.Statements) != 1 {
		t.Fatalf("root.Body does not contain %d statements. got=%d\n",
			1, len(root.Statements))
	}
	stmt, ok := root.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("root.Statements[0] is not %T. got=%T", &ast.ExpressionStatement{},
			root.Statements[0])
	}
	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not %T. got=%T", &ast.IfExpression{}, stmt.Expression)
	}
	if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
		return
	}
	if len(exp.Consequence.Statements) != 1 {
		t.Errorf("consequence is not 1 statements. got=%d\n",
			len(exp.Consequence.Statements))
	}
	consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not %T. got=%T", &ast.ExpressionStatement{},
			exp.Consequence.Statements[0])
	}
	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}
	if len(exp.Alternative.Statements) != 1 {
		t.Errorf("exp.Alternative.Statements does not contain 1 statements. got=%d\n",
			len(exp.Alternative.Statements))
	}
	alternative, ok := exp.Alternative.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not %T. got=%T", &ast.ExpressionStatement{},
			exp.Alternative.Statements[0])
	}
	if !testIdentifier(t, alternative.Expression, "y") {
		return
	}
}

func TestFunctionLiteralParsing(t *testing.T) {
	input := `func(x, y) { x + y; }`

	lxr := lexer.NewLexer(input)
	psr := NewParser(lxr)
	root := psr.ParseRootStatement()
	checkParserErrors(t, psr)

	if len(root.Statements) != 1 {
		t.Fatalf("root.Body does not contain %d statements. got=%d\n",
			1, len(root.Statements))
	}
	stmt, ok := root.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("root.Statements[0] is not *ast.ExpressionStatement. got=%T", stmt)
	}
	function, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.FunctionLiteral. got=%T", function)
	}
	if len(function.Parameters) != 2 {
		t.Errorf("function literal parameter wrong. expected 2, got=%d\n",
			len(function.Parameters))
	}
	testLiteralExpression(t, function.Parameters[0], "x")
	testLiteralExpression(t, function.Parameters[1], "y")

	if len(function.Body.Statements) != 1 {
		t.Errorf("function body does not contain 1 statements. got=%d\n",
			len(function.Body.Statements))
	}
	bodyStmt, ok := function.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("function body is not *ast.ExpressionStatement. got=%T", bodyStmt)
	}
	testInfixExpression(t, bodyStmt.Expression, "x", "+", "y")
}

func TestCallExpressionParsing(t *testing.T) {
	input := "add(1, 2 * 3, 4 + 5);"

	lxr := lexer.NewLexer(input)
	psr := NewParser(lxr)
	root := psr.ParseRootStatement()
	checkParserErrors(t, psr)

	if len(root.Statements) != 1 {
		t.Fatalf("root.Statements does not contain %d statements. got=%d\n",
			1, len(root.Statements))
	}
	stmt, ok := root.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt is not *ast.ExpressionStatement. got=%T", root.Statements[0])
	}
	expr, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not *ast.CallExpression. got=%T", stmt.Expression)
	}
	if !testIdentifier(t, expr.Function, "add") {
		return
	}
	if len(expr.Arguments) != 3 {
		t.Fatalf("wrong length of arguments. got=%d", len(expr.Arguments))
	}
	testLiteralExpression(t, expr.Arguments[0], 1)
	testInfixExpression(t, expr.Arguments[1], 2, "*", 3)
	testInfixExpression(t, expr.Arguments[2], 4, "+", 5)
}

func TestParsingArrayLiteral(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	lxr := lexer.NewLexer(input)
	psr := NewParser(lxr)
	root := psr.ParseRootStatement()
	checkParserErrors(t, psr)

	stmt, ok := root.Statements[0].(*ast.ExpressionStatement)
	arrs, ok := stmt.Expression.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("exp not %T. got=%T", &ast.ArrayLiteral{}, stmt.Expression)
	}
	if len(arrs.Elements) != 3 {
		t.Fatalf("len(arrs.Elements) not 3. got=%d", len(arrs.Elements))
	}
	testIntegerLiteral(t, arrs.Elements[0], 1)
	testInfixExpression(t, arrs.Elements[1], 2, "*", 2)
	testInfixExpression(t, arrs.Elements[2], 3, "+", 3)
}

func TestParsingIndexExpressions(t *testing.T) {
	input := "myArray[1 + 1]"

	lxr := lexer.NewLexer(input)
	psr := NewParser(lxr)
	root := psr.ParseRootStatement()
	checkParserErrors(t, psr)

	stmt, ok := root.Statements[0].(*ast.ExpressionStatement)
	indexExp, ok := stmt.Expression.(*ast.IndexExpression)
	if !ok {
		t.Fatalf("exp not %T. got=%T", &ast.IndexExpression{}, stmt.Expression)
	}
	if !testIdentifier(t, indexExp.Left, "myArray") {
		return
	}
	if !testInfixExpression(t, indexExp.Index, 1, "+", 1) {
		return
	}
}

func TestParsingHashLiteralsStringKeys(t *testing.T) {
	input := `{"one": 1, "two": 2, "three": 3}`

	lxr := lexer.NewLexer(input)
	psr := NewParser(lxr)
	root := psr.ParseRootStatement()
	checkParserErrors(t, psr)

	stmt := root.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	if !ok {
		t.Fatalf("exp is not %T. got=%T", ast.HashLiteral{}, stmt.Expression)
	}
	if len(hash.Pairs) != 3 {
		t.Errorf("hash.Pairs has wrong length. got=%d", len(hash.Pairs))
	}
	expected := map[string]int64{
		"one":   1,
		"two":   2,
		"three": 3,
	}
	for key, value := range hash.Pairs {
		literal, ok := key.(*ast.StringLiteral)
		if !ok {
			t.Errorf("key is not %T. got=%T", ast.StringLiteral{}, key)
		}
		expectedValue := expected[literal.String()]
		testIntegerLiteral(t, value, expectedValue)
	}
}

func TestParsingEmptyHashLiteral(t *testing.T) {
	input := `{}`

	lxr := lexer.NewLexer(input)
	psr := NewParser(lxr)
	root := psr.ParseRootStatement()
	checkParserErrors(t, psr)

	stmt := root.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	if !ok {
		t.Fatalf("exp is not %T. got=%T", ast.HashLiteral{}, stmt.Expression)
	}
	if len(hash.Pairs) != 0 {
		t.Errorf("hash.Pairs has wrong length. got=%d", len(hash.Pairs))
	}
}

func TestParsingHashLiteralsWithExpressions(t *testing.T) {
	input := `{"one": 0 + 1, "two": 10 - 8, "three": 15 / 5}`

	lxr := lexer.NewLexer(input)
	psr := NewParser(lxr)
	root := psr.ParseRootStatement()
	checkParserErrors(t, psr)

	stmt := root.Statements[0].(*ast.ExpressionStatement)
	hash, ok := stmt.Expression.(*ast.HashLiteral)
	if !ok {
		t.Fatalf("exp is not %T. got=%T", ast.HashLiteral{}, stmt.Expression)
	}
	if len(hash.Pairs) != 3 {
		t.Errorf("hash.Pairs has wrong length. got=%d", len(hash.Pairs))
	}
	tests := map[string]func(ast.Expression){
		"one": func(e ast.Expression) {
			testInfixExpression(t, e, 0, "+", 1)
		},
		"two": func(e ast.Expression) {
			testInfixExpression(t, e, 10, "-", 8)
		},
		"three": func(e ast.Expression) {
			testInfixExpression(t, e, 15, "/", 5)
		},
	}
	for key, value := range hash.Pairs {
		literal, ok := key.(*ast.StringLiteral)
		if !ok {
			t.Errorf("key is not %T. got=%T", ast.StringLiteral{}, key)
			continue
		}
		testFunc, ok := tests[literal.String()]
		if !ok {
			t.Errorf("No test function for key %q found", literal.String())
			continue
		}
		testFunc(value)
	}
}

func TestFunctionParameterParsing(t *testing.T) {
	tests := []struct {
		input          string
		expectedParams []string
	}{
		{input: "func() {};", expectedParams: []string{}},
		{input: "func(x) {};", expectedParams: []string{"x"}},
		{input: "func(x, y, z) {};", expectedParams: []string{"x", "y", "z"}},
	}

	for _, tt := range tests {
		lxr := lexer.NewLexer(tt.input)
		psr := NewParser(lxr)
		root := psr.ParseRootStatement()
		checkParserErrors(t, psr)

		stmt := root.Statements[0].(*ast.ExpressionStatement)
		function := stmt.Expression.(*ast.FunctionLiteral)

		if len(function.Parameters) != len(tt.expectedParams) {
			t.Errorf("length parameters wrong. want %d, got=%d\n",
				len(tt.expectedParams), len(function.Parameters))
		}
		for i, identifier := range tt.expectedParams {
			testLiteralExpression(t, function.Parameters[i], identifier)
		}
	}
}

func testInfixExpression(t *testing.T, expr ast.Expression, left interface{}, operator string,
	right interface{},
) bool {
	oprExpr, ok := expr.(*ast.InfixExpression)
	if !ok {
		t.Errorf("expression is not ast.OperatorExpression. got=%T(%s)", expr, expr)
		return false
	}
	if !testLiteralExpression(t, oprExpr.Left, left) {
		return false
	}
	if oprExpr.Operator != operator {
		t.Errorf("oprExpr.Operator is not '%s'. got=%s", operator, oprExpr.Operator)
		return false
	}
	if !testLiteralExpression(t, oprExpr.Right, right) {
		return false
	}
	return true
}

func testLiteralExpression(t *testing.T, expr ast.Expression, expected interface{}) bool {
	switch val := expected.(type) {
	case int:
		return testIntegerLiteral(t, expr, int64(val))
	case int64:
		return testIntegerLiteral(t, expr, val)
	case string:
		return testIdentifier(t, expr, val)
	case bool:
		return testBooleanLiteral(t, expr, val)
	}
	t.Errorf("unhandled type: %T", expected)
	return false
}

func testBooleanLiteral(t *testing.T, expr ast.Expression, value bool) bool {
	boolean, ok := expr.(*ast.Boolean)
	if !ok {
		t.Errorf("expression is not *ast.Boolean. got=%T", expr)
		return false
	}
	if boolean.Value != value {
		t.Errorf("boolean.Value is not %t. got=%t", value, boolean.Value)
		return false
	}
	if boolean.TokenLiteral() != fmt.Sprintf("%t", value) {
		t.Errorf("boolean.TokenLiteral is not %t. got=%s", value, boolean.TokenLiteral())
		return false
	}
	return true
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("expt not *ast.Identifier got=%T", exp)
		return false
	}
	if ident.Value != value {
		t.Errorf("ident.Value not %s. got=%T", exp, exp)
		return false
	}
	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral not %s. got=%s", value, ident.TokenLiteral())
		return false
	}
	return true
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integer, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il not *ast.IntegerLiteral. got=%T", il)
		return false
	}
	if integer.Value != value {
		t.Errorf("integer.Value not %d. got=%d", value, integer.Value)
		return false
	}
	if integer.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integer.TokenLiteral not %d. got=%s", value, integer.TokenLiteral())
		return false
	}
	return true
}
