package main

type callable interface {
	arity() int
	call(*interpreter, []any) (any, error)
}
