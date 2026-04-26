package main

type stmt interface{ isStmt() }

type exprStmt struct {
	expr expr
}

func (e *exprStmt) isStmt() {}

type printStmt struct {
	expr expr
}

func (e *printStmt) isStmt() {}

type varStmt struct {
	name        token
	initializer expr
}

func (v *varStmt) isStmt() {}

type blockStmt struct {
	statements []stmt
}

func (b *blockStmt) isStmt() {}
