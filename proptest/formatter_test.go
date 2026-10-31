package proptest

import (
	"strings"
	"testing"

	"pgregory.net/rapid"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/formatter"
	"github.com/vektah/gqlparser/v2/parser"
)

// TestFormatQueryDocumentIsStable checks that formatting reaches a fixed point: parsing
// formatted output and formatting it again produces the same text.
//
// The golden files in testdata pin the output for the documents someone thought to write
// down. This covers the shape of document nobody thought of, and it is the property that
// makes the formatter usable as a canonicaliser: a tool that formats its own output must not
// keep changing it. A counterexample is either a construct the formatter prints in a form it
// cannot reparse, or one it prints differently on a second pass.
func TestFormatQueryDocumentIsStable(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		src := genQuery().Draw(t, "query")

		once := formatQuery(t, parseQuery(t, src))
		twice := formatQuery(t, parseQuery(t, once))

		if once != twice {
			t.Fatalf("formatting is not stable\nfirst pass:\n%s\nsecond pass:\n%s", once, twice)
		}
	})
}

func parseQuery(t *rapid.T, src string) *ast.QueryDocument {
	t.Helper()

	doc, err := parser.ParseQuery(&ast.Source{Name: "property.graphql", Input: src})
	if err != nil {
		t.Fatalf("parsing\n%s\nfailed: %s", src, err)
	}
	return doc
}

func formatQuery(t *rapid.T, doc *ast.QueryDocument) string {
	t.Helper()

	var buf strings.Builder
	formatter.NewFormatter(&buf).FormatQueryDocument(doc)
	return buf.String()
}

// genQuery draws a document that is valid by construction, so that a parse error or unstable
// output is the library's doing and not the generator's. It says nothing about whether the
// document would pass validation: the formatter does not care, and keeping the generator free
// of a schema keeps it able to reach shapes a schema-aware one could not.
func genQuery() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		var doc strings.Builder

		doc.WriteString(genOperation().Draw(t, "operation"))
		for _, fragment := range rapid.SliceOfN(genFragment(), 0, 2).Draw(t, "fragments") {
			doc.WriteString(fragment)
		}
		return doc.String()
	})
}

func genOperation() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		operation := rapid.SampledFrom([]string{"", "query ", "mutation ", "subscription "}).
			Draw(t, "operationType")
		name := ""
		if operation != "" {
			name = rapid.SampledFrom([]string{"", "A", "Named"}).Draw(t, "operationName")
		}
		return operation + name + " " + genSelectionSet(2).Draw(t, "selectionSet") + "\n"
	})
}

func genFragment() *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		name := rapid.SampledFrom([]string{"F", "G"}).Draw(t, "fragmentName")
		condition := rapid.SampledFrom([]string{"T", "U"}).Draw(t, "typeCondition")
		return "fragment " + name + " on " + condition + " " +
			genSelectionSet(1).Draw(t, "fragmentSelectionSet") + "\n"
	})
}

// genSelectionSet draws a brace-delimited set holding at least one selection, nesting no more
// than depth levels further down.
func genSelectionSet(depth int) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		selections := rapid.SliceOfN(genSelection(depth), 1, 3).Draw(t, "selections")
		return "{ " + strings.Join(selections, " ") + " }"
	})
}

func genSelection(depth int) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		kinds := []string{"field", "spread"}
		if depth > 0 {
			kinds = append(kinds, "inlineFragment")
		}

		switch rapid.SampledFrom(kinds).Draw(t, "selectionKind") {
		case "spread":
			return "..." + rapid.SampledFrom([]string{"F", "G"}).Draw(t, "spreadName")
		case "inlineFragment":
			condition := rapid.SampledFrom([]string{"", "on T", "on U"}).Draw(t, "inlineCondition")
			return strings.TrimSpace("... "+condition) + " " +
				genSelectionSet(depth-1).Draw(t, "inlineSelectionSet")
		default:
			return genField(depth).Draw(t, "field")
		}
	})
}

func genField(depth int) *rapid.Generator[string] {
	return rapid.Custom(func(t *rapid.T) string {
		var field strings.Builder

		if alias := rapid.SampledFrom([]string{"", "a", "theSame"}).Draw(t, "alias"); alias != "" {
			field.WriteString(alias + ": ")
		}
		field.WriteString(rapid.SampledFrom([]string{"a", "b", "fields", "__typename"}).
			Draw(t, "fieldName"))

		arguments := rapid.SliceOfNDistinct(genArgument(), 0, 2,
			func(argument [2]string) string { return argument[0] }).Draw(t, "arguments")
		if len(arguments) > 0 {
			pairs := make([]string, len(arguments))
			for i, argument := range arguments {
				pairs[i] = argument[0] + ": " + argument[1]
			}
			field.WriteString("(" + strings.Join(pairs, ", ") + ")")
		}

		if depth > 0 && rapid.Bool().Draw(t, "hasSubSelection") {
			field.WriteString(" " + genSelectionSet(depth-1).Draw(t, "subSelectionSet"))
		}
		return field.String()
	})
}

// genArgument draws a name and a value, covering the value kinds the formatter prints
// differently: the composite ones hold their contents in Children rather than Raw.
func genArgument() *rapid.Generator[[2]string] {
	return rapid.Custom(func(t *rapid.T) [2]string {
		name := rapid.SampledFrom([]string{"arg", "codes", "nested"}).Draw(t, "argumentName")
		value := rapid.SampledFrom([]string{
			"1", "-2", "1.5", `"s"`, `""`, "true", "null", "ENUM", "$var",
			"[]", `["a", "b"]`, "{}", `{a: 1}`, `{a: {b: [1, null]}}`,
		}).Draw(t, "argumentValue")
		return [2]string{name, value}
	})
}
