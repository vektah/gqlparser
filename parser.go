package graphql_parser

import (
	"fmt"

	"github.com/vektah/graphql-parser/lexer"
)

func Parse(source string) (Document, error) {
	parser := Parser{lexer.New(source)}
	return parser.parseDocument()
}

type Parser struct {
	lex lexer.Lexer
}

func (p *Parser) parseDocument() (Document, error) {
	start := p.lex.PeekToken()

	var definitions []Definition
	for p.lex.PeekToken().Kind != lexer.EOF {
		definition, err := p.parseDefinition()
		if err != nil {
			return Document{}, err
		}
		definitions = append(definitions, definition)
	}
	return Document{
		Loc:         p.loc(start),
		Definitions: definitions,
	}, nil
}

func (p *Parser) parseDefinition() (Definition, error) {
	peek := p.lex.PeekToken()
	switch peek.Kind {
	case lexer.Name:
		switch peek.Value {
		case "query", "mutation", "subscription", "fragment":
			return p.parseExecutableDefinition()

		case "schema", "scalar", "type", "interface", "union", "enum", "input", "directive":
			return p.parseTypeSystemDefinition()

		case "extend":
			return p.parseTypeSystemExtension()
		}
	case lexer.BraceL:
		return p.parseExecutableDefininition()
	case lexer.String, lexer.BlockString:
		return p.parseTypeSystemDefinition()
	}

	return nil, p.unexpectedError()
}

func (p *Parser) parseExecutableDefinition() (Definition, error) {
	peek := p.lex.PeekToken()
	switch peek.Kind {
	case lexer.Name:
		switch peek.Value {
		case "query", "mutation", "fragment":
			return p.parseOperationDefinition()
		case "fragment":
			return p.parseFragmentDefinition()
		}
	case lexer.BraceL:
		return p.parseOperationDefinition()
	}

	return nil, p.unexpectedError()
}

func (p *Parser) parseOperationDefinition() (OperationDefinition, error) {
	start := p.lex.PeekToken()

	if p.lex.PeekToken().Kind == lexer.BraceL {
		return OperationDefinition{
			Operation:    Query,
			SelectionSet: p.parseSelectionSet(),
			Loc:          p.loc(start),
		}, nil
	}

	operation := p.parseOperationType()
	var name = ""
	if p.lex.PeekToken() == lexer.Name {
		name = p.parseName()
	}
	return OperationDefinition{
		Operation:           operation,
		Name:                name,
		VariableDefinitions: p.parseVariableDefinitions(),
		Directives:          p.parseDirectives(),
		SelectionSet:        p.parseSelectionSet(),
		Loc:                 p.loc(start),
	}, nil
}

func (p *Parser) parseOperationType() (Operation, error) {
	token, err := p.lex.ReadToken()
	if err != nil {
		return "", err
	}
	switch token.Value {
	case "query":
		return Query, nil
	case "mutation":
		return Mutation, nil
	case "subscription":
		return Subscription, nil
	}
	return "", p.unexpectedError()
}

func (p *Parser) parseVariableDefinitions() ([]VariableDefinition, error) {
	var defs []VariableDefinition

	err := p.many(lexer.ParenL, lexer.ParenR, func() error {
		def, err := p.parseVariableDefinition()
		if err != nil {
			return err
		}

		defs = append(defs, def)
		return nil
	})

	return defs, err
}

func (p *Parser) many(start lexer.Type, end lexer.Type, cb func() error) error {
	hasDef, err := p.skip(lexer.ParenL)
	if !hasDef {
		return err
	}

	for p.lex.PeekToken().Kind != lexer.ParenR {
		err := cb()
		if err != nil {
			return err
		}
	}
	_, err = p.lex.ReadToken()
	return err
}

func (p *Parser) parseVariableDefinition() (VariableDefinition, error) {
	start := p.lex.PeekToken()
	var def VariableDefinition
	var err error

	def.Variable, err = p.parseVariable()
	if err != nil {
		return def, err
	}

	p.expect(lexer.Colon)

	def.Type, err = p.parseTypeReference()
	if err != nil {
		return def, err
	}

	hasDefault, err := p.skip(lexer.Equals)
	if err != nil {
		return def, err
	}
	if hasDefault {
		def.DefaultValue, err = p.parseValueLiteral(true)
		if err != nil {
			return def, err
		}
	}
	def.Loc = p.loc(start)
	return def, nil
}

func (p *Parser) parseVariable() (Variable, error) {
	start := p.lex.PeekToken()
	if err := p.expect(lexer.Dollar); err != nil {
		return Variable{}, err
	}

	return Variable{
		Name: p.parseName(),
		Loc:  p.loc(start),
	}, nil
}

func (p *Parser) parseSelectionSet() (SelectionSet, error) {
	start := p.lex.PeekToken()

	var selections []Selection

	return SelectionSet{
		Selections: selections,
		Loc:        p.loc(start),
	}, nil
}

func (p *Parser) expect(kind lexer.Type) error {
	tok, err := p.lex.ReadToken()
	if err != nil {
		return err
	}
	if tok.Kind != kind {
		return fmt.Errorf("Expected %s, found %s", kind.String(), tok.Kind.String())
	}
	return nil
}

func (p *Parser) skip(kind lexer.Type) (bool, error) {
	tok := p.lex.PeekToken()

	if tok.Kind != kind {
		return false, nil
	}
	_, err := p.lex.ReadToken()
	return true, err
}

func (p *Parser) unexpectedError() error {
	return fmt.Errorf("Unexpected %s", p.lex.PeekToken().String())
}

func (p *Parser) loc(start lexer.Token) Location {
	end := p.lex.LastToken()
	return Location{
		StartToken: start,
		EndToken:   end,
	}
}
