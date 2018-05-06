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

	p.many(lexer.BraceL, lexer.BraceR, func() error {
		selection, err := p.parseSelection()
		if err != nil {
			return err
		}
		selections = append(selections, selection)
		return nil
	})

	return SelectionSet{
		Selections: selections,
		Loc:        p.loc(start),
	}, nil
}

func (p *Parser) parseSelection() (Selection, error) {
	if p.lex.PeekToken().Kind == lexer.Spread {
		return p.parseFragment()
	}
	return p.parseField()
}

func (p *Parser) parseField() (Field, error) {
	start := p.lex.PeekToken()

	var field Field

	nameOrAlias, err := p.parseName()
	if err != nil {
		return field, err
	}

	hasName, err := p.skip(lexer.Colon)
	if err != nil {
		return field, err
	}
	if hasName {
		field.Alias = nameOrAlias
		field.Name, err = p.parseName()
		if err != nil {
			return field, err
		}
	} else {
		field.Name = nameOrAlias
	}

	field.Arguments, err = p.parseArguments(p.parseValueLiteral)
	if err != nil {
		return field, err
	}
	field.Directives, err = p.parseDirectives(false)
	if err != nil {
		return field, err
	}
	if p.lex.PeekToken().Kind == lexer.BraceL {
		field.SelectionSet, err = p.parseSelectionSet()
		if err != nil {
			return field, err
		}
	}

	field.Loc = p.loc(start)

	return field, nil
}

func (p *Parser) parseArguments(valuer func() (Value, error)) ([]Argument, error) {
	var arguments []Argument
	err := p.many(lexer.ParenL, lexer.ParenR, func() error {
		arg, err := p.parseArgument(valuer)
		if err != nil {
			return err
		}

		arguments = append(arguments, arg)
		return nil
	})

	return arguments, err
}

func (p *Parser) parseArgument(valuer func() (Value, error)) (Argument, error) {
	start := p.lex.PeekToken()
	arg := Argument{}
	var err error

	arg.Name, err = p.parseName()
	if err != nil {
		return arg, err
	}

	if err = p.expect(lexer.Colon); err != nil {
		return arg, err
	}

	arg.Value, err = valuer()
	if err != nil {
		return arg, err
	}

	arg.Loc = p.loc(start)
	return arg, nil
}

func (p *Parser) parseFragment() (Selection, error) {
	start := p.lex.PeekToken()

	err := p.expect(lexer.Spread)
	if err != nil {
		return nil, err
	}

	if peek := p.lex.PeekToken(); peek.Kind == lexer.Name && peek.Value != "on" {
		var def FragmentSpread
		def.Name, err = p.parseFragmentName()
		if err != nil {
			return nil, err
		}

		def.Directives, err = p.parseDirectives()
		if err != nil {
			return nil, err
		}

		def.Loc = p.loc(start)
		return def, nil
	}

	var def InlineFragment
	if p.lex.PeekToken().Value == "on" {
		_, err := p.lex.ReadToken() // "on"
		if err != nil {
			return nil, err
		}

		def.TypeCondition, err = p.parseNamedType()
		if err != nil {
			return nil, err
		}
	}

	def.Directives, err = p.parseDirectives(false)
	if err != nil {
		return nil, err
	}

	def.SelectionSet, err = p.parseSelectionSet()
	if err != nil {
		return nil, err
	}

	def.Loc = p.loc(start)
	return def, nil
}

func (p *Parser) parseFragmentDefinition() (FragmentDefinition, error) {
	var def FragmentDefinition
	start, err := p.expectKeyword("fragment")
	if err != nil {
		return def, err
	}

	def.Name, err = p.parseFragmentName()
	if err != nil {
		return def, err
	}

	def.VariableDefinition, err = p.parseVariableDefinitions()
	if err != nil {
		return def, err
	}

	_, err = p.expectKeyword("on")
	if err != nil {
		return def, err
	}

	def.TypeCondition, err = p.parseNamedType()
	if err != nil {
		return def, err
	}

	def.Directives, err = p.parseDirectives(false)
	if err != nil {
		return def, err
	}

	def.Loc = p.loc(start)
	return def, nil
}

func (p *Parser) parseFragmentName() (Name, error) {
	if p.lex.PeekToken().Value == "on" {
		return Name{}, p.unexpectedError()
	}

	return p.parseName()
}

func (p *Parser) parseValueLiteral(isConst bool) (Value, error) {
	token := p.lex.PeekToken()

	switch token.Kind {
	case lexer.BracketL:
		return p.parseList(isConst)
	case lexer.BraceL:
		return p.parseObject(isCont)
	case lexer.Int:
		_, err := p.lex.ReadToken()
		if err != nil {
			return nil, err
		}
		return IntValue{
			Value: token.Value,
			Loc:   p.loc(token),
		}, nil
	case lexer.Float:
		_, err := p.lex.ReadToken()
		if err != nil {
			return nil, err
		}

		return FloatValue{
			Value: token.Value,
			Loc:   p.loc(token),
		}, nil

	case lexer.String, lexer.BlockString:
		return p.parseStringLiteral()

	case lexer.Dollar:
		if !isConst {
			return p.parseVariable()
		}
		break

	case lexer.Name:
		_, err := p.lex.ReadToken()
		if err != nil {
			return nil, err
		}
		switch token.Value {

		case "true", "false":
			return BooleanValue{
				Value: token.Value == "true",
				Loc:   p.loc(token),
			}, nil
		case "null":
			return NullValue{
				Loc: p.loc(token),
			}, nil
		default:
			return EnumValue{
				Value: token.Value,
				Loc:   p.loc(token),
			}, nil
		}
	}

	return nil, p.unexpectedError()
}

func (p *Parser) expectKeyword(value string) (lexer.Token, error) {
	tok := p.lex.PeekToken()
	if tok.Kind == lexer.Name && tok.Value == value {
		return p.lex.ReadToken()
	}

	return lexer.Token{}, fmt.Errorf("Expected %s, found %s", value, tok.Kind.String())
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
