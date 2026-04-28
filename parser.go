package main

type parser struct {
	current int
	tokens  []token
}

func (p *parser) parse() []stmt {
	var statements []stmt
	for !p.isAtEnd() {
		stmt := p.declaration()

		if stmt == nil {
			p.synchronize()
			continue
		}

		statements = append(statements, stmt)

	}

	return statements

}

func (p *parser) declaration() stmt {
	if p.match(Var) {
		return p.varDeclaration()
	}
	if p.match(Fun) {
		return p.function("function")
	}
	return p.statement()

}

func (p *parser) varDeclaration() stmt {
	name, ok := p.consume(Identifier, "expect variable name")
	if !ok {
		return nil
	}

	var initializer expr = nil
	if p.match(Equal) {
		initializer = p.expression()
		if initializer == nil {
			return nil
		}
	}

	if _, ok := p.consume(Semicolon, "expect ';' after variable declaration"); !ok {
		return nil
	}
	return &varStmt{
		name:        name,
		initializer: initializer,
	}
}

func (p *parser) function(kind string) stmt {
	name, ok := p.consume(Identifier, "expect "+kind+" name")
	if !ok {
		return nil
	}

	if _, ok := p.consume(LeftParen, "expect '(' after "+kind+" name"); !ok {
		return nil
	}

	var parameters []token
	if !p.check(RightParen) {
		for {
			if len(parameters) >= 255 {
				p.parseError(p.peek(), "can't have more than 255 parameters")
			}
			parameter, ok := p.consume(Identifier, "expect parameter name")
			if !ok {
				return nil
			}
			parameters = append(parameters, parameter)

			if !p.match(Comma) {
				break
			}
		}
	}

	if _, ok := p.consume(RightParen, "expect ')' after parameters"); !ok {
		return nil
	}
	if _, ok := p.consume(LeftBrace, "expect '{' before "+kind+" body"); !ok {
		return nil
	}

	body := p.blockStatement()
	if body == nil {
		return nil
	}

	return &functionStmt{
		name:   name,
		params: parameters,
		body:   body,
	}
}

func (p *parser) statement() stmt {
	if p.match(Print) {
		return p.printStatement()
	}
	if p.match(While) {
		return p.whileStatement()
	}
	if p.match(For) {
		return p.forStatement()
	}

	if p.match(LeftBrace) {
		statements := p.blockStatement()
		if statements == nil {
			return nil
		}
		return &blockStmt{statements: statements}
	}
	if p.match(If) {
		return p.ifStatement()
	}
	if p.match(Return) {
		return p.returnStatement()
	}

	return p.expressionStatement()
}

func (p *parser) returnStatement() stmt {
	keyword := p.previous()

	var value expr
	if !p.check(Semicolon) {
		value = p.expression()
		if value == nil {
			return nil
		}
	}

	if _, ok := p.consume(Semicolon, "expect ';' after return value"); !ok {
		return nil
	}

	return &returnStmt{
		keyword: keyword,
		value:   value,
	}
}

func (p *parser) forStatement() stmt {
	if _, ok := p.consume(LeftParen, "expect '(' after 'for'"); !ok {
		return nil
	}

	var initializer stmt
	if p.match(Semicolon) {
		// no initializer
	} else if p.match(Var) {
		initializer = p.varDeclaration()
		if initializer == nil {
			return nil
		}

	} else {
		initializer = p.expressionStatement()
		if initializer == nil {
			return nil
		}
	}

	var condition expr
	if !p.check(Semicolon) {
		condition = p.expression()
		if condition == nil {
			return nil
		}
	}
	if _, ok := p.consume(Semicolon, "expect ';' after loop condition"); !ok {
		return nil
	}

	var increment expr
	if !p.check(RightParen) {
		increment = p.expression()
		if increment == nil {
			return nil
		}
	}
	if _, ok := p.consume(RightParen, "expect ')' after for clauses"); !ok {
		return nil
	}

	body := p.statement()
	if body == nil {
		return nil
	}

	if increment != nil {
		body = &blockStmt{
			statements: []stmt{body,
				&exprStmt{
					expr: increment}}}
	}

	if condition == nil {
		condition = &literal{value: true}

	}

	body = &whileStmt{
		condition: condition,
		body:      body}

	if initializer != nil {
		body = &blockStmt{
			statements: []stmt{initializer, body},
		}
	}

	return body
}

func (p *parser) whileStatement() stmt {
	if _, ok := p.consume(LeftParen, "expect '(' after 'while'"); !ok {
		return nil
	}
	condition := p.expression()
	if condition == nil {
		return nil
	}
	if _, ok := p.consume(RightParen, "expect ')' after condition"); !ok {
		return nil
	}
	body := p.statement()
	if body == nil {
		return nil
	}
	return &whileStmt{
		condition: condition,
		body:      body}
}

func (p *parser) ifStatement() stmt {
	if _, ok := p.consume(LeftParen, "expect '(' after 'if'"); !ok {
		return nil
	}
	condition := p.expression()
	if condition == nil {
		return nil
	}
	if _, ok := p.consume(RightParen, "expect ')' after if condition"); !ok {
		return nil
	}

	thenBranch := p.statement()
	if thenBranch == nil {
		return nil
	}
	var elseBranch stmt

	if p.match(Else) {
		elseBranch = p.statement()
		if elseBranch == nil {
			return nil
		}
	}

	return &ifStmt{
		condition:  condition,
		thenBranch: thenBranch,
		elseBranch: elseBranch}
}

func (p *parser) printStatement() stmt {
	value := p.expression()
	if value == nil {
		return nil
	}

	_, ok := p.consume(Semicolon, "expect ';' after value")
	if !ok {
		return nil
	}

	return &printStmt{
		expr: value}
}

func (p *parser) blockStatement() []stmt {
	var statements []stmt

	for !p.check(RightBrace) && !p.isAtEnd() {
		stmt := p.declaration()
		if stmt == nil {
			p.synchronize()
			continue
		}

		statements = append(statements, stmt)
	}

	if _, ok := p.consume(RightBrace, "Expect '}' after block"); !ok {
		return nil
	}
	return statements
}

func (p *parser) expressionStatement() stmt {
	expr := p.expression()
	if expr == nil {
		return nil
	}
	_, ok := p.consume(Semicolon, "expect ';' after expression")
	if !ok {
		return nil
	}
	return &exprStmt{expr}
}

func (p *parser) expression() expr {
	return p.assignment()
}

func (p *parser) assignment() expr {
	expr := p.or()

	if p.match(Equal) {
		equals := p.previous()
		value := p.assignment()

		if value == nil {
			return nil
		}

		if v, ok := expr.(*variable); ok {
			name := v.name
			return &assign{name: name, value: value}
		}

		p.parseError(equals, "invalid assignment target")
		return nil
	}
	return expr
}

func (p *parser) or() expr {
	expr := p.and()
	if expr == nil {
		return nil
	}
	for p.match(Or) {
		operator := p.previous()
		right := p.and()
		if right == nil {
			return nil
		}

		expr = &logical{
			left:     expr,
			operator: operator,
			right:    right}
	}

	return expr
}

func (p *parser) and() expr {
	expr := p.equality()
	if expr == nil {
		return nil
	}
	for p.match(And) {
		operator := p.previous()
		right := p.equality()
		if right == nil {
			return nil
		}
		expr = &logical{left: expr, operator: operator, right: right}
	}
	return expr
}

func (p *parser) equality() expr {
	expr := p.comparison()
	if expr == nil {
		return nil
	}

	for p.match(BangEqual, EqualEqual) {
		operator := p.previous()
		right := p.comparison()
		if right == nil {
			return nil
		}
		expr = &binary{
			left:     expr,
			operator: operator,
			right:    right,
		}
	}
	return expr
}

func (p *parser) comparison() expr {
	expr := p.term()
	if expr == nil {
		return nil
	}
	for p.match(Greater, GreaterEqual, Less, LessEqual) {
		operator := p.previous()
		right := p.term()
		if right == nil {
			return nil
		}
		expr = &binary{
			left:     expr,
			operator: operator,
			right:    right,
		}
	}
	return expr
}

func (p *parser) term() expr {
	expr := p.factor()
	if expr == nil {
		return nil
	}
	for p.match(Minus, Plus) {
		operator := p.previous()
		right := p.factor()
		if right == nil {
			return nil
		}
		expr = &binary{
			left:     expr,
			operator: operator,
			right:    right,
		}
	}
	return expr
}

func (p *parser) factor() expr {
	expr := p.unary()
	if expr == nil {
		return nil
	}
	for p.match(Slash, Star) {
		operator := p.previous()
		right := p.unary()
		if right == nil {
			return nil
		}
		expr = &binary{
			left:     expr,
			operator: operator,
			right:    right,
		}
	}
	return expr
}

func (p *parser) unary() expr {
	if p.match(Bang, Minus) {
		operator := p.previous()
		right := p.unary()
		if right == nil {
			return nil
		}
		return &unary{
			operator: operator,
			right:    right,
		}
	}
	return p.call()
}

func (p *parser) call() expr {
	expr := p.primary()

	// parse zero or more function calls e.g getCallback()()
	for p.match(LeftParen) {
		expr = p.finishCall(expr)
	}

	return expr
}

func (p *parser) finishCall(callee expr) expr {
	var arguments []expr

	// parse comma-separated args
	if !p.check(RightParen) {
		for {
			if len(arguments) >= 255 {
				p.parseError(p.peek(), "can't have more than 255 arguments")
			}
			expr := p.expression()
			if expr == nil {
				return nil
			}

			arguments = append(arguments, expr)

			if !p.match(Comma) {
				break
			}
		}
	}
	paren, ok := p.consume(RightParen, "Expect ')' after arguments")
	if !ok {
		return nil
	}

	return &call{
		callee:    callee,
		paren:     paren,
		arguments: arguments,
	}
}

func (p *parser) primary() expr {
	if p.match(False) {
		return &literal{value: false}
	}
	if p.match(True) {
		return &literal{value: true}
	}
	if p.match(Nil) {
		return &literal{value: nil}
	}

	if p.match(Number, String) {
		return &literal{value: p.previous().literal}
	}

	if p.match(Identifier) {
		return &variable{name: p.previous()}
	}

	if p.match(LeftParen) {
		expr := p.expression()
		if expr == nil {
			return nil
		}

		if _, ok := p.consume(RightParen, "expect ')' after expression"); !ok {
			return nil
		}
		return &grouping{expr}
	}

	p.parseError(p.peek(), "expect expression")
	return nil
}

func (p *parser) match(types ...tokenType) bool {
	for _, typ := range types {
		if p.check(typ) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *parser) consume(typ tokenType, message string) (token, bool) {
	if p.check(typ) {
		return p.advance(), true
	}

	p.parseError(p.peek(), message)
	return token{}, false // consume failed
}

func (p *parser) check(typ tokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().typ == typ
}

func (p *parser) advance() token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *parser) isAtEnd() bool {
	return p.peek().typ == Eof
}

func (p *parser) peek() token {
	return p.tokens[p.current]
}

func (p *parser) previous() token {
	return p.tokens[p.current-1]
}

func (p *parser) parseError(tok token, message string) {
	reportTokenError(tok, message)
}

func (p *parser) synchronize() {
	p.advance()
	for !p.isAtEnd() {
		if p.previous().typ == Semicolon {
			return
		}

		switch p.peek().typ {
		case Class, For, Fun, If, Print, Return, Var, While:
			return
		}

		p.advance()
	}
}
