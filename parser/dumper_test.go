package parser

import (
	"fmt"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/stretchr/testify/require"
)

func TestDump(t *testing.T) {
	res := Dump(SchemaDefinition{
		Directives: []Directive{
			{
				Name:      "foo",
				Arguments: []Argument{{Name: "bar"}},
			},
			{Arguments: []Argument{}},
		},
	})

	expected := `SchemaDefinition {
  Directives: [
    Directive {
      Name: foo
      Arguments: [
        Argument {
          Name: bar
        }
      ]
    }
    Directive {}
  ]
}`

	fmt.Println(diff.LineDiff(expected, res))
	require.Equal(t, expected, res)
}
