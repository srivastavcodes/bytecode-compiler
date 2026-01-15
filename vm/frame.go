package vm

import (
	"comp/code"
	"comp/object"
)

// Frame represents a call frame in the virtual machine's call stack.
// It tracks the execution state of a compiled function, including:
//   - fn: the compiled function being executed
//   - ip: instruction pointer, the index of the next instruction to execute
//   - basePointer: the base pointer in the VM's stack for this frame's local variables
type Frame struct {
	fn *object.CompiledFunction
	ip int

	basePointer int
}

// NewFrame returns a pointer to an initialized Frame with the basePointer
// passed in.
func NewFrame(fn *object.CompiledFunction, bp int) *Frame {
	return &Frame{
		fn: fn,
		ip: -1,

		basePointer: bp,
	}
}

// Instructions returns the instructions inside the function within the
// Frame.
func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
