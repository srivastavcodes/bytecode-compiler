package compiler

import (
	"comp/ast"
	"comp/code"
	"comp/object"
	"fmt"
)

type Compiler struct {
	instructions code.Instructions
	constants    []object.Object
}

func NewCompiler() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},
	}
}

// Compile walks the AST recursively until it encounters a node that can be compiled/evaluated.
//
// works similar to the Evaluate function
func (cmp *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.RootStatement:
		for _, stmt := range node.Statements {
			err := cmp.Compile(stmt)
			if err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		err := cmp.Compile(node.Expression)
		if err != nil {
			return err
		}
		cmp.emit(code.OpPop)
	case *ast.InfixExpression:
		switch {
		case node.Operator == "<":
			err := cmp.Compile(node.Right)
			if err != nil {
				return err
			}
			err = cmp.Compile(node.Left)
			if err != nil {
				return err
			}
			cmp.emit(code.OpGreaterThan)
			return nil
		default:
			err := cmp.Compile(node.Left)
			if err != nil {
				return err
			}
			err = cmp.Compile(node.Right)
			if err != nil {
				return err
			}
			err = cmp.emitInfixOp(node)
			if err != nil {
				return err
			}
		}
	case *ast.Boolean:
		if !node.Value {
			cmp.emit(code.OpFalse)
		} else {
			cmp.emit(code.OpTrue)
		}
	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		cmp.emit(code.OpConstant, cmp.addConstant(integer))
	}
	return nil
}

// emitInfixOp emits the corresponding code.Opcode for each infix
// operator
func (cmp *Compiler) emitInfixOp(node ast.Node) error {
	switch node.(*ast.InfixExpression).Operator {
	case "+":
		cmp.emit(code.OpAdd)
	case "-":
		cmp.emit(code.OpSub)
	case "*":
		cmp.emit(code.OpMul)
	case "/":
		cmp.emit(code.OpDiv)
	case ">":
		cmp.emit(code.OpGreaterThan)
	case "==":
		cmp.emit(code.OpEqual)
	case "!=":
		cmp.emit(code.OpNotEqual)
	default:
		return fmt.Errorf(
			"unknown operator %s",
			node.(*ast.InfixExpression).Operator,
		)
	}
	return nil
}

// addConstant appends ob to the compiler's constant slice.
//
// Returns the index of the constant in the constant pool as its very own identifier
func (cmp *Compiler) addConstant(ob object.Object) int {
	cmp.constants = append(cmp.constants, ob)
	return len(cmp.constants) - 1
}

// emit generates an instruction and adds it to a collection in memory.
//
// Returns the starting position of the just emitted(added to memory) instruction.
func (cmp *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.MakeInstruction(op, operands...)
	pos := cmp.addInstruction(ins)
	return pos
}

func (cmp *Compiler) addInstruction(ins []byte) int {
	// stores the len of already present instruction and returns the same because the
	// newly added instruction will start from there
	posNewIns := len(cmp.instructions)
	cmp.instructions = append(cmp.instructions, ins...)
	return posNewIns
}

type ByteCode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func (cmp *Compiler) ByteCode() *ByteCode {
	return &ByteCode{
		Instructions: cmp.instructions,
		Constants:    cmp.constants,
	}
}
