package spec

import (
	"fmt"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser"
)

func TestDump(t *testing.T) {
	res := DumpAST(gqlparser.SchemaDefinition{
		Directives: []gqlparser.Directive{
			{
				Name:      "foo",
				Arguments: []gqlparser.Argument{{Name: "bar"}},
			},
			{Arguments: []gqlparser.Argument{}},
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
