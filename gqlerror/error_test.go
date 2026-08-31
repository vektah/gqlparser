package gqlerror

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vektah/gqlparser/v2/ast"
)

type testError struct {
	message string
}

func (e testError) Error() string {
	return e.message
}

var (
	underlyingError = testError{
		"Underlying error",
	}

	error1 = &Error{
		Message: "Some error 1",
	}
	error2 = &Error{
		Err:     underlyingError,
		Message: "Some error 2",
	}
)

func TestErrorFormatting(t *testing.T) {
	t.Run("without filename", func(t *testing.T) {
		err := ErrorLocf("", 66, 2, "kabloom")

		require.Equal(t, `input:66:2: kabloom`, err.Error())
		require.Nil(t, err.Extensions["file"])
	})

	t.Run("with filename", func(t *testing.T) {
		err := ErrorLocf("schema.graphql", 66, 2, "kabloom")

		require.Equal(t, `schema.graphql:66:2: kabloom`, err.Error())
		require.Equal(t, "schema.graphql", err.Extensions["file"])
	})

	t.Run("with path", func(t *testing.T) {
		err := ErrorPathf(
			ast.Path{ast.PathName("a"), ast.PathIndex(1), ast.PathName("b")},
			"kabloom",
		)

		require.Equal(t, `input: a[1].b kabloom`, err.Error())
	})

	t.Run("with legacy multi-location file extension", func(t *testing.T) {
		err := &Error{
			Message:    "kabloom",
			Locations:  []Location{{Line: 1, Column: 2}, {Line: 3, Column: 4}},
			Extensions: map[string]any{"file": "legacy.graphql"},
		}

		require.Equal(t, `legacy.graphql:1:2: kabloom`, err.Error())
	})

	t.Run("with overridden single-location file", func(t *testing.T) {
		err := ErrorPosf(&ast.Position{
			Src:    &ast.Source{Name: "query.graphql"},
			Line:   1,
			Column: 2,
		}, "kabloom")
		err.SetFile("override.graphql")

		require.Equal(t, `override.graphql:1:2: kabloom`, err.Error())
	})
}

func TestErrorPosition(t *testing.T) {
	t.Run("with nil position", func(t *testing.T) {
		err := ErrorLocf("", -1, -1, "kabloom")
		errNilPosition := ErrorPosf(nil, "%s", "kabloom")

		require.Equal(t, `input:-1:-1: kabloom`, err.Error())
		require.Equal(t, errNilPosition.Error(), err.Error())
		require.Nil(t, err.Extensions["file"])
		require.Nil(t, errNilPosition.Extensions["file"])
	})
}

func TestErrorWithSourcesAreNotSerialized(t *testing.T) {
	source := &ast.Source{Name: "query.graphql", Input: "query Q { field }"}
	err := NewErrorWithSources(
		&Error{
			Message:   "kabloom",
			Locations: []Location{{Line: 1, Column: 11}},
		},
		[]SourceLocation{{Line: 1, Column: 11, Source: source}},
	)

	encoded, errEncoding := json.Marshal(err)
	require.NoError(t, errEncoding)
	require.JSONEq(t, `{"message":"kabloom","locations":[{"line":1,"column":11}]}`, string(encoded))
}

func TestErrorWithSourcesUnmarshalClearsSources(t *testing.T) {
	err := NewErrorWithSources(
		&Error{Message: "first", Locations: []Location{{Line: 1, Column: 2}}},
		[]SourceLocation{{Line: 1, Column: 2, Source: &ast.Source{Name: "first.graphql"}}},
	)

	require.NoError(t, json.Unmarshal(
		[]byte(`{"message":"second","locations":[{"line":3,"column":4}]}`),
		err,
	))

	require.Equal(t, "second", err.Message)
	require.Equal(t, []SourceLocation{{Line: 3, Column: 4}}, err.Locations)
	require.Equal(t, err.Locations, err.SourceLocations())
	require.Nil(t, err.Locations[0].Source)
}

func TestErrorWithSourcesFormatsDirectCanonicalLocations(t *testing.T) {
	err := &ErrorWithSources{
		Message: "kabloom",
		Locations: []SourceLocation{
			{Line: 1, Column: 2, Source: &ast.Source{Name: "first.graphql"}},
			{Line: 3, Column: 4, Source: &ast.Source{Name: "second.graphql"}},
		},
	}

	require.Equal(t, `first.graphql:1:2: kabloom`, err.Error())
}

func TestErrorWithSourcesFormatsDirectSingleLocation(t *testing.T) {
	err := &ErrorWithSources{
		Message: "kabloom",
		Locations: []SourceLocation{
			{Line: 1, Column: 2, Source: &ast.Source{Name: "query.graphql"}},
		},
	}

	require.Equal(t, `query.graphql:1:2: kabloom`, err.Error())
}

func TestErrorWithSourcesPreservesDirectMissingPrimarySource(t *testing.T) {
	err := &ErrorWithSources{
		Message:    "kabloom",
		Extensions: map[string]any{"file": "second.graphql"},
		Locations: []SourceLocation{
			{Line: 1, Column: 2},
			{Line: 3, Column: 4, Source: &ast.Source{Name: "second.graphql"}},
		},
	}

	require.Equal(t, `input:1:2: kabloom`, err.Error())
}

func TestErrorWithSourcesReturnsDerivedCopy(t *testing.T) {
	source := &ast.Source{Name: "query.graphql"}
	err := NewErrorWithSources(
		&Error{Locations: []Location{{Line: 1, Column: 2}}},
		[]SourceLocation{{Line: 1, Column: 2, Source: source}},
	)

	locations := err.SourceLocations()
	locations[0].Line = 99
	locations[0].Source = nil

	require.Equal(t, SourceLocation{Line: 1, Column: 2, Source: source}, err.SourceLocations()[0])
	require.Same(t, source, err.SourceLocations()[0].Source)
}

func TestNewErrorWithSourcesCopiesInputLocations(t *testing.T) {
	source := &ast.Source{Name: "query.graphql"}
	locations := []SourceLocation{{Line: 1, Column: 2, Source: source}}
	err := NewErrorWithSources(
		&Error{Locations: []Location{{Line: 1, Column: 2}}},
		locations,
	)

	locations[0].Line = 99
	locations[0].Source = nil

	require.Equal(t, SourceLocation{Line: 1, Column: 2, Source: source}, err.Locations[0])
}

func TestNewErrorWithSourcesCopiesLegacyLocationsWithoutSources(t *testing.T) {
	err := NewErrorWithSources(
		&Error{
			Message:    "kabloom",
			Locations:  []Location{{Line: 1, Column: 2}, {Line: 3, Column: 4}},
			Extensions: map[string]any{"file": "legacy.graphql"},
		},
		nil,
	)

	require.Equal(t, []SourceLocation{
		{Line: 1, Column: 2},
		{Line: 3, Column: 4},
	}, err.Locations)
	require.Equal(t, `legacy.graphql:1:2: kabloom`, err.Error())
}

func TestNewErrorWithSourcesRejectsMismatchedLocations(t *testing.T) {
	require.PanicsWithValue(
		t,
		"gqlerror: source location count 2 does not match location count 1",
		func() {
			NewErrorWithSources(
				&Error{Locations: []Location{{Line: 1, Column: 2}}},
				[]SourceLocation{{Line: 1, Column: 2}, {Line: 3, Column: 4}},
			)
		},
	)
}

func TestNewErrorWithSourcesRejectsMismatchedCoordinates(t *testing.T) {
	require.PanicsWithValue(
		t,
		"gqlerror: source location 0 does not match location coordinates",
		func() {
			NewErrorWithSources(
				&Error{Locations: []Location{{Line: 1, Column: 2}}},
				[]SourceLocation{{Line: 3, Column: 4}},
			)
		},
	)
}

func TestSourceListMatchesListErrorBehavior(t *testing.T) {
	first := &ErrorWithSources{Message: "first"}
	second := &ErrorWithSources{Err: underlyingError, Message: "second"}
	errs := SourceList{first, second}

	require.EqualError(t, errs, "input: first\ninput: second\n")
	require.True(t, errs.Is(underlyingError))

	var found *ErrorWithSources
	require.True(t, errs.As(&found))
	require.Same(t, first, found)

	unwrapped := errs.Unwrap()
	require.Len(t, unwrapped, 2)
	require.Same(t, first, unwrapped[0])
	require.Same(t, second, unwrapped[1])
}

func TestList_As(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		errs        List
		target      any
		wantsTarget any
		targetFound bool
	}{
		{
			name: "Empty list",
			errs: List{},
		},
		{
			name:        "List with one error",
			errs:        List{error1},
			target:      new(*Error),
			wantsTarget: &error1,
			targetFound: true,
		},
		{
			name:        "List with multiple errors 1",
			errs:        List{error1, error2},
			target:      new(*Error),
			wantsTarget: &error1,
			targetFound: true,
		},
		{
			name:        "List with multiple errors 2",
			errs:        List{error1, error2},
			target:      new(testError),
			wantsTarget: &underlyingError,
			targetFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			targetFound := tt.errs.As(tt.target)

			if targetFound != tt.targetFound {
				t.Errorf("List.As() = %v, want %v", targetFound, tt.targetFound)
			}

			if tt.targetFound && !reflect.DeepEqual(tt.target, tt.wantsTarget) {
				t.Errorf("target = %v, want %v", tt.target, tt.wantsTarget)
			}
		})
	}
}

func TestList_Is(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		errs             List
		target           error
		hasMatchingError bool
	}{
		{
			name:             "Empty list",
			errs:             List{},
			target:           new(Error),
			hasMatchingError: false,
		},
		{
			name: "List with one error",
			errs: List{
				error1,
			},
			target:           error1,
			hasMatchingError: true,
		},
		{
			name: "List with multiple errors 1",
			errs: List{
				error1,
				error2,
			},
			target:           error2,
			hasMatchingError: true,
		},
		{
			name: "List with multiple errors 2",
			errs: List{
				error1,
				error2,
			},
			target:           underlyingError,
			hasMatchingError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hasMatchingError := tt.errs.Is(tt.target)
			if hasMatchingError != tt.hasMatchingError {
				t.Errorf("List.Is() = %v, want %v", hasMatchingError, tt.hasMatchingError)
			}
			if hasMatchingError && tt.target == nil {
				t.Errorf("List.Is() returned nil target, wants concrete error")
			}
		})
	}
}

func BenchmarkError(b *testing.B) {
	list := List([]*Error{error1, error2})
	for range b.N {
		_ = underlyingError.Error()
		_ = error1.Error()
		_ = error2.Error()
		_ = list.Error()
	}
}
