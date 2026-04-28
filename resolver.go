package main

type resolver struct {
	interpreter     *interpreter
	scopes          []map[string]bool
	currentFunction functionType
}

type functionType int

const (
	functionNone functionType = iota
	functionFunction
)

func (r *resolver) resolveLocal(e expr, name token) {
	for i := len(r.scopes) - 1; i >= 0; i-- {
		if _, ok := r.scopes[i][name.lexeme]; ok {
			distance := len(r.scopes) - 1 - i
			r.interpreter.resolve(e, distance)
			return
		}
	}
}

func (r *resolver) resolveStatements(statements []stmt) {
	for _, statement := range statements {
		r.resolveStmt(statement)
	}
}

func (r *resolver) resolveStmt(s stmt) {
	switch v := s.(type) {
	case *varStmt:
		r.declare(v.name)
		if v.initializer != nil {
			r.resolveExpr(v.initializer)
		}
		r.define(v.name)

	case *functionStmt:
		r.declare(v.name)
		r.define(v.name)
		r.resolveFunction(v, functionFunction)

	case *blockStmt:
		r.beginScope()
		r.resolveStatements(v.statements)
		r.endScope()
	case *exprStmt:
		r.resolveExpr(v.expr)

	case *printStmt:
		r.resolveExpr(v.expr)
	case *ifStmt:
		r.resolveExpr(v.condition)
		r.resolveStmt(v.thenBranch)
		if v.elseBranch != nil {
			r.resolveStmt(v.elseBranch)
		}

	case *whileStmt:
		r.resolveExpr(v.condition)
		r.resolveStmt(v.body)

	case *returnStmt:
		if r.currentFunction == functionNone {
			reportTokenError(v.keyword, "can't return from top-level code")
		}

		if v.value != nil {
			r.resolveExpr(v.value)
		}

	}

}

func (r *resolver) resolveExpr(e expr) {
	switch v := e.(type) {
	case *variable:
		if len(r.scopes) > 0 {
			if defined, ok := r.currentScope()[v.name.lexeme]; ok && !defined {
				reportTokenError(v.name, "can't read local variable in its own initializer")
			}
		}
		r.resolveLocal(v, v.name)

	case *assign:
		r.resolveExpr(v.value)
		r.resolveLocal(v, v.name)

	case *binary:
		r.resolveExpr(v.left)
		r.resolveExpr(v.right)

	case *call:
		r.resolveExpr(v.callee)
		for _, arg := range v.arguments {
			r.resolveExpr(arg)
		}

	case *grouping:
		r.resolveExpr(v.expression)

	case *logical:
		r.resolveExpr(v.left)
		r.resolveExpr(v.right)

	case *unary:
		r.resolveExpr(v.right)

	case *literal:
		// nothing to resolve
	}

}

func (r *resolver) resolveFunction(f *functionStmt, typ functionType) {
	enclosingFunction := r.currentFunction
	r.currentFunction = typ

	r.beginScope()
	for _, param := range f.params {
		r.declare(param)
		r.define(param)
	}
	r.resolveStatements(f.body)
	r.endScope()

	r.currentFunction = enclosingFunction
}

func (r *resolver) beginScope() {
	r.scopes = append(r.scopes, make(map[string]bool))

}

func (r *resolver) endScope() {
	r.scopes = r.scopes[:len(r.scopes)-1]
}

func (r *resolver) currentScope() map[string]bool {
	return r.scopes[len(r.scopes)-1]
}

func (r *resolver) declare(name token) {
	if len(r.scopes) == 0 {
		return
	}
	// duplicate check only happens inside local scopes

	scope := r.currentScope()
	if _, exists := scope[name.lexeme]; exists {
		reportTokenError(name, "already a variable with this name in this scope")
	}

	scope[name.lexeme] = false
}

func (r *resolver) define(name token) {
	if len(r.scopes) == 0 {
		return
	}
	r.currentScope()[name.lexeme] = true
}
