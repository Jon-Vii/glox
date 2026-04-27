package main

type expr interface{ isExpr() }

type binary struct {
	left     expr
	operator token
	right    expr
}

func (b *binary) isExpr() {}

type grouping struct {
	expression expr
}

func (g *grouping) isExpr() {}

type literal struct {
	value any
}

func (l *literal) isExpr() {}

type unary struct {
	operator token
	right    expr
}

func (u *unary) isExpr() {}

type variable struct {
	name token
}

func (v *variable) isExpr() {}

type assign struct {
	name  token
	value expr
}

func (a *assign) isExpr() {}

type logical struct {
	left     expr
	operator token
	right    expr
}

func (l *logical) isExpr() {}
