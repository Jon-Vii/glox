package main

type resolver struct {
	interpreter     *interpreter
	scopes          []map[string]bool
	currentFunction functionType
	currentClass    classType
}

type functionType int

const (
	functionNone functionType = iota
	functionFunction
	functionInitializer
	functionMethod
)

type classType int

const (
	classNone classType = iota
	classClass
	classSubclass
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
		} else if r.currentFunction == functionInitializer {
			reportTokenError(v.keyword, "can't return a value from an initializer")
		}

		if v.value != nil {
			r.resolveExpr(v.value)
		}

	case *classStmt:
		enclosingClass := r.currentClass
		r.currentClass = classClass

		r.declare(v.name)
		r.define(v.name)

		if v.superclass != nil && v.name.lexeme == v.superclass.name.lexeme {
			reportTokenError(v.superclass.name, "a class can't inherit from itself")
		}

		if v.superclass != nil {
			r.currentClass = classSubclass
			r.resolveExpr(v.superclass)
		}

		if v.superclass != nil {
			r.beginScope()
			r.currentScope()["super"] = true
		}

		r.beginScope()
		r.currentScope()["this"] = true

		for _, method := range v.methods {
			declaration := functionMethod
			if method.name.lexeme == "init" {
				declaration = functionInitializer
			}

			r.resolveFunction(method, declaration)
		}

		r.endScope()

		if v.superclass != nil {
			r.endScope()
		}

		r.currentClass = enclosingClass
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

	case *get:
		r.resolveExpr(v.object)

	case *set:
		r.resolveExpr(v.value)
		r.resolveExpr(v.object)

	case *super:
		if r.currentClass == classNone {
			reportTokenError(v.keyword, "can't use 'super' outside of a class")

		} else if r.currentClass != classSubclass {
			reportTokenError(v.keyword, "can't use 'super in a class with no superclass")
		}

		r.resolveLocal(v, v.keyword)

	case *this:
		if r.currentClass == classNone {
			reportTokenError(v.keyword, "can't use 'this' outside of a class")
		}
		r.resolveLocal(v, v.keyword)
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
