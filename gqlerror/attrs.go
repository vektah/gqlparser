package gqlerror

import (
	"log/slog"

	"github.com/vektah/gqlparser/v2/ast"
)

// Attrs reports the error's own fields as a single slog group named "gql": message, path,
// locations, extensions and rule, each left out when empty. The result is nil when every field
// is empty, or for a nil *Error.
//
// The group keeps these keys from colliding with a cause's: gqlgen's InvalidNullError, for one,
// reports a path of its own.
//
// Err is not reported. Attrs means an error's own attributes, never its cause's, so a walk that
// collects Attrs from every error in a chain visits the cause exactly once.
func (err *Error) Attrs() []slog.Attr {
	if err == nil {
		return nil
	}
	var locations slog.Attr
	if len(err.Locations) > 0 {
		locations = slog.Any("locations", err.Locations)
	}
	return gqlAttrs(err.Message, err.Path, locations, err.Extensions, err.Rule)
}

// Attrs reports the error's own fields as Error.Attrs does, in the same "gql" group under the
// same keys. Each location also names its source document, when it has one: in JSON as a
// "source" field beside "line" and "column", and in text as name:line:column.
//
// The *ast.Source itself is not logged. It would print as a pointer, and it holds the whole
// document.
func (err *ErrorWithSources) Attrs() []slog.Attr {
	if err == nil {
		return nil
	}
	var locations slog.Attr
	if len(err.Locations) > 0 {
		logged := make([]loggedLocation, len(err.Locations))
		for i, location := range err.Locations {
			logged[i] = loggedLocation{Line: location.Line, Column: location.Column}
			if location.Source != nil {
				logged[i].Source = location.Source.Name
			}
		}
		locations = slog.Any("locations", logged)
	}
	return gqlAttrs(err.Message, err.Path, locations, err.Extensions, err.Rule)
}

// loggedLocation is a SourceLocation as it is logged: the source document by name only.
type loggedLocation struct {
	Line   int    `json:"line,omitempty"`
	Column int    `json:"column,omitempty"`
	Source string `json:"source,omitempty"`
}

// String writes the location as Location.String does, prefixed with the source name when there
// is one, so that a text handler prints both error types' locations alike.
func (l loggedLocation) String() string {
	position := Location{Line: l.Line, Column: l.Column}.String()
	if l.Source == "" {
		return position
	}
	return l.Source + ":" + position
}

// gqlAttrs builds the "gql" group that both error types report, so the group's name, its keys,
// their order and the rule that an empty field is left out are decided in one place. locations
// is the zero Attr when there are none, because only the caller knows how its type logs them.
func gqlAttrs(
	message string,
	path ast.Path,
	locations slog.Attr,
	extensions map[string]any,
	rule string,
) []slog.Attr {
	// Sized for every field, so building the group costs one allocation rather than a regrowth
	// per field.
	fields := make([]slog.Attr, 0, 5)
	if message != "" {
		fields = append(fields, slog.String("message", message))
	}
	if len(path) > 0 {
		fields = append(fields, slog.String("path", path.String()))
	}
	if locations.Key != "" {
		fields = append(fields, locations)
	}
	if len(extensions) > 0 {
		fields = append(fields, slog.Any("extensions", extensions))
	}
	if rule != "" {
		fields = append(fields, slog.String("rule", rule))
	}
	if len(fields) == 0 {
		return nil
	}
	return []slog.Attr{{Key: "gql", Value: slog.GroupValue(fields...)}}
}
