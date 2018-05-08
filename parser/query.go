package parser

import (
	"github.com/vektah/graphql-parser"
	"github.com/vektah/graphql-parser/lexer"
)

func ParseQuery(source string) (QueryDocument, *gqlparser.Error) {
	parser := Parser{
		lexer: lexer.New(source),
	}
	return parser.parseQueryDocument(), parser.err
}

func (p *Parser) parseQueryDocument() QueryDocument {
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

func (p *Parser) parseOperationDefinition() OperationDefinition {
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

func (p *Parser) parseOperationType() Operation {
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

func (p *Parser) parseVariableDefinitions() []VariableDefinition {
	var defs []VariableDefinition
	p.many(lexer.ParenL, lexer.ParenR, func() {
		defs = append(defs, p.parseVariableDefinition())
	})

	return defs
}

func (p *Parser) parseVariableDefinition() VariableDefinition {
	var def VariableDefinition

	def.Variable = p.parseVariable()

	p.expect(lexer.Colon)

	def.Type = p.parseTypeReference()

	if p.skip(lexer.Equals) {
		def.DefaultValue = p.parseValueLiteral(true)
	}

	return def
}

func (p *Parser) parseVariable() Variable {
	p.expect(lexer.Dollar)
	return Variable(p.parseName())
}

func (p *Parser) parseSelectionSet() SelectionSet {
	var selections []Selection
	p.many(lexer.BraceL, lexer.BraceR, func() {
		selections = append(selections, p.parseSelection())
	})

	return SelectionSet(selections)
}

func (p *Parser) parseSelection() Selection {
	if p.peek().Kind == lexer.Spread {
		return p.parseFragment()
	}
	return p.parseField()
}

func (p *Parser) parseField() Field {
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

func (p *Parser) parseArguments(isConst bool) []Argument {
	var arguments []Argument
	p.many(lexer.ParenL, lexer.ParenR, func() {
		arguments = append(arguments, p.parseArgument(isConst))
	})

	return arguments
}

func (p *Parser) parseArgument(isConst bool) Argument {
	arg := Argument{}

	arg.Name = p.parseName()
	p.expect(lexer.Colon)

	arg.Value = p.parseValueLiteral(isConst)
	return arg
}

func (p *Parser) parseFragment() Selection {
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

func (p *Parser) parseFragmentDefinition() FragmentDefinition {
	var def FragmentDefinition
	p.expectKeyword("fragment")

	def.Name = p.parseFragmentName()
	def.VariableDefinition = p.parseVariableDefinitions()

	p.expectKeyword("on")

	def.TypeCondition = p.parseNamedType()
	def.Directives = p.parseDirectives(false)
	return def
}

func (p *Parser) parseFragmentName() string {
	if p.peek().Value == "on" {
		p.unexpectedError()
		return ""
	}

	return p.parseName()
}

func (p *Parser) parseValueLiteral(isConst bool) Value {
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

func (p *Parser) parseStringLiteral() Value {
	token := p.next()

	if token.Kind == lexer.BlockString {
		return BlockValue(token.Value)
	}
	return StringValue(token.Value)
}

func (p *Parser) parseList(isConst bool) ListValue {
	var values ListValue
	p.many(lexer.BracketL, lexer.BracketR, func() {
		values = append(values, p.parseValueLiteral(isConst))
	})

	return values
}

func (p *Parser) parseObject(isConst bool) ObjectValue {
	var fields ObjectValue
	p.many(lexer.BraceL, lexer.BraceR, func() {
		fields = append(fields, p.parseObjectField(isConst))
	})

	return fields
}

func (p *Parser) parseObjectField(isConst bool) ObjectField {
	field := ObjectField{}

	field.Name = p.parseName()

	p.expect(lexer.Colon)

	field.Value = p.parseValueLiteral(isConst)
	return field
}

func (p *Parser) parseDirectives(isConst bool) []Directive {
	var directives []Directive

	for p.peek().Kind == lexer.At {
		if p.err != nil {
			break
		}
		directives = append(directives, p.parseDirective(isConst))
	}
	return directives
}

func (p *Parser) parseDirective(isConst bool) Directive {
	p.expect(lexer.At)

	return Directive{
		Name:      p.parseName(),
		Arguments: p.parseArguments(isConst),
	}
}

func (p *Parser) parseTypeReference() Type {
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

func (p *Parser) parseNamedType() NamedType {
	return NamedType(p.parseName())
}

func (p *Parser) parseName() string {
	token := p.expect(lexer.Name)

	return token.Value
}
