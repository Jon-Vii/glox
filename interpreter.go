package main

import (
	"fmt"
	"strconv"
)

type interpreter struct {
	globals *environment
	env     *environment
	locals  map[expr]int
}

func newInterpreter() *interpreter {
	globals := newEnvironment()
	globals.define("clock", nativeClock{})

	return &interpreter{
		globals: globals,
		env:     globals,
		locals:  make(map[expr]int),
	}
}

func (i *interpreter) resolve(e expr, depth int) {
	i.locals[e] = depth
}

func (i *interpreter) interpret(statements []stmt) error {
	for _, stmt := range statements {
		if err := i.execute(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (i *interpreter) lookUpVariable(name token, e expr) (any, error) {
	if distance, ok := i.locals[e]; ok {
		return i.env.getAt(distance, name.lexeme), nil
	}
	return i.globals.get(name)
}

func (i *interpreter) execute(stmt stmt) error {
	switch v := stmt.(type) {
	case *printStmt:
		value, err := i.evaluate(v.expr)
		if err != nil {
			return err
		}
		fmt.Println(stringify(value))
		return nil

	case *exprStmt:
		_, err := i.evaluate(v.expr)
		return err

	case *varStmt:
		var value any = nil
		if v.initializer != nil {
			var err error
			value, err = i.evaluate(v.initializer)
			if err != nil {
				return err
			}
		}

		i.env.define(v.name.lexeme, value)
		return nil

	case *blockStmt:
		return i.executeBlock(v.statements, newEnclosedEnvironment(i.env))

	case *classStmt:
		i.env.define(v.name.lexeme, nil)
		methods := make(map[string]*functionValue)
		for _, method := range v.methods {
			function := &functionValue{
				declaration:   method,
				closure:       i.env,
				isInitializer: method.name.lexeme == "init",
			}
			methods[method.name.lexeme] = function
		}
		class := &classValue{
			name:    v.name.lexeme,
			methods: methods,
		}
		if err := i.env.assign(v.name, class); err != nil {
			return err
		}

		return nil

	case *ifStmt:
		value, err := i.evaluate(v.condition)
		if err != nil {
			return err
		}

		if isTruthy(value) {
			return i.execute(v.thenBranch)

		}

		if v.elseBranch != nil {
			return i.execute(v.elseBranch)
		}
		return nil

	case *whileStmt:
		for {
			value, err := i.evaluate(v.condition)
			if err != nil {
				return err
			}

			if !isTruthy(value) {
				break
			}
			if err := i.execute(v.body); err != nil {
				return err
			}
		}

		return nil

	case *functionStmt:
		fn := &functionValue{
			declaration:   v,
			closure:       i.env,
			isInitializer: false,
		}
		i.env.define(v.name.lexeme, fn)

		return nil

	case *returnStmt:
		var value any
		if v.value != nil {
			result, err := i.evaluate(v.value)
			if err != nil {
				return err
			}
			value = result
		}

		return &returnValue{
			value: value,
		}

	default:
		return fmt.Errorf("unknown statement type")
	}

}

func (i *interpreter) executeBlock(statements []stmt, env *environment) error {
	previous := i.env
	i.env = env

	defer func() {
		i.env = previous
	}()

	for _, stmt := range statements {
		if err := i.execute(stmt); err != nil {
			return err
		}
	}

	return nil
}

func (i *interpreter) evaluate(e expr) (any, error) {
	switch v := e.(type) {
	case *variable:
		return i.lookUpVariable(v.name, v)

	case *assign:
		value, err := i.evaluate(v.value)
		if err != nil {
			return nil, err
		}

		if distance, ok := i.locals[v]; ok {
			i.env.assignAt(distance, v.name, value)
		} else if err := i.globals.assign(v.name, value); err != nil {
			return nil, err
		}

		return value, nil

	case *call:
		callee, err := i.evaluate(v.callee)
		if err != nil {
			return nil, err
		}

		var arguments []any
		for _, argument := range v.arguments {
			value, err := i.evaluate(argument)
			if err != nil {
				return nil, err
			}
			arguments = append(arguments, value)
		}

		function, ok := callee.(callable)
		if !ok {
			return nil, fmt.Errorf("can only call functions and classes")
		}

		arity := function.arity()
		if len(arguments) != arity {
			return nil, fmt.Errorf("expected %d arguments but got %d", arity, len(arguments))
		}

		return function.call(i, arguments)

	case *get:
		object, err := i.evaluate(v.object)
		if err != nil {
			return nil, err
		}

		instance, ok := object.(*instanceValue)
		if !ok {
			return nil, fmt.Errorf("only instances have properties")
		}

		return instance.get(v.name)

	case *set:
		object, err := i.evaluate(v.object)
		if err != nil {
			return nil, err
		}

		instance, ok := object.(*instanceValue)
		if !ok {
			return nil, fmt.Errorf("only instances have fields")
		}

		value, err := i.evaluate(v.value)
		if err != nil {
			return nil, err
		}

		instance.set(v.name, value)
		return value, nil

	case *this:
		value, err := i.lookUpVariable(v.keyword, e)
		if err != nil {
			return nil, err
		}
		return value, nil

	case *logical:
		left, err := i.evaluate(v.left)
		if err != nil {
			return nil, err
		}

		if v.operator.typ == Or {
			if isTruthy(left) {
				return left, nil
			}
		} else {
			if !isTruthy(left) {
				return left, nil
			}
		}
		return i.evaluate(v.right)

	case *literal:
		return v.value, nil

	case *grouping:
		return i.evaluate(v.expression)

	case *unary:
		right, err := i.evaluate(v.right)
		if err != nil {
			return nil, err
		}

		switch v.operator.typ {
		case Bang:
			return !isTruthy(right), nil
		case Minus:
			num, ok := right.(float64)
			if !ok {
				return nil, fmt.Errorf("operand must be a number")
			}
			return -num, nil
		default:
			return nil, fmt.Errorf("unknown unary operator")
		}

	case *binary:
		left, err := i.evaluate(v.left)
		if err != nil {
			return nil, err
		}

		right, err := i.evaluate(v.right)
		if err != nil {
			return nil, err
		}

		switch v.operator.typ {
		case Greater:
			leftNum, rightNum, ok := bothNumbers(left, right)
			if !ok {
				return nil, fmt.Errorf("operands must be numbers")
			}
			return leftNum > rightNum, nil

		case GreaterEqual:
			leftNum, rightNum, ok := bothNumbers(left, right)
			if !ok {
				return nil, fmt.Errorf("operands must be numbers")
			}
			return leftNum >= rightNum, nil

		case Less:
			leftNum, rightNum, ok := bothNumbers(left, right)
			if !ok {
				return nil, fmt.Errorf("operands must be numbers")
			}
			return leftNum < rightNum, nil

		case LessEqual:
			leftNum, rightNum, ok := bothNumbers(left, right)
			if !ok {
				return nil, fmt.Errorf("operands must be numbers")
			}
			return leftNum <= rightNum, nil

		case Minus:
			leftNum, rightNum, ok := bothNumbers(left, right)
			if !ok {
				return nil, fmt.Errorf("operands must be numbers")
			}
			return leftNum - rightNum, nil

		case Slash:
			leftNum, rightNum, ok := bothNumbers(left, right)
			if !ok {
				return nil, fmt.Errorf("operands must be numbers")
			}
			return leftNum / rightNum, nil

		case Star:
			leftNum, rightNum, ok := bothNumbers(left, right)
			if !ok {
				return nil, fmt.Errorf("operands must be numbers")
			}
			return leftNum * rightNum, nil

		case Plus:
			if leftNum, rightNum, ok := bothNumbers(left, right); ok {
				return leftNum + rightNum, nil
			}

			if leftStr, rightStr, ok := bothStrings(left, right); ok {
				return leftStr + rightStr, nil
			}

			return nil, fmt.Errorf("operands must be two numbers or two strings")

		case BangEqual:
			return !isEqual(left, right), nil

		case EqualEqual:
			return isEqual(left, right), nil

		default:
			return nil, fmt.Errorf("unknown binary operator")
		}

	default:
		return nil, fmt.Errorf("unknown expression type")
	}
}

func bothNumbers(left any, right any) (float64, float64, bool) {
	leftNum, lok := left.(float64)
	rightNum, rok := right.(float64)
	return leftNum, rightNum, lok && rok
}

func bothStrings(left any, right any) (string, string, bool) {
	leftStr, lok := left.(string)
	rightStr, rok := right.(string)
	return leftStr, rightStr, lok && rok
}

func isTruthy(v any) bool {
	if v == nil {
		return false
	}

	if b, ok := v.(bool); ok {
		return b
	}

	return true
}

func isEqual(a any, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil {
		return false
	}
	return a == b
}

func stringify(v any) string {
	if v == nil {
		return "nil"
	}
	switch x := v.(type) {
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case bool:
		if x {
			return "true"
		}
		return "false"
	case string:
		return x
	default:
		return fmt.Sprint(x)
	}
}
