package gqlerror

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/vektah/gqlparser/v2/ast"
)

// attrErr is a cause that carries slog attributes of its own, in the shape of gqlgen's
// InvalidNullError, so a test can check that a wrapper does not report them.
type attrErr struct {
	attrs []slog.Attr
}

func (e *attrErr) Error() string      { return "attr error" }
func (e *attrErr) Attrs() []slog.Attr { return e.attrs }

// render formats attrs the way slog.TextHandler writes them, with the time, level and message
// left out. Comparing rendered text rather than the slices themselves is deliberate:
// reflect.DeepEqual, and so require.Equal, compares only the first byte of a slog string value,
// and slog.Value.Equal panics on the map in Extensions.
func render(t *testing.T, attrs []slog.Attr) string {
	t.Helper()

	var buf strings.Builder
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{ReplaceAttr: dropLevelAndMessage})
	// A zero time is left out by the handler.
	record := slog.NewRecord(time.Time{}, slog.LevelInfo, "", 0)
	record.AddAttrs(attrs...)
	require.NoError(t, handler.Handle(context.Background(), record))
	return strings.TrimSuffix(buf.String(), "\n")
}

// renderJSON formats attrs as slog.JSONHandler writes them, with the time, level and message
// left out.
func renderJSON(t *testing.T, attrs []slog.Attr) string {
	t.Helper()

	var buf strings.Builder
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{ReplaceAttr: dropLevelAndMessage})
	record := slog.NewRecord(time.Time{}, slog.LevelInfo, "", 0)
	record.AddAttrs(attrs...)
	require.NoError(t, handler.Handle(context.Background(), record))
	return buf.String()
}

func dropLevelAndMessage(groups []string, a slog.Attr) slog.Attr {
	if len(groups) == 0 && (a.Key == slog.LevelKey || a.Key == slog.MessageKey) {
		return slog.Attr{}
	}
	return a
}

func TestErrorAttrs(t *testing.T) {
	cases := map[string]struct {
		err  *Error
		want string
	}{
		"nil *Error": {
			err:  nil,
			want: "",
		},
		"every field empty": {
			err:  &Error{},
			want: "",
		},
		"message only": {
			err:  Errorf("kabloom"),
			want: "gql.message=kabloom",
		},
		"WrapPath": {
			err:  WrapPath(ast.Path{ast.PathName("user")}, errors.New("plain")),
			want: "gql.message=plain gql.path=user",
		},
		"every field, and a cause that is not reported": {
			err: &Error{
				Err:        &attrErr{attrs: []slog.Attr{slog.String("cause", "hidden")}},
				Message:    "kabloom",
				Path:       ast.Path{ast.PathName("field"), ast.PathIndex(0)},
				Locations:  []Location{{Line: 1, Column: 2}},
				Extensions: map[string]any{"code": "BAD"},
				Rule:       "NoUnusedFragments",
			},
			want: `gql.message=kabloom gql.path=field[0] gql.locations=[1:2] ` +
				`gql.extensions=map[code:BAD] gql.rule=NoUnusedFragments`,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := tc.err.Attrs()
			if tc.want == "" {
				require.Nil(t, got)
				return
			}
			require.Equal(t, tc.want, render(t, got))
		})
	}
}

func TestErrorWithSourcesAttrs(t *testing.T) {
	cases := map[string]struct {
		err  *ErrorWithSources
		want string
	}{
		"nil *ErrorWithSources": {
			err:  nil,
			want: "",
		},
		"every field empty": {
			err:  &ErrorWithSources{},
			want: "",
		},
		"every field, and a cause that is not reported": {
			err: &ErrorWithSources{
				Err:     &attrErr{attrs: []slog.Attr{slog.String("cause", "hidden")}},
				Message: "kabloom",
				Path:    ast.Path{ast.PathName("field"), ast.PathIndex(0)},
				Locations: []SourceLocation{
					{
						Line:   1,
						Column: 2,
						Source: &ast.Source{Name: "query.graphql", Input: "{ x }"},
					},
					{Line: 3, Column: 4},
				},
				Extensions: map[string]any{"code": "BAD"},
				Rule:       "NoUnusedFragments",
			},
			want: `gql.message=kabloom gql.path=field[0] gql.locations="[query.graphql:1:2 3:4]" ` +
				`gql.extensions=map[code:BAD] gql.rule=NoUnusedFragments`,
		},
		"built from an *Error without sources": {
			err: NewErrorWithSources(&Error{
				Message:   "kabloom",
				Locations: []Location{{Line: 1, Column: 2}},
			}, nil),
			want: "gql.message=kabloom gql.locations=[1:2]",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := tc.err.Attrs()
			if tc.want == "" {
				require.Nil(t, got)
				return
			}
			require.Equal(t, tc.want, render(t, got))
		})
	}
}

// TestErrorWithSourcesLogsLocationsAsJSON pins the JSON form, which the text cases cannot: the
// same line and column keys as Location, and the source by name rather than the document.
func TestErrorWithSourcesLogsLocationsAsJSON(t *testing.T) {
	err := &ErrorWithSources{Locations: []SourceLocation{
		{Line: 1, Column: 2, Source: &ast.Source{Name: "query.graphql", Input: "{ x }"}},
		{Line: 3, Column: 4},
	}}

	require.JSONEq(
		t,
		`{"gql":{"locations":[{"line":1,"column":2,"source":"query.graphql"},{"line":3,"column":4}]}}`,
		renderJSON(t, err.Attrs()),
	)
}

// TestBothErrorTypesLogLocationsAlike checks that converting an *Error to an *ErrorWithSources
// without sources does not change how its locations are logged, in text or in JSON.
func TestBothErrorTypesLogLocationsAlike(t *testing.T) {
	plain := &Error{
		Message:   "kabloom",
		Locations: []Location{{Line: 1, Column: 2}, {Line: 3, Column: 4}},
	}
	withSources := NewErrorWithSources(plain, nil)

	require.Equal(t, render(t, plain.Attrs()), render(t, withSources.Attrs()))
	require.JSONEq(t, renderJSON(t, plain.Attrs()), renderJSON(t, withSources.Attrs()))
}

func BenchmarkErrorAttrs(b *testing.B) {
	err := WrapPath(ast.Path{ast.PathName("user"), ast.PathIndex(0)}, errors.New("plain"))

	b.ReportAllocs()
	for b.Loop() {
		_ = err.Attrs()
	}
}

func BenchmarkErrorWithSourcesAttrs(b *testing.B) {
	err := &ErrorWithSources{
		Message: "kabloom",
		Path:    ast.Path{ast.PathName("user"), ast.PathIndex(0)},
		Locations: []SourceLocation{
			{Line: 1, Column: 2, Source: &ast.Source{Name: "query.graphql"}},
		},
	}

	b.ReportAllocs()
	for b.Loop() {
		_ = err.Attrs()
	}
}
