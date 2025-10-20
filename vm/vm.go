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
)

const StackSize = 2048

type VM struct {
	instructions code.Instructions
	constants    []object.Object

	stack []object.Object
	sp    int // Always points to the next value. Top of stack is stack[sp-1]
}

func NewVM(bytecode *compiler.ByteCode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,
		stack:        make([]object.Object, StackSize),
		sp:           0,
	}
}

func (vm *VM) LastPoppedStackElement() object.Object {
	return vm.stack[vm.sp]
}

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
			err := vm.push(True)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.executeBinaryOperation(operation)
			if err != nil {
				return err
			}
		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			err := vm.executeComparison(operation)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	var (
		right = vm.pop()
		left  = vm.pop()
	)
	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.executeBinaryIntegerOperation(op, left, right)
	}
	return fmt.Errorf(
		"invalid types for binary operation: %s %s",
		left.Type(), right.Type(),
	)
}

func (vm *VM) executeBinaryIntegerOperation(op code.Opcode, left, right object.Object) error {
	var (
		leftVal  = left.(*object.Integer).Value
		rightVal = right.(*object.Integer).Value
	)
	var result int64
	switch op {
	case code.OpAdd:
		result = leftVal + rightVal
	case code.OpSub:
		result = leftVal - rightVal
	case code.OpMul:
		result = leftVal * rightVal
	case code.OpDiv:
		result = leftVal / rightVal
	default:
		return fmt.Errorf("invalid interger operation: %d", op)
	}
	return vm.push(&object.Integer{Value: result})
}

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

func boolNativeToBoolObject(input bool) *object.Boolean {
	if input {
		return True
	} else {
		return False
	}
}

func (vm *VM) pop() object.Object {
	ob := vm.stack[vm.sp-1]
	vm.sp--
	return ob
}

func (vm *VM) push(ob object.Object) error {
	if vm.sp >= StackSize {
		return errors.New("stack overflow")
	}
	vm.stack[vm.sp] = ob
	vm.sp++
	return nil
}
