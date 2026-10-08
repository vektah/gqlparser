package gqlerror_test

import (
	"context"
	"log/slog"
	"os"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// nullError stands in for a cause that reports into the "gql" group, as gqlgen's
// InvalidNullError does.
type nullError struct{ field string }

func (e *nullError) Error() string { return e.field + " is null" }

func (e *nullError) Attrs() []slog.Attr {
	return []slog.Attr{slog.Group("gql", slog.String("field", e.field))}
}

func ExampleCollectAttrs() {
	err := gqlerror.WrapPath(ast.Path{ast.PathName("b")}, &nullError{field: "Errors.b"})

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) == 0 && (a.Key == slog.TimeKey || a.Key == slog.LevelKey) {
				return slog.Attr{} // keep the output stable and short
			}
			return a
		},
	}))
	logger.LogAttrs(context.Background(), slog.LevelError, "graphql error",
		gqlerror.CollectAttrs(err)...)

	// Output:
	// {"msg":"graphql error","gql":{"message":"Errors.b is null","path":"b","field":"Errors.b"}}
}
