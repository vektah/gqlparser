package parser

import (
	"github.com/vektah/graphql-parser"
	"github.com/vektah/graphql-parser/lexer"
)

func ParseSchema(source string) (SchemaDocument, *graphql_parser.Error) {
	parser := Parser{
		lexer: lexer.New(source),
	}
	return parser.parseSchemaDocument(), parser.err
}

func (p *Parser) parseSchemaDocument() SchemaDocument {
	var doc SchemaDocument
	for p.peek().Kind != lexer.EOF {
		if p.err != nil {
			return doc
		}

		var description string
		if p.peek().Kind == lexer.BlockString || p.peek().Kind == lexer.String {
			description = p.parseDescription()
		}

		if p.peek().Kind != lexer.Name {
			p.unexpectedError()
			break
		}

		switch p.peek().Value {
		case "schema", "scalar", "type", "interface", "union", "enum", "input", "directive":
			doc.Definitions = append(doc.Definitions, p.parseTypeSystemDefinition(description))
		case "extend":
			doc.Extensions = append(doc.Extensions, p.parseTypeSystemExtension(description))
		default:
			p.unexpectedError()
			return doc
		}
	}

	return doc
}

func (p *Parser) parseDescription() string {
	token := p.peek()

	if token.Kind != lexer.BlockString && token.Kind != lexer.String {
		return ""
	}

	return p.next().Value
}

func (p *Parser) parseTypeSystemDefinition(description string) Definition {
	tok := p.peek()
	if tok.Kind != lexer.Name {
		p.unexpectedError()
		return nil
	}

	switch tok.Value {
	case "schema":
		return p.parseSchemaDefinition(description)
	case "scalar":
		return p.parseScalarTypeDefinition(description)
	case "type":
		return p.parseObjectTypeDefinition(description)
	case "interface":
		return p.parseInterfaceTypeDefinition(description)
	case "union":
		return p.parseUnionTypeDefinition(description)
	case "enum":
		return p.parseEnumTypeDefinition(description)
	case "input":
		return p.parseInputObjectTypeDefinition(description)
	case "directive":
		return p.parseDirectiveDefinition(description)
	default:
		p.unexpectedError()
		return nil
	}
}

func (p *Parser) parseSchemaDefinition(description string) SchemaDefinition {
	p.expectKeyword("schema")

	def := SchemaDefinition{Description: description}
	def.Description = description
	def.Directives = p.parseDirectives(true)

	p.many(lexer.BraceL, lexer.BraceR, func() {
		def.OperationTypes = append(def.OperationTypes, p.parseOperationTypeDefinition())
	})
	return def
}

func (p *Parser) parseOperationTypeDefinition() OperationTypeDefinition {
	var op OperationTypeDefinition
	op.Operation = p.parseOperationType()
	p.expect(lexer.Colon)
	op.Type = p.parseNamedType()
	return op
}

func (p *Parser) parseScalarTypeDefinition(description string) ScalarTypeDefinition {
	p.expectKeyword("scalar")

	def := ScalarTypeDefinition{Description: description}
	def.Name = p.parseName()
	def.Directives = p.parseDirectives(true)
	return def
}

func (p *Parser) parseObjectTypeDefinition(description string) ObjectTypeDefinition {
	p.expectKeyword("type")

	def := ObjectTypeDefinition{Description: description}
	def.Name = p.parseName()
	def.Interfaces = p.parseImplementsInterfaces()
	def.Directives = p.parseDirectives(true)
	def.Fields = p.parseFieldsDefinition()
	return def
}

func (p *Parser) parseImplementsInterfaces() []NamedType {
	var types []NamedType
	if p.peek().Value == "implements" {
		p.next()
		// optional leading ampersand
		p.skip(lexer.Amp)

		types = append(types, p.parseNamedType())
		for p.skip(lexer.Amp) && p.err == nil {
			types = append(types, p.parseNamedType())
		}
	}
	return types
}

func (p *Parser) parseFieldsDefinition() []FieldDefinition {
	var defs []FieldDefinition
	p.many(lexer.BraceL, lexer.BraceR, func() {
		defs = append(defs, p.parseFieldDefinition())
	})
	return defs
}

func (p *Parser) parseFieldDefinition() FieldDefinition {
	var def FieldDefinition

	def.Description = p.parseDescription()
	def.Name = p.parseName()
	def.Arguments = p.parseArgumentDefs()
	p.expect(lexer.Colon)
	def.Type = p.parseTypeReference()
	def.Directives = p.parseDirectives(true)

	return def
}

func (p *Parser) parseArgumentDefs() []InputValueDefinition {
	var args []InputValueDefinition
	p.many(lexer.ParenL, lexer.ParenR, func() {
		args = append(args, p.parseInputValueDef())
	})
	return args
}

func (p *Parser) parseInputValueDef() InputValueDefinition {
	var def InputValueDefinition
	def.Description = p.parseDescription()
	def.Name = p.parseName()
	p.expect(lexer.Colon)
	def.Type = p.parseTypeReference()
	if p.skip(lexer.Equals) {
		def.DefaultValue = p.parseValueLiteral(true)
	}
	def.Directives = p.parseDirectives(true)
	return def
}

func (p *Parser) parseInterfaceTypeDefinition(description string) InterfaceTypeDefinition {
	p.expectKeyword("interface")

	def := InterfaceTypeDefinition{Description: description}
	def.Name = p.parseName()
	def.Directives = p.parseDirectives(true)
	def.Fields = p.parseFieldsDefinition()
	return def
}

func (p *Parser) parseUnionTypeDefinition(description string) UnionTypeDefinition {
	p.expectKeyword("union")

	def := UnionTypeDefinition{Description: description}
	def.Name = p.parseName()
	def.Directives = p.parseDirectives(true)
	def.Types = p.parseUnionMemberTypes()
	return def
}

func (p *Parser) parseUnionMemberTypes() []NamedType {
	var types []NamedType
	if p.skip(lexer.Equals) {
		// optional leading pipe
		p.skip(lexer.Pipe)

		types = append(types, p.parseNamedType())
		for p.skip(lexer.Pipe) && p.err == nil {
			types = append(types, p.parseNamedType())
		}
	}
	return types
}

func (p *Parser) parseEnumTypeDefinition(description string) EnumTypeDefinition {
	p.expectKeyword("enum")

	def := EnumTypeDefinition{Description: description}
	def.Name = p.parseName()
	def.Directives = p.parseDirectives(true)
	def.Values = p.parseEnumValuesDefinition()
	return def
}

func (p *Parser) parseEnumValuesDefinition() []EnumValueDefinition {
	var values []EnumValueDefinition
	p.many(lexer.BraceL, lexer.BraceR, func() {
		values = append(values, p.parseEnumValueDefinition())
	})
	return values
}

func (p *Parser) parseEnumValueDefinition() EnumValueDefinition {
	return EnumValueDefinition{
		Description: p.parseDescription(),
		Name:        p.parseName(),
		Directives:  p.parseDirectives(true),
	}
}

func (p *Parser) parseInputObjectTypeDefinition(description string) InputObjectTypeDefinition {
	p.expectKeyword("input")

	def := InputObjectTypeDefinition{Description: description}
	def.Name = p.parseName()
	def.Directives = p.parseDirectives(true)
	def.Fields = p.parseInputFieldsDefinition()
	return def
}

func (p *Parser) parseInputFieldsDefinition() []InputValueDefinition {
	var values []InputValueDefinition
	p.many(lexer.BraceL, lexer.BraceR, func() {
		values = append(values, p.parseInputValueDef())
	})
	return values
}

func (p *Parser) parseTypeSystemExtension(description string) TypeExtension {
	p.expectKeyword("extend")

	switch p.peek().Value {
	case "schema":
		return p.parseSchemaExtension(description)
	case "scalar":
		return p.parseScalarTypeExtension(description)
	case "type":
		return p.parseObjectTypeExtension(description)
	case "interface":
		return p.parseInterfaceTypeExtension(description)
	case "union":
		return p.parseUnionTypeExtension(description)
	case "enum":
		return p.parseEnumTypeExtension(description)
	case "input":
		return p.parseInputObjectTypeExtension(description)
	default:
		p.unexpectedError()
		return nil
	}
}

func (p *Parser) parseSchemaExtension(description string) SchemaExtension {
	p.expectKeyword("schema")

	def := SchemaExtension{Description: description}
	def.Directives = p.parseDirectives(true)
	p.many(lexer.BraceL, lexer.BraceR, func() {
		def.OperationTypes = append(def.OperationTypes, p.parseOperationTypeDefinition())
	})
	if len(def.Directives) == 0 && len(def.OperationTypes) == 0 {
		p.unexpectedError()
	}
	return def
}

func (p *Parser) parseScalarTypeExtension(description string) ScalarTypeExtension {
	p.expectKeyword("scalar")

	def := ScalarTypeExtension{Description: description}
	def.Directives = p.parseDirectives(true)
	if len(def.Directives) == 0 {
		p.unexpectedError()
	}
	return def
}

func (p *Parser) parseObjectTypeExtension(description string) ObjectTypeExtension {
	p.expectKeyword("type")

	def := ObjectTypeExtension{Description: description}
	def.Name = p.parseName()
	def.Interfaces = p.parseImplementsInterfaces()
	def.Directives = p.parseDirectives(true)
	def.Fields = p.parseFieldsDefinition()
	if len(def.Interfaces) == 0 && len(def.Directives) == 0 && len(def.Fields) == 0 {
		p.unexpectedError()
	}
	return def
}

func (p *Parser) parseInterfaceTypeExtension(description string) InterfaceTypeExtension {
	p.expectKeyword("interface")

	def := InterfaceTypeExtension{Description: description}
	def.Name = p.parseName()
	def.Directives = p.parseDirectives(true)
	def.Fields = p.parseFieldsDefinition()
	if len(def.Directives) == 0 && len(def.Fields) == 0 {
		p.unexpectedError()
	}
	return def
}

func (p *Parser) parseUnionTypeExtension(description string) UnionTypeExtension {
	p.expectKeyword("union")

	def := UnionTypeExtension{Description: description}
	def.Name = p.parseName()
	def.Directives = p.parseDirectives(true)
	def.Types = p.parseUnionMemberTypes()

	if len(def.Directives) == 0 && len(def.Types) == 0 {
		p.unexpectedError()
	}
	return def
}

func (p *Parser) parseEnumTypeExtension(description string) EnumTypeExtension {
	p.expectKeyword("enum")

	def := EnumTypeExtension{Description: description}
	def.Name = p.parseName()
	def.Directives = p.parseDirectives(true)
	def.Values = p.parseEnumValuesDefinition()
	if len(def.Directives) == 0 && len(def.Values) == 0 {
		p.unexpectedError()
	}
	return def
}

func (p *Parser) parseInputObjectTypeExtension(description string) InputObjectTypeExtension {
	p.expectKeyword("input")

	def := InputObjectTypeExtension{Description: description}
	def.Name = p.parseName()
	def.Directives = p.parseDirectives(false)
	def.Fields = p.parseInputFieldsDefinition()
	if len(def.Directives) == 0 && len(def.Fields) == 0 {
		p.unexpectedError()
	}
	return def
}

func (p *Parser) parseDirectiveDefinition(description string) DirectiveDefinition {
	p.expectKeyword("directive")
	p.expect(lexer.At)

	def := DirectiveDefinition{Description: description}
	def.Name = p.parseName()
	def.Arguments = p.parseArgumentDefs()

	p.expectKeyword("on")
	def.Locations = p.parseDirectiveLocations()
	return def
}

func (p *Parser) parseDirectiveLocations() []DirectiveLocation {
	p.skip(lexer.Pipe)

	locations := []DirectiveLocation{p.parseDirectiveLocation()}

	for p.skip(lexer.Pipe) && p.err == nil {
		locations = append(locations, p.parseDirectiveLocation())
	}

	return locations
}

func (p *Parser) parseDirectiveLocation() DirectiveLocation {
	name := p.expect(lexer.Name)

	switch name.Value {
	case `QUERY`:
		return LocationQuery
	case `MUTATION`:
		return LocationMutation
	case `SUBSCRIPTION`:
		return LocationSubscription
	case `FIELD`:
		return LocationField
	case `FRAGMENT_DEFINITION`:
		return LocationFragmentDefinition
	case `FRAGMENT_SPREAD`:
		return LocationFragmentSpread
	case `INLINE_FRAGMENT`:
		return LocationInlineFragment
	case `SCHEMA`:
		return LocationSchema
	case `SCALAR`:
		return LocationScalar
	case `OBJECT`:
		return LocationObject
	case `FIELD_DEFINITION`:
		return LocationFieldDefinition
	case `ARGUMENT_DEFINITION`:
		return LocationArgumentDefinition
	case `INTERFACE`:
		return LocationIinterface
	case `UNION`:
		return LocationUnion
	case `ENUM`:
		return LocationEnum
	case `ENUM_VALUE`:
		return LocationEnumValue
	case `INPUT_OBJECT`:
		return LocationInputObject
	case `INPUT_FIELD_DEFINITION`:
		return LocationInputFieldDefinition
	}

	p.unexpectedToken(name)
	return ""
}
