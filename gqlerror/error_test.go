package gqlerror

import (
	"errors"
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

func TestError_Is(t *testing.T) {
	t.Parallel()

	matchingPath := ast.Path{ast.PathName("query"), ast.PathIndex(1)}
	matchingLocations := []Location{{Line: 2, Column: 3}}
	matchingExtensions := map[string]any{"code": "INVALID", "nested": []string{"a", "b"}}

	tests := []struct {
		name    string
		err     *Error
		target  error
		want    bool
		reverse bool
	}{
		{
			name:    "identical messages",
			err:     &Error{Message: "invalid query"},
			target:  &Error{Message: "invalid query"},
			want:    true,
			reverse: true,
		},
		{
			name: "identical fields",
			err: &Error{
				Message:    "invalid query",
				Rule:       "KnownTypeNames",
				Path:       matchingPath,
				Locations:  matchingLocations,
				Extensions: matchingExtensions,
			},
			target: &Error{
				Message:    "invalid query",
				Rule:       "KnownTypeNames",
				Path:       ast.Path{ast.PathName("query"), ast.PathIndex(1)},
				Locations:  []Location{{Line: 2, Column: 3}},
				Extensions: map[string]any{"code": "INVALID", "nested": []string{"a", "b"}},
			},
			want:    true,
			reverse: true,
		},
		{
			name:   "different message",
			err:    &Error{Message: "first"},
			target: &Error{Message: "second"},
		},
		{
			name:   "different rule",
			err:    &Error{Message: "invalid", Rule: "RuleA"},
			target: &Error{Message: "invalid", Rule: "RuleB"},
		},
		{
			name:   "different path",
			err:    &Error{Message: "invalid", Path: ast.Path{ast.PathName("a")}},
			target: &Error{Message: "invalid", Path: ast.Path{ast.PathName("b")}},
		},
		{
			name:   "different locations",
			err:    &Error{Message: "invalid", Locations: []Location{{Line: 1, Column: 1}}},
			target: &Error{Message: "invalid", Locations: []Location{{Line: 1, Column: 2}}},
		},
		{
			name:   "different extensions",
			err:    &Error{Message: "invalid", Extensions: map[string]any{"code": "A"}},
			target: &Error{Message: "invalid", Extensions: map[string]any{"code": "B"}},
		},
		{
			name:   "wrapped non gqlerror target",
			err:    &Error{Err: underlyingError, Message: "wrapped"},
			target: underlyingError,
			want:   true,
		},
		{
			name: "nil target",
			err:  &Error{Message: "invalid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, errors.Is(tt.err, tt.target))
			if tt.reverse {
				require.Equal(t, tt.want, errors.Is(tt.target, tt.err))
			}
		})
	}

	var nilErr *Error
	require.True(t, nilErr.Is(nil))

	t.Run("does not match wrapped target", func(t *testing.T) {
		err := &Error{Message: "invalid query"}
		equivalent := &Error{Message: "invalid query"}

		require.True(t, errors.Is(err, equivalent))
		require.False(t, errors.Is(err, fmt.Errorf("ctx: %w", equivalent)))
	})
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
