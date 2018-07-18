package parser

import (
	"fmt"

	. "github.com/vektah/gqlparser/ast"
)

func LoadSchema(input string) (*Schema, error) {
	ast, err := ParseSchema(input)
	if err != nil {
		return nil, err
	}

	schema := Schema{
		Types:         map[string]*Definition{},
		Directives:    map[string]*DirectiveDefinition{},
		PossibleTypes: map[string][]*Definition{},
	}

	schema.Types["Int"] = &Definition{
		Kind:        Scalar,
		Description: "The `Int` scalar type represents non-fractional signed whole numeric values. Int can represent values between -(2^31) and 2^31 - 1.",
		Name:        "Int",
	}
	schema.Types["Float"] = &Definition{
		Kind:        Scalar,
		Description: "The `Float` scalar type represents signed double-precision fractional values as specified by [IEEE 754](http://en.wikipedia.org/wiki/IEEE_floating_point).",
		Name:        "Float",
	}
	schema.Types["String"] = &Definition{
		Kind:        Scalar,
		Description: "The `String` scalar type represents textual data, represented as UTF-8 character sequences. The String type is most often used by GraphQL to represent free-form human-readable text.",
		Name:        "String",
	}
	schema.Types["Boolean"] = &Definition{
		Kind:        Scalar,
		Description: "The `Boolean` scalar type represents `true` or `false`.",
		Name:        "Boolean",
	}
	schema.Types["ID"] = &Definition{
		Kind:        Scalar,
		Description: "The `ID` scalar type represents a unique identifier, often used to refetch an object or as key for a cache. The ID type appears in a JSON response as a String; however, it is not intended to be human-readable. When expected as an input type, any string (such as `\"4\"`) or integer (such as `4`) input value will be accepted as an ID.",
		Name:        "ID",
	}

	schema.Directives["skip"] = &DirectiveDefinition{
		Name:        "skip",
		Description: "Directs the executor to skip this field or fragment when the `if` argument is true.",
		Arguments: FieldList{
			&FieldDefinition{
				Name:        "if",
				Description: "Skipped when true.",
				Type:        NonNullNamedType("Boolean"),
			},
		},
		Locations: []DirectiveLocation{LocationField, LocationFragmentSpread, LocationInlineFragment},
	}
	schema.Directives["include"] = &DirectiveDefinition{
		Name:        "include",
		Description: "Directs the executor to include this field or fragment only when the `if` argument is true.",
		Arguments: FieldList{
			&FieldDefinition{
				Name:        "if",
				Description: "Included when true.",
				Type:        NonNullNamedType("Boolean"),
			},
		},
		Locations: []DirectiveLocation{LocationField, LocationFragmentSpread, LocationInlineFragment},
	}

	for i, def := range ast.Definitions {
		if schema.Types[def.Name] != nil {
			return nil, fmt.Errorf("Cannot redeclare type %s.", def.Name)
		}
		schema.Types[def.Name] = ast.Definitions[i]

		if def.Kind != Interface {
			for _, intf := range def.Interfaces {
				schema.AddPossibleType(intf, ast.Definitions[i])
			}
			schema.AddPossibleType(def.Name, ast.Definitions[i])
		}
	}

	for _, ext := range ast.Extensions {
		def := schema.Types[ext.Name]
		if def == nil {
			return nil, fmt.Errorf("Cannot extend type %s because it does not exist.", ext.Name)
		}

		if def.Kind != ext.Kind {
			return nil, fmt.Errorf("Cannot extend type %s because the base type is a %s, not %s.", ext.Name, def.Kind, ext.Kind)
		}

		def.Directives = append(def.Directives, ext.Directives...)
		def.Interfaces = append(def.Interfaces, ext.Interfaces...)
		def.Fields = append(def.Fields, ext.Fields...)
		def.Types = append(def.Types, ext.Types...)
		def.EnumValues = append(def.EnumValues, ext.EnumValues...)
	}

	for i, dir := range ast.Directives {
		if schema.Directives[dir.Name] != nil {
			return nil, fmt.Errorf("Cannot redeclare directive %s.", dir.Name)
		}
		schema.Directives[dir.Name] = ast.Directives[i]
	}

	if len(ast.Schema) > 1 {
		return nil, fmt.Errorf("Cannot have multiple schema entry points, consider schema extensions instead.")
	}

	if len(ast.Schema) == 1 {
		for _, entrypoint := range ast.Schema[0].OperationTypes {
			def := schema.Types[entrypoint.Type]
			if def == nil {
				return nil, fmt.Errorf("Schema root %s refers to a type %s that does not exist.", entrypoint.Operation, entrypoint.Type)
			}
			switch entrypoint.Operation {
			case Query:
				schema.Query = def
			case Mutation:
				schema.Mutation = def
			case Subscription:
				schema.Subscription = def
			}
		}
	}

	for _, ext := range ast.SchemaExtension {
		for _, entrypoint := range ext.OperationTypes {
			def := schema.Types[entrypoint.Type]
			if def == nil {
				return nil, fmt.Errorf("Schema root %s refers to a type %s that does not exist.", entrypoint.Operation, entrypoint.Type)
			}
			switch entrypoint.Operation {
			case Query:
				schema.Query = def
			case Mutation:
				schema.Mutation = def
			case Subscription:
				schema.Subscription = def
			}
		}
	}

	return &schema, nil
}
