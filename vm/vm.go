package vm

import (
	"comp/code"
	"comp/compiler"
	"comp/object"
	"encoding/binary"
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
)

type VM struct {
	instructions code.Instructions
	constants    []object.Object

	stack []object.Object
	sp    int // Always points to the next value. Top of stack is stack[sp-1]

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
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		stack:        make([]object.Object, StackSize),
		sp:           0,
		globals:      make([]object.Object, GlobalsSize),
	}
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
	for ip := 0; ip < len(vm.instructions); ip++ {
		operation := code.Opcode(vm.instructions[ip])

		switch operation {
		case code.OpPop:
			vm.pop()
		case code.OpConstant:
			constIndex := binary.BigEndian.Uint16(vm.instructions[ip+1:])
			ip += 2
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
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip = pos - 1
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2
			condition := vm.pop()
			if !isTruthy(condition) {
				ip = pos - 1
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
			globalIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2
			vm.globals[globalIndex] = vm.pop()
		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2
			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}
		case code.OpNull:
			if err := vm.push(Null); err != nil {
				return err
			}
		case code.OpArray:
			length := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2
			array := vm.buildArray(vm.sp-length, vm.sp)

			vm.sp = vm.sp - length
			if err := vm.push(array); err != nil {
				return err
			}
		case code.OpHash:
			length := int(code.ReadUint16(vm.instructions[ip+1:]))
			ip += 2
			hash, err := vm.buildHash(vm.sp-length, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - length
			if err := vm.push(hash); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown operation: %d", operation)
		}
	}
	return nil
}

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

// executeBinaryOperation performs binary arithmetic operations on the top two stack elements.
// Currently supports integer operations only.
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
