package object

import (
	"comp/ast"
	"comp/code"
	"fmt"
	"hash/fnv"
	"strings"
)

type ObjectType string

type BuiltInFunction func(args ...Object) Object

const (
	COLOR_RED   = "\033[31m"
	COLOR_RESET = "\033[0m"
	// todo -> add support for logging same as 'puts' with COLOR_GREEN = "\033[32m"
)

const (
	INTEGER_OBJ           = "INTEGER"
	BOOLEAN_OBJ           = "BOOLEAN"
	NULL_OBJ              = "NULL"
	RETURN_VALUE_OBJ      = "RETURN_VALUE"
	ERROR_OBJ             = "ERROR"
	FUNCTION_OBJ          = "FUNCTION"
	STRING_OBJ            = "STRING"
	BUILTIN_OBJ           = "BUILTIN"
	HASH_OBJ              = "HASH"
	ARRAY_OBJ             = "ARRAY"
	COMPILED_FUNCTION_OBJ = "COMPILED_FUNCTION"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type CompiledFunction struct {
	Instructions  code.Instructions
	NumLocals     int
	NumParameters int
}

func (cf *CompiledFunction) Type() ObjectType { return COMPILED_FUNCTION_OBJ }

func (cf *CompiledFunction) Inspect() string {
	return fmt.Sprintf("CompiledFunction[%p]", cf)
}

type Integer struct {
	Value int64
}

func (ig *Integer) Type() ObjectType { return INTEGER_OBJ }

func (ig *Integer) Inspect() string { return fmt.Sprintf("%d", ig.Value) }

type String struct {
	Value string
}

func (str *String) Type() ObjectType { return STRING_OBJ }

func (str *String) Inspect() string { return str.Value }

type Boolean struct {
	Value bool
}

func (bl *Boolean) Type() ObjectType { return BOOLEAN_OBJ }

func (bl *Boolean) Inspect() string { return fmt.Sprintf("%t", bl.Value) }

type Null struct{}

func (nl *Null) Inspect() string { return "nil" }

func (nl *Null) Type() ObjectType { return NULL_OBJ }

type Return struct {
	Value Object
}

func (rv *Return) Type() ObjectType { return RETURN_VALUE_OBJ }

func (rv *Return) Inspect() string { return rv.Value.Inspect() }

type Error struct {
	Message string
}

func (er *Error) Type() ObjectType { return ERROR_OBJ }

func (er *Error) Inspect() string {
	return fmt.Sprintf("%sERROR::%s %s", COLOR_RED, COLOR_RESET, er.Message)
}

type Function struct {
	Parameters []*ast.Identifier
	Env        *Environment
	Body       *ast.BlockStatement
}

func (fn *Function) Type() ObjectType { return FUNCTION_OBJ }

func (fn *Function) Inspect() string {
	var output strings.Builder
	var params []string

	for _, pr := range fn.Parameters {
		params = append(params, pr.String())
	}
	output.WriteString("func(")
	output.WriteString(strings.Join(params, ", "))
	output.WriteString(") {\n")
	output.WriteString(fn.Body.String() + "\n")

	return output.String()
}

type BuiltIn struct {
	Func BuiltInFunction
}

func (bl *BuiltIn) Type() ObjectType { return BUILTIN_OBJ }

func (bl *BuiltIn) Inspect() string { return "builtin function" }

type Array struct {
	Elements []Object
}

func (arr *Array) Type() ObjectType { return ARRAY_OBJ }

func (arr *Array) Inspect() string {
	var out strings.Builder

	var values []string
	for _, val := range arr.Elements {
		values = append(values, val.Inspect())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(values, ", "))
	out.WriteString("]")

	return out.String()
}

type Hashable interface {
	HashKey() HashKey // todo -> add caching to the HashKey() returned values
}

type HashKey struct {
	Type  ObjectType
	Value uint64
}

func (bl *Boolean) HashKey() HashKey {
	var value uint64
	if bl.Value {
		value = 1
	} else {
		value = 0
	}
	return HashKey{Type: bl.Type(), Value: value}
}

func (ig *Integer) HashKey() HashKey {
	return HashKey{Type: ig.Type(), Value: uint64(ig.Value)}
}

func (str *String) HashKey() HashKey {
	hash := fnv.New64a()
	hash.Write([]byte(str.Value))
	return HashKey{Type: str.Type(), Value: hash.Sum64()}
}

type HashPair struct {
	Key   Object
	Value Object
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

func (hs *Hash) Type() ObjectType { return HASH_OBJ }

func (hs *Hash) Inspect() string {
	var out strings.Builder
	var pairs []string

	for _, pair := range hs.Pairs {
		data := fmt.Sprintf("%s:%s", pair.Key.Inspect(), pair.Value.Inspect())
		pairs = append(pairs, data)
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}
