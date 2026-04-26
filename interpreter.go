package main

import (
	"fmt"
	"strconv"
)

type interpreter struct {
	env *environment
}

func (i *interpreter) interpret(statements []stmt) error {
	for _, stmt := range statements {
		if err := i.execute(stmt); err != nil {
			return err
		}
	}
	return nil
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
		return i.env.get(v.name)
	case *assign:
		value, err := i.evaluate(v.value)
		if err != nil {
			return nil, err
		}

		if err := i.env.assign(v.name, value); err != nil {
			return nil, err
		}

		return value, nil
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
