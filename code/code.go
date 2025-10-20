package code

import (
	"encoding/binary"
	"fmt"
	"strings"
)

const (
	OpConstant Opcode = iota
	OpPop
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpTrue
	OpFalse
)

type Instructions []byte

type Opcode byte

type Definition struct {
	Name         string
	OperandWidth []int
}

var definitions = map[Opcode]*Definition{
	OpConstant: {"OpConstant", []int{2}},
	OpPop:      {"OpPop", []int{}},
	OpAdd:      {"OpAdd", []int{}},
	OpSub:      {"OpAdd", []int{}},
	OpMul:      {"OpAdd", []int{}},
	OpDiv:      {"OpAdd", []int{}},
	OpTrue:     {"OpTrue", []int{}},
	OpFalse:    {"OpFalse", []int{}},
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}

// Make Returns a byte slice with the opcode as first byte followed by operands encoded
// in big-endian format according to their defined widths.
//
// Example: Make(OpConstant, 65534) returns [OpConstant, 0xFF, 0xFE]
func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}
	instructionLen := 1

	for _, w := range def.OperandWidth {
		instructionLen += w
	}
	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for i, opr := range operands {
		width := def.OperandWidth[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(opr))
		}
		offset += width
	}
	return instruction
}

func (ins Instructions) String() string {
	var out strings.Builder

	for i := 0; i < len(ins); {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}
		operands, read := ReadOperands(def, ins[i+1:])
		str := ins.fmtInstruction(def, operands)

		fmt.Fprintf(&out, "%04d %s\n", i, str)
		i += 1 + read
	}
	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidth)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: "+
			"operand len %d does not match defined %d\n", len(operands), operandCount)
	}
	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}
	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

// ReadOperands extracts operand values from bytecode instruction bytes.
// Takes a definition specifying operand widths and returns the decoded operands
// along with the total bytes consumed.
//
// Example: ReadOperands(Opcode Definition, [0xFF, 0xFE]) returns ([65534], 2)
func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidth))
	offset := 0

	for i, width := range def.OperandWidth {
		switch width {
		case 2:
			operands[i] = int(binary.BigEndian.Uint16(ins[offset:]))
		}
		offset += width
	}
	return operands, offset
}
