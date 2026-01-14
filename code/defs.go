package code

const (
	OpConstant Opcode = iota
	OpPop
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpTrue
	OpFalse
	OpEqual
	OpNotEqual
	OpGreaterThan
	OpMinus
	OpBang
	OpJumpNotTruthy
	OpJump
	OpNull
	OpGetGlobal
	OpSetGlobal
	OpArray
	OpHash
	OpIndex
	OpCall
	OpReturnValue
	OpReturn
	OpGetLocal
	OpSetLocal
)

type Instructions []byte

type Opcode byte

type Definition struct {
	Name         string
	OperandWidth []int
}

var byte0 []int

var definitions = map[Opcode]*Definition{
	OpConstant:      {"OpConstant", []int{2}},
	OpPop:           {"OpPop", byte0},
	OpAdd:           {"OpAdd", byte0},
	OpSub:           {"OpSub", byte0},
	OpMul:           {"OpMul", byte0},
	OpDiv:           {"OpDiv", byte0},
	OpTrue:          {"OpTrue", byte0},
	OpFalse:         {"OpFalse", byte0},
	OpEqual:         {"OpEqual", byte0},
	OpNotEqual:      {"OpNotEqual", byte0},
	OpGreaterThan:   {"OpGreaterThan", byte0},
	OpMinus:         {"OpMinus", byte0},
	OpBang:          {"OpBang", byte0},
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}},
	OpJump:          {"OpJump", []int{2}},
	OpNull:          {"OpNull", byte0},
	OpGetGlobal:     {"OpGetGlobal", []int{2}},
	OpSetGlobal:     {"OpSetGlobal", []int{2}},
	OpArray:         {"OpArray", []int{2}},
	OpHash:          {"OpHash", []int{2}},
	OpIndex:         {"OpIndex", byte0},
	OpCall:          {"OpCall", byte0},
	OpReturnValue:   {"OpReturnValue", byte0},
	OpReturn:        {"OpReturn", byte0},
	OpGetLocal:      {"OpGetLocal", []int{1}},
	OpSetLocal:      {"OpSetLocal", []int{1}},
}
