package parser

import (
	"fmt"

	"github.com/vektah/graphql-parser/lexer"
)

type Parser struct {
	lexer lexer.Lexer
	err   error

	peeked    bool
	peekToken lexer.Token
	peekError error

	prev lexer.Token
}

func newParser(input string) Parser {
	return Parser{lexer: lexer.New(input)}
}

func (p *Parser) peek() lexer.Token {
	if p.err != nil {
		return p.prev
	}

	if !p.peeked {
		p.peekToken, p.peekError = p.lexer.ReadToken()
		p.peeked = true
	}

	return p.peekToken
}

func (p *Parser) error(format string, args ...interface{}) {
	if p.err != nil {
		return
	}
	p.err = fmt.Errorf(format, args...)
}

func (p *Parser) next() lexer.Token {
	if p.err != nil {
		return p.prev
	}
	if p.peeked {
		p.peeked = false
		p.prev, p.err = p.peekToken, p.peekError
	} else {
		p.prev, p.err = p.lexer.ReadToken()
	}
	return p.prev
}

func (p *Parser) expectKeyword(value string) lexer.Token {
	tok := p.peek()
	if tok.Kind == lexer.Name && tok.Value == value {
		return p.next()
	}

	p.error("Expected %s, found %s", value, tok.String())
	return tok
}

func (p *Parser) expect(kind lexer.Type) lexer.Token {
	tok := p.peek()
	if tok.Kind == kind {
		return p.next()
	}

	p.error("Expected %s, found %s", kind, tok.Kind.String())
	return tok
}

func (p *Parser) skip(kind lexer.Type) bool {
	tok := p.peek()

	if tok.Kind != kind {
		return false
	}
	p.next()
	return true
}

func (p *Parser) unexpectedError() {
	p.unexpectedToken(p.peek())
}

func (p *Parser) unexpectedToken(tok lexer.Token) {
	p.error("Unexpected %s", tok.String())
}

func (p *Parser) many(start lexer.Type, end lexer.Type, cb func()) {
	hasDef := p.skip(start)
	if !hasDef {
		return
	}

	for p.peek().Kind != end {
		cb()
		if p.err != nil {
			return
		}
	}
	p.next()
}
