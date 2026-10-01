// The property tests are a module of their own so that rapid stays out of gqlparser's
// dependency graph.
//
// A test-only requirement is not free to a library's consumers. It does not reach their
// builds and `go mod tidy` does not add it to their go.mod, but it does land in their go.sum
// with a full module hash and appear in `go list -m all`, which is enough to show up in a
// supply-chain scan, an SBOM and a dependency-review bot. gqlparser is depended on widely
// enough that one more name in that list is worth a directory.
module github.com/vektah/gqlparser/v2/proptest

go 1.25

require (
	github.com/vektah/gqlparser/v2 v2.5.58
	pgregory.net/rapid v1.3.0
)

require github.com/agnivade/levenshtein v1.2.1 // indirect

// The properties are about the working tree, not the last release.
replace github.com/vektah/gqlparser/v2 => ../
