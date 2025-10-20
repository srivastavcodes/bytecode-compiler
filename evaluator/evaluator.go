package evaluator

import (
	"comp/ast"
	"comp/object"
	"fmt"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Evaluate(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.RootStatement:
		return evalRootStatement(node, env)
	case *ast.LetStatement:
		value := Evaluate(node.Value, env)
		if isError(value) {
			return value
		}
		env.Set(node.Name.Value, value)
	case *ast.ExpressionStatement:
		return Evaluate(node.Expression, env)
	case *ast.ReturnStatement:
		reVal := Evaluate(node.ReturnValue, env)
		if isError(reVal) {
			return reVal
		}
		return &object.Return{Value: reVal}
	case *ast.CallExpression:
		fn := Evaluate(node.Function, env)
		if isError(fn) {
			return fn
		}
		args := evalListExpression(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}
		return applyFunction(fn, args)

	case *ast.Identifier:
		return evalIdentifier(node, env)

	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.Boolean:
		return boolNativeToBoolObject(node.Value)
	case *ast.ArrayLiteral:
		values := evalListExpression(node.Elements, env)
		if len(values) == 1 && isError(values[0]) {
			return values[0]
		}
		return &object.Array{Elements: values}
	case *ast.HashLiteral:
		return evalHashLiteral(node, env)

	case *ast.PrefixExpression:
		right := Evaluate(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		lt := Evaluate(node.Left, env)
		if isError(lt) {
			return lt
		}
		rt := Evaluate(node.Right, env)
		if isError(rt) {
			return rt
		}
		return evalInfixExpression(node.Operator, lt, rt)
	case *ast.IndexExpression:
		lt := Evaluate(node.Left, env)
		if isError(lt) {
			return lt
		}
		idx := Evaluate(node.Index, env)
		if isError(idx) {
			return idx
		}
		return evalIndexExpression(lt, idx)

	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.IfExpression:
		return evalConditionalExpression(node, env)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Body: body, Env: env}
	}
	return nil
}

func evalRootStatement(root *ast.RootStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, stmt := range root.Statements {
		result = Evaluate(stmt, env)

		switch result := result.(type) {
		case *object.Error:
			return result
		case *object.Return:
			return result.Value
		}
	}
	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, stmt := range block.Statements {
		result = Evaluate(stmt, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}
	return result
}

func evalListExpression(args []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, arg := range args {
		value := Evaluate(arg, env)
		if isError(value) {
			return []object.Object{value}
		}
		result = append(result, value)
	}
	return result
}

func evalIndexExpression(lt, idx object.Object) object.Object {
	switch {
	case lt.Type() == object.ARRAY_OBJ && idx.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(lt, idx)
	case lt.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(lt, idx)
	default:
		return createError("index operator not supported: %s", lt.Type())
	}
}

func evalArrayIndexExpression(arr, idx object.Object) object.Object {
	index := idx.(*object.Integer).Value
	array := arr.(*object.Array)

	last := int64(len(array.Elements) - 1)
	if index < 0 || index > last {
		return NULL
	}
	return array.Elements[index]
}

func evalHashIndexExpression(hash, idx object.Object) object.Object {
	hashOb := hash.(*object.Hash)

	key, ok := idx.(object.Hashable)
	if !ok {
		return createError("unusable as hash key: %s", idx.Type())
	}
	pair, ok := hashOb.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}
	return pair.Value
}

func evalHashLiteral(hash *ast.HashLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valNode := range hash.Pairs {
		key := Evaluate(keyNode, env)
		if isError(key) {
			return key
		}
		hashKey, ok := key.(object.Hashable)
		if !ok {
			return createError("unusable as hash key: %s", key.Type())
		}
		value := Evaluate(valNode, env)
		if isError(value) {
			return value
		}
		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}
	return &object.Hash{Pairs: pairs}
}

func evalIdentifier(id *ast.Identifier, env *object.Environment) object.Object {
	if builtIn, ok := builtIns[id.Value]; ok {
		return builtIn
	}
	if val, ok := env.Get(id.Value); ok {
		return val
	}
	return createError("Identifier '" + id.Value + "' not found")
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalPrefixNegationExpression(right)
	default:
		return createError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)

	case operator == "==":
		return boolNativeToBoolObject(left == right)
	case operator == "!=":
		return boolNativeToBoolObject(left != right)

	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)

	case left.Type() != right.Type():
		return createError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return createError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, lt, rt object.Object) object.Object {
	ltVal := lt.(*object.Integer).Value
	rtVal := rt.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: ltVal + rtVal}
	case "-":
		return &object.Integer{Value: ltVal - rtVal}
	case "*":
		return &object.Integer{Value: ltVal * rtVal}
	case "/":
		return &object.Integer{Value: ltVal / rtVal}

	case "<":
		return boolNativeToBoolObject(ltVal < rtVal)
	case ">":
		return boolNativeToBoolObject(ltVal > rtVal)
	case "==":
		return boolNativeToBoolObject(ltVal == rtVal)
	case "!=":
		return boolNativeToBoolObject(ltVal != rtVal)
	default:
		return createError("unknown operator: %s %s %s", lt.Type(), operator, rt.Type())
	}
}

func evalStringInfixExpression(operator string, lt, rt object.Object) object.Object {
	ltVal := lt.(*object.String).Value
	rtVal := rt.(*object.String).Value

	switch operator {
	case "==":
		return boolNativeToBoolObject(ltVal == rtVal)
	case "+":
		return &object.String{Value: ltVal + rtVal}
	case "!=":
		return boolNativeToBoolObject(ltVal != rtVal)
	default:
		return createError("unknown operator: %s %s %s", lt.Type(), operator, rt.Type())
	}
}

func evalConditionalExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Evaluate(ie.Condition, env)
	if isError(condition) {
		return condition
	}
	if isTruthy(condition) {
		return Evaluate(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Evaluate(ie.Alternative, env)
	} else {
		return NULL
	}
}

func evalPrefixNegationExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return createError("unknown operator: -%s", right.Type())
	}
	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case FALSE:
		return TRUE
	case TRUE:
		return FALSE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func isTruthy(ob object.Object) bool {
	switch ob {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func boolNativeToBoolObject(value bool) *object.Boolean {
	if value {
		return TRUE
	} else {
		return FALSE
	}
}

func createError(format string, args ...any) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, args...)}
}

func isError(ob object.Object) bool {
	if ob != nil {
		return ob.Type() == object.ERROR_OBJ
	}
	return false
}

func applyFunction(fun object.Object, args []object.Object) object.Object {
	switch fn := fun.(type) {
	case *object.Function:
		evalOb := Evaluate(fn.Body, extendFunctionEnv(fn, args))
		return unwrapReturnValue(evalOb)
	case *object.BuiltIn:
		return fn.Func(args...)
	default:
		return createError("unknown function: %s", fn.Type())
	}
}

func unwrapReturnValue(ob object.Object) object.Object {
	if returnValue, ok := ob.(*object.Return); ok {
		return returnValue.Value
	}
	return ob
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for pIdx, param := range fn.Parameters {
		env.Set(param.Value, args[pIdx]) // binds args to param names with the help of param-index
	}
	return env
}
