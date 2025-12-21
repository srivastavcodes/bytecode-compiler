package code

import (
	"encoding/binary"
	"fmt"
	"strings"
)

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}

// MakeInstruction Returns a byte slice with the opcode as the first byte followed
// by operands encoded in big-endian format according to their defined widths.
//
// Example: MakeInstruction(OpConstant, 65534) returns [OpConstant, 0xFF, 0xFE]
func MakeInstruction(op Opcode, operands ...int) []byte {
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

func (in Instructions) String() string {
	var out strings.Builder

	for i := 0; i < len(in); {
		def, err := Lookup(in[i])
		if err != nil {
			_, _ = fmt.Fprintf(&out, "ERROR: %s\n", err)
			i++
			continue
		}
		operands, read := ReadOperands(def, in[i+1:])
		str := in.instructionFmt(def, operands)

		_, _ = fmt.Fprintf(&out, "%04d %s\n", i, str)
		i += 1 + read
	}
	return out.String()
}

func (in Instructions) instructionFmt(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidth)

	if len(operands) != operandCount {
		return fmt.Sprintf(
			"ERROR: operand len %d does not match defined %d\n",
			len(operands), operandCount,
		)
	}
	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}
	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

// ReadOperands extracts operand values from bytecode instructionFmt bytes.
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
			operands[i] = int(ReadUint16(ins[offset:]))
		}
		offset += width
	}
	return operands, offset
}

// ReadUint16 reads two consecutive bytes from the given Instructions
// and converts them back to an uint16 using big-endian byte order.
func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}
