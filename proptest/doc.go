// Package proptest holds gqlparser's property-based tests.
//
// They live outside the library's own module so that rapid, the generator library they need,
// stays out of the dependency graph every consumer of gqlparser inherits. See go.mod for why
// a test-only requirement is not free.
//
// Everything here drives gqlparser through its public API, which is the other reason the
// split costs little: a property worth stating about a parser is a property of what callers
// can see.
package proptest
