package gqlparser_test

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
	"github.com/vektah/gqlparser/v2/validator"
	"github.com/vektah/gqlparser/v2/validator/rules"
)

// An Apollo Federation v1 gateway validates each client operation against the
// supergraph that @apollo/composition builds. testdata/federation/README.md
// describes the fixtures and how to regenerate them.
func TestFederationV1Supergraph(t *testing.T) {
	supergraph, err := os.ReadFile(filepath.Join("testdata", "federation", "supergraph.graphql"))
	require.NoError(t, err)
	schema, err := gqlparser.LoadSchema(
		&ast.Source{Name: "supergraph.graphql", Input: string(supergraph)},
	)
	require.NoError(t, err)

	t.Run("keeps deprecations", func(t *testing.T) {
		search := schema.Query.Fields.ForName("search")
		require.NotNil(t, search.Arguments.ForName("limit").Directives.ForName("deprecated"))
		listKey := schema.Types["CreateStudentsInput"].Fields.ForName("listKey")
		require.NotNil(t, listKey.Directives.ForName("deprecated"))
		exercise := schema.Types["ContentKind"].EnumValues.ForName("EXERCISE")
		require.NotNil(t, exercise.Directives.ForName("deprecated"))
	})

	operations, err := os.ReadFile(filepath.Join("testdata", "federation", "operations.graphql"))
	require.NoError(t, err)
	doc, errs := gqlparser.LoadQueryWithRules(schema, string(operations), rules.NewDefaultRules())
	require.Empty(t, errs)

	cases := map[string]map[string]any{
		"MyProgress":         {},
		"DeprecatedArgument": {},
		"Search":             {"query": "fractions", "withArticles": true},
		"CreateStudents": {
			"input": map[string]any{"usernames": []any{"sal"}, "listKey": "period-1"},
		},
		"ListMutations": {},
	}
	var names []string
	for _, op := range doc.Operations {
		names = append(names, op.Name)
	}
	require.ElementsMatch(t, slices.Collect(maps.Keys(cases)), names,
		"give every operation in operations.graphql its variables")

	for name, variables := range cases {
		t.Run(name, func(t *testing.T) {
			coerced, err := validator.VariableValues(
				schema,
				doc.Operations.ForName(name),
				variables,
			)
			require.NoError(t, err)
			require.Equal(t, variables, coerced)
		})
	}
}

// The query planner turns each client operation into fetches against the
// subgraphs. Each gqlgen subgraph validates its fetches against its own schema
// plus the federation definitions gqlgen adds.
func TestFederationV1SubgraphFetches(t *testing.T) {
	federation, err := os.ReadFile(
		filepath.Join("testdata", "federation", "federation_v1.graphql"),
	)
	require.NoError(t, err)

	// The _Entity union lists each subgraph's types with a @key.
	entities := map[string]string{
		"accounts": "User",
		"content":  "Video | Article",
		"progress": "User | Video",
	}
	schemas := map[string]*ast.Schema{}
	for subgraph, members := range entities {
		sdl, err := os.ReadFile(
			filepath.Join("testdata", "federation", "subgraphs", subgraph+".graphql"),
		)
		require.NoError(t, err)
		schema, err := gqlparser.LoadSchema(
			&ast.Source{
				Name:    "federation/directives.graphql",
				Input:   string(federation),
				BuiltIn: true,
			},
			&ast.Source{Name: subgraph + ".graphql", Input: string(sdl)},
			&ast.Source{
				Name: "federation/entity.graphql",
				Input: "union _Entity = " + members + "\n" +
					"extend type Query {\n" +
					"  _entities(representations: [_Any!]!): [_Entity]!\n" +
					"  _service: _Service!\n" +
					"}\n",
				BuiltIn: true,
			},
		)
		require.NoError(t, err, subgraph)
		schemas[subgraph] = schema
	}

	cases := map[string]map[string]any{
		"MyProgress__accounts__0": {},
		"MyProgress__progress__1": {"representations": []any{
			map[string]any{"__typename": "User", "id": "1", "username": "sal"},
		}},
		"MyProgress__content__2": {"representations": []any{
			map[string]any{"__typename": "Video", "id": "video-1"},
		}},
		"DeprecatedArgument__accounts__0": {},
		"Search__content__0":              {"query": "fractions", "withArticles": true},
		"CreateStudents__accounts__0": {
			"input": map[string]any{"usernames": []any{"sal"}, "listKey": "period-1"},
		},
		"CreateStudents__progress__1": {"representations": []any{
			map[string]any{"__typename": "User", "id": "1"},
		}},
	}

	raw, err := os.ReadFile(filepath.Join("testdata", "federation", "fetches.json"))
	require.NoError(t, err)
	var fetches []struct {
		Subgraph string `json:"subgraph"`
		Query    string `json:"query"`
	}
	require.NoError(t, json.Unmarshal(raw, &fetches))
	require.Len(t, fetches, len(cases), "give every fetch in fetches.json its variables")

	for _, fetch := range fetches {
		parsed, err := parser.ParseQuery(&ast.Source{Input: fetch.Query})
		require.NoError(t, err)
		name := parsed.Operations[0].Name
		t.Run(name, func(t *testing.T) {
			variables, ok := cases[name]
			require.True(t, ok, "give fetch %s its variables", name)
			schema := schemas[fetch.Subgraph]

			doc, errs := gqlparser.LoadQueryWithRules(schema, fetch.Query, rules.NewDefaultRules())
			require.Empty(t, errs)
			coerced, err := validator.VariableValues(schema, doc.Operations[0], variables)
			require.NoError(t, err)
			require.Equal(t, variables, coerced)
		})
	}
}
