package gqlerror_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

func TestLocationSupportsUnkeyedLiteral(t *testing.T) {
	location := gqlerror.Location{1, 2}

	require.Equal(t, 1, location.Line)
	require.Equal(t, 2, location.Column)
}

func TestErrorSupportsUnkeyedLiteral(t *testing.T) {
	err := gqlerror.Error{nil, "kabloom", nil, nil, nil, ""}

	require.Equal(t, "kabloom", err.Message)
}

func TestErrorWithSourcesIsAvailableOutsidePackage(t *testing.T) {
	source := &ast.Source{Name: "query.graphql"}
	err := gqlerror.NewErrorWithSources(
		&gqlerror.Error{Locations: []gqlerror.Location{{Line: 1, Column: 2}}},
		[]gqlerror.SourceLocation{{Line: 1, Column: 2, Source: source}},
	)

	require.Len(t, err.Locations, 1)
	require.Same(t, source, err.Locations[0].Source)
	require.Equal(t, gqlerror.SourceLocation{Line: 1, Column: 2, Source: source}, err.Locations[0])
}
