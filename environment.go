package main

import (
	"fmt"
)

type environment struct {
	values map[string]any
}

func newEnvironment() *environment {
	return &environment{
		values: make(map[string]any),
	}
}

func (e *environment) define(name string, value any) {
	e.values[name] = value
}

func (e *environment) get(name token) (any, error) {
	value, ok := e.values[name.lexeme]
	if !ok {
		return nil, fmt.Errorf("undefined variable '%s'", name.lexeme)
	}
	return value, nil
}
