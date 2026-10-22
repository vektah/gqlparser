package validator

import (
	_ "embed"
	"sync"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

//go:embed imported/prelude.graphql
var preludeGraphql string

// Prelude is the GraphQL source of the types and directives the spec declares
// implicitly. gqlparser.LoadSchema adds it ahead of the caller's sources.
var Prelude = &ast.Source{
	Name:    "prelude.graphql",
	Input:   preludeGraphql,
	BuiltIn: true,
}

// PreludeDocument returns the parsed Prelude: the built-in scalars, directives and
// introspection types that LoadSchema adds to every schema. Use it to find out what
// this version of gqlparser declares, instead of keeping a copy of those names.
//
// Each call parses Prelude again, so the caller owns the result and may modify it.
// It panics if the embedded prelude does not parse, which is a gqlparser bug.
func PreludeDocument() *ast.SchemaDocument {
	doc, err := parser.ParseSchema(Prelude)
	if err != nil {
		panic("gqlparser: embedded prelude does not parse: " + err.Error())
	}
	return doc
}

// QueryIntrospectionFields returns the __schema and __type fields that the spec
// adds to the query root type. ValidateSchemaDocument appends them to the query
// type it finds; a caller that supplies its own query type needs them too.
//
// Each call returns new definitions, so the caller owns the result.
func QueryIntrospectionFields() ast.FieldList {
	return ast.FieldList{
		{
			Name: "__schema",
			Type: ast.NonNullNamedType("__Schema", nil),
		},
		{
			Name: "__type",
			Type: ast.NamedType("__Type", nil),
			Arguments: ast.ArgumentDefinitionList{
				{Name: "name", Type: ast.NonNullNamedType("String", nil)},
			},
		},
	}
}

// preludeDirectiveNames reports which directive names the prelude declares. Only a
// schema that redeclares a directive needs it, so it is computed on first use and
// then reused, sparing every other LoadSchema a second parse of the prelude. The
// set is never modified, so sharing it is safe.
var preludeDirectiveNames = sync.OnceValue(func() map[string]bool {
	names := map[string]bool{}
	for _, dir := range PreludeDocument().Directives {
		names[dir.Name] = true
	}
	return names
})
