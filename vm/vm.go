package vm

import (
	"comp/code"
	"comp/compiler"
	"comp/object"
	"errors"
	"fmt"
)

var (
	True  = &object.Boolean{Value: true}
	False = &object.Boolean{Value: false}
	Null  = &object.Null{}
)

const (
	StackSize   = 2048
	GlobalsSize = 65536
	MaxFrames   = 1024
)

type VM struct {
	constants []object.Object

	stack []object.Object
	sp    int // Always points to the next value. Top of stack is stack[sp-1]

	frames     []*Frame
	frameIndex int

	globals []object.Object
}

// NewVMWithGlobalsStore creates a new VM instance initialized with existing global variables.
// This is useful for resuming execution or sharing state across multiple VM instances.
func NewVMWithGlobalsStore(bytecode *compiler.ByteCode, globals []object.Object) *VM {
	vm := NewVM(bytecode)
	vm.globals = globals
	return vm
}

// NewVM creates and returns a new VM instance initialized with the provided bytecode.
// This is the standard entry point for creating a VM from compiled bytecode.
func NewVM(bytecode *compiler.ByteCode) *VM {
	var (
		mainFn    = &object.CompiledFunction{Instructions: bytecode.Instructions}
		mainFrame = NewFrame(mainFn)
		frames    = make([]*Frame, MaxFrames)
	)
	frames[0] = mainFrame
	return &VM{
		constants:  bytecode.Constants,
		stack:      make([]object.Object, StackSize),
		sp:         0,
		globals:    make([]object.Object, GlobalsSize),
		frames:     frames,
		frameIndex: 1,
	}
}

// currentFrame returns the Frame most likely at the top.
func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.frameIndex-1]
}

// pushFrame pushes the passed frame at the top of the Frame stack.
func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.frameIndex] = f
	vm.frameIndex++
}

// popFrame removes the top Frame from the frame stack and returns it.
func (vm *VM) popFrame() *Frame {
	vm.frameIndex--
	return vm.frames[vm.frameIndex]
}

// LastPoppedStackElement returns the most recently popped element from the stack.
// The element remains in the stack array at position sp but is no longer
// considered part of the active stack.
func (vm *VM) LastPoppedStackElement() object.Object {
	return vm.stack[vm.sp]
}

// RunVM executes the bytecode instructions stored in the VM. It loops through
// instructions, decodes opcodes, and performs corresponding operations.
// Returns an error if execution fails at any point.
func (vm *VM) RunVM() error {
	var (
		ins       code.Instructions
		ip        int
		operation code.Opcode
	)
	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++
		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()

		operation = code.Opcode(ins[ip])
		switch operation {
		case code.OpPop:
			vm.pop()
		case code.OpConstant:
			constIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}
		case code.OpTrue:
			if err := vm.push(True); err != nil {
				return err
			}
		case code.OpFalse:
			if err := vm.push(False); err != nil {
				return err
			}
		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}
		case code.OpJump:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.executeBinaryOperation(operation)
			if err != nil {
				return err
			}
		case code.OpMinus:
			err := vm.executeMinusOperation()
			if err != nil {
				return err
			}
		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			err := vm.executeComparison(operation)
			if err != nil {
				return err
			}
		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			vm.globals[globalIndex] = vm.pop()
		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}
		case code.OpNull:
			if err := vm.push(Null); err != nil {
				return err
			}
		case code.OpArray:
			length := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			array := vm.buildArray(vm.sp-length, vm.sp)

			vm.sp = vm.sp - length
			if err := vm.push(array); err != nil {
				return err
			}
		case code.OpHash:
			length := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			hash, err := vm.buildHash(vm.sp-length, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - length
			if err := vm.push(hash); err != nil {
				return err
			}
		case code.OpIndex:
			var (
				index = vm.pop()
				left  = vm.pop()
			)
			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown operation: %d", operation)
		}
	}
	return nil
}

// buildHash creates a new hash object from a range of stack elements.
func (vm *VM) buildHash(startIndex, endIndex int) (object.Object, error) {
	pairs := make(map[object.HashKey]object.HashPair, (endIndex-startIndex)/2)

	for i := startIndex; i < endIndex; i += 2 {
		var (
			key  = vm.stack[i]
			val  = vm.stack[i+1]
			pair = object.HashPair{Key: key, Value: val}
		)
		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}
		pairs[hashKey.HashKey()] = pair
	}
	return &object.Hash{Pairs: pairs}, nil
}

// buildArray creates a new array object from a range of stack elements.
func (vm *VM) buildArray(startIndex, endIndex int) object.Object {
	elements := make([]object.Object, endIndex-startIndex)

	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]
	}
	return &object.Array{Elements: elements}
}

// executeIndexExpression performs an indexing operation on the provided object.
func (vm *VM) executeIndexExpression(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)

	case left.Type() == object.HASH_OBJ:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported for type: %s", left.Type())
	}
}

// executeArrayIndex performs sanity checks and pushes the element at the given
// index or null on the top of the stack.
func (vm *VM) executeArrayIndex(left, index object.Object) error {
	var (
		arrayOb = left.(*object.Array)
		idx     = index.(*object.Integer).Value
		maxIdx  = int64(len(arrayOb.Elements) - 1)
	)
	if idx < 0 || idx > maxIdx {
		return vm.push(Null)
	}
	return vm.push(arrayOb.Elements[idx])
}

// executeHashIndex checks if the key is hashable and pushes the value for
// the corresponding key if exists, or pushes Null.
func (vm *VM) executeHashIndex(left, keyOb object.Object) error {
	hashOb := left.(*object.Hash)

	key, ok := keyOb.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", keyOb.Type())
	}
	pairs, ok := hashOb.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}
	return vm.push(pairs.Value)
}

// executeBinaryOperation performs binary arithmetic/concatenation operation on
// the top two stack elements depending on the object type.
func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return vm.executeBinaryIntegerOperation(op, left, right)

	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return vm.executeBinaryStringOperation(op, left, right)
	default:
		return fmt.Errorf("invalid types for binary operation: %s %s",
			left.Type(), right.Type(),
		)
	}
}

// executeBinaryIntegerOperation performs arithmetic operations (add, subtract, multiply, divide)
// on two integer operands and pushes the result onto the stack.
func (vm *VM) executeBinaryIntegerOperation(op code.Opcode, left, right object.Object) error {
	var (
		lval = left.(*object.Integer).Value
		rval = right.(*object.Integer).Value
	)
	var result int64
	switch op {
	case code.OpAdd:
		result = lval + rval
	case code.OpSub:
		result = lval - rval
	case code.OpMul:
		result = lval * rval
	case code.OpDiv:
		if rval == 0 {
			return fmt.Errorf("division by zero")
		}
		result = lval / rval
	default:
		return fmt.Errorf("invalid integer operation: %d", op)
	}
	return vm.push(&object.Integer{Value: result})
}

// executeBinaryStringOperation concatenates two strings together.
func (vm *VM) executeBinaryStringOperation(op code.Opcode, left, right object.Object) error {
	if op != code.OpAdd {
		return fmt.Errorf("invalid string operation: %d", op)
	}
	var (
		lval = left.(*object.String).Value
		rval = right.(*object.String).Value
	)
	return vm.push(&object.String{Value: lval + rval})
}

// executeBangOperator performs logical negation on the top stack element.
// Returns False for True, True for False and Null, and False for all other values.
func (vm *VM) executeBangOperator() error {
	operand := vm.pop()

	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	case Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

// executeMinusOperation negates the top stack element. Only works with integer
// objects.
func (vm *VM) executeMinusOperation() error {
	operand := vm.pop()

	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf(
			"invalid object type for negation: %s",
			operand.Type(),
		)
	}
	value := operand.(*object.Integer).Value
	return vm.push(&object.Integer{Value: -value})
}

// executeComparison performs comparison operations on the top two stack elements.
// Handles both integer and pointer equality comparisons.
func (vm *VM) executeComparison(op code.Opcode) error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)
	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.executeIntegerComparison(op, left, right)
	}
	switch op {
	case code.OpEqual:
		return vm.push(boolNativeToBoolObject(right == left))
	case code.OpNotEqual:
		return vm.push(boolNativeToBoolObject(right != left))
	default:
		return fmt.Errorf(
			"invalid operator: %d (%s %s)",
			op, left.Type(), right.Type(),
		)
	}
}

// executeIntegerComparison performs comparison operations (greater than, equal, not equal)
// on two integer operands and pushes the boolean result onto the stack.
func (vm *VM) executeIntegerComparison(op code.Opcode, left, right object.Object) error {
	var (
		leftVal  = left.(*object.Integer).Value
		rightVal = right.(*object.Integer).Value
	)
	switch op {
	case code.OpGreaterThan:
		return vm.push(boolNativeToBoolObject(leftVal > rightVal))
	case code.OpEqual:
		return vm.push(boolNativeToBoolObject(leftVal == rightVal))
	case code.OpNotEqual:
		return vm.push(boolNativeToBoolObject(leftVal != rightVal))
	default:
		return fmt.Errorf("invalid operator: %d", op)
	}
}

// isTruthy determines whether an object evaluates to true in a boolean context.
// Returns false for False and Null, true for all other values.
func isTruthy(condition object.Object) bool {
	switch ob := condition.(type) {
	case *object.Boolean:
		return ob.Value
	case *object.Null:
		return false
	default:
		return true
	}
}

// boolNativeToBoolObject converts a native Go boolean to a shared Boolean object.
// Uses singleton instances to avoid unnecessary allocations.
func boolNativeToBoolObject(input bool) *object.Boolean {
	if input {
		return True
	}
	return False
}

// pop removes and returns the top element from the stack.
// Decrements the stack pointer but does not clear the stack array slot.
func (vm *VM) pop() object.Object {
	ob := vm.stack[vm.sp-1]
	vm.sp--
	return ob
}

// push adds an object to the top of the stack.
// Returns an error if the stack is full.
func (vm *VM) push(ob object.Object) error {
	if vm.sp >= StackSize {
		return errors.New("stack overflow")
	}
	vm.stack[vm.sp] = ob
	vm.sp++
	return nil
}
