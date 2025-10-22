package compiler

import (
	"comp/ast"
	"comp/code"
	"comp/object"
	"fmt"
)

type EmittedInstruction struct {
	OpCode   code.Opcode
	Position int
}

type Compiler struct {
	instructions        code.Instructions
	constants           []object.Object
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

func NewCompiler() *Compiler {
	return &Compiler{
		instructions:        code.Instructions{},
		constants:           []object.Object{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
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
	case *ast.BlockStatement:
		for _, stmt := range node.Statements {
			err := cmp.Compile(stmt)
			if err != nil {
				return err
			}
		}
	case *ast.PrefixExpression:
		err := cmp.Compile(node.Right)
		if err != nil {
			return err
		}
		switch node.Operator {
		case "-":
			cmp.emit(code.OpMinus)
		case "!":
			cmp.emit(code.OpBang)
		default:
			return fmt.Errorf("invalid operation: %s", node.Operator)
		}
	case *ast.InfixExpression:
		err := cmp.compileInfix(node)
		if err != nil {
			return err
		}
	case *ast.IfExpression:
		err := cmp.Compile(node.Condition)
		if err != nil {
			return err
		}
		posJumpNotTruthy := cmp.emit(code.OpJumpNotTruthy, 1000)

		err = cmp.Compile(node.Consequence)
		if err != nil {
			return err
		}
		if cmp.lastInstructionIsPop() {
			cmp.removeLastPop()
		}
		return cmp.handleJump(node, posJumpNotTruthy)
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

// changeOperand creates the instruction for the given operand and swaps the old
// instruction for the new one - including the operand.
func (cmp *Compiler) changeOperand(pos int, operand int) {
	op := code.Opcode(cmp.instructions[pos])
	ins := code.MakeInstruction(op, operand)
	cmp.replaceInstruction(pos, ins)
}

// replaceInstruction replaces an instruction at (pos)[position:] with the
// provided instruction
func (cmp *Compiler) replaceInstruction(pos int, instruction []byte) {
	for i := 0; i < len(instruction); i++ {
		cmp.instructions[pos+i] = instruction[i]
	}
}

// handleJump handles jump operations over conditionals depending on resulting
// truthy value or lack thereof.
func (cmp *Compiler) handleJump(node *ast.IfExpression, posJumpNotTruthy int) error {
	posJump := cmp.emit(code.OpJump, 1000)

	posAfterConsequence := len(cmp.instructions)
	cmp.changeOperand(posJumpNotTruthy, posAfterConsequence)

	if node.Alternative == nil {
		cmp.emit(code.OpNull)
	} else {
		err := cmp.Compile(node.Alternative)
		if err != nil {
			return err
		}
		if cmp.lastInstructionIsPop() {
			cmp.removeLastPop()
		}
	}
	posAfterAlternative := len(cmp.instructions)
	cmp.changeOperand(posJump, posAfterAlternative)
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
	cmp.setLastInstruction(op, pos)
	return pos
}

func (cmp *Compiler) setLastInstruction(op code.Opcode, pos int) {
	cmp.previousInstruction = cmp.lastInstruction
	last := EmittedInstruction{
		OpCode:   op,
		Position: pos,
	}
	cmp.lastInstruction = last
}

// addInstruction adds the given (ins) instruction and stores the len of
// already present instruction and returns the same as that will be the
// instruction's offset
func (cmp *Compiler) addInstruction(ins []byte) int {
	posNewIns := len(cmp.instructions)
	cmp.instructions = append(cmp.instructions, ins...)
	return posNewIns
}

// is pop?
func (cmp *Compiler) lastInstructionIsPop() bool {
	return cmp.lastInstruction.OpCode == code.OpPop
}

// removeLastPop removes the last instruction from the instructions
// slice and updates the lastInstruction field
func (cmp *Compiler) removeLastPop() {
	cmp.instructions = cmp.instructions[:cmp.lastInstruction.Position]
	cmp.lastInstruction = cmp.previousInstruction
}

// compileInfix performs the same recursive compilation that Compile does.
func (cmp *Compiler) compileInfix(node *ast.InfixExpression) error {
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
	return nil
}

// emitInfixOp emits the corresponding code.Opcode for each infix operator
func (cmp *Compiler) emitInfixOp(infixExpr *ast.InfixExpression) error {
	switch infixExpr.Operator {
	case "+":
		cmp.emit(code.OpAdd)
	case "-":
		cmp.emit(code.OpSub)
	case "*":
		cmp.emit(code.OpMul)
	case "/":
		cmp.emit(code.OpDiv)
	case "!=":
		cmp.emit(code.OpNotEqual)
	case "==":
		cmp.emit(code.OpEqual)
	case ">":
		cmp.emit(code.OpGreaterThan)
	default:
		return fmt.Errorf("unknown operator %s", infixExpr.Operator)
	}
	return nil
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
