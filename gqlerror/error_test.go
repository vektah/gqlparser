package gqlerror

import (
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
		err := ErrorPathf(ast.Path{ast.PathName("a"), ast.PathIndex(1), ast.PathName("b")}, "kabloom")

		require.Equal(t, `input: a[1].b kabloom`, err.Error())
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
		tt := tt

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
		tt := tt
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
	for i := 0; i < b.N; i++ {
		_ = underlyingError.Error()
		_ = error1.Error()
		_ = error2.Error()
		_ = list.Error()
	}
}
