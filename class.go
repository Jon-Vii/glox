package main

import "fmt"

type classValue struct {
	name       string
	superclass *classValue
	methods    map[string]*functionValue
}

func (c *classValue) arity() int {
	initializer, ok := c.findMethod("init")
	if !ok {
		return 0
	}
	return initializer.arity()
}

func (c *classValue) call(i *interpreter, args []any) (any, error) {
	instance := &instanceValue{
		class:  c,
		fields: make(map[string]any),
	}

	initializer, ok := c.findMethod("init")
	if ok {
		_, err := initializer.bind(instance).call(i, args)
		if err != nil {
			return nil, err
		}
	}

	return instance, nil
}

func (c *classValue) findMethod(name string) (*functionValue, bool) {
	method, ok := c.methods[name]
	if ok {
		return method, true
	}

	if c.superclass != nil {
		return c.superclass.findMethod(name)
	}

	return nil, false

}

func (c *classValue) String() string {
	return c.name
}

type instanceValue struct {
	class  *classValue
	fields map[string]any
}

func (i *instanceValue) get(name token) (any, error) {
	if value, ok := i.fields[name.lexeme]; ok {
		return value, nil
	}
	if method, ok := i.class.findMethod(name.lexeme); ok {
		return method.bind(i), nil
	}

	return nil, fmt.Errorf("undefined property '%s'", name.lexeme)
}

func (i *instanceValue) set(name token, value any) {
	i.fields[name.lexeme] = value
}

func (i *instanceValue) String() string {
	return i.class.name + " instance"
}
