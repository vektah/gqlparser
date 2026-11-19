package validator_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/validator"
	"github.com/vektah/gqlparser/v2/validator/rules"
)

func TestVariableValuesNestedListCoercion(t *testing.T) {
	schema := gqlparser.MustLoadSchema(&ast.Source{Input: `
scalar JSON
enum Choice { A B }
input Filter { xs: [[String]], children: [Filter!], label: String }
type Query {
  strings(value: [[String]]): Int
  deep(value: [[[String]]]): Int
  input(value: Filter): Int
  objects(value: [[Filter]]): Int
  choices(value: [[Choice]]): Int
  json(value: JSON): Int
}
`})
	value := "a"
	tests := []struct {
		name, definition, field string
		variables               map[string]any
		want                    string
		wantErr                 string
	}{
		{
			name: "scalar", definition: "$v: [[String]]", field: "strings",
			variables: map[string]any{"v": "a"},
			want:      `{"v":[["a"]]}`,
		},
		{
			name: "three dimensions", definition: "$v: [[[String]]]", field: "deep",
			variables: map[string]any{"v": "a"},
			want:      `{"v":[[["a"]]]}`,
		},
		{
			name: "flat JSON list", definition: "$v: [[String]]", field: "strings",
			variables: map[string]any{"v": []any{"a", "b"}},
			want:      `{"v":[["a"],["b"]]}`,
		},
		{
			name: "flat typed list", definition: "$v: [[String]]", field: "strings",
			variables: map[string]any{"v": []string{"a", "b"}},
			want:      `{"v":[["a"],["b"]]}`,
		},
		{
			name: "typed list needs two dimensions", definition: "$v: [[[String]]]", field: "deep",
			variables: map[string]any{"v": []string{"a", "b"}},
			want:      `{"v":[[["a"]],[["b"]]]}`,
		},
		{
			name:       "nested typed list needs another dimension",
			definition: "$v: [[[String]]]",
			field:      "deep",
			variables:  map[string]any{"v": [][]string{{"a", "b"}, {"c"}}},
			want:       `{"v":[[["a"],["b"]],[["c"]]]}`,
		},
		{
			name: "explicit typed lists", definition: "$v: [[String]]", field: "strings",
			variables: map[string]any{"v": [][]string{{"a", "b"}, {"c"}}},
			want:      `{"v":[["a","b"],["c"]]}`,
		},
		{
			name: "mixed list", definition: "$v: [[String]]", field: "strings",
			variables: map[string]any{"v": []any{"a", []any{"b", "c"}, nil, []any{}}},
			want:      `{"v":[["a"],["b","c"],null,[]]}`,
		},
		{
			name: "null before a typed element", definition: "$v: [[String]]", field: "strings",
			variables: map[string]any{"v": []*string{nil, &value}},
			want:      `{"v":[null,["a"]]}`,
		},
		{
			name: "null after a typed element", definition: "$v: [[String]]", field: "strings",
			variables: map[string]any{"v": []*string{&value, nil}},
			want:      `{"v":[["a"],null]}`,
		},
		{
			name: "scalar default", definition: `$v: [[String]] = "a"`, field: "strings",
			want: `{"v":[["a"]]}`,
		},
		{
			name: "flat default", definition: `$v: [[String]] = ["a", "b"]`, field: "strings",
			want: `{"v":[["a"],["b"]]}`,
		},
		{
			name: "null overrides default", definition: `$v: [[String]] = "a"`, field: "strings",
			variables: map[string]any{"v": nil},
			want:      `{"v":null}`,
		},
		{
			name: "null default", definition: "$v: [[String]] = null", field: "strings",
			want: `{"v":null}`,
		},
		{
			name: "omitted", definition: "$v: [[String]]", field: "strings",
			want: `{}`,
		},
		{
			name: "input object", definition: "$v: Filter", field: "input",
			variables: map[string]any{"v": map[string]any{
				"xs": "a", "children": map[string]any{"xs": "b"},
			}},
			want: `{"v":{"xs":[["a"]],"children":[{"xs":[["b"]]}]}}`,
		},
		{
			name: "typed input map", definition: "$v: Filter", field: "input",
			variables: map[string]any{"v": map[string]string{"xs": "a"}},
			want:      `{"v":{"xs":[["a"]]}}`,
		},
		{
			name: "typed input maps in a list", definition: "$v: [[Filter]]", field: "objects",
			variables: map[string]any{"v": []map[string]string{{"xs": "a"}, {"xs": "b"}}},
			want:      `{"v":[[{"xs":[["a"]]}],[{"xs":[["b"]]}]]}`,
		},
		{
			name:       "typed input map inside an explicit list",
			definition: "$v: [[Filter]]",
			field:      "objects",
			variables:  map[string]any{"v": [][]map[string]string{{{"xs": "a"}}}},
			want:       `{"v":[[{"xs":[["a"]]}]]}`,
		},
		{
			name:       "typed input map preserves siblings",
			definition: "$v: Filter",
			field:      "input",
			variables: map[string]any{"v": map[string]string{
				"xs": "a", "label": "keep", "__typename": "Filter",
			}},
			want: `{"v":{"xs":[["a"]],"label":"keep","__typename":"Filter"}}`,
		},
		{
			name:       "widened list preserves prior input objects",
			definition: "$v: [[Filter]]",
			field:      "objects",
			variables: map[string]any{"v": [][]map[string]string{
				{{"label": "first"}, {"xs": "a"}},
			}},
			want: `{"v":[[{"label":"first"},{"xs":[["a"]]}]]}`,
		},
		{
			name:       "input object default",
			definition: `$v: Filter = {xs: "a", children: {xs: "b"}}`,
			field:      "input",
			want:       `{"v":{"xs":[["a"]],"children":[{"xs":[["b"]]}]}}`,
		},
		{
			name: "enum", definition: "$v: [[Choice]]", field: "choices",
			variables: map[string]any{"v": "A"},
			want:      `{"v":[["A"]]}`,
		},
		{
			name: "opaque custom scalar", definition: "$v: JSON", field: "json",
			variables: map[string]any{"v": map[string]any{"xs": "a", "list": []any{"b", nil}}},
			want:      `{"v":{"xs":"a","list":["b",null]}}`,
		},
		{
			name: "required variable omitted", definition: "$v: [[String!]!]!", field: "strings",
			wantErr: "input: variable.v must be defined",
		},
		{
			name: "required variable null", definition: "$v: [[String!]!]!", field: "strings",
			variables: map[string]any{"v": nil},
			wantErr:   "input: variable.v cannot be null",
		},
		{
			name: "required inner list null", definition: "$v: [[String!]!]!", field: "strings",
			variables: map[string]any{"v": []any{nil}},
			wantErr:   "input: variable.v[0] cannot be null",
		},
		{
			name: "required leaf null", definition: "$v: [[String!]!]!", field: "strings",
			variables: map[string]any{"v": []any{[]any{nil}}},
			wantErr:   "input: variable.v[0][0] cannot be null",
		},
		{
			name: "invalid scalar", definition: "$v: [[String]]", field: "strings",
			variables: map[string]any{"v": []int{1}},
			wantErr:   "input: variable.v[0][0] cannot use int as String",
		},
		{
			name: "invalid enum", definition: "$v: [[Choice]]", field: "choices",
			variables: map[string]any{"v": "unknown"},
			wantErr:   "input: variable.v[0][0] unknown is not a valid Choice",
		},
		{
			name: "unknown input field", definition: "$v: Filter", field: "input",
			variables: map[string]any{"v": map[string]any{"unknown": true}},
			wantErr:   "input: variable.v.unknown unknown field",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := gqlparser.MustLoadQueryWithRules(schema,
				"query("+tt.definition+"){"+tt.field+"(value:$v)}", rules.NewDefaultRules())
			variables, err := validator.VariableValues(schema, query.Operations[0], tt.variables)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			encoded, err := json.Marshal(variables)
			require.NoError(t, err)
			require.JSONEq(t, tt.want, string(encoded))
		})
	}
}
