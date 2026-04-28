package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	var err error

	switch len(os.Args) {
	case 1:
		err = runPrompt()
	case 2:
		err = runFile(os.Args[1])
	default:
		err = fmt.Errorf("usage: %s [script]", os.Args[0])
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(source string, intr *interpreter) {
	hadError = false

	scanner := scanner{
		source: source,
		line:   1,
	}

	scanner.scanTokens()

	if hadError {
		return
	}

	parser := parser{
		current: 0,
		tokens:  scanner.tokens,
	}

	statements := parser.parse()
	if hadError {
		return
	}

	if err := intr.interpret(statements); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	/*
		fmt.Printf("%T %#v\n", expr, expr)*/

}

func runFile(path string) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	intr := newInterpreter()
	run(string(bytes), intr)

	if hadError {
		return fmt.Errorf("source had syntax error(s)")
	}

	return nil
}

func runPrompt() error {
	input := bufio.NewScanner(os.Stdin)
	intr := newInterpreter()

	for {
		fmt.Print("> ")

		if !input.Scan() {
			fmt.Println()
			return input.Err()
		}

		run(input.Text(), intr)
	}
}
