package gqlerror

import (
	"encoding/json"
	"fmt"
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

func TestNilErrorUnwrapsToNil(t *testing.T) {
	// WrapPath returns a nil *Error for a nil cause, and stored in an error it is not == nil.
	var err error = WrapPath(nil, nil)

	require.NotErrorIs(t, err, underlyingError)
}

func TestWrapPosKeepsVerbsInTheWrappedMessage(t *testing.T) {
	wrapped := testError{"100% of %s queries failed"}

	t.Run("with nil position", func(t *testing.T) {
		err := WrapPos(nil, wrapped)

		require.Equal(t, wrapped.Error(), err.Message)
	})

	t.Run("with a position", func(t *testing.T) {
		pos := &ast.Position{
			Src:    &ast.Source{Name: "query.graphql"},
			Line:   2,
			Column: 3,
		}
		err := WrapPos(pos, wrapped)

		require.Equal(t, wrapped.Error(), err.Message)
		require.Equal(t, "query.graphql", err.Extensions["file"])
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

// TestErrorWithSourcesKeepsAnOverriddenSingleLocationFile matches the *Error case in
// TestErrorFormatting: with one location, a file set by SetFile takes precedence over the source's
// name, and converting the error must not undo that.
func TestErrorWithSourcesKeepsAnOverriddenSingleLocationFile(t *testing.T) {
	source := &ast.Source{Name: "query.graphql"}
	plain := ErrorPosf(&ast.Position{Src: source, Line: 1, Column: 2}, "kabloom")
	plain.SetFile("override.graphql")

	err := NewErrorWithSources(plain, []SourceLocation{{Line: 1, Column: 2, Source: source}})

	require.Equal(t, `override.graphql:1:2: kabloom`, err.Error())
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
	cases := map[string]struct {
		err  *Error
		want string
	}{
		"more sources than locations": {
			err:  &Error{Locations: []Location{{Line: 1, Column: 2}}},
			want: "gqlerror: source location count 2 does not match location count 1",
		},
		// Sources annotate an error's locations; they cannot supply locations it never had.
		"sources for an error with no locations": {
			err:  &Error{Message: "kabloom"},
			want: "gqlerror: source location count 2 does not match location count 0",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			require.PanicsWithValue(t, tc.want, func() {
				NewErrorWithSources(
					tc.err,
					[]SourceLocation{{Line: 1, Column: 2}, {Line: 3, Column: 4}},
				)
			})
		})
	}
}

func TestNewErrorWithSourcesRejectsMismatchedCoordinates(t *testing.T) {
	// Either coordinate differing is a mismatch on its own, not only both together.
	cases := map[string]SourceLocation{
		"both differ":         {Line: 3, Column: 4},
		"only line differs":   {Line: 3, Column: 2},
		"only column differs": {Line: 1, Column: 4},
	}
	for name, source := range cases {
		t.Run(name, func(t *testing.T) {
			require.PanicsWithValue(
				t,
				"gqlerror: source location 0 does not match location coordinates",
				func() {
					NewErrorWithSources(
						&Error{Locations: []Location{{Line: 1, Column: 2}}},
						[]SourceLocation{source},
					)
				},
			)
		})
	}
}

func TestNewErrorWithSourcesKeepsEveryField(t *testing.T) {
	path := ast.Path{ast.PathName("field")}
	extensions := map[string]any{"code": "BAD"}
	err := NewErrorWithSources(&Error{
		Err:        underlyingError,
		Message:    "kabloom",
		Path:       path,
		Extensions: extensions,
		Rule:       "NoUnusedFragments",
	}, nil)

	require.ErrorIs(t, err, underlyingError)
	require.Equal(t, "kabloom", err.Message)
	require.Equal(t, path, err.Path)
	require.Equal(t, extensions, err.Extensions)
	require.Equal(t, "NoUnusedFragments", err.Rule)
}

func TestWrappersKeepTheCause(t *testing.T) {
	pos := &ast.Position{Src: &ast.Source{Name: "query.graphql"}, Line: 2, Column: 3}
	cases := map[string]error{
		"Wrap":                        Wrap(underlyingError),
		"WrapPath":                    WrapPath(ast.Path{ast.PathName("field")}, underlyingError),
		"WrapPos with a position":     WrapPos(pos, underlyingError),
		"WrapPos without a position":  WrapPos(nil, underlyingError),
		"WrapIfUnwrapped over plain":  WrapIfUnwrapped(underlyingError),
		"NewErrorWithSources wrapped": NewErrorWithSources(Wrap(underlyingError), nil),
	}
	for name, err := range cases {
		t.Run(name, func(t *testing.T) {
			require.ErrorIs(t, err, underlyingError)
		})
	}
}

func TestWrappersCopyTheCauseMessage(t *testing.T) {
	cases := map[string]*Error{
		"Wrap":                       Wrap(underlyingError),
		"WrapPath":                   WrapPath(ast.Path{ast.PathName("field")}, underlyingError),
		"WrapIfUnwrapped over plain": WrapIfUnwrapped(underlyingError),
	}
	for name, err := range cases {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, underlyingError.Error(), err.Message)
		})
	}
}

func TestWrapPosWithoutPositionReportsUnknownLocation(t *testing.T) {
	err := WrapPos(nil, testError{"kabloom"})

	require.Equal(t, []Location{{Line: -1, Column: -1}}, err.Locations)
	require.Equal(t, "input:-1:-1: kabloom", err.Error())
}

func TestSetFileKeepsOtherExtensions(t *testing.T) {
	err := &Error{Extensions: map[string]any{"code": "BAD"}}

	err.SetFile("query.graphql")

	require.Equal(t, map[string]any{"code": "BAD", "file": "query.graphql"}, err.Extensions)
}

func TestErrorWithSourcesFormatsPath(t *testing.T) {
	err := &ErrorWithSources{Message: "kabloom", Path: ast.Path{ast.PathName("field")}}

	require.Equal(t, "input: field kabloom", err.Error())
}

func TestErrorWithSourcesFormatsSingleLocationWithoutSource(t *testing.T) {
	err := &ErrorWithSources{
		Message:   "kabloom",
		Locations: []SourceLocation{{Line: 1, Column: 2}},
	}

	require.Equal(t, "input:1:2: kabloom", err.Error())
}

// TestErrorWithSourcesUnmarshalKeepsLegacyFile checks that an error decoded from JSON, which
// cannot carry sources, still names the file its extensions record. Without the legacy flag
// UnmarshalJSON sets, several locations with no source would clear it.
func TestErrorWithSourcesUnmarshalKeepsLegacyFile(t *testing.T) {
	var err ErrorWithSources
	require.NoError(t, json.Unmarshal([]byte(`{
		"message": "kabloom",
		"locations": [{"line": 1, "column": 2}, {"line": 3, "column": 4}],
		"extensions": {"file": "legacy.graphql"}
	}`), &err))

	require.Equal(t, "legacy.graphql:1:2: kabloom", err.Error())
}

func TestErrorWithSourcesSourceLocationsIsNilWhenEmpty(t *testing.T) {
	cases := map[string]*ErrorWithSources{
		"nil *ErrorWithSources": nil,
		"no locations":          {Message: "kabloom"},
	}
	for name, err := range cases {
		t.Run(name, func(t *testing.T) {
			require.Nil(t, err.SourceLocations())
		})
	}
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

func TestConstructorsReturnNilForANilCause(t *testing.T) {
	pos := &ast.Position{Src: &ast.Source{Name: "query.graphql"}, Line: 2, Column: 3}
	cases := map[string]any{
		"Wrap":                    Wrap(nil),
		"WrapPath":                WrapPath(ast.Path{ast.PathName("field")}, nil),
		"WrapPos":                 WrapPos(pos, nil),
		"WrapIfUnwrapped":         WrapIfUnwrapped(nil),
		"NewErrorWithSources nil": NewErrorWithSources(nil, nil),
	}
	for name, got := range cases {
		t.Run(name, func(t *testing.T) {
			require.Nil(t, got)
		})
	}
}

// TestNilReceiversAreSafe covers the methods a nil pointer reaches once it is stored in an error,
// where it is not == nil. AsError exists to turn such a pointer back into a nil error, so the
// checks compare against the nil interface, which a typed nil does not equal.
func TestNilReceiversAreSafe(t *testing.T) {
	var plain *Error
	var withSources *ErrorWithSources

	require.Empty(t, plain.Error())
	require.Empty(t, withSources.Error())
	require.NoError(t, plain.Unwrap())
	require.NoError(t, withSources.Unwrap())
	require.NoError(t, plain.AsError())
	require.NoError(t, withSources.AsError())
}

func TestAsErrorReturnsTheSameError(t *testing.T) {
	plain := &Error{Message: "kabloom"}
	withSources := &ErrorWithSources{Message: "kabloom"}

	require.Same(t, plain, plain.AsError())
	require.Same(t, withSources, withSources.AsError())
}

func TestSetFile(t *testing.T) {
	cases := map[string]struct {
		err  *Error
		file string
		want map[string]any
	}{
		"no extensions yet": {
			err:  &Error{},
			file: "query.graphql",
			want: map[string]any{"file": "query.graphql"},
		},
		"an empty name leaves no extensions": {
			err:  &Error{},
			file: "",
			want: nil,
		},
		"an empty name keeps the file already set": {
			err:  &Error{Extensions: map[string]any{"file": "query.graphql"}},
			file: "",
			want: map[string]any{"file": "query.graphql"},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tc.err.SetFile(tc.file)

			require.Equal(t, tc.want, tc.err.Extensions)
		})
	}
}

func TestListFormatsAndUnwrapsEachError(t *testing.T) {
	first := &Error{Message: "first"}
	second := &Error{Err: underlyingError, Message: "second"}
	errs := List{first, second}

	require.EqualError(t, errs, "input: first\ninput: second\n")

	unwrapped := errs.Unwrap()
	require.Len(t, unwrapped, 2)
	require.Same(t, first, unwrapped[0])
	require.Same(t, second, unwrapped[1])
}

func TestWrapIfUnwrappedReturnsTheErrorAlreadyInTheChain(t *testing.T) {
	gqlErr := &Error{Message: "kabloom"}

	require.Same(t, gqlErr, WrapIfUnwrapped(fmt.Errorf("resolve: %w", gqlErr)))
}
