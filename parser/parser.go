package parser

import (
	"fmt"
	"strconv"

	"evanescence/lexer"
)

type Parser struct {
	tokens []lexer.Token
	pos    int
}

func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

func (p *Parser) Parse() (*Program, error) {
	prog := &Program{}
	for !p.atEnd() {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		prog.Statements = append(prog.Statements, stmt)
	}
	return prog, nil
}

// ----- Statements -----

func (p *Parser) parseStatement() (Statement, error) {
	tok := p.peek()
	switch tok.Type {
	case lexer.TOKEN_VAR:
		return p.parseVarDecl(false)
	case lexer.TOKEN_CONST:
		return p.parseVarDecl(true)
	case lexer.TOKEN_IF:
		return p.parseIf()
	case lexer.TOKEN_WHILE:
		return p.parseWhile()
	case lexer.TOKEN_FOR:
		return p.parseFor()
	case lexer.TOKEN_FUNC:
		return p.parseFunc()
	case lexer.TOKEN_RETURN:
		return p.parseReturn()
	case lexer.TOKEN_BREAK:
		p.advance()
		if err := p.expect(lexer.TOKEN_SEMICOLON, "';' after break"); err != nil {
			return nil, err
		}
		return &BreakStmt{}, nil
	case lexer.TOKEN_CONTINUE:
		p.advance()
		if err := p.expect(lexer.TOKEN_SEMICOLON, "';' after continue"); err != nil {
			return nil, err
		}
		return &ContinueStmt{}, nil
	case lexer.TOKEN_LBRACE:
		return p.parseBlock()
	}

	if tok.Type == lexer.TOKEN_IDENT && p.pos+1 < len(p.tokens) {
		next := p.tokens[p.pos+1].Type
		if next == lexer.TOKEN_ASSIGN || next == lexer.TOKEN_PLUSEQ ||
			next == lexer.TOKEN_MINUSEQ || next == lexer.TOKEN_STAREQ ||
			next == lexer.TOKEN_SLASHEQ {
			return p.parseAssign()
		}
	}
	return p.parseExprStmt()
}

func (p *Parser) parseVarDecl(isConst bool) (Statement, error) {
	p.advance() // var | const
	name, err := p.expectIdent()
	if err != nil {
		return nil, err
	}
	if err := p.expect(lexer.TOKEN_ASSIGN, "'=' in declaration"); err != nil {
		return nil, err
	}
	val, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.expect(lexer.TOKEN_SEMICOLON, "';' after declaration"); err != nil {
		return nil, err
	}
	return &VarDecl{Name: name, Value: val, IsConst: isConst}, nil
}

func (p *Parser) parseAssign() (Statement, error) {
	name := p.advance().Lexeme
	op := p.advance().Lexeme
	val, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.expect(lexer.TOKEN_SEMICOLON, "';' after assignment"); err != nil {
		return nil, err
	}
	return &AssignStmt{Name: name, Op: op, Value: val}, nil
}

func (p *Parser) parseIf() (Statement, error) {
	p.advance() // if
	if err := p.expect(lexer.TOKEN_LPAREN, "'(' after 'if'"); err != nil {
		return nil, err
	}
	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.expect(lexer.TOKEN_RPAREN, "')' after if condition"); err != nil {
		return nil, err
	}
	then, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	stmt := &IfStmt{Cond: cond, Then: then.(*BlockStmt)}

	if p.peek().Type == lexer.TOKEN_ELSE {
		p.advance()
		if p.peek().Type == lexer.TOKEN_IF {
			elseIf, err := p.parseIf()
			if err != nil {
				return nil, err
			}
			stmt.Else = elseIf
		} else {
			block, err := p.parseBlock()
			if err != nil {
				return nil, err
			}
			stmt.Else = block
		}
	}
	return stmt, nil
}

func (p *Parser) parseWhile() (Statement, error) {
	p.advance() // while
	if err := p.expect(lexer.TOKEN_LPAREN, "'(' after 'while'"); err != nil {
		return nil, err
	}
	cond, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.expect(lexer.TOKEN_RPAREN, "')' after while condition"); err != nil {
		return nil, err
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return &WhileStmt{Cond: cond, Body: body.(*BlockStmt)}, nil
}

func (p *Parser) parseFor() (Statement, error) {
	p.advance() // for
	name, err := p.expectIdent()
	if err != nil {
		return nil, err
	}
	if err := p.expect(lexer.TOKEN_IN, "'in' in for loop"); err != nil {
		return nil, err
	}

	stmt := &ForStmt{VarName: name}
	if p.peek().Type == lexer.TOKEN_RANGE {
		p.advance() // range
		if err := p.expect(lexer.TOKEN_LPAREN, "'(' after 'range'"); err != nil {
			return nil, err
		}
		first, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		args := []Expression{first}
		for p.peek().Type == lexer.TOKEN_COMMA {
			p.advance()
			e, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			args = append(args, e)
			if len(args) > 3 {
				return nil, p.errorf("range() takes at most 3 arguments")
			}
		}
		if err := p.expect(lexer.TOKEN_RPAREN, "')' after range arguments"); err != nil {
			return nil, err
		}
		stmt.RangeArgs = args
	} else {
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		stmt.Iterable = expr
	}

	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	stmt.Body = body.(*BlockStmt)
	return stmt, nil
}

func (p *Parser) parseFunc() (Statement, error) {
	p.advance() // func
	name, err := p.expectIdent()
	if err != nil {
		return nil, err
	}
	if err := p.expect(lexer.TOKEN_LPAREN, "'(' after function name"); err != nil {
		return nil, err
	}
	var params []string
	if p.peek().Type != lexer.TOKEN_RPAREN {
		first, err := p.expectIdent()
		if err != nil {
			return nil, err
		}
		params = append(params, first)
		for p.peek().Type == lexer.TOKEN_COMMA {
			p.advance()
			next, err := p.expectIdent()
			if err != nil {
				return nil, err
			}
			params = append(params, next)
		}
	}
	if err := p.expect(lexer.TOKEN_RPAREN, "')' after parameters"); err != nil {
		return nil, err
	}
	body, err := p.parseBlock()
	if err != nil {
		return nil, err
	}
	return &FuncDecl{Name: name, Params: params, Body: body.(*BlockStmt)}, nil
}

func (p *Parser) parseReturn() (Statement, error) {
	p.advance() // return
	stmt := &ReturnStmt{}
	if p.peek().Type != lexer.TOKEN_SEMICOLON {
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		stmt.Value = expr
	}
	if err := p.expect(lexer.TOKEN_SEMICOLON, "';' after return"); err != nil {
		return nil, err
	}
	return stmt, nil
}

func (p *Parser) parseBlock() (Statement, error) {
	if err := p.expect(lexer.TOKEN_LBRACE, "'{' to start block"); err != nil {
		return nil, err
	}
	block := &BlockStmt{}
	for p.peek().Type != lexer.TOKEN_RBRACE && !p.atEnd() {
		s, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		block.Statements = append(block.Statements, s)
	}
	if err := p.expect(lexer.TOKEN_RBRACE, "'}' to close block"); err != nil {
		return nil, err
	}
	return block, nil
}

func (p *Parser) parseExprStmt() (Statement, error) {
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if err := p.expect(lexer.TOKEN_SEMICOLON, "';' after expression"); err != nil {
		return nil, err
	}
	return &ExprStmt{Expr: expr}, nil
}

func (p *Parser) parseExpression() (Expression, error) { return p.parseLogicOr() }

func (p *Parser) parseLogicOr() (Expression, error) {
	left, err := p.parseLogicAnd()
	if err != nil {
		return nil, err
	}
	for p.peek().Type == lexer.TOKEN_OR {
		op := p.advance().Lexeme
		right, err := p.parseLogicAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseLogicAnd() (Expression, error) {
	left, err := p.parseEquality()
	if err != nil {
		return nil, err
	}
	for p.peek().Type == lexer.TOKEN_AND {
		op := p.advance().Lexeme
		right, err := p.parseEquality()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseEquality() (Expression, error) {
	left, err := p.parseComparison()
	if err != nil {
		return nil, err
	}
	for p.peek().Type == lexer.TOKEN_EQ || p.peek().Type == lexer.TOKEN_NEQ {
		op := p.advance().Lexeme
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseComparison() (Expression, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for {
		t := p.peek().Type
		if t != lexer.TOKEN_LT && t != lexer.TOKEN_LTE && t != lexer.TOKEN_GT && t != lexer.TOKEN_GTE {
			break
		}
		op := p.advance().Lexeme
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseTerm() (Expression, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}
	for p.peek().Type == lexer.TOKEN_PLUS || p.peek().Type == lexer.TOKEN_MINUS {
		op := p.advance().Lexeme
		right, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseFactor() (Expression, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for {
		t := p.peek().Type
		if t != lexer.TOKEN_STAR && t != lexer.TOKEN_SLASH &&
			t != lexer.TOKEN_FLOORDIV && t != lexer.TOKEN_PERCENT {
			break
		}
		op := p.advance().Lexeme
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseUnary() (Expression, error) {
	if p.peek().Type == lexer.TOKEN_BANG || p.peek().Type == lexer.TOKEN_MINUS {
		op := p.advance().Lexeme
		operand, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryExpr{Op: op, Operand: operand}, nil
	}
	return p.parsePower()
}

func (p *Parser) parsePower() (Expression, error) {
	left, err := p.parseCall()
	if err != nil {
		return nil, err
	}
	if p.peek().Type == lexer.TOKEN_POWER {
		op := p.advance().Lexeme
		// right-associative: recurse into unary so -2**3 still parses sensibly
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &BinaryExpr{Op: op, Left: left, Right: right}, nil
	}
	return left, nil
}

func (p *Parser) parseCall() (Expression, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	for {
		switch p.peek().Type {
		case lexer.TOKEN_LPAREN:
			p.advance()
			var args []Expression
			if p.peek().Type != lexer.TOKEN_RPAREN {
				first, err := p.parseExpression()
				if err != nil {
					return nil, err
				}
				args = append(args, first)
				for p.peek().Type == lexer.TOKEN_COMMA {
					p.advance()
					a, err := p.parseExpression()
					if err != nil {
						return nil, err
					}
					args = append(args, a)
				}
			}
			if err := p.expect(lexer.TOKEN_RPAREN, "')' after arguments"); err != nil {
				return nil, err
			}
			expr = &CallExpr{Callee: expr, Args: args}
		case lexer.TOKEN_LBRACKET:
			p.advance()
			idx, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			if err := p.expect(lexer.TOKEN_RBRACKET, "']' after index"); err != nil {
				return nil, err
			}
			expr = &IndexExpr{Collection: expr, Index: idx}
		default:
			return expr, nil
		}
	}
}

func (p *Parser) parsePrimary() (Expression, error) {
	tok := p.peek()
	switch tok.Type {
	case lexer.TOKEN_NUMBER:
		p.advance()
		v, err := strconv.ParseFloat(tok.Lexeme, 64)
		if err != nil {
			return nil, p.errorf("invalid number %q", tok.Lexeme)
		}
		return &NumberLiteral{Value: v}, nil
	case lexer.TOKEN_STRING:
		p.advance()
		return &StringLiteral{Value: tok.Lexeme}, nil
	case lexer.TOKEN_TRUE:
		p.advance()
		return &BoolLiteral{Value: true}, nil
	case lexer.TOKEN_FALSE:
		p.advance()
		return &BoolLiteral{Value: false}, nil
	case lexer.TOKEN_NULL:
		p.advance()
		return &NullLiteral{}, nil
	case lexer.TOKEN_IDENT:
		p.advance()
		return &Identifier{Name: tok.Lexeme}, nil
	case lexer.TOKEN_LPAREN:
		p.advance()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if err := p.expect(lexer.TOKEN_RPAREN, "')' after expression"); err != nil {
			return nil, err
		}
		return expr, nil
	case lexer.TOKEN_LBRACKET:
		return p.parseListLiteral()
	}
	return nil, p.errorf("unexpected token %s (%q)", tok.Type, tok.Lexeme)
}

func (p *Parser) parseListLiteral() (Expression, error) {
	p.advance() // [
	list := &ListLiteral{}
	if p.peek().Type != lexer.TOKEN_RBRACKET {
		first, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		list.Elements = append(list.Elements, first)
		for p.peek().Type == lexer.TOKEN_COMMA {
			p.advance()
			e, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			list.Elements = append(list.Elements, e)
		}
	}
	if err := p.expect(lexer.TOKEN_RBRACKET, "']' to close list literal"); err != nil {
		return nil, err
	}
	return list, nil
}

// ----- Helpers -----

func (p *Parser) peek() lexer.Token { return p.tokens[p.pos] }

func (p *Parser) advance() lexer.Token {
	tok := p.tokens[p.pos]
	if !p.atEnd() {
		p.pos++
	}
	return tok
}

func (p *Parser) atEnd() bool { return p.tokens[p.pos].Type == lexer.TOKEN_EOF }

func (p *Parser) expect(tt lexer.TokenType, what string) error {
	if p.peek().Type != tt {
		return p.errorf("expected %s, got %s (%q)", what, p.peek().Type, p.peek().Lexeme)
	}
	p.advance()
	return nil
}

func (p *Parser) expectIdent() (string, error) {
	if p.peek().Type != lexer.TOKEN_IDENT {
		return "", p.errorf("expected identifier, got %s (%q)", p.peek().Type, p.peek().Lexeme)
	}
	return p.advance().Lexeme, nil
}

func (p *Parser) errorf(format string, args ...interface{}) error {
	tok := p.peek()
	return fmt.Errorf("parse error at line %d:%d: %s", tok.Line, tok.Column, fmt.Sprintf(format, args...))
}
