# SDL Validation Gaps: graphql-js vs gqlparser `ValidateSchemaDocument`

## Architectural context

Two facts apply throughout this analysis.

**Fail-fast vs. accumulate.** gqlparser returns on the first error encountered; graphql-js collects and reports all errors in one pass. A schema with multiple violations will always surface only one error in gqlparser.

**Isolated validation.** graphql-js SDL rules receive a `SDLValidationContext` carrying a pre-existing schema object, enabling checks like "this type already exists in the schema you're extending." gqlparser validates a single `SchemaDocument` in isolation. Checks that require pre-existing schema context are not gaps — they are an architectural difference.

---

## Unintentional gaps

None currently known.

---

## Intentional divergences

### `PossibleTypeExtensions` — allows extensions of undefined types

When an extension references a type that doesn't exist, gqlparser creates a synthetic `Definition` for it (`schema.go:41–48`) and continues. graphql-js rejects with:
> `Cannot extend type "X" because it is not defined.`

This is **intentional**: `schema_test.yml` has an explicit test case "can extend non existant types" asserting no error. The practical use case is federation-style schemas where types are extended without a local base definition.

The consequence worth noting: the resulting ghost type does pass through `validateDefinition`. If the extension body provides at least one field and all types referenced in it exist, the ghost type becomes a valid Object type in the compiled schema. A typo in an extension's type name therefore produces a new, unexpected type rather than an error.

---

### `UniqueDirectiveNames` — builtin redeclaration silently accepted

For non-builtin directives, gqlparser correctly returns an error (`schema.go:107`), tested by `schema_test.yml`'s "cannot redeclare directives" case. For the six builtins — `include`, `skip`, `deprecated`, `specifiedBy`, `defer`, `oneOf` — a redeclaration is silently accepted with the first definition kept. `schema_test.yml` has an explicit "can redeclare builtin directives" test asserting this.

graphql-js rejects any directive redefinition, including builtins:
> `Directive "@skip" already exists in the schema. It cannot be redefined.`

The rationale is documented at `schema.go:95`: servers may ship directive definitions from an older or divergent spec version, and validating definition equivalence is considered more work than it's worth. The practical consequence is that a schema with a conflicting `@deprecated` or `@skip` definition loads without error.

---

## Confirmed covered

**`UniqueTypeNames`** — the first-pass map insertion at `schema.go:30–34` catches any type defined more than once in the document, returning `"Cannot redeclare type X."` Tested by `schema_test.yml`.

**`UniqueFieldDefinitionNames`** — field merging (`schema.go:63`) is followed by an O(n²) pair-scan at `schema.go:386–397`, catching duplicates within a definition, across definition + extension, and across multiple extensions. Tested by three cases in `schema_test.yml`.

**`UniqueEnumValueNames`** — enum value merging (`schema.go:65`) is followed by an O(n²) pair-scan at `schema.go:399–410`, mirroring the field check, catching duplicates within a definition and across extensions, returning `"Enum value X.Y can only be defined once."` Tested by two cases in `schema_test.yml` (same definition and across an extension).

**`UniqueArgumentDefinitionNames`** — `validateArgNames` runs a pair-scan over each field's argument list (from `validateDefinition`) and each directive definition's argument list (from `validateDirective`), returning `"Argument Query.field(id:) can only be defined once."` or `"Argument @dir(id:) can only be defined once."`, matching the graphql-js wording without the quotes. Field arguments cannot be split across extensions, so a per-field check is sufficient. Tested by `schema_test.yml` cases for object fields, interface fields and directive definitions, plus positive cases for the same argument name on different fields and directives.

**`UniqueDirectivesPerLocation` (SDL)** — `validateDirectives` (`schema.go:468`) tracks seen directive names per call and rejects a repeated non-repeatable directive with `"The directive X can only be used once at this location."` It is gated by a `singleLocation` flag (`schema.go:479`) so it applies only to single authored locations — fields, enum values, arguments, and the `schema` / `extend schema` directive lists. A type's own directives are exempt: they are merged across the base definition and every extension (`schema.go:65`-style append for `def.Directives`), which the spec treats as distinct locations, so the merged list validated at `schema.go:422` passes `singleLocation: false`. Consequence worth noting: a non-repeatable directive repeated within a single type definition (e.g. `type T @x @x` with no extension) is **not** caught, because provenance is lost once the base and extension directive lists are merged — directive definitions aren't even registered until `schema.go:112`, after the merge. graphql-js catches this by validating pre-merge AST nodes. Tested by four cases in `schema_test.yml` (non-repeatable directive repeated on a field and on an enum value; positive cases for a repeatable directive, the same directive on distinct field locations, and a directive on a type plus its extension).

**`UniqueOperationTypes`** — `ValidateSchemaDocument` tracks the operation types seen across the `schema {}` block and every `extend schema` block, returning `"There can be only one query type in schema."` (graphql-js wording) for the second occurrence. This covers a duplicate within one block, the same operation in two extensions, and an extension re-specifying an operation from the base block. graphql-js reports the last case as `"Type for query already defined in the schema. It cannot be redefined."` only when extending a pre-existing schema object, which is an isolated-validation architectural difference. Tested by three cases in `schema_test.yml`, plus a positive case adding each operation type once across extensions.

**`UniqueInputFieldNames` (SDL)** — `validateInputFieldNames` walks an input object value, including objects nested in other objects and lists, and rejects a field set more than once with `There can be only one input field named "x".`, the same message as the query-side rule. It runs on directive argument values at every schema location (from `validateDirectives`) and on default values of field arguments, directive definition arguments and input fields. Tested by `schema_test.yml` cases for each of those, plus a positive case for the same field name at different depths.

**`LoneSchemaDefinition`** — `len(sd.Schema) > 1` is checked at `schema.go:115`. The graphql-js check for "schema already defined in prior context" is an isolated-validation architectural difference, not a gap.

---

## Summary

| Rule | Status | Nature |
|---|---|---|
| `PossibleTypeExtensions` | Intentional divergence | Allows ghost types; federation use case |
| `UniqueDirectiveNames` (builtins) | Intentional divergence | Explicit test documents the choice |
| `UniqueTypeNames` | Covered | — |
| `UniqueFieldDefinitionNames` | Covered | — |
| `UniqueArgumentDefinitionNames` | Covered | Field and directive argument lists; tested |
| `UniqueEnumValueNames` | Covered | Pair-scan mirroring the field check; tested |
| `UniqueDirectivesPerLocation` (SDL) | Covered (with caveat) | Per single authored location; merged type-level list exempt |
| `UniqueOperationTypes` | Covered | Across `schema` and `extend schema` blocks; tested |
| `UniqueInputFieldNames` (SDL) | Covered | Directive argument values and default values; tested |
| `LoneSchemaDefinitionRule` | Covered / arch. difference | Within-doc check present |
