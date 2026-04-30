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

type ifStmt struct {
	condition  expr
	thenBranch stmt
	elseBranch stmt
}

func (i *ifStmt) isStmt() {}

type whileStmt struct {
	condition expr
	body      stmt
}

func (w *whileStmt) isStmt() {}

type functionStmt struct {
	name   token
	params []token
	body   []stmt
}

func (f *functionStmt) isStmt() {}

type returnStmt struct {
	keyword token
	value   expr
}

func (r *returnStmt) isStmt() {}

type classStmt struct {
	name    token
	methods []*functionStmt
}

func (c *classStmt) isStmt() {}
