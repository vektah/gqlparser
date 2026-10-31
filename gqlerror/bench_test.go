package gqlerror_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// The validator builds one error per violation, and a server formats, converts and matches each
// error it returns. None of that is a hot path on a valid request, but a malicious or broken
// client can send a query with thousands of violations, and then these costs are paid per
// violation. The benchmarks below measure each step on the shapes that reach it.

var (
	benchSource   = &ast.Source{Name: "query.graphql"}
	benchPos      = &ast.Position{Src: benchSource, Line: 2, Column: 3}
	benchPath     = ast.Path{ast.PathName("user"), ast.PathIndex(0), ast.PathName("name")}
	errBenchCause = errors.New("resolver failed")
)

// BenchmarkErrorString measures formatting an *Error, which every log line and every error
// message the validator returns pays for. The shapes add one component each: a file and a
// location, then a path.
func BenchmarkErrorString(b *testing.B) {
	located := gqlerror.ErrorPosf(benchPos, "Cannot query field %q on type %q.", "nme", "User")
	cases := []struct {
		name string
		err  *gqlerror.Error
	}{
		{"message", gqlerror.Errorf("kabloom")},
		{"located", located},
		{"located with path", &gqlerror.Error{
			Message:    located.Message,
			Path:       benchPath,
			Locations:  located.Locations,
			Extensions: located.Extensions,
		}},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = tc.err.Error()
			}
		})
	}
}

// BenchmarkErrorWithSourcesString measures formatting the source-aware error, which decides the
// file name from its locations' sources before formatting as *Error does. Each shape takes a
// different branch of that decision.
func BenchmarkErrorWithSourcesString(b *testing.B) {
	other := &ast.Source{Name: "fragments.graphql"}
	cases := []struct {
		name string
		err  *gqlerror.ErrorWithSources
	}{
		{"one location with a source", &gqlerror.ErrorWithSources{
			Message:   "kabloom",
			Locations: []gqlerror.SourceLocation{{Line: 1, Column: 2, Source: benchSource}},
		}},
		{"one location and a file extension", &gqlerror.ErrorWithSources{
			Message:    "kabloom",
			Extensions: map[string]any{"file": "override.graphql", "code": "BAD"},
			Locations:  []gqlerror.SourceLocation{{Line: 1, Column: 2, Source: benchSource}},
		}},
		{"several locations with sources", &gqlerror.ErrorWithSources{
			Message: "kabloom",
			Locations: []gqlerror.SourceLocation{
				{Line: 1, Column: 2, Source: benchSource},
				{Line: 3, Column: 4, Source: other},
			},
		}},
		{"several locations without sources", &gqlerror.ErrorWithSources{
			Message:   "kabloom",
			Locations: []gqlerror.SourceLocation{{Line: 1, Column: 2}, {Line: 3, Column: 4}},
		}},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = tc.err.Error()
			}
		})
	}
}

// BenchmarkNewErrorWithSources measures the conversion the source-aware validator applies to
// every error, with sources to check against the locations and without any.
func BenchmarkNewErrorWithSources(b *testing.B) {
	err := &gqlerror.Error{
		Message:   "kabloom",
		Locations: []gqlerror.Location{{Line: 1, Column: 2}, {Line: 3, Column: 4}},
	}
	sources := []gqlerror.SourceLocation{
		{Line: 1, Column: 2, Source: benchSource},
		{Line: 3, Column: 4, Source: benchSource},
	}
	cases := []struct {
		name    string
		sources []gqlerror.SourceLocation
	}{
		{"with sources", sources},
		{"without sources", nil},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = gqlerror.NewErrorWithSources(err, tc.sources)
			}
		})
	}
}

// BenchmarkConstructors measures building an error, once per violation in the validator and
// once per failed field in a server. The wrappers copy their cause's message, so they pay for
// its Error as well.
func BenchmarkConstructors(b *testing.B) {
	wrapped := gqlerror.WrapPath(benchPath, errBenchCause)
	inChain := fmt.Errorf("resolve: %w", wrapped)
	cases := []struct {
		name  string
		build func() *gqlerror.Error
	}{
		{"Errorf", func() *gqlerror.Error { return gqlerror.Errorf("kabloom %d", 1) }},
		{"ErrorPathf", func() *gqlerror.Error { return gqlerror.ErrorPathf(benchPath, "kabloom") }},
		{"ErrorPosf", func() *gqlerror.Error { return gqlerror.ErrorPosf(benchPos, "kabloom") }},
		{"ErrorPosf without a position", func() *gqlerror.Error {
			return gqlerror.ErrorPosf(nil, "kabloom")
		}},
		{"ErrorLocf", func() *gqlerror.Error {
			return gqlerror.ErrorLocf("query.graphql", 2, 3, "kabloom")
		}},
		{"Wrap", func() *gqlerror.Error { return gqlerror.Wrap(errBenchCause) }},
		{"WrapPath", func() *gqlerror.Error { return gqlerror.WrapPath(benchPath, errBenchCause) }},
		{"WrapPos", func() *gqlerror.Error { return gqlerror.WrapPos(benchPos, errBenchCause) }},
		{"WrapIfUnwrapped over a plain error", func() *gqlerror.Error {
			return gqlerror.WrapIfUnwrapped(errBenchCause)
		}},
		{"WrapIfUnwrapped over a chain holding an *Error", func() *gqlerror.Error {
			return gqlerror.WrapIfUnwrapped(inChain)
		}},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = tc.build()
			}
		})
	}
}

// BenchmarkListString measures formatting every error a request returns, which grows with the
// number of violations.
func BenchmarkListString(b *testing.B) {
	for _, size := range []int{1, 16} {
		list := make(gqlerror.List, size)
		sourceList := make(gqlerror.SourceList, size)
		for i := range size {
			list[i] = gqlerror.ErrorPosf(benchPos, "violation %d", i)
			sourceList[i] = gqlerror.NewErrorWithSources(list[i], nil)
		}

		b.Run("List/"+strconv.Itoa(size), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = list.Error()
			}
		})
		b.Run("SourceList/"+strconv.Itoa(size), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = sourceList.Error()
			}
		})
	}
}

// BenchmarkListMatching measures errors.Is and errors.As over a list of errors, where a miss
// walks every error and every cause. The miss is the expensive case, and the common one when a
// server checks for a sentinel the request did not produce.
func BenchmarkListMatching(b *testing.B) {
	sentinel := errors.New("sentinel")
	list := make(gqlerror.List, 16)
	sourceList := make(gqlerror.SourceList, 16)
	for i := range list {
		list[i] = gqlerror.WrapPath(benchPath, errBenchCause)
		sourceList[i] = gqlerror.NewErrorWithSources(list[i], nil)
	}

	b.Run("List/Is miss", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = errors.Is(list, sentinel)
		}
	})
	b.Run("List/As hit", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var target *gqlerror.Error
			_ = errors.As(list, &target)
		}
	})
	b.Run("SourceList/Is miss", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = errors.Is(sourceList, sentinel)
		}
	})
	b.Run("SourceList/As hit", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			var target *gqlerror.ErrorWithSources
			_ = errors.As(sourceList, &target)
		}
	})
	b.Run("SourceList/Unwrap", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = sourceList.Unwrap()
		}
	})
}

// BenchmarkErrorWithSourcesUnmarshalJSON measures decoding an error from a response, which a
// client does for every error a server returns.
func BenchmarkErrorWithSourcesUnmarshalJSON(b *testing.B) {
	data := []byte(`{"message":"kabloom","path":["user",0,"name"],` +
		`"locations":[{"line":1,"column":2},{"line":3,"column":4}],` +
		`"extensions":{"file":"query.graphql","code":"BAD"}}`)

	b.ReportAllocs()
	for b.Loop() {
		var err gqlerror.ErrorWithSources
		if unmarshalErr := json.Unmarshal(data, &err); unmarshalErr != nil {
			b.Fatal(unmarshalErr)
		}
	}
}

// BenchmarkCopies measures the two operations that allocate on behalf of a caller: naming the
// file on an error that has no extensions yet, and copying out the source locations.
func BenchmarkCopies(b *testing.B) {
	withSources := gqlerror.NewErrorWithSources(
		&gqlerror.Error{Locations: []gqlerror.Location{{Line: 1, Column: 2}, {Line: 3, Column: 4}}},
		[]gqlerror.SourceLocation{
			{Line: 1, Column: 2, Source: benchSource},
			{Line: 3, Column: 4, Source: benchSource},
		},
	)

	b.Run("SetFile", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			err := &gqlerror.Error{Message: "kabloom"}
			err.SetFile("query.graphql")
		}
	})
	b.Run("SourceLocations", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			_ = withSources.SourceLocations()
		}
	})
}
