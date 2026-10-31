package core_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
	"github.com/vektah/gqlparser/v2/validator/core"
)

// Each fragment spreads the next two, and the walker walks every fragment
// definition on its own as well as from the operation. With a linear lookup
// per spread that is cubic in the fragment count: 1 500 fragments, about
// 50 KB, took several seconds.
func TestWalkIsNotCubicInFragments(t *testing.T) {
	const n = 1500
	s := gqlparser.MustLoadSchema(
		&ast.Source{Input: `type Query { me: User } type User { id: ID, friend: User }`},
	)
	var b strings.Builder
	b.WriteString("{ me { ...F0 } }\n")
	for i := range n {
		fmt.Fprintf(&b, "fragment F%d on User { id ...F%d ...F%d }\n", i, i+1, i+2)
	}
	fmt.Fprintf(&b, "fragment F%d on User { id }\nfragment F%d on User { id }\n", n, n+1)
	doc, err := parser.ParseQuery(&ast.Source{Input: b.String()})
	require.NoError(t, err)

	start := time.Now()
	core.Walk(s, doc, &core.Events{})
	require.Less(t, time.Since(start), 2*time.Second)
}
