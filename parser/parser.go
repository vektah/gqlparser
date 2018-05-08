package parser

import (
	"fmt"
	"strconv"

	"github.com/vektah/graphql-parser"
	"github.com/vektah/graphql-parser/lexer"
)

type Parser struct {
	lexer lexer.Lexer
	err   *gqlparser.Error

	peeked    bool
	peekToken lexer.Token
	peekError *gqlparser.Error

	prev lexer.Token
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

func (p *Parser) error(tok lexer.Token, format string, args ...interface{}) {
	if p.err != nil {
		return
	}
	p.err = &gqlparser.Error{
		Message: fmt.Sprintf(format, args...),
		Locations: []gqlparser.Location{
			{Line: tok.Line, Column: tok.Column},
		},
	}
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

	p.error(tok, "Expected %s, found %s", strconv.Quote(value), tok.String())
	return tok
}

func (p *Parser) expect(kind lexer.Type) lexer.Token {
	tok := p.peek()
	if tok.Kind == kind {
		return p.next()
	}

	p.error(tok, "Expected %s, found %s", kind, tok.Kind.String())
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
	p.error(tok, "Unexpected %s", tok.String())
}

func (p *Parser) many(start lexer.Type, end lexer.Type, cb func()) {
	hasDef := p.skip(start)
	if !hasDef {
		return
	}

	for p.peek().Kind != end && p.err == nil {
		cb()
	}
	p.next()
}
