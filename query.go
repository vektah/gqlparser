package gqlparser

import (
	"github.com/vektah/gqlparser/errors"
	"github.com/vektah/gqlparser/lexer"
)

func ParseQuery(source string) (QueryDocument, *errors.Syntax) {
	p := parser{
		lexer: lexer.New(source),
	}
	return p.parseQueryDocument(), p.err
}

func (p *parser) parseQueryDocument() QueryDocument {
	var doc QueryDocument
	for p.peek().Kind != lexer.EOF {
		if p.err != nil {
			return doc
		}
		switch p.peek().Kind {
		case lexer.Name:
			switch p.peek().Value {
			case "query", "mutation", "subscription":
				doc.Operations = append(doc.Operations, p.parseOperationDefinition())
			case "fragment":
				doc.Fragments = append(doc.Fragments, p.parseFragmentDefinition())
			default:
				p.unexpectedError()
			}
		case lexer.BraceL:
			doc.Operations = append(doc.Operations, p.parseOperationDefinition())
		default:
			p.unexpectedError()
		}
	}

	return doc
}

func (p *parser) parseOperationDefinition() OperationDefinition {
	if p.peek().Kind == lexer.BraceL {
		return OperationDefinition{
			SelectionSet: p.parseSelectionSet(),
		}
	}

	var od OperationDefinition

	od.Operation = p.parseOperationType()

	if p.peek().Kind == lexer.Name {
		od.Name = p.next().Value
	}

	od.VariableDefinitions = p.parseVariableDefinitions()
	od.Directives = p.parseDirectives(false)
	od.SelectionSet = p.parseSelectionSet()

	return od
}

func (p *parser) parseOperationType() Operation {
	tok := p.next()
	switch tok.Value {
	case "query":
		return Query
	case "mutation":
		return Mutation
	case "subscription":
		return Subscription
	}
	p.unexpectedToken(tok)
	return ""
}

func (p *parser) parseVariableDefinitions() []VariableDefinition {
	var defs []VariableDefinition
	p.many(lexer.ParenL, lexer.ParenR, func() {
		defs = append(defs, p.parseVariableDefinition())
	})

	return defs
}

func (p *parser) parseVariableDefinition() VariableDefinition {
	var def VariableDefinition

	def.Variable = p.parseVariable()

	p.expect(lexer.Colon)

	def.Type = p.parseTypeReference()

	if p.skip(lexer.Equals) {
		def.DefaultValue = p.parseValueLiteral(true)
	}

	return def
}

func (p *parser) parseVariable() Variable {
	p.expect(lexer.Dollar)
	return Variable(p.parseName())
}

func (p *parser) parseSelectionSet() SelectionSet {
	var selections []Selection
	p.many(lexer.BraceL, lexer.BraceR, func() {
		selections = append(selections, p.parseSelection())
	})

	return SelectionSet(selections)
}

func (p *parser) parseSelection() Selection {
	if p.peek().Kind == lexer.Spread {
		return p.parseFragment()
	}
	return p.parseField()
}

func (p *parser) parseField() Field {
	var field Field

	nameOrAlias := p.parseName()

	if p.skip(lexer.Colon) {
		field.Alias = nameOrAlias
		field.Name = p.parseName()
	} else {
		field.Name = nameOrAlias
	}

	field.Arguments = p.parseArguments(false)
	field.Directives = p.parseDirectives(false)
	if p.peek().Kind == lexer.BraceL {
		field.SelectionSet = p.parseSelectionSet()
	}

	return field
}

func (p *parser) parseArguments(isConst bool) []Argument {
	var arguments []Argument
	p.many(lexer.ParenL, lexer.ParenR, func() {
		arguments = append(arguments, p.parseArgument(isConst))
	})

	return arguments
}

func (p *parser) parseArgument(isConst bool) Argument {
	arg := Argument{}

	arg.Name = p.parseName()
	p.expect(lexer.Colon)

	arg.Value = p.parseValueLiteral(isConst)
	return arg
}

func (p *parser) parseFragment() Selection {
	p.expect(lexer.Spread)

	if peek := p.peek(); peek.Kind == lexer.Name && peek.Value != "on" {
		return FragmentSpread{
			Name:       p.parseFragmentName(),
			Directives: p.parseDirectives(false),
		}
	}

	var def InlineFragment
	if p.peek().Value == "on" {
		p.next() // "on"

		def.TypeCondition = p.parseNamedType()
	}

	def.Directives = p.parseDirectives(false)
	def.SelectionSet = p.parseSelectionSet()
	return def
}

func (p *parser) parseFragmentDefinition() FragmentDefinition {
	var def FragmentDefinition
	p.expectKeyword("fragment")

	def.Name = p.parseFragmentName()
	def.VariableDefinition = p.parseVariableDefinitions()

	p.expectKeyword("on")

	def.TypeCondition = p.parseNamedType()
	def.Directives = p.parseDirectives(false)
	def.SelectionSet = p.parseSelectionSet()
	return def
}

func (p *parser) parseFragmentName() string {
	if p.peek().Value == "on" {
		p.unexpectedError()
		return ""
	}

	return p.parseName()
}

func (p *parser) parseValueLiteral(isConst bool) Value {
	token := p.peek()

	switch token.Kind {
	case lexer.BracketL:
		return p.parseList(isConst)
	case lexer.BraceL:
		return p.parseObject(isConst)
	case lexer.Int:
		p.next()
		return IntValue(token.Value)
	case lexer.Float:
		p.next()

		return FloatValue(token.Value)

	case lexer.String, lexer.BlockString:
		return p.parseStringLiteral()

	case lexer.Dollar:
		if !isConst {
			return p.parseVariable()
		}

	case lexer.Name:
		p.next()
		switch token.Value {

		case "true", "false":
			return BooleanValue(token.Value == "true")
		case "null":
			return NullValue{}
		default:
			return EnumValue(token.Value)
		}
	}

	p.unexpectedError()
	return nil
}

func (p *parser) parseStringLiteral() Value {
	token := p.next()

	if token.Kind == lexer.BlockString {
		return BlockValue(token.Value)
	}
	return StringValue(token.Value)
}

func (p *parser) parseList(isConst bool) ListValue {
	var values ListValue
	p.many(lexer.BracketL, lexer.BracketR, func() {
		values = append(values, p.parseValueLiteral(isConst))
	})

	return values
}

func (p *parser) parseObject(isConst bool) ObjectValue {
	var fields ObjectValue
	p.many(lexer.BraceL, lexer.BraceR, func() {
		fields = append(fields, p.parseObjectField(isConst))
	})

	return fields
}

func (p *parser) parseObjectField(isConst bool) ObjectField {
	field := ObjectField{}

	field.Name = p.parseName()

	p.expect(lexer.Colon)

	field.Value = p.parseValueLiteral(isConst)
	return field
}

func (p *parser) parseDirectives(isConst bool) []Directive {
	var directives []Directive

	for p.peek().Kind == lexer.At {
		if p.err != nil {
			break
		}
		directives = append(directives, p.parseDirective(isConst))
	}
	return directives
}

func (p *parser) parseDirective(isConst bool) Directive {
	p.expect(lexer.At)

	return Directive{
		Name:      p.parseName(),
		Arguments: p.parseArguments(isConst),
	}
}

func (p *parser) parseTypeReference() Type {
	var typ Type

	if p.skip(lexer.BracketL) {
		typ = p.parseTypeReference()

		typ = ListType{
			Type: typ,
		}
		p.expect(lexer.BracketR)
	} else {
		typ = p.parseNamedType()
	}

	if p.skip(lexer.Bang) {
		typ = NonNullType{
			Type: typ,
		}
	}
	return typ
}

func (p *parser) parseNamedType() NamedType {
	return NamedType(p.parseName())
}

func (p *parser) parseName() string {
	token := p.expect(lexer.Name)

	return token.Value
}
