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
	symbolTable         *SymbolTable
}

// NewWithState creates a new Compiler instance initialized with the existing state.
// This is useful for resuming compilation or reusing the compiler state across
// multiple compilation passes.
//
// Returns a pointer to a newly created Compiler with the provided state injected.
func NewWithState(sym *SymbolTable, consts []object.Object) *Compiler {
	compiler := NewCompiler()
	compiler.symbolTable = sym
	compiler.constants = consts
	return compiler
}

// NewCompiler creates and returns a new Compiler instance with an empty / default state.
// This is the standard entry point for starting a new compilation process from scratch.
//
// Returns a pointer to the newly created Compiler instance.
func NewCompiler() *Compiler {
	return &Compiler{
		instructions:        code.Instructions{},
		constants:           []object.Object{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
		symbolTable:         NewSymbolTable(),
	}
}

// Compile walks the AST recursively until it encounters a node that can be compiled/evaluated.
//
// Works similar to the Evaluate function
func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.RootStatement:
		for _, stmt := range node.Statements {
			err := c.Compile(stmt)
			if err != nil {
				return err
			}
		}
	case *ast.LetStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		symbol := c.symbolTable.Define(node.Name.Value)
		c.emit(code.OpSetGlobal, symbol.Index)
	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable: %s", node.Value)
		}
		c.emit(code.OpGetGlobal, symbol.Index)
	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)
	case *ast.BlockStatement:
		for _, stmt := range node.Statements {
			err := c.Compile(stmt)
			if err != nil {
				return err
			}
		}
	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}
		switch node.Operator {
		case "-":
			c.emit(code.OpMinus)
		case "!":
			c.emit(code.OpBang)
		default:
			return fmt.Errorf("invalid operation: %s", node.Operator)
		}
	case *ast.InfixExpression:
		err := c.compileInfix(node)
		if err != nil {
			return err
		}
	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}
		posJumpNotTruthy := c.emit(code.OpJumpNotTruthy, 1000)

		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}
		if c.lastInstructionIsPop() {
			c.removeLastPop()
		}
		return c.handleJump(node, posJumpNotTruthy)
	case *ast.Boolean:
		if !node.Value {
			c.emit(code.OpFalse)
		} else {
			c.emit(code.OpTrue)
		}
	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))
	}
	return nil
}

// changeOperand creates the instruction for the given operand and swaps the old
// instruction for the new one - including the operand.
func (c *Compiler) changeOperand(pos int, operand int) {
	op := code.Opcode(c.instructions[pos])
	ins := code.MakeInstruction(op, operand)
	c.replaceInstruction(pos, ins)
}

// replaceInstruction replaces an instruction at (pos)[position:] with the
// provided instruction
func (c *Compiler) replaceInstruction(pos int, instruction []byte) {
	for i := 0; i < len(instruction); i++ {
		c.instructions[pos+i] = instruction[i]
	}
}

// handleJump handles jump operations over conditionals depending on resulting
// truthy value or lack thereof.
func (c *Compiler) handleJump(node *ast.IfExpression, posJumpNotTruthy int) error {
	posJump := c.emit(code.OpJump, 1000)

	posAfterConsequence := len(c.instructions)
	c.changeOperand(posJumpNotTruthy, posAfterConsequence)

	if node.Alternative == nil {
		c.emit(code.OpNull)
	} else {
		err := c.Compile(node.Alternative)
		if err != nil {
			return err
		}
		if c.lastInstructionIsPop() {
			c.removeLastPop()
		}
	}
	posAfterAlternative := len(c.instructions)
	c.changeOperand(posJump, posAfterAlternative)
	return nil
}

// addConstant appends ob to the compiler's constant slice.
//
// Returns the index of the constant in the constant pool as its very own identifier
func (c *Compiler) addConstant(ob object.Object) int {
	c.constants = append(c.constants, ob)
	return len(c.constants) - 1
}

// emit generates an instruction and adds it to a collection in memory.
//
// Returns the starting position of the just emitted(added to memory) instruction.
func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.MakeInstruction(op, operands...)
	pos := c.addInstruction(ins)
	c.setLastInstruction(op, pos)
	return pos
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	c.previousInstruction = c.lastInstruction
	last := EmittedInstruction{
		OpCode:   op,
		Position: pos,
	}
	c.lastInstruction = last
}

// addInstruction adds the given (ins) instruction and stores the len of
// already present instruction and returns the same as that will be the
// instruction's offset
func (c *Compiler) addInstruction(ins []byte) int {
	posNewIns := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return posNewIns
}

// is pop?
func (c *Compiler) lastInstructionIsPop() bool {
	return c.lastInstruction.OpCode == code.OpPop
}

// removeLastPop removes the last instruction from the instructions
// slice and updates the lastInstruction field
func (c *Compiler) removeLastPop() {
	c.instructions = c.instructions[:c.lastInstruction.Position]
	c.lastInstruction = c.previousInstruction
}

// compileInfix performs the same recursive compilation that Compile does.
func (c *Compiler) compileInfix(node *ast.InfixExpression) error {
	switch {
	case node.Operator == "<":
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}
		err = c.Compile(node.Left)
		if err != nil {
			return err
		}
		c.emit(code.OpGreaterThan)
		return nil
	default:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}
		err = c.Compile(node.Right)
		if err != nil {
			return err
		}
		err = c.emitInfixOp(node)
		if err != nil {
			return err
		}
	}
	return nil
}

// emitInfixOp emits the corresponding code.Opcode for each infix operator
func (c *Compiler) emitInfixOp(infixExpr *ast.InfixExpression) error {
	switch infixExpr.Operator {
	case "+":
		c.emit(code.OpAdd)
	case "-":
		c.emit(code.OpSub)
	case "*":
		c.emit(code.OpMul)
	case "/":
		c.emit(code.OpDiv)
	case "!=":
		c.emit(code.OpNotEqual)
	case "==":
		c.emit(code.OpEqual)
	case ">":
		c.emit(code.OpGreaterThan)
	default:
		return fmt.Errorf("unknown operator %s", infixExpr.Operator)
	}
	return nil
}

type ByteCode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func (c *Compiler) ByteCode() *ByteCode {
	return &ByteCode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}
