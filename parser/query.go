package parser

import "github.com/vektah/graphql-parser/lexer"

func Parse(source string) (Document, error) {
	parser := Parser{
		lexer: lexer.New(source),
	}
	return parser.parseDocument()
}

func (p *Parser) parseDocument() (Document, error) {
	var definitions []Definition
	for p.peek().Kind != lexer.EOF {
		definitions = append(definitions, p.parseDefinition())
	}
	return Document{
		Definitions: definitions,
	}, p.err
}

func (p *Parser) parseDefinition() Definition {
	peek := p.peek()
	switch peek.Kind {
	case lexer.Name:
		switch peek.Value {
		case "query", "mutation", "subscription", "fragment":
			return p.parseExecutableDefinition()

			//case "schema", "scalar", "type", "interface", "union", "enum", "input", "directive":
			//	return p.parseTypeSystemDefinition()

			//case "extend":
			//	return p.parseTypeSystemExtension()
		}
	case lexer.BraceL:
		return p.parseExecutableDefinition()
		//case lexer.String, lexer.BlockString:
		//	return p.parseTypeSystemDefinition()
	}

	p.unexpectedError()
	return nil
}

func (p *Parser) parseExecutableDefinition() Definition {
	switch p.peek().Kind {
	case lexer.Name:
		switch p.peek().Value {
		case "query", "mutation":
			return p.parseOperationDefinition()
		case "fragment":
			return p.parseFragmentDefinition()
		}
	case lexer.BraceL:
		return p.parseOperationDefinition()
	}

	p.unexpectedError()
	return nil
}

func (p *Parser) parseOperationDefinition() OperationDefinition {
	if p.peek().Kind == lexer.BraceL {
		return OperationDefinition{
			SelectionSet: p.parseSelectionSet(),
		}
	}

	return OperationDefinition{
		Operation:           p.parseOperationType(),
		VariableDefinitions: p.parseVariableDefinitions(),
		Directives:          p.parseDirectives(false),
		SelectionSet:        p.parseSelectionSet(),
	}
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

func (p *Parser) parseFragmentName() Name {
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
		break

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
	return NamedType{
		Name: p.parseName(),
	}
}

func (p *Parser) parseName() Name {
	token := p.expect(lexer.Name)

	return Name(token.Value)
}
