package ast

import (
	"comp/token"
	"testing"
)

func TestString(t *testing.T) {
	root := &RootStatement{
		Statements: []Statement{
			&LetStatement{
				Token: token.Token{Type: token.LET, Literal: "let"},
				Name: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "myVar"},
					Value: "myVar",
				},
				Value: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "anotherVar"},
					Value: "anotherVar",
				},
			},
		},
	}
	if root.String() != "let myVar = anotherVar;" {
		t.Errorf("root.String() wrong. got=%q", root.String())
	}
}
