package main

type parser struct {
	current int
	tokens  []token
}

func (p *parser) parse() expr {
	expr := p.expression()

	if expr == nil {
		return nil
	}
	if !p.isAtEnd() {
		p.parseError(p.peek(), "expect end of expression")
		return nil
	}

	return expr
}

func (p *parser) expression() expr {
	return p.equality()
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
	return p.primary()
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
