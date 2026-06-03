package ast

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDump(t *testing.T) {
	res := Dump(SchemaDefinition{
		Directives: []*Directive{
			{
				Name:      "foo",
				Arguments: []*Argument{{Name: "bar"}},
			},
			{Arguments: []*Argument{}},
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

	require.Equal(t, expected, res)
}
