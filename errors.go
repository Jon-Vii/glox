package main

import (
	"fmt"
	"os"
)

var hadError bool

func report(line int, where string, message string) {
	fmt.Fprintf(os.Stderr, "[line %d] Error%s: %s\n", line, where, message)
	hadError = true
}

func reportLineError(line int, message string) {
	report(line, "", message)
}

func reportTokenError(token token, message string) {
	if token.typ == Eof {
		report(token.line, " at end", message)
	} else {
		report(token.line, fmt.Sprintf(" at %q", token.lexeme), message)
	}
}
