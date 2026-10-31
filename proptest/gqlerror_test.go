package proptest

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"pgregory.net/rapid"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// TestListLogValueKeepsEverySibling checks that logging a gqlerror.List writes one JSON object
// per error, in order, each with that error's own message and path.
//
// A List is several independent errors. Merging their fields, as CollectAttrs does, keeps the
// first error's message and drops the rest, so a log line would show one failure where the
// response reported several. Any string can be a message, so the property also covers the JSON
// escaping of the array that List.LogValue assembles around slog's output.
//
// A counterexample is a list that logs as invalid JSON, with the wrong number of objects, or
// with an object whose message or path is not its own error's.
func TestListLogValueKeepsEverySibling(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		list := rapid.SliceOfN(rapid.Custom(func(t *rapid.T) *gqlerror.Error {
			if rapid.IntRange(0, 9).Draw(t, "nil") == 0 {
				return nil
			}
			err := &gqlerror.Error{Message: rapid.String().Draw(t, "message")}
			if name := rapid.StringMatching(`[a-z]{0,3}`).Draw(t, "path"); name != "" {
				err.Path = ast.Path{ast.PathName(name)}
			}
			return err
		}), 0, 5).Draw(t, "list")

		var buf bytes.Buffer
		record := slog.NewRecord(time.Time{}, slog.LevelError, "", 0)
		record.AddAttrs(slog.Any("errors", gqlerror.List(list)))
		if err := slog.NewJSONHandler(&buf, nil).Handle(context.Background(), record); err != nil {
			t.Fatalf("handle: %v", err)
		}

		var line struct {
			Errors []struct {
				GQL struct {
					Message string `json:"message"`
					Path    string `json:"path"`
				} `json:"gql"`
			} `json:"errors"`
		}
		if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
			t.Fatalf("invalid JSON %s: %v", buf.Bytes(), err)
		}
		if len(line.Errors) != len(list) {
			t.Fatalf("%d objects for %d errors: %s", len(line.Errors), len(list), buf.Bytes())
		}
		for i, err := range list {
			var wantMessage, wantPath string
			if err != nil {
				wantMessage, wantPath = err.Message, err.Path.String()
			}
			if got := line.Errors[i].GQL; got.Message != wantMessage || got.Path != wantPath {
				t.Fatalf("object %d is (%q, %q), want (%q, %q): %s",
					i, got.Message, got.Path, wantMessage, wantPath, buf.Bytes())
			}
		}
	})
}
