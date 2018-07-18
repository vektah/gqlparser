package spec

import (
	"fmt"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/ast"
)

func TestDump(t *testing.T) {
	res := DumpAST(ast.SchemaDefinition{
		Directives: []*ast.Directive{
			{
				Name:      "foo",
				Arguments: []*ast.Argument{{Name: "bar"}},
			},
			{Arguments: []*ast.Argument{}},
		},
	})

	expected := `<SchemaDefinition>
  Directives: [Directive]
  - <Directive>
      Name: "foo"
      Arguments: [Argument]
      - <Argument>
          Name: "bar"
  - <Directive>`

	fmt.Println(diff.LineDiff(expected, res))
	require.Equal(t, expected, res)
}
