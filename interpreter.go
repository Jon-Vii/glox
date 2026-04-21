package main

import (
	"fmt"
	"strconv"
)

type interpreter struct{}

func (i *interpreter) interpret(e expr) error {
	value, err := i.evaluate(e)
	if err != nil {
		return err
	}
	fmt.Println(stringify(value))
	return nil

}

func (i *interpreter) evaluate(e expr) (any, error) {
	switch v := e.(type) {
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
