package main

import (
	"fmt"
)

type environment struct {
	values    map[string]any
	enclosing *environment
}

func newEnvironment() *environment {
	return &environment{
		values: make(map[string]any),
	}
}

func newEnclosedEnvironment(enclosing *environment) *environment {
	return &environment{
		values:    make(map[string]any),
		enclosing: enclosing,
	}
}

func (e *environment) define(name string, value any) {
	e.values[name] = value
}

func (e *environment) get(name token) (any, error) {

	value, ok := e.values[name.lexeme]
	if ok {
		return value, nil
	}

	if e.enclosing != nil {
		return e.enclosing.get(name)
	}

	return nil, fmt.Errorf("undefined variable '%s'", name.lexeme)
}

func (e *environment) assign(name token, value any) error {
	if _, ok := e.values[name.lexeme]; ok {
		e.values[name.lexeme] = value
		return nil
	}

	if e.enclosing != nil {
		return e.enclosing.assign(name, value)
	}

	return fmt.Errorf("%s", "undefined variable '"+name.lexeme+"'")
}

func (e *environment) getAt(distance int, name string) any {
	return e.ancestor(distance).values[name]
}

func (e *environment) assignAt(distance int, name token, value any) {
	e.ancestor(distance).values[name.lexeme] = value
}

func (e *environment) ancestor(distance int) *environment {
	env := e
	for range distance {
		env = env.enclosing
	}
	return env
}
