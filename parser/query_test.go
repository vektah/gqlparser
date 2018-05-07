package parser

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseQuery(t *testing.T) {
	doc, err := Parse("{ foo bar { baz } }")
	require.NoError(t, err)

	b, err := json.MarshalIndent(doc, "", "  ")
	require.NoError(t, err)
	fmt.Println(string(b))
	require.Equal(t, nil, doc)
}
