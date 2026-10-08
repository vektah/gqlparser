package gqlerror

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/vektah/gqlparser/v2/ast"
)

// attrErr is a cause that carries slog attributes of its own, in the shape of gqlgen's
// InvalidNullError, so a test can check that a wrapper does not report them. Its cause, when set,
// lets a test build a chain of such errors.
type attrErr struct {
	attrs []slog.Attr
	cause error
}

func (e *attrErr) Error() string      { return "attr error" }
func (e *attrErr) Attrs() []slog.Attr { return e.attrs }
func (e *attrErr) Unwrap() error      { return e.cause }

// fieldValuer supplies a "gql" group only when slog resolves it.
type fieldValuer struct{}

func (fieldValuer) LogValue() slog.Value { return slog.GroupValue(slog.String("field", "f")) }

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

// invalidNull is a cause in the shape of gqlgen's InvalidNullError reporting into the "gql"
// group, with the same path as the *Error wrapping it.
func invalidNull() *attrErr {
	return &attrErr{
		attrs: []slog.Attr{slog.Group("gql",
			slog.String("field", "Errors.b"),
			slog.String("type", "Error!"),
			slog.String("null_source", "resolver"),
			slog.String("path", "errors.b"),
		)},
	}
}

func TestCollectAttrs(t *testing.T) {
	errorsB := ast.Path{ast.PathName("errors"), ast.PathName("b")}
	reachedTwice := &attrErr{attrs: []slog.Attr{slog.String("k", "v")}}

	cases := map[string]struct {
		err  error
		want string
	}{
		"nil error": {
			err:  nil,
			want: "",
		},
		"no Attrs and no cause": {
			err:  errors.New("plain"),
			want: "",
		},
		"an *Error alone reports its own attributes": {
			err:  WrapPath(ast.Path{ast.PathName("user")}, errors.New("plain")),
			want: "gql.message=plain gql.path=user",
		},
		"a cause's gql group joins this package's, after its keys, with one path": {
			err: WrapPath(errorsB, invalidNull()),
			want: `gql.message="attr error" gql.path=errors.b gql.field=Errors.b ` +
				`gql.type=Error! gql.null_source=resolver`,
		},
		"a key reported twice keeps the outermost value": {
			err: &Error{
				Message: "kabloom",
				Path:    ast.Path{ast.PathName("presented")},
				Err:     invalidNull(),
			},
			want: `gql.message=kabloom gql.path=presented gql.field=Errors.b ` +
				`gql.type=Error! gql.null_source=resolver`,
		},
		"a wrapper above the *Error is walked through": {
			err:  fmt.Errorf("resolve: %w", WrapPath(errorsB, errors.New("plain"))),
			want: "gql.message=plain gql.path=errors.b",
		},
		"a joined tree is walked left to right, and the left value wins": {
			err: errors.Join(
				&attrErr{attrs: []slog.Attr{slog.String("k", "left"), slog.String("l", "1")}},
				&attrErr{attrs: []slog.Attr{slog.String("k", "right"), slog.String("r", "2")}},
			),
			want: "k=left l=1 r=2",
		},
		"an error reached twice is reported once": {
			err:  errors.Join(reachedTwice, fmt.Errorf("again: %w", reachedTwice)),
			want: "k=v",
		},
		"nested groups merge at every level": {
			err: &attrErr{
				attrs: []slog.Attr{slog.Group("gql", slog.Group("sub", slog.Int("a", 1)))},
				cause: &attrErr{attrs: []slog.Attr{slog.Group("gql",
					slog.Group("sub", slog.Int("b", 2)),
					slog.Int("c", 3),
				)}},
			},
			want: "gql.sub.a=1 gql.sub.b=2 gql.c=3",
		},
		"a value before a group under the same key wins": {
			err: &attrErr{
				attrs: []slog.Attr{slog.String("gql", "flat")},
				cause: Errorf("kabloom"),
			},
			want: "gql=flat",
		},
		"a group before a value under the same key wins": {
			err: &Error{
				Message: "kabloom",
				Err:     &attrErr{attrs: []slog.Attr{slog.String("gql", "flat")}},
			},
			want: "gql.message=kabloom",
		},
		"a group with an empty key is inlined": {
			err: &attrErr{
				attrs: []slog.Attr{
					slog.Group("", slog.String("k", "inlined")),
					slog.String("after", "kept"),
				},
				cause: &attrErr{attrs: []slog.Attr{slog.String("k", "cause")}},
			},
			want: "k=inlined after=kept",
		},
		"a LogValuer is resolved, so its group merges": {
			err: &attrErr{
				attrs: []slog.Attr{slog.Any("gql", fieldValuer{})},
				cause: Errorf("kabloom"),
			},
			want: "gql.field=f gql.message=kabloom",
		},
		"an empty group does not hide a later value": {
			err: &attrErr{
				attrs: []slog.Attr{slog.Group("k"), slog.String("after", "kept")},
				cause: &attrErr{attrs: []slog.Attr{slog.String("k", "v")}},
			},
			want: "after=kept k=v",
		},
		"a zero Attr is dropped": {
			err:  &attrErr{attrs: []slog.Attr{{}, slog.String("k", "v")}},
			want: "k=v",
		},
		"a nil value under a key is kept, as slog keeps it": {
			err:  &attrErr{attrs: []slog.Attr{slog.Any("k", nil)}},
			want: "k=<nil>",
		},
		"a value under an empty key is kept, as slog keeps it": {
			err:  &attrErr{attrs: []slog.Attr{slog.String("", "v"), slog.Any("", 1)}},
			want: `""=v`,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := CollectAttrs(tc.err)
			if tc.want == "" {
				require.Nil(t, got)
				return
			}
			require.Equal(t, tc.want, render(t, got))
			requireUniqueKeys(t, got)
		})
	}
}

// requireUniqueKeys checks the guarantee CollectAttrs exists for: no key twice at any level, and
// nothing left for a handler to inline or drop.
func requireUniqueKeys(t *testing.T, attrs []slog.Attr) {
	t.Helper()

	seen := make(map[string]bool, len(attrs))
	for _, a := range attrs {
		require.False(t, seen[a.Key], "key %q reported twice", a.Key)
		seen[a.Key] = true

		value := a.Value.Resolve()
		if value.Kind() != slog.KindGroup {
			require.False(t, a.Key == "" && value.Any() == nil, "a zero Attr is dropped by slog")
			continue
		}
		require.NotEmpty(t, a.Key, "a group with an empty key is inlined by slog")
		require.NotEmpty(t, value.Group(), "group %q is empty", a.Key)
		requireUniqueKeys(t, value.Group())
	}
}

// TestCollectAttrsWritesOneGQLObject pins the JSON form, which is where a repeated "gql" key
// loses data. The comparison is exact on purpose: require.JSONEq decodes both sides, and decoding
// is what hides a repeated key.
func TestCollectAttrsWritesOneGQLObject(t *testing.T) {
	err := WrapPath(ast.Path{ast.PathName("errors"), ast.PathName("b")}, invalidNull())

	require.Equal(
		t,
		`{"gql":{"message":"attr error","path":"errors.b","field":"Errors.b",`+
			`"type":"Error!","null_source":"resolver"}}`+"\n",
		renderJSON(t, CollectAttrs(err)),
	)
}

// TestCollectAttrsLeavesErrorsUnchanged merges a cause's nested group into the outer error's,
// which is a write into the outer group's slice unless CollectAttrs copied it first.
func TestCollectAttrsLeavesErrorsUnchanged(t *testing.T) {
	err := &attrErr{
		attrs: []slog.Attr{slog.Group("gql", slog.Group("sub", slog.Int("a", 1)))},
		cause: &attrErr{attrs: []slog.Attr{slog.Group("gql", slog.Group("sub", slog.Int("b", 2)))}},
	}
	before := render(t, err.Attrs())

	require.Equal(t, "gql.sub.a=1 gql.sub.b=2", render(t, CollectAttrs(err)))
	require.Equal(t, before, render(t, err.Attrs()))
}

func BenchmarkCollectAttrs(b *testing.B) {
	err := WrapPath(ast.Path{ast.PathName("errors"), ast.PathName("b")}, invalidNull())

	b.ReportAllocs()
	for b.Loop() {
		_ = CollectAttrs(err)
	}
}

// BenchmarkLogCollectedAttrs measures logging an error the way a server's middleware would:
// collect the chain's attributes, then write them through slog. The handlers format locations
// with the String methods this package defines, which nothing else here measures.
func BenchmarkLogCollectedAttrs(b *testing.B) {
	plain := &Error{
		Message:   "kabloom",
		Path:      ast.Path{ast.PathName("errors"), ast.PathName("b")},
		Locations: []Location{{Line: 1, Column: 2}},
		Err:       invalidNull(),
	}
	withSources := &ErrorWithSources{
		Message: "kabloom",
		Locations: []SourceLocation{
			{Line: 1, Column: 2, Source: &ast.Source{Name: "query.graphql"}},
		},
	}
	cases := []struct {
		name    string
		err     error
		handler slog.Handler
	}{
		{"text", plain, slog.NewTextHandler(io.Discard, nil)},
		{"json", plain, slog.NewJSONHandler(io.Discard, nil)},
		{"text with sources", withSources, slog.NewTextHandler(io.Discard, nil)},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			logger := slog.New(tc.handler)
			ctx := context.Background()

			b.ReportAllocs()
			for b.Loop() {
				logger.LogAttrs(ctx, slog.LevelError, "graphql error", CollectAttrs(tc.err)...)
			}
		})
	}
}
