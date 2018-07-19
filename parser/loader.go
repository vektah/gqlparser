package parser

import (
	. "github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/gqlerror"
)

func LoadSchema(inputs ...*Source) (*Schema, *gqlerror.Error) {
	ast := &SchemaDocument{}
	for _, input := range inputs {
		inputAst, err := ParseSchema(input)
		if err != nil {
			return nil, err
		}

		ast.Merge(inputAst)
	}

	schema := Schema{
		Types:         map[string]*Definition{},
		Directives:    map[string]*DirectiveDefinition{},
		PossibleTypes: map[string][]*Definition{},
	}

	for i, def := range ast.Definitions {
		if schema.Types[def.Name] != nil {
			return nil, gqlerror.Errorf("Cannot redeclare type %s.", def.Name)
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
			return nil, gqlerror.Errorf("Cannot extend type %s because it does not exist.", ext.Name)
		}

		if def.Kind != ext.Kind {
			return nil, gqlerror.Errorf("Cannot extend type %s because the base type is a %s, not %s.", ext.Name, def.Kind, ext.Kind)
		}

		def.Directives = append(def.Directives, ext.Directives...)
		def.Interfaces = append(def.Interfaces, ext.Interfaces...)
		def.Fields = append(def.Fields, ext.Fields...)
		def.Types = append(def.Types, ext.Types...)
		def.EnumValues = append(def.EnumValues, ext.EnumValues...)
	}

	for i, dir := range ast.Directives {
		if schema.Directives[dir.Name] != nil {
			return nil, gqlerror.Errorf("Cannot redeclare directive %s.", dir.Name)
		}
		schema.Directives[dir.Name] = ast.Directives[i]
	}

	if len(ast.Schema) > 1 {
		return nil, gqlerror.Errorf("Cannot have multiple schema entry points, consider schema extensions instead.")
	}

	if len(ast.Schema) == 1 {
		for _, entrypoint := range ast.Schema[0].OperationTypes {
			def := schema.Types[entrypoint.Type]
			if def == nil {
				return nil, gqlerror.Errorf("Schema root %s refers to a type %s that does not exist.", entrypoint.Operation, entrypoint.Type)
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
				return nil, gqlerror.Errorf("Schema root %s refers to a type %s that does not exist.", entrypoint.Operation, entrypoint.Type)
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

	if schema.Query == nil && schema.Types["Query"] != nil {
		schema.Query = schema.Types["Query"]
	}

	if schema.Mutation == nil && schema.Types["Mutation"] != nil {
		schema.Mutation = schema.Types["Mutation"]
	}

	if schema.Subscription == nil && schema.Types["Subscription"] != nil {
		schema.Subscription = schema.Types["Subscription"]
	}

	if schema.Query != nil {
		schema.Query.Fields = append(
			schema.Query.Fields,
			&FieldDefinition{
				Name: "__schema",
				Type: NonNullNamedType("__Schema"),
			},
			&FieldDefinition{
				Name: "__type",
				Type: NonNullNamedType("__Type"),
				Arguments: FieldList{
					{Name: "name", Type: NamedType("String")},
				},
			},
		)
	}

	return &schema, nil
}
