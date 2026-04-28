package main

import "time"

type nativeClock struct{}

func (nativeClock) arity() int {
	return 0
}

func (nativeClock) call(i *interpreter, arguments []any) (any, error) {
	return float64(time.Now().UnixMilli()) / 1000.0, nil
}
