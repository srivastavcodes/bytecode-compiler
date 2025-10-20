package vm

import (
	"comp/ast"
	"comp/compiler"
	"comp/lexer"
	"comp/object"
	"comp/parser"
	"fmt"
	"testing"
)

type vmTestCase struct {
	input    string
	expected interface{}
}

func parse(input string) *ast.RootStatement {
	lxr := lexer.NewLexer(input)
	psr := parser.NewParser(lxr)
	return psr.ParseRootStatement()
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1", 1},
		{"2", 2},
		{"1 + 2", 3},
		{"1 - 2", -1},
		{"1 * 2", 2},
		{"4 / 2", 2},
		{"50 / 2 * 2 + 10 - 5", 55},
		{"5 * (2 + 10)", 60},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"5 * (2 + 10)", 60},
		// {"-5", -5},
		// {"-10", -10},
		// {"-50 + 100 + -50", 0},
		// {"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}
	runVmTests(t, tests)
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)
		comp := compiler.NewCompiler()

		err := comp.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}
		vm := NewVM(comp.ByteCode())
		err = vm.RunVM()
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}
		stackElem := vm.LastPoppedStackElement()
		testExpectedObject(t, tt.expected, stackElem)
	}
}

func testExpectedObject(t *testing.T, expected interface{}, actual object.Object) {
	t.Helper()

	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(int64(expected), actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
		/*	case bool:
				err := testBooleanObject(bool(expected), actual)
				if err != nil {
					t.Errorf("testBooleanObject failed: %s", err)
				}
			case string:
				err := testStringObject(expected, actual)
				if err != nil {
					t.Errorf("testStringObject failed: %s", err)
				}
			case []int:
				array, ok := actual.(*object.Array)
				if !ok {
					t.Errorf("object not Array: %T (%+v)", actual, actual)
					return
				}
				if len(array.Elements) != len(expected) {
					t.Errorf("wrong num of elements. want=%d, got=%d", len(expected), len(array.Elements))
					return
				}
				for i, expectedElm := range expected {
					err := testIntegerObject(int64(expectedElm), array.Elements[i])
					if err != nil {
						t.Errorf("testIntegerObject failed: %s", err)
					}
				}
			case map[object.HashKey]int64:
				hash, ok := actual.(*object.Hash)
				if !ok {
					t.Errorf("object is not Hash. got=%T (%+v)", actual, actual)
					return
				}
				if len(hash.Pairs) != len(expected) {
					t.Errorf("hash has wrong number of Pairs. want=%d, got=%d", len(expected), len(hash.Pairs))
					return
				}
				for expectedKey, expectedValue := range expected {
					pair, ok := hash.Pairs[expectedKey]
					if !ok {
						t.Errorf("no pair for given key in pairs")
					}
					err := testIntegerObject(expectedValue, pair.Value)
					if err != nil {
						t.Errorf("testIntegerObject failed: %s", err)
					}
				}
			case *object.Null:
				if actual != Null {
					t.Errorf("object is not Null: %T (%+v)", actual, actual)
				}
			case *object.Error:
				errObj, ok := actual.(*object.Error)
				if !ok {
					t.Errorf("object is not Error: %T (%+v)", actual, actual)
					return
				}
				if errObj.Message != expected.Message {
					t.Errorf("wrong error message. expected=%q, got=%q", expected.Message, errObj.Message)
				}
		*/
	}
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%+v)", actual, actual)
	}
	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%d, want=%d",
			result.Value, expected)
	}
	return nil
}
